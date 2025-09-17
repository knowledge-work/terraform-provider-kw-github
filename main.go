package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/knowledge-work/knowledgework/terraform-provider/github-repository-rule/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/knowledge-work/kw-github",
	})
	if err != nil {
		log.Fatal(err)
	}
}
