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
| `polymarket_event` | data source | `GET /events/{id}` | 🔜 |
| `polymarket_events` | data source | `GET /events` | 🔜 |
| `polymarket_series` | data source | `GET /series/{id}` | ⬜ |
| `polymarket_tags` | data source | `GET /tags` | ⬜ |

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
| `polymarket_clob_market` | data source | `GET /markets/{condition_id}` | ⬜ |
| `polymarket_order_book` | data source | `GET /book?token_id=` | ⬜ |
| `polymarket_price` | data source | `GET /price?token_id=&side=` | ⬜ |
| `polymarket_midpoint` | data source | `GET /midpoint?token_id=` | ⬜ |
| `polymarket_spread` | data source | `GET /spread?token_id=` | ⬜ |

**Tasks**

1. Refactor `internal/client` into `gamma` and `clob` sub-clients sharing a base
   transport (retry, rate-limit, error decoding).
2. Add a `clob_endpoint` provider attribute (default
   `https://clob.polymarket.com`, env `POLYMARKET_CLOB_ENDPOINT`).
3. Model order-book levels as a nested list of `{ price, size }`.

**Exit criteria:** a config can resolve a market's CLOB token IDs (Phase 1) and
read live bids/asks/midpoint for them.

---

## Phase 3 — Data API read coverage (portfolio)

Public on-chain analytics keyed by wallet address or market.

| Terraform object | Kind | Endpoint | Status |
| --- | --- | --- | --- |
| `polymarket_positions` | data source | `GET /positions?user=` | ⬜ |
| `polymarket_holders` | data source | `GET /holders?market=` | ⬜ |
| `polymarket_trades` | data source | `GET /trades` | ⬜ |
| `polymarket_portfolio_value` | data source | `GET /value?user=` | ⬜ |

**Exit criteria:** given a wallet address, a config can report current positions
and realized/unrealized value — read-only, no signing.

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

1. Add `go-ethereum` for ECDSA + EIP-712 typed-data hashing.
2. Implement **L1 auth**: sign the CLOB auth payload to derive/create API
   credentials (`POST /auth/api-key`, `GET /auth/derive-api-key`).
3. Implement **L2 auth**: HMAC request signing (`POLY_ADDRESS`, `POLY_SIGNATURE`,
   `POLY_TIMESTAMP`, `POLY_API_KEY`, `POLY_PASSPHRASE` headers).
4. Implement CTF Exchange **order struct EIP-712 signing** (salt, maker, taker,
   tokenId, makerAmount, takerAmount, side, expiration, nonce, feeRateBps).
5. Surface `polymarket_api_credentials` as a data source so users can inspect the
   derived key set without writing.

**Exit criteria:** the provider can perform an authenticated, signed GET against
the CLOB and round-trip a signed (but unsubmitted) order payload in unit tests
against known vectors.

---

## Phase 5 — Trading resources (writes) 🔒

Depends entirely on Phase 4. This is where Terraform's CRUD model meets an order
book, which is an impedance mismatch worth calling out explicitly.

| Terraform object | Kind | Endpoint | Status |
| --- | --- | --- | --- |
| `polymarket_api_key` | resource | `POST /auth/api-key` / `DELETE` | 🔒 |
| `polymarket_order` | resource | `POST /order`, `DELETE /order` | 🔒 |
| `polymarket_allowance` | resource | on-chain ERC-20 / CTF `approve` | 🔒 |

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

- **Acceptance tests** (`TF_ACC=1`) for every data source against the live public
  APIs; mocked HTTP for signing/order logic.
- **Rate limiting & retries** with backoff in the shared transport.
- **Pagination helpers** that transparently follow `next_cursor` / offset.
- **`tfplugindocs`** generation wired into CI (`make docs`); every attribute
  already carries descriptions so docs render fully.
- **CI**: `go test`, `go vet`, `gofmt -l`, `golangci-lint`, and a Terraform
  `validate` over `examples/`.
- **Release**: GoReleaser + GPG signing (`.goreleaser.yml`) and Terraform Registry
  publication under `Rangertaha/polymarket`.

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
