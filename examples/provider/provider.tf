terraform {
  required_providers {
    polymarket = {
      source  = "rangertaha/polymarket"
      version = "~> 0.1"
    }
  }
}

# The public Gamma API requires no credentials; the defaults are sufficient.
provider "polymarket" {
  # endpoint = "https://gamma-api.polymarket.com" # optional override
  # api_key  = var.polymarket_api_key             # optional, also POLYMARKET_API_KEY
}
