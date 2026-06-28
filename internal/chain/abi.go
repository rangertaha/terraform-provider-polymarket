// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package chain

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// Well-known Polygon mainnet contract addresses used for Polymarket trading.
const (
	// USDCAddress is the USDC.e (PoS) token that collateralizes markets.
	USDCAddress = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
	// ConditionalTokensAddress is the Gnosis CTF ERC-1155 outcome-token contract.
	ConditionalTokensAddress = "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"
)

// erc20ABI is the minimal ERC-20 surface needed to manage USDC allowances.
const erc20ABI = `[
  {"name":"approve","type":"function","stateMutability":"nonpayable",
   "inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],
   "outputs":[{"name":"","type":"bool"}]},
  {"name":"allowance","type":"function","stateMutability":"view",
   "inputs":[{"name":"owner","type":"address"},{"name":"spender","type":"address"}],
   "outputs":[{"name":"","type":"uint256"}]}
]`

// erc1155ABI is the minimal ERC-1155 surface needed to manage CTF approvals.
const erc1155ABI = `[
  {"name":"setApprovalForAll","type":"function","stateMutability":"nonpayable",
   "inputs":[{"name":"operator","type":"address"},{"name":"approved","type":"bool"}],
   "outputs":[]},
  {"name":"isApprovedForAll","type":"function","stateMutability":"view",
   "inputs":[{"name":"account","type":"address"},{"name":"operator","type":"address"}],
   "outputs":[{"name":"","type":"bool"}]}
]`

// mustParseABI parses a static ABI definition, panicking on malformed input
// (the definitions are compile-time constants, so any error is a programmer bug).
func mustParseABI(def string) abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(def))
	if err != nil {
		panic("chain: invalid ABI: " + err.Error())
	}
	return parsed
}
