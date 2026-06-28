// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// polymarket_prices (batch)
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = (*pricesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*pricesDataSource)(nil)
)

type pricesDataSource struct{ client *client.Client }

type pricesDataSourceModel struct {
	TokenIDs types.List         `tfsdk:"token_ids"`
	Prices   []tokenPricesModel `tfsdk:"prices"`
}

type tokenPricesModel struct {
	TokenID types.String `tfsdk:"token_id"`
	Buy     types.String `tfsdk:"buy"`
	Sell    types.String `tfsdk:"sell"`
}

// NewPricesDataSource is the data source constructor registered on the provider.
func NewPricesDataSource() datasource.DataSource { return &pricesDataSource{} }

func (d *pricesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prices"
}

func (d *pricesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the best buy and sell price for several CLOB outcome tokens in a " +
			"single request.",
		MarkdownDescription: "Fetches the best buy and sell price for several CLOB outcome " +
			"tokens in a single request.",
		Attributes: map[string]schema.Attribute{
			"token_ids": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				Description:         "CLOB ERC-1155 token IDs to quote, from markets' clob_token_ids.",
				MarkdownDescription: "CLOB ERC-1155 token IDs to quote, from markets' `clob_token_ids`.",
			},
			"prices": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Best prices per token, in the same order as token_ids.",
				MarkdownDescription: "Best prices per token, in the same order as `token_ids`.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token_id": schema.StringAttribute{
							Computed:            true,
							Description:         "Token the prices are for.",
							MarkdownDescription: "Token the prices are for.",
						},
						"buy": schema.StringAttribute{
							Computed:            true,
							Description:         "Best buy price (lowest ask), in [0, 1], as a decimal string.",
							MarkdownDescription: "Best buy price (lowest ask), in `[0, 1]`, as a decimal string.",
						},
						"sell": schema.StringAttribute{
							Computed:            true,
							Description:         "Best sell price (highest bid), in [0, 1], as a decimal string.",
							MarkdownDescription: "Best sell price (highest bid), in `[0, 1]`, as a decimal string.",
						},
					},
				},
			},
		},
	}
}

func (d *pricesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *pricesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data pricesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tokenIDs []string
	resp.Diagnostics.Append(data.TokenIDs.ElementsAs(ctx, &tokenIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prices, err := d.client.GetPrices(ctx, tokenIDs)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Polymarket prices", err.Error())
		return
	}

	data.Prices = make([]tokenPricesModel, 0, len(prices))
	for _, p := range prices {
		data.Prices = append(data.Prices, tokenPricesModel{
			TokenID: types.StringValue(p.TokenID),
			Buy:     types.StringValue(p.Buy),
			Sell:    types.StringValue(p.Sell),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------------------------------------------------------------------------
// polymarket_order_books (batch)
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = (*orderBooksDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*orderBooksDataSource)(nil)
)

type orderBooksDataSource struct{ client *client.Client }

type orderBooksDataSourceModel struct {
	TokenIDs types.List           `tfsdk:"token_ids"`
	Books    []orderBookItemModel `tfsdk:"books"`
}

type orderBookItemModel struct {
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

// NewOrderBooksDataSource is the data source constructor registered on the provider.
func NewOrderBooksDataSource() datasource.DataSource { return &orderBooksDataSource{} }

func (d *orderBooksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_order_books"
}

func (d *orderBooksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	levelAttrs := map[string]schema.Attribute{
		"price": schema.StringAttribute{
			Computed:            true,
			Description:         "Price of this level, in [0, 1], as a decimal string.",
			MarkdownDescription: "Price of this level, in `[0, 1]`, as a decimal string.",
		},
		"size": schema.StringAttribute{
			Computed:            true,
			Description:         "Total size resting at this price level, in shares.",
			MarkdownDescription: "Total size resting at this price level, in shares.",
		},
	}

	resp.Schema = schema.Schema{
		Description:         "Fetches the live order books for several CLOB outcome tokens in a single request.",
		MarkdownDescription: "Fetches the live order books for several CLOB outcome tokens in a single request.",
		Attributes: map[string]schema.Attribute{
			"token_ids": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				Description:         "CLOB ERC-1155 token IDs to fetch, from markets' clob_token_ids.",
				MarkdownDescription: "CLOB ERC-1155 token IDs to fetch, from markets' `clob_token_ids`.",
			},
			"books": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Order books, one per requested token.",
				MarkdownDescription: "Order books, one per requested token.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token_id": schema.StringAttribute{
							Computed:            true,
							Description:         "Token this book is for (its asset ID).",
							MarkdownDescription: "Token this book is for (its asset ID).",
						},
						"market": schema.StringAttribute{
							Computed:            true,
							Description:         "Condition ID of the parent market.",
							MarkdownDescription: "Condition ID of the parent market.",
						},
						"timestamp": schema.StringAttribute{
							Computed:            true,
							Description:         "Server snapshot time, in milliseconds since the Unix epoch.",
							MarkdownDescription: "Server snapshot time, in milliseconds since the Unix epoch.",
						},
						"hash": schema.StringAttribute{
							Computed:            true,
							Description:         "Content hash of the snapshot.",
							MarkdownDescription: "Content hash of the snapshot.",
						},
						"min_order_size": schema.StringAttribute{
							Computed:            true,
							Description:         "Minimum order size accepted for this token, in shares.",
							MarkdownDescription: "Minimum order size accepted for this token, in shares.",
						},
						"tick_size": schema.StringAttribute{
							Computed:            true,
							Description:         "Smallest price increment accepted for this token.",
							MarkdownDescription: "Smallest price increment accepted for this token.",
						},
						"neg_risk": schema.BoolAttribute{
							Computed:            true,
							Description:         "Whether the token is in a negative-risk market.",
							MarkdownDescription: "Whether the token is in a negative-risk market.",
						},
						"last_trade_price": schema.StringAttribute{
							Computed:            true,
							Description:         "Most recent trade price for this token, in [0, 1].",
							MarkdownDescription: "Most recent trade price for this token, in `[0, 1]`.",
						},
						"bids": schema.ListNestedAttribute{
							Computed:            true,
							Description:         "Buy orders, ascending by price.",
							MarkdownDescription: "Buy orders, ascending by price.",
							NestedObject:        schema.NestedAttributeObject{Attributes: levelAttrs},
						},
						"asks": schema.ListNestedAttribute{
							Computed:            true,
							Description:         "Sell orders, descending by price.",
							MarkdownDescription: "Sell orders, descending by price.",
							NestedObject:        schema.NestedAttributeObject{Attributes: levelAttrs},
						},
					},
				},
			},
		},
	}
}

func (d *orderBooksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *orderBooksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data orderBooksDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tokenIDs []string
	resp.Diagnostics.Append(data.TokenIDs.ElementsAs(ctx, &tokenIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	books, err := d.client.GetOrderBooks(ctx, tokenIDs)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Polymarket order books", err.Error())
		return
	}

	data.Books = make([]orderBookItemModel, 0, len(books))
	for _, b := range books {
		data.Books = append(data.Books, orderBookItemModel{
			TokenID:        types.StringValue(b.AssetID),
			Market:         types.StringValue(b.Market),
			Timestamp:      types.StringValue(b.Timestamp),
			Hash:           types.StringValue(b.Hash),
			MinOrderSize:   types.StringValue(b.MinOrderSize),
			TickSize:       types.StringValue(b.TickSize),
			NegRisk:        types.BoolValue(b.NegRisk),
			LastTradePrice: types.StringValue(b.LastTradePrice),
			Bids:           flattenLevels(b.Bids),
			Asks:           flattenLevels(b.Asks),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
