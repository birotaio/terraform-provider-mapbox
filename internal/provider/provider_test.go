package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mapbox": providerserver.NewProtocol6WithError(New("test")()),
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
