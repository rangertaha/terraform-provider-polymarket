data "polymarket_market" "m" { id = "540817" }

data "polymarket_price_history" "yes" {
  token_id = data.polymarket_market.m.clob_token_ids[0]
  interval = "1w"
  fidelity = 60 # one sample per hour
}

output "samples" {
  value = length(data.polymarket_price_history.yes.history)
}
