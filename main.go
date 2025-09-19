package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/knowledge-work/terraform-provider-kw-github/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/knowledge-work/kw-github",
	})
	if err != nil {
		log.Fatal(err)
	}
}
