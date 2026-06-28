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
// polymarket_trades
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = (*tradesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*tradesDataSource)(nil)
)

type tradesDataSource struct{ client *client.Client }

type tradesDataSourceModel struct {
	User   types.String      `tfsdk:"user"`
	Market types.String      `tfsdk:"market"`
	Limit  types.Int64       `tfsdk:"limit"`
	Offset types.Int64       `tfsdk:"offset"`
	Trades []tradeEntryModel `tfsdk:"trades"`
}

type tradeEntryModel struct {
	Side            types.String  `tfsdk:"side"`
	Asset           types.String  `tfsdk:"asset"`
	ConditionID     types.String  `tfsdk:"condition_id"`
	Size            types.Float64 `tfsdk:"size"`
	Price           types.Float64 `tfsdk:"price"`
	Timestamp       types.Int64   `tfsdk:"timestamp"`
	Title           types.String  `tfsdk:"title"`
	Slug            types.String  `tfsdk:"slug"`
	EventSlug       types.String  `tfsdk:"event_slug"`
	Outcome         types.String  `tfsdk:"outcome"`
	OutcomeIndex    types.Int64   `tfsdk:"outcome_index"`
	TransactionHash types.String  `tfsdk:"transaction_hash"`
}

// NewTradesDataSource is the data source constructor registered on the provider.
func NewTradesDataSource() datasource.DataSource { return &tradesDataSource{} }

func (d *tradesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trades"
}

func (d *tradesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Lists the executed trades for a Polymarket wallet, most recent first.",
		MarkdownDescription: "Lists the executed trades for a Polymarket wallet, most recent first.",
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				Required:            true,
				Description:         "Wallet address whose trades to list (0x-prefixed). The only required input.",
				MarkdownDescription: "Wallet address whose trades to list (`0x`-prefixed). The only required input.",
			},
			"market": schema.StringAttribute{
				Optional:            true,
				Description:         "When set, restricts results to trades in the market with this condition ID.",
				MarkdownDescription: "When set, restricts results to trades in the market with this condition ID.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				Description:         "Maximum number of trades to return. Combine with offset to paginate.",
				MarkdownDescription: "Maximum number of trades to return. Combine with `offset` to paginate.",
			},
			"offset": schema.Int64Attribute{
				Optional:            true,
				Description:         "Number of trades to skip before collecting results, for pagination.",
				MarkdownDescription: "Number of trades to skip before collecting results, for pagination.",
			},
			"trades": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Executed trades for the wallet.",
				MarkdownDescription: "Executed trades for the wallet.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"side": schema.StringAttribute{
							Computed:            true,
							Description:         "Trade direction, either \"BUY\" or \"SELL\".",
							MarkdownDescription: "Trade direction, either `BUY` or `SELL`.",
						},
						"asset": schema.StringAttribute{
							Computed:            true,
							Description:         "CLOB token ID of the traded outcome.",
							MarkdownDescription: "CLOB token ID of the traded outcome.",
						},
						"condition_id": schema.StringAttribute{
							Computed:            true,
							Description:         "Condition ID of the parent market.",
							MarkdownDescription: "Condition ID of the parent market.",
						},
						"size": schema.Float64Attribute{
							Computed:            true,
							Description:         "Number of shares traded.",
							MarkdownDescription: "Number of shares traded.",
						},
						"price": schema.Float64Attribute{
							Computed:            true,
							Description:         "Execution price, in [0, 1].",
							MarkdownDescription: "Execution price, in `[0, 1]`.",
						},
						"timestamp": schema.Int64Attribute{
							Computed:            true,
							Description:         "Execution time, in Unix seconds.",
							MarkdownDescription: "Execution time, in Unix seconds.",
						},
						"title": schema.StringAttribute{
							Computed:            true,
							Description:         "Title of the market that was traded.",
							MarkdownDescription: "Title of the market that was traded.",
						},
						"slug": schema.StringAttribute{
							Computed:            true,
							Description:         "Slug of the market that was traded.",
							MarkdownDescription: "Slug of the market that was traded.",
						},
						"event_slug": schema.StringAttribute{
							Computed:            true,
							Description:         "Slug of the parent event.",
							MarkdownDescription: "Slug of the parent event.",
						},
						"outcome": schema.StringAttribute{
							Computed:            true,
							Description:         "Name of the traded outcome.",
							MarkdownDescription: "Name of the traded outcome.",
						},
						"outcome_index": schema.Int64Attribute{
							Computed:            true,
							Description:         "Zero-based index of the traded outcome within the market.",
							MarkdownDescription: "Zero-based index of the traded outcome within the market.",
						},
						"transaction_hash": schema.StringAttribute{
							Computed:            true,
							Description:         "On-chain transaction hash that settled the trade.",
							MarkdownDescription: "On-chain transaction hash that settled the trade.",
						},
					},
				},
			},
		},
	}
}

func (d *tradesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *tradesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tradesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	trades, err := d.client.ListTrades(ctx, client.TradeFilter{
		User:   data.User.ValueString(),
		Market: data.Market.ValueString(),
		Limit:  data.Limit.ValueInt64(),
		Offset: data.Offset.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Polymarket trades", err.Error())
		return
	}

	data.Trades = make([]tradeEntryModel, 0, len(trades))
	for _, t := range trades {
		data.Trades = append(data.Trades, tradeEntryModel{
			Side:            types.StringValue(t.Side),
			Asset:           types.StringValue(t.Asset),
			ConditionID:     types.StringValue(t.ConditionID),
			Size:            types.Float64Value(t.Size),
			Price:           types.Float64Value(t.Price),
			Timestamp:       types.Int64Value(t.Timestamp),
			Title:           types.StringValue(t.Title),
			Slug:            types.StringValue(t.Slug),
			EventSlug:       types.StringValue(t.EventSlug),
			Outcome:         types.StringValue(t.Outcome),
			OutcomeIndex:    types.Int64Value(t.OutcomeIndex),
			TransactionHash: types.StringValue(t.TransactionHash),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------------------------------------------------------------------------
// polymarket_portfolio_value
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = (*portfolioValueDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*portfolioValueDataSource)(nil)
)

type portfolioValueDataSource struct{ client *client.Client }

type portfolioValueDataSourceModel struct {
	User  types.String  `tfsdk:"user"`
	Value types.Float64 `tfsdk:"value"`
}

// NewPortfolioValueDataSource is the data source constructor registered on the provider.
func NewPortfolioValueDataSource() datasource.DataSource { return &portfolioValueDataSource{} }

func (d *portfolioValueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_portfolio_value"
}

func (d *portfolioValueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reads the total current portfolio value of a Polymarket wallet, in USDC.",
		MarkdownDescription: "Reads the total current portfolio value of a Polymarket wallet, in USDC.",
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				Required:            true,
				Description:         "Wallet address to value (0x-prefixed). The only required input.",
				MarkdownDescription: "Wallet address to value (`0x`-prefixed). The only required input.",
			},
			"value": schema.Float64Attribute{
				Computed:            true,
				Description:         "Total mark-to-market value of all the wallet's positions, in USDC.",
				MarkdownDescription: "Total mark-to-market value of all the wallet's positions, in USDC.",
			},
		},
	}
}

func (d *portfolioValueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *portfolioValueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data portfolioValueDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value, err := d.client.GetValue(ctx, data.User.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Polymarket portfolio value", err.Error())
		return
	}

	data.Value = types.Float64Value(value.Value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------------------------------------------------------------------------
// polymarket_holders
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = (*holdersDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*holdersDataSource)(nil)
)

type holdersDataSource struct{ client *client.Client }

type holdersDataSourceModel struct {
	Market types.String       `tfsdk:"market"`
	Limit  types.Int64        `tfsdk:"limit"`
	Tokens []holderGroupModel `tfsdk:"tokens"`
}

type holderGroupModel struct {
	Token   types.String       `tfsdk:"token"`
	Holders []holderEntryModel `tfsdk:"holders"`
}

type holderEntryModel struct {
	ProxyWallet  types.String  `tfsdk:"proxy_wallet"`
	Amount       types.Float64 `tfsdk:"amount"`
	OutcomeIndex types.Int64   `tfsdk:"outcome_index"`
	Name         types.String  `tfsdk:"name"`
	Pseudonym    types.String  `tfsdk:"pseudonym"`
	Verified     types.Bool    `tfsdk:"verified"`
}

// NewHoldersDataSource is the data source constructor registered on the provider.
func NewHoldersDataSource() datasource.DataSource { return &holdersDataSource{} }

func (d *holdersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_holders"
}

func (d *holdersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists the top holders of each outcome token in a market, identified by " +
			"its condition ID.",
		MarkdownDescription: "Lists the top holders of each outcome token in a market, identified " +
			"by its condition ID.",
		Attributes: map[string]schema.Attribute{
			"market": schema.StringAttribute{
				Required: true,
				Description: "Condition ID of the market whose holders to list (a 0x-prefixed " +
					"hash). This is the only required input.",
				MarkdownDescription: "Condition ID of the market whose holders to list (a " +
					"`0x`-prefixed hash). This is the only required input.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				Description:         "Maximum number of holders to return per outcome token.",
				MarkdownDescription: "Maximum number of holders to return per outcome token.",
			},
			"tokens": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Per-outcome-token groups, each listing that token's top holders.",
				MarkdownDescription: "Per-outcome-token groups, each listing that token's top holders.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Computed:            true,
							Description:         "CLOB token ID this group of holders is for.",
							MarkdownDescription: "CLOB token ID this group of holders is for.",
						},
						"holders": schema.ListNestedAttribute{
							Computed:            true,
							Description:         "Top holders of the token, ordered by amount held descending.",
							MarkdownDescription: "Top holders of the token, ordered by amount held descending.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"proxy_wallet": schema.StringAttribute{
										Computed:            true,
										Description:         "On-chain proxy wallet address of the holder.",
										MarkdownDescription: "On-chain proxy wallet address of the holder.",
									},
									"amount": schema.Float64Attribute{
										Computed:            true,
										Description:         "Number of shares of the token held.",
										MarkdownDescription: "Number of shares of the token held.",
									},
									"outcome_index": schema.Int64Attribute{
										Computed:            true,
										Description:         "Zero-based index of the outcome the token represents.",
										MarkdownDescription: "Zero-based index of the outcome the token represents.",
									},
									"name": schema.StringAttribute{
										Computed:            true,
										Description:         "Display name of the holder, when public.",
										MarkdownDescription: "Display name of the holder, when public.",
									},
									"pseudonym": schema.StringAttribute{
										Computed:            true,
										Description:         "Auto-generated pseudonym of the holder.",
										MarkdownDescription: "Auto-generated pseudonym of the holder.",
									},
									"verified": schema.BoolAttribute{
										Computed:            true,
										Description:         "Whether the holder's account is verified.",
										MarkdownDescription: "Whether the holder's account is verified.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *holdersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *holdersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data holdersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := d.client.GetHolders(ctx, data.Market.ValueString(), data.Limit.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Polymarket holders", err.Error())
		return
	}

	data.Tokens = make([]holderGroupModel, 0, len(groups))
	for _, g := range groups {
		holders := make([]holderEntryModel, 0, len(g.Holders))
		for _, h := range g.Holders {
			holders = append(holders, holderEntryModel{
				ProxyWallet:  types.StringValue(h.ProxyWallet),
				Amount:       types.Float64Value(h.Amount),
				OutcomeIndex: types.Int64Value(h.OutcomeIndex),
				Name:         types.StringValue(h.Name),
				Pseudonym:    types.StringValue(h.Pseudonym),
				Verified:     types.BoolValue(h.Verified),
			})
		}
		data.Tokens = append(data.Tokens, holderGroupModel{
			Token:   types.StringValue(g.Token),
			Holders: holders,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
