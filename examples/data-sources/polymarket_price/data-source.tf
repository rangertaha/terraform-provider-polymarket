data "polymarket_market" "m" { id = "540817" }

data "polymarket_price" "buy" {
  token_id = data.polymarket_market.m.clob_token_ids[0]
  side     = "buy"
}

output "best_buy_price" { value = data.polymarket_price.buy.price }
