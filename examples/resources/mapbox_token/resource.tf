resource "mapbox_token" "example" {
  note   = "My API token"
  scopes = ["styles:read", "fonts:read"]

  allowed_urls = [
    "https://example.com",
  ]
}

output "token_id" {
  value = mapbox_token.example.id
}
