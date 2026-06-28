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
	"sync"
	"time"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/chain"
	"github.com/Rangertaha/terraform-provider-polymarket/internal/sign"
)

// DefaultEndpoint is the public Gamma Markets API base URL.
const DefaultEndpoint = "https://gamma-api.polymarket.com"

// DefaultClobEndpoint is the public CLOB (order book) API base URL.
const DefaultClobEndpoint = "https://clob.polymarket.com"

// DefaultDataEndpoint is the public Data API base URL (positions, trades, etc.).
const DefaultDataEndpoint = "https://data-api.polymarket.com"

// Client is a configured Polymarket API client. It addresses the Gamma catalog
// API, the CLOB order-book API, and the Data (portfolio/analytics) API.
type Client struct {
	endpoint     string
	clobEndpoint string
	dataEndpoint string
	apiKey       string
	signer       *sign.Signer
	chain        *chain.Client
	httpClient   *http.Client

	credsMu sync.Mutex
	creds   *APICredentials // lazily derived L2 credentials, cached
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

// WithClobEndpoint overrides the CLOB API base URL.
func WithClobEndpoint(endpoint string) Option {
	return func(c *Client) {
		if endpoint != "" {
			c.clobEndpoint = strings.TrimRight(endpoint, "/")
		}
	}
}

// WithDataEndpoint overrides the Data API base URL.
func WithDataEndpoint(endpoint string) Option {
	return func(c *Client) {
		if endpoint != "" {
			c.dataEndpoint = strings.TrimRight(endpoint, "/")
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
		endpoint:     DefaultEndpoint,
		clobEndpoint: DefaultClobEndpoint,
		dataEndpoint: DefaultDataEndpoint,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
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

// OrderBookLevel is a single price level in the order book. Price and Size are
// decimal strings to preserve full precision.
type OrderBookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// OrderBook is the live CLOB order book for a single outcome token.
type OrderBook struct {
	Market         string           `json:"market"`    // condition ID of the parent market
	AssetID        string           `json:"asset_id"`  // CLOB token ID of the outcome
	Timestamp      string           `json:"timestamp"` // server timestamp in ms since epoch
	Hash           string           `json:"hash"`      // content hash of the book snapshot
	Bids           []OrderBookLevel `json:"bids"`      // buy orders, ascending by price
	Asks           []OrderBookLevel `json:"asks"`      // sell orders, descending by price
	MinOrderSize   string           `json:"min_order_size"`
	TickSize       string           `json:"tick_size"`
	NegRisk        bool             `json:"neg_risk"`
	LastTradePrice string           `json:"last_trade_price"`
}

// GetOrderBook fetches the live order book for a CLOB outcome token.
func (c *Client) GetOrderBook(ctx context.Context, tokenID string) (*OrderBook, error) {
	q := url.Values{}
	q.Set("token_id", tokenID)
	var b OrderBook
	if err := c.getCLOB(ctx, "/book", q, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// GetPrice fetches the best price for a token on the given side ("buy"/"sell").
func (c *Client) GetPrice(ctx context.Context, tokenID, side string) (string, error) {
	q := url.Values{}
	q.Set("token_id", tokenID)
	q.Set("side", side)
	var resp struct {
		Price string `json:"price"`
	}
	if err := c.getCLOB(ctx, "/price", q, &resp); err != nil {
		return "", err
	}
	return resp.Price, nil
}

// GetMidpoint fetches the midpoint price (halfway between best bid and ask).
func (c *Client) GetMidpoint(ctx context.Context, tokenID string) (string, error) {
	q := url.Values{}
	q.Set("token_id", tokenID)
	var resp struct {
		Mid string `json:"mid"`
	}
	if err := c.getCLOB(ctx, "/midpoint", q, &resp); err != nil {
		return "", err
	}
	return resp.Mid, nil
}

// GetSpread fetches the current bid-ask spread for a token.
func (c *Client) GetSpread(ctx context.Context, tokenID string) (string, error) {
	q := url.Values{}
	q.Set("token_id", tokenID)
	var resp struct {
		Spread string `json:"spread"`
	}
	if err := c.getCLOB(ctx, "/spread", q, &resp); err != nil {
		return "", err
	}
	return resp.Spread, nil
}

// Position is a wallet's holding in a single market outcome.
type Position struct {
	ProxyWallet        string  `json:"proxyWallet"`
	Asset              string  `json:"asset"`       // CLOB token ID held
	ConditionID        string  `json:"conditionId"` // parent market condition ID
	Size               float64 `json:"size"`        // shares held
	AvgPrice           float64 `json:"avgPrice"`
	InitialValue       float64 `json:"initialValue"`
	CurrentValue       float64 `json:"currentValue"`
	CashPnl            float64 `json:"cashPnl"`
	PercentPnl         float64 `json:"percentPnl"`
	TotalBought        float64 `json:"totalBought"`
	RealizedPnl        float64 `json:"realizedPnl"`
	PercentRealizedPnl float64 `json:"percentRealizedPnl"`
	CurPrice           float64 `json:"curPrice"`
	Redeemable         bool    `json:"redeemable"`
	Mergeable          bool    `json:"mergeable"`
	Title              string  `json:"title"`
	Slug               string  `json:"slug"`
	Icon               string  `json:"icon"`
	EventID            string  `json:"eventId"`
	EventSlug          string  `json:"eventSlug"`
	Outcome            string  `json:"outcome"`
	OutcomeIndex       int64   `json:"outcomeIndex"`
	OppositeOutcome    string  `json:"oppositeOutcome"`
	OppositeAsset      string  `json:"oppositeAsset"`
	EndDate            string  `json:"endDate"`
	NegativeRisk       bool    `json:"negativeRisk"`
}

// Trade is a single executed trade by a wallet.
type Trade struct {
	ProxyWallet     string  `json:"proxyWallet"`
	Side            string  `json:"side"` // BUY or SELL
	Asset           string  `json:"asset"`
	ConditionID     string  `json:"conditionId"`
	Size            float64 `json:"size"`
	Price           float64 `json:"price"`
	Timestamp       int64   `json:"timestamp"` // Unix seconds
	Title           string  `json:"title"`
	Slug            string  `json:"slug"`
	Icon            string  `json:"icon"`
	EventSlug       string  `json:"eventSlug"`
	Outcome         string  `json:"outcome"`
	OutcomeIndex    int64   `json:"outcomeIndex"`
	Name            string  `json:"name"`
	Pseudonym       string  `json:"pseudonym"`
	TransactionHash string  `json:"transactionHash"`
}

// Holder is one wallet's holding within a HolderGroup.
type Holder struct {
	ProxyWallet           string  `json:"proxyWallet"`
	Asset                 string  `json:"asset"`
	Amount                float64 `json:"amount"`
	OutcomeIndex          int64   `json:"outcomeIndex"`
	Name                  string  `json:"name"`
	Pseudonym             string  `json:"pseudonym"`
	Bio                   string  `json:"bio"`
	DisplayUsernamePublic bool    `json:"displayUsernamePublic"`
	ProfileImage          string  `json:"profileImage"`
	Verified              bool    `json:"verified"`
}

// HolderGroup is the set of top holders for a single outcome token.
type HolderGroup struct {
	Token   string   `json:"token"`
	Holders []Holder `json:"holders"`
}

// Value is a wallet's total portfolio value, in USDC.
type Value struct {
	User  string  `json:"user"`
	Value float64 `json:"value"`
}

// PositionFilter narrows a ListPositions query. User is required.
type PositionFilter struct {
	User   string
	Market string // optional condition ID to filter to one market
	Limit  int64
	Offset int64
}

// ListPositions fetches the open positions held by a wallet.
func (c *Client) ListPositions(ctx context.Context, f PositionFilter) ([]Position, error) {
	q := url.Values{}
	q.Set("user", f.User)
	if f.Market != "" {
		q.Set("market", f.Market)
	}
	if f.Limit > 0 {
		q.Set("limit", strconv.FormatInt(f.Limit, 10))
	}
	if f.Offset > 0 {
		q.Set("offset", strconv.FormatInt(f.Offset, 10))
	}

	var positions []Position
	if err := c.getData(ctx, "/positions", q, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}

// TradeFilter narrows a ListTrades query. User is required.
type TradeFilter struct {
	User   string
	Market string // optional condition ID to filter to one market
	Limit  int64
	Offset int64
}

// ListTrades fetches the executed trades for a wallet.
func (c *Client) ListTrades(ctx context.Context, f TradeFilter) ([]Trade, error) {
	q := url.Values{}
	q.Set("user", f.User)
	if f.Market != "" {
		q.Set("market", f.Market)
	}
	if f.Limit > 0 {
		q.Set("limit", strconv.FormatInt(f.Limit, 10))
	}
	if f.Offset > 0 {
		q.Set("offset", strconv.FormatInt(f.Offset, 10))
	}

	var trades []Trade
	if err := c.getData(ctx, "/trades", q, &trades); err != nil {
		return nil, err
	}
	return trades, nil
}

// GetValue fetches a wallet's total portfolio value. The endpoint returns a
// single-element list; this returns a zero Value when the wallet is unknown.
func (c *Client) GetValue(ctx context.Context, user string) (Value, error) {
	q := url.Values{}
	q.Set("user", user)
	var values []Value
	if err := c.getData(ctx, "/value", q, &values); err != nil {
		return Value{}, err
	}
	if len(values) == 0 {
		return Value{User: user}, nil
	}
	return values[0], nil
}

// Activity is a single entry in a wallet's unified on-chain activity feed
// (trades, rewards, splits, merges, redemptions, conversions).
type Activity struct {
	ProxyWallet     string  `json:"proxyWallet"`
	Timestamp       int64   `json:"timestamp"`
	ConditionID     string  `json:"conditionId"`
	Type            string  `json:"type"` // TRADE, REWARD, SPLIT, MERGE, REDEEM, CONVERSION
	Size            float64 `json:"size"`
	UsdcSize        float64 `json:"usdcSize"`
	TransactionHash string  `json:"transactionHash"`
	Price           float64 `json:"price"`
	Asset           string  `json:"asset"`
	Side            string  `json:"side"`
	OutcomeIndex    int64   `json:"outcomeIndex"`
	Title           string  `json:"title"`
	Slug            string  `json:"slug"`
	EventSlug       string  `json:"eventSlug"`
	Outcome         string  `json:"outcome"`
}

// ActivityFilter narrows a ListActivity query. User is required.
type ActivityFilter struct {
	User   string
	Type   string // optional activity type to filter to, e.g. "TRADE"
	Market string // optional condition ID to filter to one market
	Limit  int64
	Offset int64
}

// ListActivity fetches a wallet's unified on-chain activity feed.
func (c *Client) ListActivity(ctx context.Context, f ActivityFilter) ([]Activity, error) {
	q := url.Values{}
	q.Set("user", f.User)
	if f.Type != "" {
		q.Set("type", f.Type)
	}
	if f.Market != "" {
		q.Set("market", f.Market)
	}
	if f.Limit > 0 {
		q.Set("limit", strconv.FormatInt(f.Limit, 10))
	}
	if f.Offset > 0 {
		q.Set("offset", strconv.FormatInt(f.Offset, 10))
	}

	var activity []Activity
	if err := c.getData(ctx, "/activity", q, &activity); err != nil {
		return nil, err
	}
	return activity, nil
}

// GetHolders fetches the top holders for each outcome token of a market,
// identified by its condition ID.
func (c *Client) GetHolders(ctx context.Context, market string, limit int64) ([]HolderGroup, error) {
	q := url.Values{}
	q.Set("market", market)
	if limit > 0 {
		q.Set("limit", strconv.FormatInt(limit, 10))
	}

	var groups []HolderGroup
	if err := c.getData(ctx, "/holders", q, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// get performs a GET against the Gamma API and decodes JSON into out.
func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	return c.getFrom(ctx, c.endpoint, path, query, out)
}

// getCLOB performs a GET against the CLOB API and decodes JSON into out.
func (c *Client) getCLOB(ctx context.Context, path string, query url.Values, out any) error {
	return c.getFrom(ctx, c.clobEndpoint, path, query, out)
}

// getData performs a GET against the Data API and decodes JSON into out.
func (c *Client) getData(ctx context.Context, path string, query url.Values, out any) error {
	return c.getFrom(ctx, c.dataEndpoint, path, query, out)
}

// getFrom performs a GET against the given base URL and decodes JSON into out.
func (c *Client) getFrom(ctx context.Context, base, path string, query url.Values, out any) error {
	u := base + path
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("polymarket API returned status %d for %s", resp.StatusCode, u)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decoding response from %s: %w", u, err)
	}
	return nil
}
