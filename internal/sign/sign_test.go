// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package sign

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// testKey is the canonical Ethereum test private key (scalar 1). Its address is
// a well-known fixed value, making it a stable vector for address derivation.
const (
	testKey     = "0x0000000000000000000000000000000000000000000000000000000000000001"
	testAddress = "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf"
)

func newTestSigner(t *testing.T) *Signer {
	t.Helper()
	s, err := NewSigner(testKey, 137, 0, "")
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	return s
}

func TestNewSignerAddress(t *testing.T) {
	s := newTestSigner(t)
	if s.Address() != testAddress {
		t.Errorf("Address() = %s, want %s", s.Address(), testAddress)
	}
	// With no funder supplied, the funder defaults to the signing address.
	if s.Funder() != testAddress {
		t.Errorf("Funder() = %s, want %s", s.Funder(), testAddress)
	}
}

func TestNewSignerFunderOverride(t *testing.T) {
	funder := "0x000000000000000000000000000000000000dEaD"
	s, err := NewSigner(testKey, 137, 1, funder)
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	if !common.IsHexAddress(s.Funder()) || common.HexToAddress(s.Funder()) != common.HexToAddress(funder) {
		t.Errorf("Funder() = %s, want %s", s.Funder(), funder)
	}
}

func TestNewSignerRejectsBadKey(t *testing.T) {
	if _, err := NewSigner("not-a-key", 137, 0, ""); err == nil {
		t.Fatal("expected error for invalid private key")
	}
}

// recoverSigner recovers the address that produced a 0x-prefixed 27/28-V
// signature over digest, mirroring how Polymarket's contracts verify it.
func recoverSigner(t *testing.T, digest []byte, sigHex string) common.Address {
	t.Helper()
	sig := common.FromHex(sigHex)
	if len(sig) != 65 {
		t.Fatalf("signature length = %d, want 65", len(sig))
	}
	// Undo the 27/28 normalization for recovery.
	recSig := make([]byte, 65)
	copy(recSig, sig)
	recSig[64] -= 27
	pub, err := crypto.SigToPub(digest, recSig)
	if err != nil {
		t.Fatalf("SigToPub: %v", err)
	}
	return crypto.PubkeyToAddress(*pub)
}

func TestSignClobAuthRoundTrip(t *testing.T) {
	s := newTestSigner(t)
	sig, err := s.SignClobAuth("1700000000", 0)
	if err != nil {
		t.Fatalf("SignClobAuth: %v", err)
	}

	// Rebuild the exact digest the signer hashed, then recover and compare.
	digest := clobAuthDigest(t, s)
	if got := recoverSigner(t, digest, sig); got != s.address {
		t.Errorf("recovered %s, want %s", got.Hex(), s.address.Hex())
	}
}

func TestSignOrderRoundTrip(t *testing.T) {
	s := newTestSigner(t)
	order := Order{
		Salt:          big.NewInt(123456789),
		Maker:         s.funder,
		Signer:        s.address,
		Taker:         common.HexToAddress("0x0000000000000000000000000000000000000000"),
		TokenID:       mustBig(t, "71321045679252212594626385532706912750332728571942532289631379312455583992563"),
		MakerAmount:   big.NewInt(1_000_000),
		TakerAmount:   big.NewInt(2_000_000),
		Expiration:    big.NewInt(0),
		Nonce:         big.NewInt(0),
		FeeRateBps:    big.NewInt(0),
		Side:          0,
		SignatureType: 0,
	}

	sig, err := s.SignOrder(order, ExchangeAddress)
	if err != nil {
		t.Fatalf("SignOrder: %v", err)
	}

	digest := orderDigest(t, s, order, ExchangeAddress)
	if got := recoverSigner(t, digest, sig); got != s.address {
		t.Errorf("recovered %s, want %s", got.Hex(), s.address.Hex())
	}

	// Signing against a different verifying contract must change the signature,
	// proving the domain separator is bound into the digest.
	negSig, err := s.SignOrder(order, NegRiskExchangeAddress)
	if err != nil {
		t.Fatalf("SignOrder neg-risk: %v", err)
	}
	if negSig == sig {
		t.Error("signature did not change across verifying contracts")
	}
}

func TestBuildHMACSignatureKnownVector(t *testing.T) {
	// Vector computed independently (Python hmac/base64.urlsafe).
	const want = "oDoesRVlUq0RCXhd-ze00jwPwGcsZqjM9C5-jo9C2GY="
	got, err := BuildHMACSignature("aaaa", "1700000000", "GET", "/order", "")
	if err != nil {
		t.Fatalf("BuildHMACSignature: %v", err)
	}
	if got != want {
		t.Errorf("BuildHMACSignature = %q, want %q", got, want)
	}
}

func TestBuildHMACSignatureRejectsBadSecret(t *testing.T) {
	if _, err := BuildHMACSignature("!!!not-base64!!!", "1", "GET", "/", ""); err == nil {
		t.Fatal("expected error for invalid base64 secret")
	}
}

// --- helpers that rebuild digests using the package's own typed-data layout ---

func clobAuthDigest(t *testing.T, s *Signer) []byte {
	t.Helper()
	d, err := EIP712Hash(clobAuthTypedData(s, "1700000000", 0))
	if err != nil {
		t.Fatalf("EIP712Hash: %v", err)
	}
	return d
}

func orderDigest(t *testing.T, s *Signer, o Order, verifyingContract string) []byte {
	t.Helper()
	d, err := EIP712Hash(orderTypedData(s, o, verifyingContract))
	if err != nil {
		t.Fatalf("EIP712Hash: %v", err)
	}
	return d
}

func mustBig(t *testing.T, s string) *big.Int {
	t.Helper()
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		t.Fatalf("bad big int %q", s)
	}
	return v
}
