data "polymarket_trades" "example" {
  user  = "0x349606c1b77f3ba668879cbc9347f15a44cf8fc4"
  limit = 10
}

output "trades" {
  value = [for t in data.polymarket_trades.example.trades : "${t.side} ${t.size} @ ${t.price}"]
}
