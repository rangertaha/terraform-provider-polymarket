# Changelog

All notable changes to this provider are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-06-28

Initial release. Covers Polymarket's three public APIs (Gamma catalog, CLOB
order book, and Data/portfolio), authenticated CLOB trading, and on-chain
trading approvals.

### Added

#### Provider configuration
- `endpoint`, `clob_endpoint`, `data_endpoint` — API base URLs (with
  `POLYMARKET_*` environment-variable fallbacks).
- `api_key` — optional bearer token for the Gamma API.
- `private_key`, `funder_address`, `signature_type`, `chain_id` — wallet
  signing configuration for authenticated CLOB requests and orders.
- `rpc_endpoint` — Polygon JSON-RPC endpoint for on-chain approvals.

#### Data sources (19)
- **Gamma catalog:** `polymarket_market`, `polymarket_markets`,
  `polymarket_event`, `polymarket_events`, `polymarket_series`,
  `polymarket_tags`.
- **CLOB pricing:** `polymarket_order_book`, `polymarket_order_books` (batch),
  `polymarket_price`, `polymarket_prices` (batch), `polymarket_midpoint`,
  `polymarket_spread`, `polymarket_price_history`.
- **Data / portfolio:** `polymarket_positions`, `polymarket_trades`,
  `polymarket_activity`, `polymarket_portfolio_value`, `polymarket_holders`.
- **Auth:** `polymarket_api_credentials` (derives L2 credentials).

#### Resources (3)
- `polymarket_order` — place and cancel CLOB limit orders.
- `polymarket_api_key` — provision and revoke L2 API key sets.
- `polymarket_allowance` — manage on-chain ERC-20 / ERC-1155 trading approvals.

#### Internals
- EIP-712 signing (`ClobAuth` L1, CTF Exchange orders) and L2 HMAC request
  signing.
- Three-host HTTP client with lazily-derived, cached L2 credentials.
- On-chain client for ERC-20 `approve` / ERC-1155 `setApprovalForAll`.
- `CancelOrders` and `CancelAll` client/SDK methods.
- Generated Terraform Registry documentation, CI, and a signed-release
  workflow.

[Unreleased]: https://github.com/rangertaha/terraform-provider-polymarket/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/rangertaha/terraform-provider-polymarket/releases/tag/v0.1.0
