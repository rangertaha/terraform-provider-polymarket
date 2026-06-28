// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*positionsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*positionsDataSource)(nil)
)

// positionsDataSource lists the open positions held by a wallet.
type positionsDataSource struct {
	client *client.Client
}

type positionsDataSourceModel struct {
	User      types.String         `tfsdk:"user"`
	Market    types.String         `tfsdk:"market"`
	Limit     types.Int64          `tfsdk:"limit"`
	Offset    types.Int64          `tfsdk:"offset"`
	Positions []positionEntryModel `tfsdk:"positions"`
}

type positionEntryModel struct {
	Asset           types.String  `tfsdk:"asset"`
	ConditionID     types.String  `tfsdk:"condition_id"`
	Size            types.Float64 `tfsdk:"size"`
	AvgPrice        types.Float64 `tfsdk:"avg_price"`
	InitialValue    types.Float64 `tfsdk:"initial_value"`
	CurrentValue    types.Float64 `tfsdk:"current_value"`
	CashPnl         types.Float64 `tfsdk:"cash_pnl"`
	PercentPnl      types.Float64 `tfsdk:"percent_pnl"`
	RealizedPnl     types.Float64 `tfsdk:"realized_pnl"`
	CurPrice        types.Float64 `tfsdk:"cur_price"`
	Redeemable      types.Bool    `tfsdk:"redeemable"`
	Title           types.String  `tfsdk:"title"`
	Slug            types.String  `tfsdk:"slug"`
	EventSlug       types.String  `tfsdk:"event_slug"`
	Outcome         types.String  `tfsdk:"outcome"`
	OutcomeIndex    types.Int64   `tfsdk:"outcome_index"`
	OppositeOutcome types.String  `tfsdk:"opposite_outcome"`
	EndDate         types.String  `tfsdk:"end_date"`
	NegativeRisk    types.Bool    `tfsdk:"negative_risk"`
}

// NewPositionsDataSource is the data source constructor registered on the provider.
func NewPositionsDataSource() datasource.DataSource { return &positionsDataSource{} }

func (d *positionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_positions"
}

func (d *positionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists the open positions held by a Polymarket wallet, with cost basis " +
			"and profit-and-loss for each market outcome.",
		MarkdownDescription: "Lists the open positions held by a Polymarket wallet, with cost " +
			"basis and profit-and-loss for each market outcome.",
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{ethAddress()},
				Description: "Wallet address whose positions to list (the on-chain proxy wallet, " +
					"a 0x-prefixed address). This is the only required input.",
				MarkdownDescription: "Wallet address whose positions to list (the on-chain proxy " +
					"wallet, a `0x`-prefixed address). This is the only required input.",
			},
			"market": schema.StringAttribute{
				Optional:            true,
				Description:         "When set, restricts results to positions in the market with this condition ID.",
				MarkdownDescription: "When set, restricts results to positions in the market with this condition ID.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				Description:         "Maximum number of positions to return. Combine with offset to paginate.",
				MarkdownDescription: "Maximum number of positions to return. Combine with `offset` to paginate.",
			},
			"offset": schema.Int64Attribute{
				Optional:            true,
				Description:         "Number of positions to skip before collecting results, for pagination.",
				MarkdownDescription: "Number of positions to skip before collecting results, for pagination.",
			},
			"positions": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Open positions held by the wallet.",
				MarkdownDescription: "Open positions held by the wallet.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"asset": schema.StringAttribute{
							Computed:            true,
							Description:         "CLOB token ID of the held outcome.",
							MarkdownDescription: "CLOB token ID of the held outcome.",
						},
						"condition_id": schema.StringAttribute{
							Computed:            true,
							Description:         "Condition ID of the parent market.",
							MarkdownDescription: "Condition ID of the parent market.",
						},
						"size": schema.Float64Attribute{
							Computed:            true,
							Description:         "Number of shares held.",
							MarkdownDescription: "Number of shares held.",
						},
						"avg_price": schema.Float64Attribute{
							Computed:            true,
							Description:         "Average entry price across all buys, in [0, 1].",
							MarkdownDescription: "Average entry price across all buys, in `[0, 1]`.",
						},
						"initial_value": schema.Float64Attribute{
							Computed:            true,
							Description:         "Cost basis of the position, in USDC.",
							MarkdownDescription: "Cost basis of the position, in USDC.",
						},
						"current_value": schema.Float64Attribute{
							Computed:            true,
							Description:         "Current mark-to-market value of the position, in USDC.",
							MarkdownDescription: "Current mark-to-market value of the position, in USDC.",
						},
						"cash_pnl": schema.Float64Attribute{
							Computed:            true,
							Description:         "Unrealized profit or loss in USDC (current_value minus initial_value).",
							MarkdownDescription: "Unrealized profit or loss in USDC (`current_value` minus `initial_value`).",
						},
						"percent_pnl": schema.Float64Attribute{
							Computed:            true,
							Description:         "Unrealized profit or loss as a percentage of cost basis.",
							MarkdownDescription: "Unrealized profit or loss as a percentage of cost basis.",
						},
						"realized_pnl": schema.Float64Attribute{
							Computed:            true,
							Description:         "Realized profit or loss already locked in, in USDC.",
							MarkdownDescription: "Realized profit or loss already locked in, in USDC.",
						},
						"cur_price": schema.Float64Attribute{
							Computed:            true,
							Description:         "Current market price of the held outcome, in [0, 1].",
							MarkdownDescription: "Current market price of the held outcome, in `[0, 1]`.",
						},
						"redeemable": schema.BoolAttribute{
							Computed:            true,
							Description:         "Whether the position is in a resolved market and can be redeemed for USDC.",
							MarkdownDescription: "Whether the position is in a resolved market and can be redeemed for USDC.",
						},
						"title": schema.StringAttribute{
							Computed:            true,
							Description:         "Title of the market the position is in.",
							MarkdownDescription: "Title of the market the position is in.",
						},
						"slug": schema.StringAttribute{
							Computed:            true,
							Description:         "Slug of the market the position is in.",
							MarkdownDescription: "Slug of the market the position is in.",
						},
						"event_slug": schema.StringAttribute{
							Computed:            true,
							Description:         "Slug of the parent event.",
							MarkdownDescription: "Slug of the parent event.",
						},
						"outcome": schema.StringAttribute{
							Computed:            true,
							Description:         "Name of the held outcome, e.g. \"Yes\" or \"No\".",
							MarkdownDescription: "Name of the held outcome, e.g. `Yes` or `No`.",
						},
						"outcome_index": schema.Int64Attribute{
							Computed:            true,
							Description:         "Zero-based index of the held outcome within the market.",
							MarkdownDescription: "Zero-based index of the held outcome within the market.",
						},
						"opposite_outcome": schema.StringAttribute{
							Computed:            true,
							Description:         "Name of the opposing outcome in the market.",
							MarkdownDescription: "Name of the opposing outcome in the market.",
						},
						"end_date": schema.StringAttribute{
							Computed:            true,
							Description:         "Date on which the market is scheduled to resolve.",
							MarkdownDescription: "Date on which the market is scheduled to resolve.",
						},
						"negative_risk": schema.BoolAttribute{
							Computed:            true,
							Description:         "Whether the position is in a negative-risk (multi-outcome) market.",
							MarkdownDescription: "Whether the position is in a negative-risk (multi-outcome) market.",
						},
					},
				},
			},
		},
	}
}

func (d *positionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *positionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data positionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	positions, err := d.client.ListPositions(ctx, client.PositionFilter{
		User:   data.User.ValueString(),
		Market: data.Market.ValueString(),
		Limit:  data.Limit.ValueInt64(),
		Offset: data.Offset.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Polymarket positions", err.Error())
		return
	}

	data.Positions = make([]positionEntryModel, 0, len(positions))
	for _, p := range positions {
		data.Positions = append(data.Positions, positionEntryModel{
			Asset:           types.StringValue(p.Asset),
			ConditionID:     types.StringValue(p.ConditionID),
			Size:            types.Float64Value(p.Size),
			AvgPrice:        types.Float64Value(p.AvgPrice),
			InitialValue:    types.Float64Value(p.InitialValue),
			CurrentValue:    types.Float64Value(p.CurrentValue),
			CashPnl:         types.Float64Value(p.CashPnl),
			PercentPnl:      types.Float64Value(p.PercentPnl),
			RealizedPnl:     types.Float64Value(p.RealizedPnl),
			CurPrice:        types.Float64Value(p.CurPrice),
			Redeemable:      types.BoolValue(p.Redeemable),
			Title:           types.StringValue(p.Title),
			Slug:            types.StringValue(p.Slug),
			EventSlug:       types.StringValue(p.EventSlug),
			Outcome:         types.StringValue(p.Outcome),
			OutcomeIndex:    types.Int64Value(p.OutcomeIndex),
			OppositeOutcome: types.StringValue(p.OppositeOutcome),
			EndDate:         types.StringValue(p.EndDate),
			NegativeRisk:    types.BoolValue(p.NegativeRisk),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
