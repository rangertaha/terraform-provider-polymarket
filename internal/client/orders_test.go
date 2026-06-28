// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/sign"
)

// testSecret is the base64url-encoded API secret the mock CLOB hands out.
const testSecret = "c2VjcmV0" // "secret"

// newMockCLOB returns an httptest server emulating the CLOB trading endpoints,
// validating L2 HMAC auth on each authenticated request.
func newMockCLOB(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/auth/derive-api-key":
			writeJSON(w, map[string]string{"apiKey": "key-1", "secret": testSecret, "passphrase": "pass-1"})

		case r.URL.Path == "/order" && r.Method == http.MethodPost:
			body := readBody(t, r)
			verifyHMAC(t, r, "POST", "/order", body)

			var sob signedOrderBody
			if err := json.Unmarshal(body, &sob); err != nil {
				t.Fatalf("decoding order body: %v", err)
			}
			if sob.Owner != "key-1" {
				t.Errorf("owner = %q, want key-1", sob.Owner)
			}
			if sob.Order.Signature == "" {
				t.Error("order signature missing")
			}
			if sob.Order.Side != "BUY" {
				t.Errorf("side = %q, want BUY", sob.Order.Side)
			}
			writeJSON(w, map[string]any{"success": true, "orderID": "0xabc", "status": "live"})

		case r.URL.Path == "/data/order/0xabc":
			verifyHMAC(t, r, "GET", "/data/order/0xabc", nil)
			writeJSON(w, map[string]any{"id": "0xabc", "status": "LIVE", "side": "BUY", "size_matched": "0"})

		case r.URL.Path == "/data/order/missing":
			w.WriteHeader(http.StatusNotFound)

		case r.URL.Path == "/order" && r.Method == http.MethodDelete:
			verifyHMAC(t, r, "DELETE", "/order", readBody(t, r))
			writeJSON(w, map[string]any{"canceled": []string{"0xabc"}})

		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
}

func newOrderClient(t *testing.T, url string) *Client {
	t.Helper()
	signer, err := sign.NewSigner(testKey, 137, 0, "")
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	return New(WithClobEndpoint(url), WithSigner(signer))
}

func TestPlaceOrderSignsAndAuthenticates(t *testing.T) {
	srv := newMockCLOB(t)
	defer srv.Close()
	c := newOrderClient(t, srv.URL)

	placed, err := c.PlaceOrder(context.Background(), OrderArgs{
		TokenID: "71321045679252212594626385532706912750332728571942532289631379312455583992563",
		Side:    "BUY",
		Price:   0.55,
		Size:    100,
	})
	if err != nil {
		t.Fatalf("PlaceOrder: %v", err)
	}
	if !placed.Success || placed.OrderID != "0xabc" {
		t.Errorf("unexpected placed order: %+v", placed)
	}
}

func TestGetOrderFoundAndMissing(t *testing.T) {
	srv := newMockCLOB(t)
	defer srv.Close()
	c := newOrderClient(t, srv.URL)

	status, ok, err := c.GetOrder(context.Background(), "0xabc")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if !ok || status.Status != "LIVE" {
		t.Errorf("got ok=%v status=%+v, want LIVE", ok, status)
	}

	_, ok, err = c.GetOrder(context.Background(), "missing")
	if err != nil {
		t.Fatalf("GetOrder(missing): %v", err)
	}
	if ok {
		t.Error("expected ok=false for a missing order")
	}
}

func TestCancelOrder(t *testing.T) {
	srv := newMockCLOB(t)
	defer srv.Close()
	c := newOrderClient(t, srv.URL)

	if err := c.CancelOrder(context.Background(), "0xabc"); err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}
}

func TestOrderAmounts(t *testing.T) {
	// BUY 100 shares @ 0.55 -> maker = 55 USDC, taker = 100 shares (base 1e6).
	maker, taker, err := orderAmounts("BUY", 0.55, 100)
	if err != nil {
		t.Fatalf("orderAmounts: %v", err)
	}
	if maker.String() != "55000000" || taker.String() != "100000000" {
		t.Errorf("BUY amounts = (%s, %s), want (55000000, 100000000)", maker, taker)
	}

	// SELL flips maker/taker.
	maker, taker, err = orderAmounts("SELL", 0.55, 100)
	if err != nil {
		t.Fatalf("orderAmounts: %v", err)
	}
	if maker.String() != "100000000" || taker.String() != "55000000" {
		t.Errorf("SELL amounts = (%s, %s), want (100000000, 55000000)", maker, taker)
	}

	if _, _, err := orderAmounts("BUY", 1.5, 10); err == nil {
		t.Error("expected error for out-of-range price")
	}
}

// --- test helpers ---

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func readBody(t *testing.T, r *http.Request) []byte {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	return b
}

// verifyHMAC recomputes the expected L2 signature and compares it to the header.
func verifyHMAC(t *testing.T, r *http.Request, method, path string, body []byte) {
	t.Helper()
	if r.Header.Get("POLY_API_KEY") != "key-1" {
		t.Errorf("POLY_API_KEY = %q", r.Header.Get("POLY_API_KEY"))
	}
	if r.Header.Get("POLY_PASSPHRASE") != "pass-1" {
		t.Errorf("POLY_PASSPHRASE = %q", r.Header.Get("POLY_PASSPHRASE"))
	}
	want, err := sign.BuildHMACSignature(testSecret, r.Header.Get("POLY_TIMESTAMP"), method, path, string(body))
	if err != nil {
		t.Fatalf("BuildHMACSignature: %v", err)
	}
	if got := r.Header.Get("POLY_SIGNATURE"); got != want {
		t.Errorf("POLY_SIGNATURE = %q, want %q", got, want)
	}
}
