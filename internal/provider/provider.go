// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

// Package provider implements the Polymarket Terraform provider.
package provider

import (
	"context"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure polymarketProvider satisfies the provider.Provider interface.
var _ provider.Provider = (*polymarketProvider)(nil)

// polymarketProvider is the provider implementation.
type polymarketProvider struct {
	// version is set to the provider release version, or "dev" for local builds.
	version string
}

// polymarketProviderModel maps provider schema attributes to Go types.
type polymarketProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

// New returns a function that constructs the provider, capturing the build
// version so it can be reported to Terraform.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &polymarketProvider{version: version}
	}
}

// Metadata sets the provider type name and version. The TypeName becomes the
// prefix for every resource and data source (e.g. "polymarket_market").
func (p *polymarketProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "polymarket"
	resp.Version = p.version
}

// Schema defines the provider-level configuration block.
func (p *polymarketProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Polymarket provider reads prediction-market data from the " +
			"public Polymarket Gamma Markets API. Configure the optional endpoint and " +
			"API key, or rely on the public defaults which require no credentials.",
		MarkdownDescription: "The Polymarket provider reads prediction-market data from the " +
			"public [Polymarket Gamma Markets API](https://docs.polymarket.com). Configure " +
			"the optional `endpoint` and `api_key`, or rely on the public defaults which " +
			"require no credentials.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the Polymarket Gamma API. Defaults to " +
					"\"https://gamma-api.polymarket.com\". May also be set with the " +
					"POLYMARKET_ENDPOINT environment variable. Override this to target a " +
					"proxy or a mock server in tests.",
				MarkdownDescription: "Base URL of the Polymarket Gamma API. Defaults to " +
					"`https://gamma-api.polymarket.com`. May also be set with the " +
					"`POLYMARKET_ENDPOINT` environment variable. Override this to target a " +
					"proxy or a mock server in tests.",
			},
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "Optional API key forwarded as a bearer token. The public " +
					"Gamma endpoints do not require authentication, so this is only needed " +
					"for authenticated or rate-lifted access. May also be set with the " +
					"POLYMARKET_API_KEY environment variable.",
				MarkdownDescription: "Optional API key forwarded as a bearer token. The public " +
					"Gamma endpoints do not require authentication, so this is only needed " +
					"for authenticated or rate-lifted access. May also be set with the " +
					"`POLYMARKET_API_KEY` environment variable.",
			},
		},
	}
}

// Configure builds the shared API client from provider configuration and
// environment variables, then hands it to data sources and resources.
func (p *polymarketProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config polymarketProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration value precedence: explicit config > environment > default.
	endpoint := firstNonEmpty(config.Endpoint, "POLYMARKET_ENDPOINT", client.DefaultEndpoint)
	apiKey := firstNonEmpty(config.APIKey, "POLYMARKET_API_KEY", "")

	c := client.New(
		client.WithEndpoint(endpoint),
		client.WithAPIKey(apiKey),
	)

	// Make the client available to data sources and resources.
	resp.DataSourceData = c
	resp.ResourceData = c
}

// DataSources registers the provider's data sources.
func (p *polymarketProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMarketDataSource,
		NewMarketsDataSource,
		NewEventDataSource,
		NewEventsDataSource,
		NewSeriesDataSource,
		NewTagsDataSource,
	}
}

// Resources registers the provider's managed resources. Polymarket data is
// read-only today, so none are exposed yet; add order/position resources here.
func (p *polymarketProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}
