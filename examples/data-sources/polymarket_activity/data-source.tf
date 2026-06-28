# A wallet's unified on-chain activity: trades, rewards, splits, merges,
# redemptions, and conversions.
data "polymarket_activity" "recent" {
  user  = "0x349606c1b77f3ba668879cbc9347f15a44cf8fc4"
  limit = 20
}

# Just the reward entries, for example.
output "rewards" {
  value = [for a in data.polymarket_activity.recent.activity : a if a.type == "REWARD"]
}
