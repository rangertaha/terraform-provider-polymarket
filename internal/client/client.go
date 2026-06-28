// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

// Package client is a thin HTTP client for the Polymarket Gamma Markets API.
//
// The Gamma API is the public, read-only data API that powers the Polymarket
// website. It exposes markets and events and does not require authentication.
// See https://docs.polymarket.com for the upstream documentation.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultEndpoint is the public Gamma Markets API base URL.
const DefaultEndpoint = "https://gamma-api.polymarket.com"

// Client is a configured Polymarket Gamma API client.
type Client struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

// Option customizes a Client during construction.
type Option func(*Client)

// WithEndpoint overrides the API base URL.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		if endpoint != "" {
			c.endpoint = strings.TrimRight(endpoint, "/")
		}
	}
}

// WithAPIKey sets an optional API key sent as a bearer token. The public Gamma
// endpoints do not require it, but it is forwarded when provided.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithHTTPClient overrides the underlying *http.Client (useful for tests).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// New constructs a Client with sane defaults.
func New(opts ...Option) *Client {
	c := &Client{
		endpoint:   DefaultEndpoint,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Market mirrors the subset of the Gamma API market object that this provider
// surfaces. Polymarket returns several numeric and array fields as JSON-encoded
// strings; those are decoded into native types by the data sources.
type Market struct {
	ID            string `json:"id"`
	Question      string `json:"question"`
	Slug          string `json:"slug"`
	Description   string `json:"description"`
	Active        bool   `json:"active"`
	Closed        bool   `json:"closed"`
	Archived      bool   `json:"archived"`
	Liquidity     string `json:"liquidity"`
	Volume        string `json:"volume"`
	StartDate     string `json:"startDate"`
	EndDate       string `json:"endDate"`
	ConditionID   string `json:"conditionId"`
	QuestionID    string `json:"questionID"`
	Outcomes      string `json:"outcomes"`      // JSON-encoded array, e.g. "[\"Yes\",\"No\"]"
	OutcomePrices string `json:"outcomePrices"` // JSON-encoded array, e.g. "[\"0.6\",\"0.4\"]"
}

// MarketFilter narrows a ListMarkets query. Zero-value fields are omitted.
type MarketFilter struct {
	Limit  int64
	Offset int64
	Active *bool
	Closed *bool
	Slug   string
}

// GetMarket fetches a single market by its numeric ID.
func (c *Client) GetMarket(ctx context.Context, id string) (*Market, error) {
	var m Market
	if err := c.get(ctx, "/markets/"+url.PathEscape(id), nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ListMarkets fetches markets matching the supplied filter.
func (c *Client) ListMarkets(ctx context.Context, f MarketFilter) ([]Market, error) {
	q := url.Values{}
	if f.Limit > 0 {
		q.Set("limit", strconv.FormatInt(f.Limit, 10))
	}
	if f.Offset > 0 {
		q.Set("offset", strconv.FormatInt(f.Offset, 10))
	}
	if f.Active != nil {
		q.Set("active", strconv.FormatBool(*f.Active))
	}
	if f.Closed != nil {
		q.Set("closed", strconv.FormatBool(*f.Closed))
	}
	if f.Slug != "" {
		q.Set("slug", f.Slug)
	}

	var markets []Market
	if err := c.get(ctx, "/markets", q, &markets); err != nil {
		return nil, err
	}
	return markets, nil
}

// get performs a GET request and decodes the JSON response into out.
func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	u := c.endpoint + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting %s: %w", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("polymarket API returned status %d for %s", resp.StatusCode, u)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decoding response from %s: %w", u, err)
	}
	return nil
}
