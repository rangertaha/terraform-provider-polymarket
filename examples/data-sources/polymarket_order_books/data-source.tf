data "polymarket_market" "m" { id = "540817" }

# Fetch the order book for every outcome of the market in one request.
data "polymarket_order_books" "all" {
  token_ids = data.polymarket_market.m.clob_token_ids
}

output "best_asks" {
  value = [for b in data.polymarket_order_books.all.books : try(b.asks[0].price, null)]
}
