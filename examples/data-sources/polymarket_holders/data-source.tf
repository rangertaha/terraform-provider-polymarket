# Top holders of each outcome token in a market, by condition ID.
data "polymarket_market" "m" { id = "540817" }

data "polymarket_holders" "top" {
  market = data.polymarket_market.m.condition_id
  limit  = 5
}

output "holder_count" {
  value = sum([for g in data.polymarket_holders.top.tokens : length(g.holders)])
}
