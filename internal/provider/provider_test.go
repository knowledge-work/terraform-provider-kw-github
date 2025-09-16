
package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "kw_github" {
	owner = "test"
}
`
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"kw_github": providerserver.NewProtocol6WithError(New()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("GITHUB_TOKEN"); v == "" {
		t.Fatal("GITHUB_TOKEN must be set for acceptance tests")
	}
}
