// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

// Package sign implements the cryptography Polymarket's CLOB requires for
// authenticated and order-placing requests:
//
//   - L1 authentication: an EIP-712 "ClobAuth" signature proving control of a
//     wallet, used to derive or create API credentials.
//   - L2 authentication: an HMAC-SHA256 request signature using those derived
//     credentials.
//   - Order signing: an EIP-712 signature over a CTF Exchange order struct.
//
// All signing is local; nothing here performs network I/O.
package sign

import (
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// CTF Exchange verifying-contract addresses on Polygon mainnet. Orders against a
// negative-risk market must be signed against the NegRisk exchange.
const (
	ExchangeAddress        = "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"
	NegRiskExchangeAddress = "0xC5d563A36AE78145C45a50134d48A1215220f80a"
)

// clobAuthMessage is the fixed attestation string signed for L1 auth.
const clobAuthMessage = "This message attests that I control the given wallet"

// Signer holds a wallet private key and the chain context needed to produce
// Polymarket signatures.
type Signer struct {
	key     *ecdsa.PrivateKey
	address common.Address
	funder  common.Address
	chainID int64
	sigType uint8
}

// NewSigner builds a Signer from a hex-encoded private key (with or without the
// 0x prefix). funder is the wallet holding USDC (for proxy/email accounts); when
// empty it defaults to the key's own address. sigType is the Polymarket
// signature type: 0 = EOA, 1 = email/magic proxy, 2 = browser/Gnosis proxy.
func NewSigner(privKeyHex string, chainID int64, sigType uint8, funder string) (*Signer, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}
	address := crypto.PubkeyToAddress(key.PublicKey)

	funderAddr := address
	if funder != "" {
		if !common.IsHexAddress(funder) {
			return nil, fmt.Errorf("funder %q is not a valid address", funder)
		}
		funderAddr = common.HexToAddress(funder)
	}

	return &Signer{
		key:     key,
		address: address,
		funder:  funderAddr,
		chainID: chainID,
		sigType: sigType,
	}, nil
}

// Address returns the checksummed wallet address derived from the private key.
func (s *Signer) Address() string { return s.address.Hex() }

// Funder returns the checksummed funding wallet address.
func (s *Signer) Funder() string { return s.funder.Hex() }

// SignatureType returns the Polymarket signature type this signer uses.
func (s *Signer) SignatureType() uint8 { return s.sigType }

// SignClobAuth produces the EIP-712 L1 signature used to derive or create API
// credentials. timestamp is Unix seconds as a string; nonce is typically 0.
func (s *Signer) SignClobAuth(timestamp string, nonce int64) (string, error) {
	return s.signTypedData(clobAuthTypedData(s, timestamp, nonce))
}

// clobAuthTypedData builds the EIP-712 payload for L1 authentication.
func clobAuthTypedData(s *Signer, timestamp string, nonce int64) apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"ClobAuth": {
				{Name: "address", Type: "address"},
				{Name: "timestamp", Type: "string"},
				{Name: "nonce", Type: "uint256"},
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: "ClobAuth",
		Domain: apitypes.TypedDataDomain{
			Name:    "ClobAuthDomain",
			Version: "1",
			ChainId: math.NewHexOrDecimal256(s.chainID),
		},
		Message: apitypes.TypedDataMessage{
			"address":   s.address.Hex(),
			"timestamp": timestamp,
			"nonce":     big.NewInt(nonce),
			"message":   clobAuthMessage,
		},
	}
}

// Order is a CTF Exchange order in the exact field order EIP-712 hashes.
type Order struct {
	Salt          *big.Int
	Maker         common.Address // funder that holds the assets
	Signer        common.Address // wallet that signs (the EOA)
	Taker         common.Address // allowed taker; zero address for public orders
	TokenID       *big.Int       // CLOB ERC-1155 outcome token ID
	MakerAmount   *big.Int       // amount the maker offers (base units)
	TakerAmount   *big.Int       // amount the maker requests (base units)
	Expiration    *big.Int       // Unix seconds; 0 for no expiry (GTC)
	Nonce         *big.Int       // maker nonce
	FeeRateBps    *big.Int       // fee in basis points
	Side          uint8          // 0 = BUY, 1 = SELL
	SignatureType uint8          // matches the Signer's signature type
}

// SignOrder produces the EIP-712 signature over an order, verified against the
// given exchange contract (use NegRiskExchangeAddress for negative-risk markets).
func (s *Signer) SignOrder(o Order, verifyingContract string) (string, error) {
	return s.signTypedData(orderTypedData(s, o, verifyingContract))
}

// orderTypedData builds the EIP-712 payload for a CTF Exchange order.
func orderTypedData(s *Signer, o Order, verifyingContract string) apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": {
				{Name: "salt", Type: "uint256"},
				{Name: "maker", Type: "address"},
				{Name: "signer", Type: "address"},
				{Name: "taker", Type: "address"},
				{Name: "tokenId", Type: "uint256"},
				{Name: "makerAmount", Type: "uint256"},
				{Name: "takerAmount", Type: "uint256"},
				{Name: "expiration", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
				{Name: "feeRateBps", Type: "uint256"},
				{Name: "side", Type: "uint8"},
				{Name: "signatureType", Type: "uint8"},
			},
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              "Polymarket CTF Exchange",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(s.chainID),
			VerifyingContract: verifyingContract,
		},
		Message: apitypes.TypedDataMessage{
			"salt":          o.Salt,
			"maker":         o.Maker.Hex(),
			"signer":        o.Signer.Hex(),
			"taker":         o.Taker.Hex(),
			"tokenId":       o.TokenID,
			"makerAmount":   o.MakerAmount,
			"takerAmount":   o.TakerAmount,
			"expiration":    o.Expiration,
			"nonce":         o.Nonce,
			"feeRateBps":    o.FeeRateBps,
			"side":          big.NewInt(int64(o.Side)),
			"signatureType": big.NewInt(int64(o.SignatureType)),
		},
	}
}

// EIP712Hash returns the 32-byte EIP-712 digest of typedData (the value that is
// signed). Exposed so tests and callers can verify hashing independently.
func EIP712Hash(typedData apitypes.TypedData) ([]byte, error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, fmt.Errorf("hashing domain: %w", err)
	}
	structHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, fmt.Errorf("hashing message: %w", err)
	}
	raw := append([]byte{0x19, 0x01}, domainSeparator...)
	raw = append(raw, structHash...)
	return crypto.Keccak256(raw), nil
}

// signTypedData hashes typedData per EIP-712 and signs it, returning a
// 0x-prefixed 65-byte signature with V normalized to 27/28 (Ethereum convention).
func (s *Signer) signTypedData(typedData apitypes.TypedData) (string, error) {
	digest, err := EIP712Hash(typedData)
	if err != nil {
		return "", err
	}
	sig, err := crypto.Sign(digest, s.key)
	if err != nil {
		return "", fmt.Errorf("signing digest: %w", err)
	}
	// crypto.Sign yields V in {0,1}; Polymarket expects {27,28}.
	sig[64] += 27
	return "0x" + common.Bytes2Hex(sig), nil
}

// BuildHMACSignature computes the L2 request signature: a base64url-encoded
// HMAC-SHA256 over timestamp+method+path+body, keyed by the base64url-decoded
// API secret. method should be upper-case; body is "" for GETs.
func BuildHMACSignature(secret, timestamp, method, path, body string) (string, error) {
	key, err := base64.URLEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("decoding API secret: %w", err)
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(timestamp + method + path + body))
	return base64.URLEncoding.EncodeToString(mac.Sum(nil)), nil
}
