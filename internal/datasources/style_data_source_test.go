package datasources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccStyleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStyleDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.mapbox_style.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("tf-acc-test-ds-style"),
					),
				},
			},
		},
	})
}

const testAccStyleDataSourceConfig = `
resource "mapbox_style" "test" {
  name    = "tf-acc-test-ds-style"
  version = 8
  sources = jsonencode({})
  layers  = jsonencode([])
}

data "mapbox_style" "test" {
  id = mapbox_style.test.id
}
`
