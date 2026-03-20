package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccStyleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccStyleResourceConfig("tf-acc-test-style"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mapbox_style.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("tf-acc-test-style"),
					),
					statecheck.ExpectKnownValue(
						"mapbox_style.test",
						tfjsonpath.New("version"),
						knownvalue.Int64Exact(8),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:            "mapbox_style.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sources", "layers", "metadata"},
			},
			// Update and Read testing
			{
				Config: testAccStyleResourceConfig("tf-acc-test-style-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mapbox_style.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("tf-acc-test-style-updated"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccStyleResourceConfig(name string) string {
	return `
resource "mapbox_style" "test" {
  name    = "` + name + `"
  version = 8
  sources = jsonencode({})
  layers  = jsonencode([])
}
`
}
