# Read a live order book. Token IDs come from a market's clob_token_ids.
data "polymarket_market" "m" { id = "540817" }

data "polymarket_order_book" "yes" {
  token_id = data.polymarket_market.m.clob_token_ids[0]
}

output "best_bid" { value = try(data.polymarket_order_book.yes.bids[0], null) }
output "best_ask" { value = try(data.polymarket_order_book.yes.asks[0], null) }
