# Fetch a single market by its numeric ID.
data "polymarket_market" "example" {
  id = "253123"
}

output "market_question" {
  value = data.polymarket_market.example.question
}

output "market_outcomes" {
  # Pairs each outcome with its current implied probability.
  value = zipmap(
    data.polymarket_market.example.outcomes,
    data.polymarket_market.example.outcome_prices,
  )
}
