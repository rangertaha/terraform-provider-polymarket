# Terraform Provider for Polymarket

A [Terraform](https://www.terraform.io) provider for reading prediction-market
data from [Polymarket](https://polymarket.com) via its public
[Gamma Markets API](https://docs.polymarket.com). Built on the modern
[terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework).

> Status: **scaffold / early development.** Read-only data sources are
> implemented today. See [API Coverage Plan](#api-coverage-plan) for the roadmap.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://go.dev/dl/) >= 1.24 (to build the provider)

## Using the provider

```hcl
terraform {
  required_providers {
    polymarket = {
      source  = "Rangertaha/polymarket"
      version = "~> 0.1"
    }
  }
}

provider "polymarket" {}

data "polymarket_markets" "active" {
  active = true
  closed = false
  limit  = 10
}

output "questions" {
  value = [for m in data.polymarket_markets.active.markets : m.question]
}
```

### Provider configuration

| Argument   | Env var               | Default                            | Description                                   |
| ---------- | --------------------- | ---------------------------------- | --------------------------------------------- |
| `endpoint` | `POLYMARKET_ENDPOINT` | `https://gamma-api.polymarket.com` | Base URL of the Gamma API.                    |
| `api_key`  | `POLYMARKET_API_KEY`  | _(none)_                           | Optional bearer token; not needed for public data. |

## Data sources

- `polymarket_market` — fetch a single market by numeric ID.
- `polymarket_markets` — list markets with filtering and pagination.

See [`examples/`](./examples) for runnable configurations.

## Developing

```sh
make build      # compile the provider binary
make test       # unit tests
make testacc    # acceptance tests (hit the live API)
make docs       # regenerate registry docs from schema descriptions
make fmt vet    # format and vet
```

To test locally, add a dev override to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "Rangertaha/polymarket" = "/path/to/your/GOBIN"
  }
  direct {}
}
```

## API Coverage Plan

The full roadmap toward complete Polymarket API coverage lives in
[`docs/API_COVERAGE_PLAN.md`](./docs/API_COVERAGE_PLAN.md).

## License

[Mozilla Public License 2.0](./LICENSE)
