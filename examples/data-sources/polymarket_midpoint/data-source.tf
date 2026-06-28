data "polymarket_market" "m" { id = "540817" }

data "polymarket_midpoint" "yes" {
  token_id = data.polymarket_market.m.clob_token_ids[0]
}

output "midpoint" { value = data.polymarket_midpoint.yes.midpoint }
