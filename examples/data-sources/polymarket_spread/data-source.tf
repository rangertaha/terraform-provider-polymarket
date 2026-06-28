data "polymarket_market" "m" { id = "540817" }

data "polymarket_spread" "yes" {
  token_id = data.polymarket_market.m.clob_token_ids[0]
}

output "spread" { value = data.polymarket_spread.yes.spread }
