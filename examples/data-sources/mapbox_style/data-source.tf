data "mapbox_style" "example" {
  id = "cj3kbeqzo00052rp1pklhzqwu"
}

output "style_name" {
  value = data.mapbox_style.example.name
}
