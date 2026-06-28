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

var (
	_ datasource.DataSource              = (*activityDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*activityDataSource)(nil)
)

// activityDataSource lists a wallet's unified on-chain activity feed.
type activityDataSource struct {
	client *client.Client
}

type activityDataSourceModel struct {
	User     types.String         `tfsdk:"user"`
	Type     types.String         `tfsdk:"type"`
	Market   types.String         `tfsdk:"market"`
	Limit    types.Int64          `tfsdk:"limit"`
	Offset   types.Int64          `tfsdk:"offset"`
	Activity []activityEntryModel `tfsdk:"activity"`
}

type activityEntryModel struct {
	Type            types.String  `tfsdk:"type"`
	Timestamp       types.Int64   `tfsdk:"timestamp"`
	ConditionID     types.String  `tfsdk:"condition_id"`
	Asset           types.String  `tfsdk:"asset"`
	Side            types.String  `tfsdk:"side"`
	Size            types.Float64 `tfsdk:"size"`
	UsdcSize        types.Float64 `tfsdk:"usdc_size"`
	Price           types.Float64 `tfsdk:"price"`
	Outcome         types.String  `tfsdk:"outcome"`
	OutcomeIndex    types.Int64   `tfsdk:"outcome_index"`
	Title           types.String  `tfsdk:"title"`
	Slug            types.String  `tfsdk:"slug"`
	EventSlug       types.String  `tfsdk:"event_slug"`
	TransactionHash types.String  `tfsdk:"transaction_hash"`
}

// NewActivityDataSource is the data source constructor registered on the provider.
func NewActivityDataSource() datasource.DataSource { return &activityDataSource{} }

func (d *activityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_activity"
}

func (d *activityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists a wallet's unified on-chain activity feed — trades, rewards, " +
			"splits, merges, redemptions, and conversions — most recent first.",
		MarkdownDescription: "Lists a wallet's unified on-chain activity feed — trades, rewards, " +
			"splits, merges, redemptions, and conversions — most recent first.",
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				Required:            true,
				Description:         "Wallet address whose activity to list (0x-prefixed). The only required input.",
				MarkdownDescription: "Wallet address whose activity to list (`0x`-prefixed). The only required input.",
			},
			"type": schema.StringAttribute{
				Optional: true,
				Description: "When set, returns only entries of this type, e.g. \"TRADE\", " +
					"\"REWARD\", \"SPLIT\", \"MERGE\", \"REDEEM\", or \"CONVERSION\".",
				MarkdownDescription: "When set, returns only entries of this type, e.g. `TRADE`, " +
					"`REWARD`, `SPLIT`, `MERGE`, `REDEEM`, or `CONVERSION`.",
			},
			"market": schema.StringAttribute{
				Optional:            true,
				Description:         "When set, restricts results to activity in the market with this condition ID.",
				MarkdownDescription: "When set, restricts results to activity in the market with this condition ID.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				Description:         "Maximum number of entries to return. Combine with offset to paginate.",
				MarkdownDescription: "Maximum number of entries to return. Combine with `offset` to paginate.",
			},
			"offset": schema.Int64Attribute{
				Optional:            true,
				Description:         "Number of entries to skip before collecting results, for pagination.",
				MarkdownDescription: "Number of entries to skip before collecting results, for pagination.",
			},
			"activity": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Activity entries for the wallet.",
				MarkdownDescription: "Activity entries for the wallet.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Computed:            true,
							Description:         "Entry type: TRADE, REWARD, SPLIT, MERGE, REDEEM, or CONVERSION.",
							MarkdownDescription: "Entry type: `TRADE`, `REWARD`, `SPLIT`, `MERGE`, `REDEEM`, or `CONVERSION`.",
						},
						"timestamp": schema.Int64Attribute{
							Computed:            true,
							Description:         "Time of the activity, in Unix seconds.",
							MarkdownDescription: "Time of the activity, in Unix seconds.",
						},
						"condition_id": schema.StringAttribute{
							Computed:            true,
							Description:         "Condition ID of the market involved.",
							MarkdownDescription: "Condition ID of the market involved.",
						},
						"asset": schema.StringAttribute{
							Computed:            true,
							Description:         "CLOB token ID of the outcome involved.",
							MarkdownDescription: "CLOB token ID of the outcome involved.",
						},
						"side": schema.StringAttribute{
							Computed:            true,
							Description:         "For trades, the direction (\"BUY\" or \"SELL\"); empty for non-trade activity.",
							MarkdownDescription: "For trades, the direction (`BUY` or `SELL`); empty for non-trade activity.",
						},
						"size": schema.Float64Attribute{
							Computed:            true,
							Description:         "Number of shares involved.",
							MarkdownDescription: "Number of shares involved.",
						},
						"usdc_size": schema.Float64Attribute{
							Computed:            true,
							Description:         "USDC value of the activity.",
							MarkdownDescription: "USDC value of the activity.",
						},
						"price": schema.Float64Attribute{
							Computed:            true,
							Description:         "For trades, the execution price, in [0, 1].",
							MarkdownDescription: "For trades, the execution price, in `[0, 1]`.",
						},
						"outcome": schema.StringAttribute{
							Computed:            true,
							Description:         "Name of the outcome involved.",
							MarkdownDescription: "Name of the outcome involved.",
						},
						"outcome_index": schema.Int64Attribute{
							Computed:            true,
							Description:         "Zero-based index of the outcome within the market.",
							MarkdownDescription: "Zero-based index of the outcome within the market.",
						},
						"title": schema.StringAttribute{
							Computed:            true,
							Description:         "Title of the market involved.",
							MarkdownDescription: "Title of the market involved.",
						},
						"slug": schema.StringAttribute{
							Computed:            true,
							Description:         "Slug of the market involved.",
							MarkdownDescription: "Slug of the market involved.",
						},
						"event_slug": schema.StringAttribute{
							Computed:            true,
							Description:         "Slug of the parent event.",
							MarkdownDescription: "Slug of the parent event.",
						},
						"transaction_hash": schema.StringAttribute{
							Computed:            true,
							Description:         "On-chain transaction hash for the activity.",
							MarkdownDescription: "On-chain transaction hash for the activity.",
						},
					},
				},
			},
		},
	}
}

func (d *activityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *activityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data activityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := d.client.ListActivity(ctx, client.ActivityFilter{
		User:   data.User.ValueString(),
		Type:   data.Type.ValueString(),
		Market: data.Market.ValueString(),
		Limit:  data.Limit.ValueInt64(),
		Offset: data.Offset.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Polymarket activity", err.Error())
		return
	}

	data.Activity = make([]activityEntryModel, 0, len(entries))
	for _, a := range entries {
		data.Activity = append(data.Activity, activityEntryModel{
			Type:            types.StringValue(a.Type),
			Timestamp:       types.Int64Value(a.Timestamp),
			ConditionID:     types.StringValue(a.ConditionID),
			Asset:           types.StringValue(a.Asset),
			Side:            types.StringValue(a.Side),
			Size:            types.Float64Value(a.Size),
			UsdcSize:        types.Float64Value(a.UsdcSize),
			Price:           types.Float64Value(a.Price),
			Outcome:         types.StringValue(a.Outcome),
			OutcomeIndex:    types.Int64Value(a.OutcomeIndex),
			Title:           types.StringValue(a.Title),
			Slug:            types.StringValue(a.Slug),
			EventSlug:       types.StringValue(a.EventSlug),
			TransactionHash: types.StringValue(a.TransactionHash),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
