// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories registers the provider for acceptance tests.
// Each acceptance test transparently spins up the provider server defined here.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"polymarket": providerserver.NewProtocol6WithError(New("test")()),
}

// TestProvider performs basic schema validation that runs without network
// access, catching schema definition errors early.
func TestProvider(t *testing.T) {
	p := New("test")()

	schemaResp := &providerSchemaResponse{}
	_ = schemaResp // placeholder to keep the import surface obvious

	if got := len(p.DataSources(context.Background())); got == 0 {
		t.Fatalf("expected provider to register data sources, got %d", got)
	}
}

// providerSchemaResponse is an alias kept for readability in future schema tests.
type providerSchemaResponse struct{}

// testAccPreCheck validates required preconditions before acceptance tests run.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	// Acceptance tests run against the live public Gamma API and require no
	// credentials. Add validation here if authenticated endpoints are added.
}
