data "polymarket_portfolio_value" "example" {
  user = "0x349606c1b77f3ba668879cbc9347f15a44cf8fc4"
}

output "portfolio_value_usdc" {
  value = data.polymarket_portfolio_value.example.value
}
