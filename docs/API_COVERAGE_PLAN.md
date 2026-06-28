# Polymarket Provider — Full API Coverage Plan

This document maps every Polymarket API surface onto Terraform data sources and
resources and sequences the work into phases. It is the roadmap from the current
read-only scaffold to full coverage.

## Background: Polymarket's three APIs

Polymarket is not traditional infrastructure, so coverage is framed as **reads**
(data sources) over public data and **writes** (resources) over authenticated,
wallet-signed trading actions.

| API | Base URL | Auth | Purpose |
| --- | --- | --- | --- |
| **Gamma Markets** | `https://gamma-api.polymarket.com` | none | Markets, events, series, tags (UI-facing catalog data) |
| **CLOB** | `https://clob.polymarket.com` | L1 (EIP-712 wallet sig) + L2 (API key/HMAC) | Order book, prices, order placement & cancellation, trades |
| **Data** | `https://data-api.polymarket.com` | none | On-chain positions, holders, portfolio value, activity |

Writes require a Polygon (chain ID `137`) wallet private key and EIP-712 signing
of CTF Exchange orders. That signing layer is the largest single dependency and
gates all resources.

## Status legend

- ✅ Implemented
- 🔜 Next up
- ⬜ Planned
- 🔒 Blocked on the auth/signing layer

---

## Phase 1 — Gamma read coverage (catalog)

Pure HTTP GETs against the public Gamma API. No auth. Extends the existing client.

| Terraform object | Kind | Endpoint | Status |
| --- | --- | --- | --- |
| `polymarket_market` | data source | `GET /markets/{id}` | ✅ |
| `polymarket_markets` | data source | `GET /markets` | ✅ |
| `polymarket_event` | data source | `GET /events/{id}` | ✅ |
| `polymarket_events` | data source | `GET /events` | ✅ |
| `polymarket_series` | data source | `GET /series/{id}` | ✅ |
| `polymarket_tags` | data source | `GET /tags` | ✅ |

> **Phase 1 complete.** Market schema expanded with live pricing (`best_bid`,
> `best_ask`, `last_trade_price`, `spread`), rolling volumes
> (`volume_24hr/1wk/1mo/1yr`), `clob_token_ids`, order-book flags, and
> timestamps. Events embed their full markets and tags; series embed their
> events. All six data sources verified against the live Gamma API.

**Tasks**

1. Add `Event`, `Series`, `Tag` structs to `internal/client` plus `GetEvent`,
   `ListEvents`, `GetSeries`, `ListTags`.
2. Expand the market schema with the remaining Gamma fields already returned by
   the API: `image`, `icon`, `resolution_source`, `volume_24hr`, `competitive`,
   `spread`, `best_bid`, `best_ask`, `clob_token_ids`, `accepting_orders`.
3. Add richer list filters: `tag_id`, `order` (sort key), `ascending`,
   `start_date_min/max`, `liquidity_min`, `volume_min`.
4. Model the market→event relationship (an event groups many markets).

**Exit criteria:** every Gamma catalog object is queryable; all schema
attributes carry both `Description` and `MarkdownDescription`.

---

## Phase 2 — CLOB read coverage (live pricing)

Public CLOB read endpoints — live order book and pricing. Still no auth, but a
second base URL, so generalize the client to address multiple hosts.

| Terraform object | Kind | Endpoint | Status |
| --- | --- | --- | --- |
| `polymarket_order_book` | data source | `GET /book?token_id=` | ✅ |
| `polymarket_price` | data source | `GET /price?token_id=&side=` | ✅ |
| `polymarket_midpoint` | data source | `GET /midpoint?token_id=` | ✅ |
| `polymarket_spread` | data source | `GET /spread?token_id=` | ✅ |
| `polymarket_clob_market` | data source | `GET /markets/{condition_id}` | ⬜ (deferred — largely redundant with the Gamma `market` + `order_book`) |

**Tasks**

1. ✅ Added a `clob_endpoint` provider attribute (default
   `https://clob.polymarket.com`, env `POLYMARKET_CLOB_ENDPOINT`).
2. ✅ Generalized the client transport to address both hosts (`getFrom`).
3. ✅ Modeled order-book levels as a nested list of `{ price, size }`.
4. Future: split `internal/client` into `gamma`/`clob` sub-packages once auth
   (Phase 4) adds more CLOB surface; the shared `getFrom` transport already
   carries retry/error-decode and is the seam for that split.

**Exit criteria (met):** a config resolves a market's CLOB token IDs (Phase 1)
and reads live bids/asks/price/midpoint/spread for them — verified end-to-end.

> **Extended coverage added.** `polymarket_price_history` (`GET /prices-history`)
> for charting, plus batch reads `polymarket_prices` (`POST /prices`) and
> `polymarket_order_books` (`POST /books`) that quote/fetch many tokens in one
> request — all verified live. At the client/SDK layer, `CancelOrders`
> (`DELETE /orders`) and `CancelAll` (`DELETE /cancel-all`) round out order
> management (mock-tested); they are not surfaced as Terraform objects because a
> batch-cancel is an imperative action with no declarative resource analogue.

---

## Phase 3 — Data API read coverage (portfolio)

Public on-chain analytics keyed by wallet address or market.

| Terraform object | Kind | Endpoint | Status |
| --- | --- | --- | --- |
| `polymarket_positions` | data source | `GET /positions?user=` | ✅ |
| `polymarket_holders` | data source | `GET /holders?market=` | ✅ |
| `polymarket_trades` | data source | `GET /trades?user=` | ✅ |
| `polymarket_portfolio_value` | data source | `GET /value?user=` | ✅ |

> **Phase 3 complete.** Added the `data_endpoint` provider attribute (env
> `POLYMARKET_DATA_ENDPOINT`) and a third `getData` transport. Positions carry
> full cost-basis and P&L; trades carry side/price/size/tx hash; holders group
> top wallets per outcome token. All verified live against the Data API.

**Exit criteria (met):** given a wallet address, a config reports current
positions, executed trades, and realized/unrealized value — read-only, no signing.

> **Extended coverage added.** `polymarket_activity` (`GET /activity`) surfaces a
> wallet's unified on-chain feed — trades, rewards, splits, merges, redemptions,
> and conversions — with optional type/market filters. Verified live.

---

## Phase 4 — Authentication & signing layer (foundation for writes) 🔒

The enabling work for every resource. No user-facing objects ship in this phase;
it builds the cryptographic plumbing.

**Provider configuration additions**

| Attribute | Env var | Description |
| --- | --- | --- |
| `private_key` | `POLYMARKET_PRIVATE_KEY` | Polygon wallet key used for EIP-712 signing (sensitive). |
| `funder_address` | `POLYMARKET_FUNDER` | Proxy/funder wallet that holds USDC (for email/magic accounts). |
| `signature_type` | `POLYMARKET_SIGNATURE_TYPE` | `0` EOA, `1` email/magic proxy, `2` browser/Gnosis proxy. |
| `chain_id` | `POLYMARKET_CHAIN_ID` | Defaults to `137` (Polygon mainnet); `80002` Amoy testnet. |

**Tasks**

1. ✅ Added `go-ethereum` for ECDSA + EIP-712 typed-data hashing.
2. ✅ Implemented **L1 auth** (`internal/sign.SignClobAuth`) and wired
   `DeriveAPIKey` / `CreateAPIKey` (`GET /auth/derive-api-key`,
   `POST /auth/api-key`) with the `POLY_ADDRESS/SIGNATURE/TIMESTAMP/NONCE` headers.
3. ✅ Implemented **L2 HMAC** request signing (`internal/sign.BuildHMACSignature`,
   base64url HMAC-SHA256 over `timestamp+method+path+body`).
4. ✅ Implemented CTF Exchange **order EIP-712 signing**
   (`internal/sign.SignOrder`), with EOA + NegRisk verifying contracts and the
   full order struct (salt, maker, signer, taker, tokenId, maker/takerAmount,
   expiration, nonce, feeRateBps, side, signatureType).
5. ✅ Added provider config: `private_key`, `funder_address`, `signature_type`,
   `chain_id` (each with env fallback), building a `*sign.Signer` only when a key
   is present so read-only use needs no credentials.
6. ✅ Surfaced `polymarket_api_credentials` as a (sensitive) data source.

**Verification:** EIP-712 signatures round-trip via signer recovery (sign →
`SigToPub` → address match), domain separation is asserted across verifying
contracts, the address vector matches the canonical scalar-1 test key, the HMAC
matches an independent reference vector, and `DeriveAPIKey` is exercised against
an `httptest` server that checks the L1 headers. Live verification against the
real CLOB still needs a registered/funded Polygon wallet.

**Exit criteria (met):** the provider signs L1 auth and round-trips a signed
order payload under test; an authenticated GET is verified against a mock CLOB.

---

## Phase 5 — Trading resources (writes) 🔒

Depends entirely on Phase 4. This is where Terraform's CRUD model meets an order
book, which is an impedance mismatch worth calling out explicitly.

| Terraform object | Kind | Endpoint | Status |
| --- | --- | --- | --- |
| `polymarket_api_key` | resource | `POST /auth/api-key` / `DELETE` | ✅ |
| `polymarket_order` | resource | `POST /order`, `GET /data/order/{id}`, `DELETE /order` | ✅ |
| `polymarket_allowance` | resource | on-chain ERC-20 `approve` / ERC-1155 `setApprovalForAll` | ✅ (live tx needs a real Polygon RPC + funded wallet) |

> **Phase 5 (HTTP resources) complete.** `polymarket_order` places a signed
> order (Create), refreshes status via `GET /data/order/{id}` and self-removes
> from state when the order is gone (Read), and cancels on destroy (Delete);
> every input forces replacement to model out-of-band fills. `polymarket_api_key`
> creates and revokes an L2 key set. The client derives L2 credentials lazily and
> caches them. Order amount conversion, L2 HMAC auth, and the full
> place/get/cancel flow are unit-tested against a mock CLOB, and the complete
> apply→refresh→destroy lifecycle is verified through real Terraform against a
> local mock.
>
> **`polymarket_allowance` complete.** A new `internal/chain` package (go-ethereum
> `ethclient`/`bind`) manages ERC-20 USDC allowances (`approve`/`allowance`) and
> ERC-1155 CTF operator approvals (`setApprovalForAll`/`isApprovedForAll`) against
> a Polygon RPC endpoint configured via `rpc_endpoint`. The resource creates,
> reads (drift detection), updates the ERC-20 amount, and revokes on destroy. ABI
> selectors and calldata are pinned by unit tests; the allowance read path is
> tested against a mock JSON-RPC server; and the provider plumbing (missing-RPC
> diagnostic, clean configure with an RPC set) is verified through Terraform.
> Live transaction submission still needs a real Polygon RPC and a funded wallet —
> best exercised on the Amoy testnet (`chain_id = 80002`).

**Design notes & open questions**

- **Order lifecycle vs. Terraform state.** An order can fill, partially fill, or
  expire out-of-band — state Terraform did not author. Plan: treat `Create` =
  place order, `Read` = refresh status (`live`/`matched`/`cancelled`), `Delete` =
  cancel. A filled order becomes immutable; `Update` forces replace. Document
  loudly that this is not idempotent infrastructure.
- **GTC vs. FOK/FAK.** Fill-or-kill orders complete or vanish immediately and map
  poorly to a persistent resource; restrict the `order` resource to GTC/GTD and
  expose FOK/FAK via a `polymarket_market_order` ephemeral/action instead.
- **Allowances** are genuinely declarative (idempotent on-chain approvals) and are
  the best-fit resource — implement first within this phase.
- Consider Terraform **ephemeral resources** (TF ≥ 1.10) for derived API keys so
  secrets never land in state.

**Exit criteria:** an authorized user can manage USDC/CTF allowances and place &
cancel a GTC limit order through Terraform, with status reflected on refresh.

---

## Phase 6 — Hardening & release

Cross-cutting quality work, run continuously but gated here for 1.0.

- ✅ **`tfplugindocs`** wired in (`//go:generate`, `make docs`, tool dependency);
  full `docs/` reference generated for all 15 data sources + 3 resources.
- ✅ **CI** (`.github/workflows/test.yml`): build, `go vet`, `gofmt -l`,
  `go test`, `golangci-lint` (0 issues), and a docs-up-to-date check.
- ✅ **Release** (`.github/workflows/release.yml`): GoReleaser + GPG signing on
  `v*` tags, ready for Terraform Registry publication under `rangertaha/polymarket`.
- ✅ **Acceptance test** (`TF_ACC=1`) for the markets data source against the
  live public API; mocked HTTP/RPC for signing, order, and approval logic.
- Future: rate limiting + retries with backoff in the shared transport;
  pagination helpers that follow `next_cursor`/offset; acceptance tests for the
  remaining data sources; testnet (Amoy) verification of the write resources.

---

## Dependency graph

```
Phase 1 (Gamma reads) ──┐
Phase 2 (CLOB reads) ────┼──> Phase 6 (harden/release)
Phase 3 (Data reads) ───┘
Phase 4 (auth/signing) ──> Phase 5 (writes) ──> Phase 6
```

Phases 1–3 are independent and parallelizable. Phase 4 is the critical path for
all writes; Phase 5 cannot start until it lands.

## Client architecture target

```
internal/client/
  transport.go   # shared HTTP: base URL, retries, rate limit, error decode
  gamma/         # markets, events, series, tags
  clob/          # book, price, midpoint, spread, orders, auth (L1/L2)
  data/          # positions, holders, trades, value
  sign/          # EIP-712 typed-data + HMAC request signing
```
