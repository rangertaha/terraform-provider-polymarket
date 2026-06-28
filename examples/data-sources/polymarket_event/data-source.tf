# Fetch a single event (a group of related markets) by its numeric ID.
data "polymarket_event" "example" {
  id = "644839"
}

output "event_title" {
  value = data.polymarket_event.example.title
}

output "event_market_questions" {
  value = [for m in data.polymarket_event.example.markets : m.question]
}
