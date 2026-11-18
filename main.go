package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/coreplanelabs/terraform-provider-parasail/internal/buildinfo"
	"github.com/coreplanelabs/terraform-provider-parasail/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New(buildinfo.Version), providerserver.ServeOpts{
		Address: "registry.terraform.io/coreplanelabs/parasail",
	})
	if err != nil {
		log.Fatal(err)
	}
}
