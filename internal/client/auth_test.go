// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/sign"
)

const (
	testKey     = "0x0000000000000000000000000000000000000000000000000000000000000001"
	testAddress = "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf"
)

func TestDeriveAPIKeySignsAndDecodes(t *testing.T) {
	var gotHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		if r.URL.Path != "/auth/derive-api-key" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiKey":"key-1","secret":"c2VjcmV0","passphrase":"pass-1"}`))
	}))
	defer srv.Close()

	signer, err := sign.NewSigner(testKey, 137, 0, "")
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	c := New(WithClobEndpoint(srv.URL), WithSigner(signer))

	creds, err := c.DeriveAPIKey(context.Background())
	if err != nil {
		t.Fatalf("DeriveAPIKey: %v", err)
	}
	if creds.APIKey != "key-1" || creds.Secret != "c2VjcmV0" || creds.Passphrase != "pass-1" {
		t.Errorf("unexpected creds: %+v", creds)
	}

	// The wallet address header must match the configured key.
	if gotHeaders.Get("POLY_ADDRESS") != testAddress {
		t.Errorf("POLY_ADDRESS = %s, want %s", gotHeaders.Get("POLY_ADDRESS"), testAddress)
	}
	if gotHeaders.Get("POLY_NONCE") != "0" {
		t.Errorf("POLY_NONCE = %s, want 0", gotHeaders.Get("POLY_NONCE"))
	}

	// crypto.Sign is deterministic (RFC 6979), so re-signing the timestamp the
	// server received must reproduce the exact signature the client sent. This
	// proves the client signed the correct EIP-712 payload with the right key.
	wantSig, err := signer.SignClobAuth(gotHeaders.Get("POLY_TIMESTAMP"), 0)
	if err != nil {
		t.Fatalf("SignClobAuth: %v", err)
	}
	if got := gotHeaders.Get("POLY_SIGNATURE"); got != wantSig {
		t.Errorf("POLY_SIGNATURE = %s, want %s", got, wantSig)
	}
}

func TestAuthRequiresSigner(t *testing.T) {
	c := New() // no signer configured
	if c.HasSigner() {
		t.Fatal("HasSigner should be false without a signer")
	}
	if _, err := c.DeriveAPIKey(context.Background()); err != ErrNoSigner {
		t.Fatalf("expected ErrNoSigner, got %v", err)
	}
}
