package resources_test

import (
	"fmt"
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

func TestAccTokenResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTokenResourceConfig("tf-acc-test-token", []string{"styles:read", "fonts:read"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mapbox_token.test",
						tfjsonpath.New("note"),
						knownvalue.StringExact("tf-acc-test-token"),
					),
					statecheck.ExpectKnownValue(
						"mapbox_token.test",
						tfjsonpath.New("usage"),
						knownvalue.StringExact("pk"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:            "mapbox_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
			// Update and Read testing
			{
				Config: testAccTokenResourceConfig("tf-acc-test-token-updated", []string{"styles:read", "fonts:read", "styles:list"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mapbox_token.test",
						tfjsonpath.New("note"),
						knownvalue.StringExact("tf-acc-test-token-updated"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTokenResourceConfig(note string, scopes []string) string {
	scopesList := ""
	for i, s := range scopes {
		if i > 0 {
			scopesList += ", "
		}
		scopesList += fmt.Sprintf("%q", s)
	}

	return fmt.Sprintf(`
resource "mapbox_token" "test" {
  note   = %q
  scopes = [%s]
}
`, note, scopesList)
}
