package main

import (
	"context"
	"log"

	"github.com/acecloud/terraform-provider-acecloud/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is set at build time via ldflags.
var version = "dev"

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/acecloud/acecloud",
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
