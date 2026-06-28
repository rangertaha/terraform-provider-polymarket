# Requires the provider's private_key (or POLYMARKET_PRIVATE_KEY).
provider "polymarket" {
  private_key = var.polymarket_private_key # sensitive
}

# Look up the outcome token to trade.
data "polymarket_market" "m" { id = "540817" }

# Place a good-till-cancelled limit buy for 100 shares of "Yes" at 0.55.
resource "polymarket_order" "buy_yes" {
  token_id = data.polymarket_market.m.clob_token_ids[0]
  side     = "BUY"
  price    = 0.55
  size     = 100
}

output "order_status" {
  value = polymarket_order.buy_yes.status
}
