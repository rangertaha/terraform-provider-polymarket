# List up to 10 active, open events with their embedded markets.
data "polymarket_events" "active" {
  active = true
  closed = false
  limit  = 10
}

output "event_titles" {
  value = [for e in data.polymarket_events.active.events : e.title]
}
