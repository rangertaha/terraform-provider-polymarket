# List up to 25 active, open markets.
data "polymarket_markets" "active" {
  active = true
  closed = false
  limit  = 25
}

output "active_market_questions" {
  value = [for m in data.polymarket_markets.active.markets : m.question]
}
