package datasources_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/birotaio/terraform-provider-mapbox/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mapbox": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("MAPBOX_ACCESS_TOKEN") == "" {
		t.Fatal("MAPBOX_ACCESS_TOKEN must be set for acceptance tests")
	}
	if os.Getenv("MAPBOX_USERNAME") == "" {
		t.Fatal("MAPBOX_USERNAME must be set for acceptance tests")
	}
}

func TestAccTokenDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTokenDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.mapbox_token.test",
						tfjsonpath.New("note"),
						knownvalue.StringExact("tf-acc-test-ds-token"),
					),
				},
			},
		},
	})
}

const testAccTokenDataSourceConfig = `
resource "mapbox_token" "test" {
  note   = "tf-acc-test-ds-token"
  scopes = ["styles:read", "fonts:read"]
}

data "mapbox_token" "test" {
  id = mapbox_token.test.id
}
`
