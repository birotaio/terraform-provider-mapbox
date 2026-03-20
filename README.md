# Terraform Provider for Mapbox

A [Terraform](https://www.terraform.io/) provider for managing [Mapbox](https://www.mapbox.com/) resources including access tokens and map styles.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (to build the provider)

## Resources

| Resource       | Description                                                                |
| -------------- | -------------------------------------------------------------------------- |
| `mapbox_token` | Manages Mapbox access tokens with configurable scopes and URL restrictions |
| `mapbox_style` | Manages Mapbox map styles with sources, layers, and rendering rules        |

## Data Sources

| Data Source    | Description                                 |
| -------------- | ------------------------------------------- |
| `mapbox_token` | Reads an existing Mapbox access token by ID |
| `mapbox_style` | Reads an existing Mapbox map style by ID    |

## Usage

```terraform
terraform {
  required_providers {
    mapbox = {
      source = "birotaio/mapbox"
    }
  }
}

provider "mapbox" {
  access_token = var.mapbox_access_token  # Or set MAPBOX_ACCESS_TOKEN
  username     = var.mapbox_username      # Or set MAPBOX_USERNAME
}

resource "mapbox_token" "api" {
  note   = "API access token"
  scopes = ["styles:read", "fonts:read"]
}

resource "mapbox_style" "main" {
  name    = "Production Style"
  version = 8
  sources = jsonencode({})
  layers  = jsonencode([])
}
```

## Authentication

The provider requires a Mapbox **secret access token** (`sk.*`) and your **username**. These can be set in the provider configuration or via environment variables:

| Configuration  | Environment Variable  | Description                                 |
| -------------- | --------------------- | ------------------------------------------- |
| `access_token` | `MAPBOX_ACCESS_TOKEN` | Mapbox secret token with appropriate scopes |
| `username`     | `MAPBOX_USERNAME`     | Mapbox account username                     |

## Development

### Building

```shell
make build
```

### Testing

Run unit tests:

```shell
make test
```

Run acceptance tests (requires valid Mapbox credentials):

```shell
export MAPBOX_ACCESS_TOKEN="sk.your-secret-token"
export MAPBOX_USERNAME="your-username"
make testacc
```

### Generating Documentation

```shell
make generate
```

## Project Structure

```
├── main.go                       # Provider entry point
├── internal/
│   ├── provider/                 # Provider configuration and registration
│   ├── mapbox/                   # Mapbox API client (no Terraform dependencies)
│   ├── resources/                # Terraform resource implementations
│   └── datasources/             # Terraform data source implementations
├── docs/                         # Provider documentation
├── examples/                     # HCL usage examples
└── tools/                        # Documentation generation tools
```

## License

MIT License. See [LICENSE](LICENSE) for details.
