// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// These are set at build time via -ldflags by goreleaser.
var (
	// version is the provider version, overridden during release builds.
	version string = "dev"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name polymarket

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// The registry address used to source the provider in Terraform
		// configurations (e.g. required_providers).
		Address: "registry.terraform.io/rangertaha/polymarket",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
