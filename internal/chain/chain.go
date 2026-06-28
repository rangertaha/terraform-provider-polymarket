// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

// Package chain submits and reads the on-chain ERC-20 / ERC-1155 approvals that
// Polymarket trading requires: USDC spend approval for the exchange and CTF
// outcome-token operator approval. It talks to a Polygon JSON-RPC endpoint and
// signs transactions with the configured wallet key.
package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// MaxUint256 is the conventional "unlimited" ERC-20 approval amount.
var MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

// Client manages on-chain approvals for a single wallet.
type Client struct {
	eth        *ethclient.Client
	erc20      abi.ABI
	erc1155    abi.ABI
	key        *ecdsa.PrivateKey
	from       common.Address
	chainID    *big.Int
	rpcAddress string
}

// New dials the JSON-RPC endpoint and prepares a signer from the private key.
// For HTTP endpoints the dial is lazy, so this does not require connectivity.
func New(ctx context.Context, rpcURL, privKeyHex string, chainID int64) (*Client, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	eth, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dialing RPC endpoint: %w", err)
	}

	return &Client{
		eth:        eth,
		erc20:      mustParseABI(erc20ABI),
		erc1155:    mustParseABI(erc1155ABI),
		key:        key,
		from:       crypto.PubkeyToAddress(key.PublicKey),
		chainID:    big.NewInt(chainID),
		rpcAddress: rpcURL,
	}, nil
}

// From returns the wallet address that signs approval transactions.
func (c *Client) From() common.Address { return c.from }

// transactor builds signed-transaction options bound to this wallet and chain.
func (c *Client) transactor(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(c.key, c.chainID)
	if err != nil {
		return nil, fmt.Errorf("building transactor: %w", err)
	}
	opts.Context = ctx
	return opts, nil
}

// Close releases the underlying RPC connection.
func (c *Client) Close() {
	if c.eth != nil {
		c.eth.Close()
	}
}

// AllowanceERC20 reads the USDC allowance the owner has granted the spender.
func (c *Client) AllowanceERC20(ctx context.Context, token, owner, spender common.Address) (*big.Int, error) {
	var out []any
	contract := bind.NewBoundContract(token, c.erc20, c.eth, c.eth, c.eth)
	if err := contract.Call(&bind.CallOpts{Context: ctx}, &out, "allowance", owner, spender); err != nil {
		return nil, fmt.Errorf("reading allowance: %w", err)
	}
	amount, ok := out[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected allowance type %T", out[0])
	}
	return amount, nil
}

// IsApprovedForAll reads whether the owner has approved the operator on a CTF
// (ERC-1155) token contract.
func (c *Client) IsApprovedForAll(ctx context.Context, token, owner, operator common.Address) (bool, error) {
	var out []any
	contract := bind.NewBoundContract(token, c.erc1155, c.eth, c.eth, c.eth)
	if err := contract.Call(&bind.CallOpts{Context: ctx}, &out, "isApprovedForAll", owner, operator); err != nil {
		return false, fmt.Errorf("reading isApprovedForAll: %w", err)
	}
	approved, ok := out[0].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected isApprovedForAll type %T", out[0])
	}
	return approved, nil
}

// ApproveERC20 submits an ERC-20 approve and waits for it to be mined, returning
// the transaction hash.
func (c *Client) ApproveERC20(ctx context.Context, token, spender common.Address, amount *big.Int) (string, error) {
	return c.transact(ctx, token, c.erc20, "approve", spender, amount)
}

// SetApprovalForAll submits an ERC-1155 setApprovalForAll and waits for it to be
// mined, returning the transaction hash.
func (c *Client) SetApprovalForAll(ctx context.Context, token, operator common.Address, approved bool) (string, error) {
	return c.transact(ctx, token, c.erc1155, "setApprovalForAll", operator, approved)
}

// transact sends a signed contract call and waits for a successful receipt.
func (c *Client) transact(ctx context.Context, to common.Address, contractABI abi.ABI, method string, args ...any) (string, error) {
	opts, err := c.transactor(ctx)
	if err != nil {
		return "", err
	}
	contract := bind.NewBoundContract(to, contractABI, c.eth, c.eth, c.eth)
	tx, err := contract.Transact(opts, method, args...)
	if err != nil {
		return "", fmt.Errorf("submitting %s: %w", method, err)
	}
	receipt, err := bind.WaitMined(ctx, c.eth, tx)
	if err != nil {
		return tx.Hash().Hex(), fmt.Errorf("waiting for %s to mine: %w", method, err)
	}
	if receipt.Status != 1 {
		return tx.Hash().Hex(), fmt.Errorf("transaction %s reverted", tx.Hash().Hex())
	}
	return tx.Hash().Hex(), nil
}
