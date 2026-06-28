data "polymarket_market" "m" { id = "540817" }

# Best buy/sell for every outcome of the market in one request.
data "polymarket_prices" "all" {
  token_ids = data.polymarket_market.m.clob_token_ids
}

output "prices" {
  value = { for p in data.polymarket_prices.all.prices : p.token_id => "${p.buy}/${p.sell}" }
}
