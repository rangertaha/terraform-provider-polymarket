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

// Ensure the data source satisfies the required interfaces.
var (
	_ datasource.DataSource              = (*marketDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*marketDataSource)(nil)
)

// marketDataSource looks up a single Polymarket market by ID.
type marketDataSource struct {
	client *client.Client
}

// marketDataSourceModel maps the data source schema to Go types.
type marketDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Question      types.String `tfsdk:"question"`
	Slug          types.String `tfsdk:"slug"`
	Description   types.String `tfsdk:"description"`
	Active        types.Bool   `tfsdk:"active"`
	Closed        types.Bool   `tfsdk:"closed"`
	Archived      types.Bool   `tfsdk:"archived"`
	Liquidity     types.String `tfsdk:"liquidity"`
	Volume        types.String `tfsdk:"volume"`
	StartDate     types.String `tfsdk:"start_date"`
	EndDate       types.String `tfsdk:"end_date"`
	ConditionID   types.String `tfsdk:"condition_id"`
	Outcomes      types.List   `tfsdk:"outcomes"`
	OutcomePrices types.List   `tfsdk:"outcome_prices"`
}

// NewMarketDataSource is the data source constructor registered on the provider.
func NewMarketDataSource() datasource.DataSource {
	return &marketDataSource{}
}

// Metadata sets the data source type name ("polymarket_market").
func (d *marketDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_market"
}

// Schema describes a single market. Every attribute is documented so the schema
// renders cleanly in the Terraform Registry documentation.
func (d *marketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Polymarket prediction market by its numeric ID, " +
			"including its outcomes, current prices, and trading statistics.",
		MarkdownDescription: "Fetches a single Polymarket prediction market by its numeric " +
			"`id`, including its outcomes, current prices, and trading statistics.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
				Description: "Numeric identifier of the market to fetch, as assigned by " +
					"Polymarket (for example \"253123\"). This is the only required input.",
				MarkdownDescription: "Numeric identifier of the market to fetch, as assigned " +
					"by Polymarket (for example `253123`). This is the only required input.",
			},
			"question": schema.StringAttribute{
				Computed: true,
				Description: "Human-readable question the market resolves, e.g. " +
					"\"Will candidate X win the 2024 election?\".",
				MarkdownDescription: "Human-readable question the market resolves, e.g. " +
					"`Will candidate X win the 2024 election?`.",
			},
			"slug": schema.StringAttribute{
				Computed:            true,
				Description:         "URL-friendly slug identifying the market on polymarket.com.",
				MarkdownDescription: "URL-friendly slug identifying the market on polymarket.com.",
			},
			"description": schema.StringAttribute{
				Computed: true,
				Description: "Long-form description and resolution criteria explaining how " +
					"the market settles.",
				MarkdownDescription: "Long-form description and resolution criteria explaining " +
					"how the market settles.",
			},
			"active": schema.BoolAttribute{
				Computed: true,
				Description: "Whether the market is currently active and accepting trades. " +
					"A market may be inactive without yet being closed.",
				MarkdownDescription: "Whether the market is currently active and accepting " +
					"trades. A market may be inactive without yet being closed.",
			},
			"closed": schema.BoolAttribute{
				Computed: true,
				Description: "Whether the market has closed. A closed market no longer " +
					"accepts trades and is either resolved or awaiting resolution.",
				MarkdownDescription: "Whether the market has closed. A closed market no longer " +
					"accepts trades and is either resolved or awaiting resolution.",
			},
			"archived": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the market has been archived and hidden from the default UI.",
				MarkdownDescription: "Whether the market has been archived and hidden from the default UI.",
			},
			"liquidity": schema.StringAttribute{
				Computed: true,
				Description: "Total liquidity available in the market, in USDC. Returned as " +
					"a decimal string to preserve full precision.",
				MarkdownDescription: "Total liquidity available in the market, in USDC. " +
					"Returned as a decimal string to preserve full precision.",
			},
			"volume": schema.StringAttribute{
				Computed: true,
				Description: "Cumulative trading volume of the market, in USDC. Returned as " +
					"a decimal string to preserve full precision.",
				MarkdownDescription: "Cumulative trading volume of the market, in USDC. " +
					"Returned as a decimal string to preserve full precision.",
			},
			"start_date": schema.StringAttribute{
				Computed:            true,
				Description:         "ISO-8601 timestamp at which the market opened for trading.",
				MarkdownDescription: "ISO-8601 timestamp at which the market opened for trading.",
			},
			"end_date": schema.StringAttribute{
				Computed:            true,
				Description:         "ISO-8601 timestamp at which the market is scheduled to close.",
				MarkdownDescription: "ISO-8601 timestamp at which the market is scheduled to close.",
			},
			"condition_id": schema.StringAttribute{
				Computed: true,
				Description: "On-chain condition ID (a 0x-prefixed hash) linking the market " +
					"to its CTF conditional-token contract.",
				MarkdownDescription: "On-chain condition ID (a `0x`-prefixed hash) linking the " +
					"market to its CTF conditional-token contract.",
			},
			"outcomes": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Ordered list of outcome names, e.g. [\"Yes\", \"No\"]. The order " +
					"aligns positionally with outcome_prices.",
				MarkdownDescription: "Ordered list of outcome names, e.g. `[\"Yes\", \"No\"]`. " +
					"The order aligns positionally with `outcome_prices`.",
			},
			"outcome_prices": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Current market price of each outcome as a decimal string in " +
					"[0, 1], representing the implied probability. Aligns positionally with outcomes.",
				MarkdownDescription: "Current market price of each outcome as a decimal string " +
					"in `[0, 1]`, representing the implied probability. Aligns positionally with `outcomes`.",
			},
		},
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
