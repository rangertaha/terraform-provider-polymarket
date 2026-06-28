// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/sign"
	"github.com/ethereum/go-ethereum/common"
)

// amountScale is the fixed-point scale Polymarket uses for order amounts: both
// USDC and conditional-token amounts are expressed in units of 1e-6.
const amountScale = 1_000_000

// zeroAddress is the taker address for public (anyone-can-fill) orders.
const zeroAddress = "0x0000000000000000000000000000000000000000"

// OrderArgs describes a limit order to place in high-level terms. The client
// converts price and size into the signed maker/taker base amounts.
type OrderArgs struct {
	TokenID    string  // CLOB ERC-1155 outcome token ID
	Side       string  // "BUY" or "SELL"
	Price      float64 // price per share, in [0, 1]
	Size       float64 // number of shares
	FeeRateBps int64   // fee in basis points (usually 0)
	Expiration int64   // Unix seconds; 0 for a GTC order with no expiry
	NegRisk    bool    // true if the market uses negative-risk settlement
	OrderType  string  // "GTC" (default), "GTD", "FOK", or "FAK"
}

// PlacedOrder is the CLOB response to a successful order placement.
type PlacedOrder struct {
	Success  bool   `json:"success"`
	OrderID  string `json:"orderID"`
	Status   string `json:"status"`
	ErrorMsg string `json:"errorMsg"`
}

// OrderStatus is the current state of a resting or completed order.
type OrderStatus struct {
	ID           string `json:"id"`
	Status       string `json:"status"` // e.g. LIVE, MATCHED, CANCELED
	Side         string `json:"side"`
	Price        string `json:"price"`
	OriginalSize string `json:"original_size"`
	SizeMatched  string `json:"size_matched"`
	AssetID      string `json:"asset_id"`
	Market       string `json:"market"`
	Expiration   string `json:"expiration"`
	OrderType    string `json:"order_type"`
}

// PlaceOrder signs and submits an order, returning the CLOB's acknowledgement.
func (c *Client) PlaceOrder(ctx context.Context, args OrderArgs) (*PlacedOrder, error) {
	if c.signer == nil {
		return nil, ErrNoSigner
	}

	creds, err := c.ensureCreds(ctx)
	if err != nil {
		return nil, err
	}

	body, err := c.buildSignedOrder(args, creds.APIKey)
	if err != nil {
		return nil, err
	}

	var placed PlacedOrder
	if err := c.l2Request(ctx, http.MethodPost, "/order", body, &placed); err != nil {
		return nil, err
	}
	if !placed.Success && placed.ErrorMsg != "" {
		return &placed, fmt.Errorf("order rejected: %s", placed.ErrorMsg)
	}
	return &placed, nil
}

// GetOrder fetches the current status of an order by ID. ok is false when the
// order is unknown to the CLOB (e.g. fully matched, cancelled, or expired).
func (c *Client) GetOrder(ctx context.Context, id string) (status *OrderStatus, ok bool, err error) {
	if c.signer == nil {
		return nil, false, ErrNoSigner
	}
	var s OrderStatus
	code, err := c.l2RequestStatus(ctx, http.MethodGet, "/data/order/"+id, nil, &s)
	if err != nil {
		return nil, false, err
	}
	if code == http.StatusNotFound {
		return nil, false, nil
	}
	return &s, true, nil
}

// CancelOrder cancels a resting order by ID.
func (c *Client) CancelOrder(ctx context.Context, id string) error {
	if c.signer == nil {
		return ErrNoSigner
	}
	body, err := json.Marshal(map[string]string{"orderID": id})
	if err != nil {
		return err
	}
	return c.l2Request(ctx, http.MethodDelete, "/order", body, nil)
}

// signedOrderBody is the JSON shape POST /order expects.
type signedOrderBody struct {
	Order     signedOrder `json:"order"`
	Owner     string      `json:"owner"`
	OrderType string      `json:"orderType"`
}

type signedOrder struct {
	Salt          string `json:"salt"`
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	Taker         string `json:"taker"`
	TokenID       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Expiration    string `json:"expiration"`
	Nonce         string `json:"nonce"`
	FeeRateBps    string `json:"feeRateBps"`
	Side          string `json:"side"`
	SignatureType uint8  `json:"signatureType"`
	Signature     string `json:"signature"`
}

// buildSignedOrder converts high-level OrderArgs into a signed order body. owner
// is the API key that owns the order.
func (c *Client) buildSignedOrder(args OrderArgs, owner string) ([]byte, error) {
	tokenID, ok := new(big.Int).SetString(args.TokenID, 10)
	if !ok {
		return nil, fmt.Errorf("invalid token_id %q", args.TokenID)
	}

	makerAmount, takerAmount, err := orderAmounts(args.Side, args.Price, args.Size)
	if err != nil {
		return nil, err
	}

	salt, err := randomSalt()
	if err != nil {
		return nil, err
	}

	sideCode := uint8(0) // BUY
	if args.Side == "SELL" {
		sideCode = 1
	}

	order := sign.Order{
		Salt:          salt,
		Maker:         common.HexToAddress(c.signer.Funder()),
		Signer:        common.HexToAddress(c.signer.Address()),
		Taker:         common.HexToAddress(zeroAddress),
		TokenID:       tokenID,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		Expiration:    big.NewInt(args.Expiration),
		Nonce:         big.NewInt(0),
		FeeRateBps:    big.NewInt(args.FeeRateBps),
		Side:          sideCode,
		SignatureType: c.signer.SignatureType(),
	}

	verifyingContract := sign.ExchangeAddress
	if args.NegRisk {
		verifyingContract = sign.NegRiskExchangeAddress
	}
	signature, err := c.signer.SignOrder(order, verifyingContract)
	if err != nil {
		return nil, err
	}

	orderType := args.OrderType
	if orderType == "" {
		orderType = "GTC"
	}

	return json.Marshal(signedOrderBody{
		Order: signedOrder{
			Salt:          salt.String(),
			Maker:         order.Maker.Hex(),
			Signer:        order.Signer.Hex(),
			Taker:         order.Taker.Hex(),
			TokenID:       tokenID.String(),
			MakerAmount:   makerAmount.String(),
			TakerAmount:   takerAmount.String(),
			Expiration:    strconv.FormatInt(args.Expiration, 10),
			Nonce:         "0",
			FeeRateBps:    strconv.FormatInt(args.FeeRateBps, 10),
			Side:          args.Side,
			SignatureType: c.signer.SignatureType(),
			Signature:     signature,
		},
		Owner:     owner,
		OrderType: orderType,
	})
}

// orderAmounts converts a price/size into integer maker and taker base amounts.
// For a BUY the maker offers USDC for shares; for a SELL the reverse.
func orderAmounts(side string, price, size float64) (maker, taker *big.Int, err error) {
	if price <= 0 || price >= 1 {
		return nil, nil, fmt.Errorf("price %v must be in the open interval (0, 1)", price)
	}
	if size <= 0 {
		return nil, nil, fmt.Errorf("size %v must be positive", size)
	}

	sizeBase := roundToBase(size)
	costBase := roundToBase(price * size)

	switch side {
	case "BUY":
		return costBase, sizeBase, nil
	case "SELL":
		return sizeBase, costBase, nil
	default:
		return nil, nil, fmt.Errorf("side %q must be BUY or SELL", side)
	}
}

// roundToBase scales a decimal amount to integer base units (1e6).
func roundToBase(v float64) *big.Int {
	scaled := v * amountScale
	// Round to nearest to avoid systematic truncation bias.
	return big.NewInt(int64(scaled + 0.5))
}

// randomSalt returns a cryptographically random 256-bit salt.
func randomSalt() (*big.Int, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("generating salt: %w", err)
	}
	return new(big.Int).SetBytes(buf), nil
}

// l2RequestStatus performs an L2-authenticated request and returns the HTTP
// status code so callers can distinguish 404 (not found) from success.
func (c *Client) l2RequestStatus(ctx context.Context, method, path string, body []byte, out any) (int, error) {
	creds, err := c.ensureCreds(ctx)
	if err != nil {
		return 0, err
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := sign.BuildHMACSignature(creds.Secret, timestamp, method, path, string(body))
	if err != nil {
		return 0, err
	}

	var reqBody *bytes.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	} else {
		reqBody = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.clobEndpoint+path, reqBody)
	if err != nil {
		return 0, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("POLY_ADDRESS", c.signer.Address())
	req.Header.Set("POLY_API_KEY", creds.APIKey)
	req.Header.Set("POLY_PASSPHRASE", creds.Passphrase)
	req.Header.Set("POLY_TIMESTAMP", timestamp)
	req.Header.Set("POLY_SIGNATURE", signature)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("requesting %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return resp.StatusCode, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("polymarket CLOB returned status %d for %s", resp.StatusCode, path)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, fmt.Errorf("decoding response: %w", err)
		}
	}
	return resp.StatusCode, nil
}

// l2Request is l2RequestStatus without the status code, treating 404 as an error.
func (c *Client) l2Request(ctx context.Context, method, path string, body []byte, out any) error {
	code, err := c.l2RequestStatus(ctx, method, path, body, out)
	if err != nil {
		return err
	}
	if code == http.StatusNotFound {
		return fmt.Errorf("polymarket CLOB returned status 404 for %s", path)
	}
	return nil
}
