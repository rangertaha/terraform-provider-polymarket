# Fetch a recurring series (a group of events) by its numeric ID.
data "polymarket_series" "example" {
  id = "10000"
}

output "series_title" {
  value = data.polymarket_series.example.title
}

output "series_event_count" {
  value = length(data.polymarket_series.example.events)
}
