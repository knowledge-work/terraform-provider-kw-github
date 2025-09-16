package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/knowledge-work/knowledgework/terraform-provider/github-repository-rule/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "knowledge-work/kwgithub",
	})
	if err != nil {
		log.Fatal(err)
	}
}
