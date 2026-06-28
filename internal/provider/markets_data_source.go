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

var (
	_ datasource.DataSource              = (*marketsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*marketsDataSource)(nil)
)

// marketsDataSource lists Polymarket markets matching optional filters.
type marketsDataSource struct {
	client *client.Client
}

// marketsDataSourceModel is the top-level model: filter inputs plus the
// resulting list of markets.
type marketsDataSourceModel struct {
	// Filter inputs.
	Limit  types.Int64  `tfsdk:"limit"`
	Offset types.Int64  `tfsdk:"offset"`
	Active types.Bool   `tfsdk:"active"`
	Closed types.Bool   `tfsdk:"closed"`
	Slug   types.String `tfsdk:"slug"`

	// Computed results.
	Markets []marketDataSourceModel `tfsdk:"markets"`
}

// NewMarketsDataSource is the data source constructor registered on the provider.
func NewMarketsDataSource() datasource.DataSource {
	return &marketsDataSource{}
}

// Metadata sets the data source type name ("polymarket_markets").
func (d *marketsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_markets"
}

// Schema describes the filter inputs and the nested list of returned markets.
func (d *marketsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Polymarket prediction markets, optionally filtered by activity, " +
			"closed state, or slug, with pagination support.",
		MarkdownDescription: "Lists Polymarket prediction markets, optionally filtered by " +
			"activity, closed state, or slug, with pagination support.",
		Attributes: map[string]schema.Attribute{
			"limit": schema.Int64Attribute{
				Optional: true,
				Description: "Maximum number of markets to return. When omitted, the API " +
					"default page size is used. Combine with offset to paginate.",
				MarkdownDescription: "Maximum number of markets to return. When omitted, the " +
					"API default page size is used. Combine with `offset` to paginate.",
			},
			"offset": schema.Int64Attribute{
				Optional: true,
				Description: "Number of markets to skip before collecting results, used " +
					"together with limit for pagination.",
				MarkdownDescription: "Number of markets to skip before collecting results, used " +
					"together with `limit` for pagination.",
			},
			"active": schema.BoolAttribute{
				Optional: true,
				Description: "When set, returns only markets whose active flag matches this " +
					"value. Omit to return markets regardless of activity.",
				MarkdownDescription: "When set, returns only markets whose `active` flag matches " +
					"this value. Omit to return markets regardless of activity.",
			},
			"closed": schema.BoolAttribute{
				Optional: true,
				Description: "When set, returns only markets whose closed flag matches this " +
					"value. Omit to return markets regardless of closed state.",
				MarkdownDescription: "When set, returns only markets whose `closed` flag matches " +
					"this value. Omit to return markets regardless of closed state.",
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				Description:         "When set, returns only the market with this exact URL slug.",
				MarkdownDescription: "When set, returns only the market with this exact URL slug.",
			},
			"markets": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "List of markets matching the supplied filters.",
				MarkdownDescription: "List of markets matching the supplied filters.",
				NestedObject: schema.NestedAttributeObject{
					// Shared with the single-market data source so both surface
					// identical, fully documented fields.
					Attributes: marketAttributes(false),
				},
			},
		},
	}
}

// Configure receives the shared API client from the provider.
func (d *marketsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read applies the filters, fetches matching markets, and maps them to state.
func (d *marketsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data marketsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter := client.MarketFilter{
		Limit:  data.Limit.ValueInt64(),
		Offset: data.Offset.ValueInt64(),
		Slug:   data.Slug.ValueString(),
	}
	if !data.Active.IsNull() {
		v := data.Active.ValueBool()
		filter.Active = &v
	}
	if !data.Closed.IsNull() {
		v := data.Closed.ValueBool()
		filter.Closed = &v
	}

	markets, err := d.client.ListMarkets(ctx, filter)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Polymarket markets", err.Error())
		return
	}

	data.Markets = make([]marketDataSourceModel, 0, len(markets))
	for _, m := range markets {
		var entry marketDataSourceModel
		resp.Diagnostics.Append(flattenMarket(ctx, m, &entry)...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Markets = append(data.Markets, entry)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
