# Requires the provider's private_key (or POLYMARKET_PRIVATE_KEY).
provider "polymarket" {
  private_key = var.polymarket_private_key # sensitive
}

# Provision an L2 API key set for the configured wallet.
resource "polymarket_api_key" "trading" {}

output "api_key" {
  value = polymarket_api_key.trading.api_key
}
