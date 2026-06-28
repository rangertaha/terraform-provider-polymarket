// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const testKey = "0x0000000000000000000000000000000000000000000000000000000000000001"

// TestABISelectors pins the function selectors so a mistyped ABI signature
// (which would send the wrong calldata on-chain) fails loudly.
func TestABISelectors(t *testing.T) {
	erc20 := mustParseABI(erc20ABI)
	erc1155 := mustParseABI(erc1155ABI)

	cases := []struct {
		name     string
		selector string // first 4 bytes, hex
		got      []byte
	}{
		{"approve", "095ea7b3", erc20.Methods["approve"].ID},
		{"allowance", "dd62ed3e", erc20.Methods["allowance"].ID},
		{"setApprovalForAll", "a22cb465", erc1155.Methods["setApprovalForAll"].ID},
		{"isApprovedForAll", "e985e9c5", erc1155.Methods["isApprovedForAll"].ID},
	}
	for _, tc := range cases {
		if got := common.Bytes2Hex(tc.got); got != tc.selector {
			t.Errorf("%s selector = %s, want %s", tc.name, got, tc.selector)
		}
	}
}

// TestApproveCalldata verifies the full packed calldata for an approve call.
func TestApproveCalldata(t *testing.T) {
	erc20 := mustParseABI(erc20ABI)
	spender := common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E")
	data, err := erc20.Pack("approve", spender, big.NewInt(1000000))
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}
	// selector (4) + address (32) + uint256 (32) = 68 bytes.
	if len(data) != 68 {
		t.Fatalf("calldata length = %d, want 68", len(data))
	}
	if common.Bytes2Hex(data[:4]) != "095ea7b3" {
		t.Errorf("selector = %s", common.Bytes2Hex(data[:4]))
	}
	// The spender address occupies the right-most 20 bytes of the first arg word.
	if common.BytesToAddress(data[4:36]) != spender {
		t.Errorf("encoded spender = %s, want %s", common.BytesToAddress(data[4:36]).Hex(), spender.Hex())
	}
}

// TestAllowanceERC20ReadsViaRPC exercises the read path against a mock JSON-RPC
// endpoint, proving the eth_call result is decoded into the allowance value.
func TestAllowanceERC20ReadsViaRPC(t *testing.T) {
	const want = 1234567

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		var result string
		switch req.Method {
		case "eth_chainId":
			result = "0x89" // 137
		case "eth_call":
			// ABI-encode a single uint256 as a 32-byte word.
			result = hexutil.Encode(common.LeftPadBytes(big.NewInt(want).Bytes(), 32))
		default:
			result = "0x"
		}
		_, _ = fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":%q}`, req.ID, result)
	}))
	defer srv.Close()

	c, err := New(context.Background(), srv.URL, testKey, 137)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	owner := common.HexToAddress("0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf")
	spender := common.HexToAddress(ConditionalTokensAddress)
	got, err := c.AllowanceERC20(context.Background(), common.HexToAddress(USDCAddress), owner, spender)
	if err != nil {
		t.Fatalf("AllowanceERC20: %v", err)
	}
	if got.Int64() != want {
		t.Errorf("allowance = %s, want %d", got, want)
	}
}

func TestNewRejectsBadKey(t *testing.T) {
	if _, err := New(context.Background(), "http://127.0.0.1:1", "bad-key", 137); err == nil {
		t.Fatal("expected error for invalid private key")
	}
}
