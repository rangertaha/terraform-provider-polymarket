# Requires the provider's private_key and rpc_endpoint. Submits real Polygon
# transactions that cost gas.
provider "polymarket" {
  private_key  = var.polymarket_private_key # sensitive
  rpc_endpoint = "https://polygon-rpc.com"
}

locals {
  usdc     = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174" # USDC.e
  ctf      = "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045" # Conditional Tokens
  exchange = "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E" # CTF Exchange
}

# Approve the exchange to spend USDC (unlimited, the default).
resource "polymarket_allowance" "usdc_spend" {
  token   = local.usdc
  spender = local.exchange
}

# Approve the exchange as an operator of the CTF outcome tokens (ERC-1155).
resource "polymarket_allowance" "ctf_operator" {
  token   = local.ctf
  spender = local.exchange
  erc1155 = true
}
