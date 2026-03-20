data "mapbox_token" "example" {
  id = "cilnk7n1y00opstl3g1qe4ick"
}

output "token_note" {
  value = data.mapbox_token.example.note
}
