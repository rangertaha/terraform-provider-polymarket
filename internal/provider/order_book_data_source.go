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
	_ datasource.DataSource              = (*orderBookDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*orderBookDataSource)(nil)
)

// orderBookDataSource reads the live CLOB order book for an outcome token.
type orderBookDataSource struct {
	client *client.Client
}

// orderBookDataSourceModel maps the order book schema to Go types.
type orderBookDataSourceModel struct {
	TokenID        types.String          `tfsdk:"token_id"`
	Market         types.String          `tfsdk:"market"`
	Timestamp      types.String          `tfsdk:"timestamp"`
	Hash           types.String          `tfsdk:"hash"`
	MinOrderSize   types.String          `tfsdk:"min_order_size"`
	TickSize       types.String          `tfsdk:"tick_size"`
	NegRisk        types.Bool            `tfsdk:"neg_risk"`
	LastTradePrice types.String          `tfsdk:"last_trade_price"`
	Bids           []orderBookLevelModel `tfsdk:"bids"`
	Asks           []orderBookLevelModel `tfsdk:"asks"`
}

// orderBookLevelModel maps a single price level to Go types.
type orderBookLevelModel struct {
	Price types.String `tfsdk:"price"`
	Size  types.String `tfsdk:"size"`
}

// NewOrderBookDataSource is the data source constructor registered on the provider.
func NewOrderBookDataSource() datasource.DataSource {
	return &orderBookDataSource{}
}

// Metadata sets the data source type name ("polymarket_order_book").
func (d *orderBookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_order_book"
}

// Schema describes the order book inputs and the resulting bid/ask ladders.
func (d *orderBookDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	levelAttrs := map[string]schema.Attribute{
		"price": schema.StringAttribute{
			Computed:            true,
			Description:         "Price of this level, in [0, 1], as a decimal string.",
			MarkdownDescription: "Price of this level, in `[0, 1]`, as a decimal string.",
		},
		"size": schema.StringAttribute{
			Computed:            true,
			Description:         "Total size resting at this price level, in shares, as a decimal string.",
			MarkdownDescription: "Total size resting at this price level, in shares, as a decimal string.",
		},
	}

	resp.Schema = schema.Schema{
		Description: "Reads the live Polymarket CLOB order book for a single outcome token, " +
			"returning the full bid and ask ladders. Token IDs come from a market's clob_token_ids.",
		MarkdownDescription: "Reads the live Polymarket CLOB order book for a single outcome " +
			"token, returning the full bid and ask ladders. Token IDs come from a market's `clob_token_ids`.",
		Attributes: map[string]schema.Attribute{
			"token_id": schema.StringAttribute{
				Required: true,
				Description: "CLOB ERC-1155 token ID of the outcome to read. Obtain it from a " +
					"market's clob_token_ids attribute. This is the only required input.",
				MarkdownDescription: "CLOB ERC-1155 token ID of the outcome to read. Obtain it from " +
					"a market's `clob_token_ids` attribute. This is the only required input.",
			},
			"market": schema.StringAttribute{
				Computed:            true,
				Description:         "Condition ID of the parent market the token belongs to.",
				MarkdownDescription: "Condition ID of the parent market the token belongs to.",
			},
			"timestamp": schema.StringAttribute{
				Computed:            true,
				Description:         "Server timestamp of the snapshot, in milliseconds since the Unix epoch.",
				MarkdownDescription: "Server timestamp of the snapshot, in milliseconds since the Unix epoch.",
			},
			"hash": schema.StringAttribute{
				Computed:            true,
				Description:         "Content hash uniquely identifying this order book snapshot.",
				MarkdownDescription: "Content hash uniquely identifying this order book snapshot.",
			},
			"min_order_size": schema.StringAttribute{
				Computed:            true,
				Description:         "Minimum order size accepted for this token, in shares.",
				MarkdownDescription: "Minimum order size accepted for this token, in shares.",
			},
			"tick_size": schema.StringAttribute{
				Computed:            true,
				Description:         "Smallest price increment accepted for this token (e.g. 0.01).",
				MarkdownDescription: "Smallest price increment accepted for this token (e.g. `0.01`).",
			},
			"neg_risk": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the token participates in a negative-risk (multi-outcome) market.",
				MarkdownDescription: "Whether the token participates in a negative-risk (multi-outcome) market.",
			},
			"last_trade_price": schema.StringAttribute{
				Computed:            true,
				Description:         "Price at which this token most recently traded, in [0, 1].",
				MarkdownDescription: "Price at which this token most recently traded, in `[0, 1]`.",
			},
			"bids": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Buy orders resting in the book, ordered ascending by price.",
				MarkdownDescription: "Buy orders resting in the book, ordered ascending by price.",
				NestedObject:        schema.NestedAttributeObject{Attributes: levelAttrs},
			},
			"asks": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Sell orders resting in the book, ordered descending by price.",
				MarkdownDescription: "Sell orders resting in the book, ordered descending by price.",
				NestedObject:        schema.NestedAttributeObject{Attributes: levelAttrs},
			},
		},
	}
}

// Configure receives the shared API client from the provider.
func (d *orderBookDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches the order book and maps it onto Terraform state.
func (d *orderBookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data orderBookDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	book, err := d.client.GetOrderBook(ctx, data.TokenID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Polymarket order book",
			fmt.Sprintf("Could not fetch order book for token %q: %s", data.TokenID.ValueString(), err),
		)
		return
	}

	data.Market = types.StringValue(book.Market)
	data.Timestamp = types.StringValue(book.Timestamp)
	data.Hash = types.StringValue(book.Hash)
	data.MinOrderSize = types.StringValue(book.MinOrderSize)
	data.TickSize = types.StringValue(book.TickSize)
	data.NegRisk = types.BoolValue(book.NegRisk)
	data.LastTradePrice = types.StringValue(book.LastTradePrice)
	data.Bids = flattenLevels(book.Bids)
	data.Asks = flattenLevels(book.Asks)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// flattenLevels maps API order book levels onto their Terraform model.
func flattenLevels(levels []client.OrderBookLevel) []orderBookLevelModel {
	out := make([]orderBookLevelModel, 0, len(levels))
	for _, l := range levels {
		out = append(out, orderBookLevelModel{
			Price: types.StringValue(l.Price),
			Size:  types.StringValue(l.Size),
		})
	}
	return out
}
