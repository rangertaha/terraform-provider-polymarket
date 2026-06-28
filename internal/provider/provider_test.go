// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories registers the provider for acceptance tests.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"polymarket": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates preconditions before acceptance tests run. The
// public Gamma/CLOB/Data read APIs require no credentials, so there is nothing
// to assert yet; authenticated tests would validate POLYMARKET_PRIVATE_KEY here.
func testAccPreCheck(t *testing.T) {
	t.Helper()
}

// TestProvider verifies the provider registers its data sources and resources.
// It runs without network access, catching registration/schema wiring errors.
func TestProvider(t *testing.T) {
	p := New("test")()

	if got := len(p.DataSources(context.Background())); got != 15 {
		t.Errorf("expected 15 data sources, got %d", got)
	}
	if got := len(p.Resources(context.Background())); got != 3 {
		t.Errorf("expected 3 resources, got %d", got)
	}
}

// TestAccMarketsDataSource exercises the markets data source against the live
// public Gamma API. Gated behind TF_ACC, so it is skipped by a plain `go test`.
func TestAccMarketsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "polymarket_markets" "test" {
  limit = 1
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.polymarket_markets.test", "markets.#"),
				),
			},
		},
	})
}
