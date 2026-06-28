data "polymarket_positions" "example" {
  user  = "0x349606c1b77f3ba668879cbc9347f15a44cf8fc4"
  limit = 10
}

output "position_titles" {
  value = [for p in data.polymarket_positions.example.positions : p.title]
}
