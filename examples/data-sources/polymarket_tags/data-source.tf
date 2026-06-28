# List the first 20 category tags.
data "polymarket_tags" "all" {
  limit = 20
}

output "tag_labels" {
  value = [for t in data.polymarket_tags.all.tags : t.label]
}
