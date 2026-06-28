// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/sign"
)

// ErrNoSigner is returned by authenticated calls when the provider was
// configured without a private key.
var ErrNoSigner = errors.New("polymarket: no private_key configured for authenticated requests")

// WithSigner attaches a wallet signer, enabling L1-authenticated CLOB requests.
func WithSigner(s *sign.Signer) Option {
	return func(c *Client) { c.signer = s }
}

// HasSigner reports whether the client can make authenticated requests.
func (c *Client) HasSigner() bool { return c.signer != nil }

// APICredentials are the L2 API key set derived from an L1 signature. The secret
// and passphrase are sensitive and authenticate subsequent HMAC-signed requests.
type APICredentials struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// DeriveAPIKey returns the deterministic API credentials for the configured
// wallet, creating them server-side on first use (GET /auth/derive-api-key).
func (c *Client) DeriveAPIKey(ctx context.Context) (*APICredentials, error) {
	return c.l1Request(ctx, http.MethodGet, "/auth/derive-api-key")
}

// CreateAPIKey provisions a fresh API key set for the configured wallet
// (POST /auth/api-key).
func (c *Client) CreateAPIKey(ctx context.Context) (*APICredentials, error) {
	return c.l1Request(ctx, http.MethodPost, "/auth/api-key")
}

// l1Request performs an L1 (EIP-712) authenticated request against the CLOB and
// decodes the resulting API credentials.
func (c *Client) l1Request(ctx context.Context, method, path string) (*APICredentials, error) {
	if c.signer == nil {
		return nil, ErrNoSigner
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	const nonce = 0
	signature, err := c.signer.SignClobAuth(timestamp, nonce)
	if err != nil {
		return nil, fmt.Errorf("signing L1 auth: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.clobEndpoint+path, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("POLY_ADDRESS", c.signer.Address())
	req.Header.Set("POLY_SIGNATURE", signature)
	req.Header.Set("POLY_TIMESTAMP", timestamp)
	req.Header.Set("POLY_NONCE", strconv.Itoa(nonce))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("polymarket CLOB auth returned status %d for %s", resp.StatusCode, path)
	}

	var creds APICredentials
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return nil, fmt.Errorf("decoding credentials: %w", err)
	}
	return &creds, nil
}
