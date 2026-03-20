resource "mapbox_style" "example" {
  name    = "My Custom Style"
  version = 8

  sources = jsonencode({
    "mapbox-streets" = {
      type = "vector"
      url  = "mapbox://mapbox.mapbox-streets-v8"
    }
  })

  layers = jsonencode([
    {
      id   = "background"
      type = "background"
      paint = {
        "background-color" = "#f8f4f0"
      }
    }
  ])

  visibility = "private"
}

output "style_id" {
  value = mapbox_style.example.id
}
