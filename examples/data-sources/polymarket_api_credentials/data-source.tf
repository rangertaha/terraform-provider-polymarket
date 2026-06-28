# Requires the provider's private_key (or POLYMARKET_PRIVATE_KEY) to be set.
provider "polymarket" {
  private_key = var.polymarket_private_key # sensitive
}

# Derives the deterministic L2 API credentials for the configured wallet.
data "polymarket_api_credentials" "creds" {}

output "api_key" {
  value = data.polymarket_api_credentials.creds.api_key
}

# secret and passphrase are sensitive; reference them in downstream providers
# rather than printing them.
