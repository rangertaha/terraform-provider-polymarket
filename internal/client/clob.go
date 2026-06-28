// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// PricePoint is one sample in a token's price history.
type PricePoint struct {
	T int64   `json:"t"` // Unix seconds
	P float64 `json:"p"` // price in [0, 1]
}

// priceHistoryResponse wraps the /prices-history payload.
type priceHistoryResponse struct {
	History []PricePoint `json:"history"`
}

// GetPriceHistory fetches the historical price series for an outcome token.
// interval is a window like "1h", "6h", "1d", "1w", or "max"; fidelity is the
// resolution in minutes (0 lets the server choose).
func (c *Client) GetPriceHistory(ctx context.Context, tokenID, interval string, fidelity int64) ([]PricePoint, error) {
	q := url.Values{}
	q.Set("market", tokenID)
	if interval != "" {
		q.Set("interval", interval)
	}
	if fidelity > 0 {
		q.Set("fidelity", strconv.FormatInt(fidelity, 10))
	}

	var resp priceHistoryResponse
	if err := c.getCLOB(ctx, "/prices-history", q, &resp); err != nil {
		return nil, err
	}
	return resp.History, nil
}

// TokenPrices holds the best buy and sell price for a token.
type TokenPrices struct {
	TokenID string
	Buy     string
	Sell    string
}

// GetPrices fetches the best buy and sell price for several tokens in one call.
func (c *Client) GetPrices(ctx context.Context, tokenIDs []string) ([]TokenPrices, error) {
	type query struct {
		TokenID string `json:"token_id"`
		Side    string `json:"side"`
	}
	body := make([]query, 0, len(tokenIDs)*2)
	for _, id := range tokenIDs {
		body = append(body, query{TokenID: id, Side: "BUY"}, query{TokenID: id, Side: "SELL"})
	}

	// Response is keyed by token ID: {tokenID: {"BUY": "..", "SELL": ".."}}.
	var raw map[string]map[string]string
	if err := c.postCLOBPublic(ctx, "/prices", body, &raw); err != nil {
		return nil, err
	}

	out := make([]TokenPrices, 0, len(tokenIDs))
	for _, id := range tokenIDs {
		sides := raw[id]
		out = append(out, TokenPrices{TokenID: id, Buy: sides["BUY"], Sell: sides["SELL"]})
	}
	return out, nil
}

// GetOrderBooks fetches the order books for several tokens in one call.
func (c *Client) GetOrderBooks(ctx context.Context, tokenIDs []string) ([]OrderBook, error) {
	type query struct {
		TokenID string `json:"token_id"`
	}
	body := make([]query, 0, len(tokenIDs))
	for _, id := range tokenIDs {
		body = append(body, query{TokenID: id})
	}

	var books []OrderBook
	if err := c.postCLOBPublic(ctx, "/books", body, &books); err != nil {
		return nil, err
	}
	return books, nil
}

// postCLOBPublic issues an unauthenticated JSON POST against the CLOB API.
func (c *Client) postCLOBPublic(ctx context.Context, path string, body, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.clobEndpoint+path, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("polymarket CLOB returned status %d for %s", resp.StatusCode, path)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decoding response from %s: %w", path, err)
	}
	return nil
}
