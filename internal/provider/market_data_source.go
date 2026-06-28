// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the data source satisfies the required interfaces.
var (
	_ datasource.DataSource              = (*marketDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*marketDataSource)(nil)
)

// marketDataSource looks up a single Polymarket market by ID. The model type
// (marketDataSourceModel) and attribute schema (marketAttributes) live in
// market_schema.go and are shared with the markets list data source.
type marketDataSource struct {
	client *client.Client
}

// NewMarketDataSource is the data source constructor registered on the provider.
func NewMarketDataSource() datasource.DataSource {
	return &marketDataSource{}
}

// Metadata sets the data source type name ("polymarket_market").
func (d *marketDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_market"
}

// Schema describes a single market. The attribute set is shared with the
// markets list data source via marketAttributes so the two never diverge.
func (d *marketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Polymarket prediction market by its numeric ID, " +
			"including its outcomes, current prices, and trading statistics.",
		MarkdownDescription: "Fetches a single Polymarket prediction market by its numeric " +
			"`id`, including its outcomes, current prices, and trading statistics.",
		Attributes: marketAttributes(true),
	}
}

// Configure receives the shared API client from the provider.
func (d *marketDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = c
}

// Read fetches the market and maps it onto Terraform state.
func (d *marketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data marketDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	market, err := d.client.GetMarket(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Polymarket market",
			fmt.Sprintf("Could not fetch market %q: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(flattenMarket(ctx, *market, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
