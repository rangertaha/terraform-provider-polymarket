// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// configureClient is the shared Configure logic for the lightweight CLOB quote
// data sources, returning the typed client or recording a diagnostic.
func configureClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *client.Client {
	if req.ProviderData == nil {
		return nil
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return nil
	}
	return c
}

// tokenIDAttribute is the shared required input for CLOB quote data sources.
func tokenIDAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Required: true,
		Description: "CLOB ERC-1155 token ID of the outcome to quote. Obtain it from a " +
			"market's clob_token_ids attribute.",
		MarkdownDescription: "CLOB ERC-1155 token ID of the outcome to quote. Obtain it from a " +
			"market's `clob_token_ids` attribute.",
	}
}

// ---------------------------------------------------------------------------
// polymarket_price
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = (*priceDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*priceDataSource)(nil)
)

type priceDataSource struct{ client *client.Client }

type priceDataSourceModel struct {
	TokenID types.String `tfsdk:"token_id"`
	Side    types.String `tfsdk:"side"`
	Price   types.String `tfsdk:"price"`
}

// NewPriceDataSource is the data source constructor registered on the provider.
func NewPriceDataSource() datasource.DataSource { return &priceDataSource{} }

func (d *priceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_price"
}

func (d *priceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the best available CLOB price for an outcome token on a given " +
			"side (buy or sell).",
		MarkdownDescription: "Reads the best available CLOB price for an outcome token on a " +
			"given side (`buy` or `sell`).",
		Attributes: map[string]schema.Attribute{
			"token_id": tokenIDAttribute(),
			"side": schema.StringAttribute{
				Required: true,
				Description: "Side to quote: \"buy\" returns the best ask (lowest sell price); " +
					"\"sell\" returns the best bid (highest buy price).",
				MarkdownDescription: "Side to quote: `buy` returns the best ask (lowest sell " +
					"price); `sell` returns the best bid (highest buy price).",
			},
			"price": schema.StringAttribute{
				Computed:            true,
				Description:         "Best price for the requested side, in [0, 1], as a decimal string.",
				MarkdownDescription: "Best price for the requested side, in `[0, 1]`, as a decimal string.",
			},
		},
	}
}

func (d *priceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *priceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data priceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	price, err := d.client.GetPrice(ctx, data.TokenID.ValueString(), data.Side.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Polymarket price",
			fmt.Sprintf("Could not fetch %s price for token %q: %s", data.Side.ValueString(), data.TokenID.ValueString(), err),
		)
		return
	}

	data.Price = types.StringValue(price)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------------------------------------------------------------------------
// polymarket_midpoint
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = (*midpointDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*midpointDataSource)(nil)
)

type midpointDataSource struct{ client *client.Client }

type midpointDataSourceModel struct {
	TokenID  types.String `tfsdk:"token_id"`
	Midpoint types.String `tfsdk:"midpoint"`
}

// NewMidpointDataSource is the data source constructor registered on the provider.
func NewMidpointDataSource() datasource.DataSource { return &midpointDataSource{} }

func (d *midpointDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_midpoint"
}

func (d *midpointDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reads the CLOB midpoint price (halfway between best bid and best ask) for an outcome token.",
		MarkdownDescription: "Reads the CLOB midpoint price (halfway between best bid and best ask) for an outcome token.",
		Attributes: map[string]schema.Attribute{
			"token_id": tokenIDAttribute(),
			"midpoint": schema.StringAttribute{
				Computed:            true,
				Description:         "Midpoint between the best bid and best ask, in [0, 1], as a decimal string.",
				MarkdownDescription: "Midpoint between the best bid and best ask, in `[0, 1]`, as a decimal string.",
			},
		},
	}
}

func (d *midpointDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *midpointDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data midpointDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mid, err := d.client.GetMidpoint(ctx, data.TokenID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Polymarket midpoint",
			fmt.Sprintf("Could not fetch midpoint for token %q: %s", data.TokenID.ValueString(), err),
		)
		return
	}

	data.Midpoint = types.StringValue(mid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------------------------------------------------------------------------
// polymarket_spread
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = (*spreadDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*spreadDataSource)(nil)
)

type spreadDataSource struct{ client *client.Client }

type spreadDataSourceModel struct {
	TokenID types.String `tfsdk:"token_id"`
	Spread  types.String `tfsdk:"spread"`
}

// NewSpreadDataSource is the data source constructor registered on the provider.
func NewSpreadDataSource() datasource.DataSource { return &spreadDataSource{} }

func (d *spreadDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spread"
}

func (d *spreadDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reads the current CLOB bid-ask spread for an outcome token.",
		MarkdownDescription: "Reads the current CLOB bid-ask spread for an outcome token.",
		Attributes: map[string]schema.Attribute{
			"token_id": tokenIDAttribute(),
			"spread": schema.StringAttribute{
				Computed:            true,
				Description:         "Difference between the best ask and best bid, as a decimal string.",
				MarkdownDescription: "Difference between the best ask and best bid, as a decimal string.",
			},
		},
	}
}

func (d *spreadDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *spreadDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data spreadDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spread, err := d.client.GetSpread(ctx, data.TokenID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Polymarket spread",
			fmt.Sprintf("Could not fetch spread for token %q: %s", data.TokenID.ValueString(), err),
		)
		return
	}

	data.Spread = types.StringValue(spread)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
