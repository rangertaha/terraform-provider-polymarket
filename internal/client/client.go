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
	ID               string `json:"id"`
	Question         string `json:"question"`
	Slug             string `json:"slug"`
	Description      string `json:"description"`
	ResolutionSource string `json:"resolutionSource"`
	Active           bool   `json:"active"`
	Closed           bool   `json:"closed"`
	Archived         bool   `json:"archived"`
	Liquidity        string `json:"liquidity"`
	Volume           string `json:"volume"`
	StartDate        string `json:"startDate"`
	EndDate          string `json:"endDate"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
	ConditionID      string `json:"conditionId"`
	QuestionID       string `json:"questionID"`
	GroupItemTitle   string `json:"groupItemTitle"`
	Image            string `json:"image"`
	Icon             string `json:"icon"`
	Outcomes         string `json:"outcomes"`      // JSON-encoded array, e.g. "[\"Yes\",\"No\"]"
	OutcomePrices    string `json:"outcomePrices"` // JSON-encoded array, e.g. "[\"0.6\",\"0.4\"]"
	ClobTokenIDs     string `json:"clobTokenIds"`  // JSON-encoded array of CLOB ERC-1155 token IDs

	// Trading statistics and live pricing (native JSON numbers).
	Volume24hr            float64 `json:"volume24hr"`
	Volume1wk             float64 `json:"volume1wk"`
	Volume1mo             float64 `json:"volume1mo"`
	Volume1yr             float64 `json:"volume1yr"`
	Spread                float64 `json:"spread"`
	BestBid               float64 `json:"bestBid"`
	BestAsk               float64 `json:"bestAsk"`
	LastTradePrice        float64 `json:"lastTradePrice"`
	Competitive           float64 `json:"competitive"`
	OrderMinSize          float64 `json:"orderMinSize"`
	OrderPriceMinTickSize float64 `json:"orderPriceMinTickSize"`

	// Order-book status flags.
	EnableOrderBook bool `json:"enableOrderBook"`
	AcceptingOrders bool `json:"acceptingOrders"`
}

// Tag is a category label attached to events and markets. CreatedAt/UpdatedAt
// are populated by the /tags endpoint and empty when embedded in an event.
type Tag struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Series groups recurring events under a common theme (e.g. a weekly market).
type Series struct {
	ID              string  `json:"id"`
	Ticker          string  `json:"ticker"`
	Slug            string  `json:"slug"`
	Title           string  `json:"title"`
	SeriesType      string  `json:"seriesType"`
	Recurrence      string  `json:"recurrence"`
	Image           string  `json:"image"`
	Icon            string  `json:"icon"`
	Active          bool    `json:"active"`
	Closed          bool    `json:"closed"`
	Archived        bool    `json:"archived"`
	Featured        bool    `json:"featured"`
	Restricted      bool    `json:"restricted"`
	CommentsEnabled bool    `json:"commentsEnabled"`
	Competitive     string  `json:"competitive"`
	Volume24hr      float64 `json:"volume24hr"`
	Volume          float64 `json:"volume"`
	Liquidity       float64 `json:"liquidity"`
	StartDate       string  `json:"startDate"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
	CommentCount    int64   `json:"commentCount"`
	Events          []Event `json:"events"`
}

// Event groups one or more related markets (e.g. all outcomes of an election).
type Event struct {
	ID               string   `json:"id"`
	Ticker           string   `json:"ticker"`
	Slug             string   `json:"slug"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	ResolutionSource string   `json:"resolutionSource"`
	StartDate        string   `json:"startDate"`
	CreationDate     string   `json:"creationDate"`
	EndDate          string   `json:"endDate"`
	Image            string   `json:"image"`
	Icon             string   `json:"icon"`
	Active           bool     `json:"active"`
	Closed           bool     `json:"closed"`
	Archived         bool     `json:"archived"`
	New              bool     `json:"new"`
	Featured         bool     `json:"featured"`
	Restricted       bool     `json:"restricted"`
	CreatedAt        string   `json:"createdAt"`
	UpdatedAt        string   `json:"updatedAt"`
	EnableOrderBook  bool     `json:"enableOrderBook"`
	NegRisk          bool     `json:"negRisk"`
	CommentCount     int64    `json:"commentCount"`
	SeriesSlug       string   `json:"seriesSlug"`
	Markets          []Market `json:"markets"`
	Tags             []Tag    `json:"tags"`
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

// EventFilter narrows a ListEvents query. Zero-value fields are omitted.
type EventFilter struct {
	Limit  int64
	Offset int64
	Active *bool
	Closed *bool
	Slug   string
}

// GetEvent fetches a single event by its numeric ID.
func (c *Client) GetEvent(ctx context.Context, id string) (*Event, error) {
	var e Event
	if err := c.get(ctx, "/events/"+url.PathEscape(id), nil, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// ListEvents fetches events matching the supplied filter.
func (c *Client) ListEvents(ctx context.Context, f EventFilter) ([]Event, error) {
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

	var events []Event
	if err := c.get(ctx, "/events", q, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// GetSeries fetches a single series by its numeric ID.
func (c *Client) GetSeries(ctx context.Context, id string) (*Series, error) {
	var s Series
	if err := c.get(ctx, "/series/"+url.PathEscape(id), nil, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// TagFilter narrows a ListTags query. Zero-value fields are omitted.
type TagFilter struct {
	Limit  int64
	Offset int64
}

// ListTags fetches category tags matching the supplied filter.
func (c *Client) ListTags(ctx context.Context, f TagFilter) ([]Tag, error) {
	q := url.Values{}
	if f.Limit > 0 {
		q.Set("limit", strconv.FormatInt(f.Limit, 10))
	}
	if f.Offset > 0 {
		q.Set("offset", strconv.FormatInt(f.Offset, 10))
	}

	var tags []Tag
	if err := c.get(ctx, "/tags", q, &tags); err != nil {
		return nil, err
	}
	return tags, nil
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
