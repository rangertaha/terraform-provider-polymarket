// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*seriesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*seriesDataSource)(nil)
)

// seriesDataSource looks up a single Polymarket series (a recurring group of
// events) by its numeric ID.
type seriesDataSource struct {
	client *client.Client
}

// seriesDataSourceModel maps every series attribute to a Go type.
type seriesDataSourceModel struct {
	ID              types.String  `tfsdk:"id"`
	Ticker          types.String  `tfsdk:"ticker"`
	Slug            types.String  `tfsdk:"slug"`
	Title           types.String  `tfsdk:"title"`
	SeriesType      types.String  `tfsdk:"series_type"`
	Recurrence      types.String  `tfsdk:"recurrence"`
	Image           types.String  `tfsdk:"image"`
	Icon            types.String  `tfsdk:"icon"`
	Active          types.Bool    `tfsdk:"active"`
	Closed          types.Bool    `tfsdk:"closed"`
	Archived        types.Bool    `tfsdk:"archived"`
	Featured        types.Bool    `tfsdk:"featured"`
	Restricted      types.Bool    `tfsdk:"restricted"`
	CommentsEnabled types.Bool    `tfsdk:"comments_enabled"`
	Competitive     types.String  `tfsdk:"competitive"`
	Volume24hr      types.Float64 `tfsdk:"volume_24hr"`
	Volume          types.Float64 `tfsdk:"volume"`
	Liquidity       types.Float64 `tfsdk:"liquidity"`
	StartDate       types.String  `tfsdk:"start_date"`
	CreatedAt       types.String  `tfsdk:"created_at"`
	UpdatedAt       types.String  `tfsdk:"updated_at"`
	CommentCount    types.Int64   `tfsdk:"comment_count"`

	Events []eventDataSourceModel `tfsdk:"events"`
}

// NewSeriesDataSource is the data source constructor registered on the provider.
func NewSeriesDataSource() datasource.DataSource {
	return &seriesDataSource{}
}

// Metadata sets the data source type name ("polymarket_series").
func (d *seriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_series"
}

// Schema describes a single series and its embedded events.
func (d *seriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Polymarket series by its numeric ID. A series groups " +
			"recurring events under a common theme, e.g. a market that repeats weekly.",
		MarkdownDescription: "Fetches a single Polymarket series by its numeric `id`. A series " +
			"groups recurring events under a common theme, e.g. a market that repeats weekly.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				Description:         "Numeric identifier of the series to fetch, as assigned by Polymarket. The only required input.",
				MarkdownDescription: "Numeric identifier of the series to fetch, as assigned by Polymarket. The only required input.",
			},
			"ticker": schema.StringAttribute{
				Computed:            true,
				Description:         "Short ticker symbol identifying the series.",
				MarkdownDescription: "Short ticker symbol identifying the series.",
			},
			"slug": schema.StringAttribute{
				Computed:            true,
				Description:         "URL-friendly slug identifying the series on polymarket.com.",
				MarkdownDescription: "URL-friendly slug identifying the series on polymarket.com.",
			},
			"title": schema.StringAttribute{
				Computed:            true,
				Description:         "Human-readable title of the series.",
				MarkdownDescription: "Human-readable title of the series.",
			},
			"series_type": schema.StringAttribute{
				Computed:            true,
				Description:         "Type of series, e.g. \"single\", describing how its events are structured.",
				MarkdownDescription: "Type of series, e.g. `single`, describing how its events are structured.",
			},
			"recurrence": schema.StringAttribute{
				Computed:            true,
				Description:         "Cadence at which new events are created, e.g. \"weekly\" or \"daily\".",
				MarkdownDescription: "Cadence at which new events are created, e.g. `weekly` or `daily`.",
			},
			"image": schema.StringAttribute{
				Computed:            true,
				Description:         "URL of the series' banner image.",
				MarkdownDescription: "URL of the series' banner image.",
			},
			"icon": schema.StringAttribute{
				Computed:            true,
				Description:         "URL of the series' icon image.",
				MarkdownDescription: "URL of the series' icon image.",
			},
			"active": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the series is currently active.",
				MarkdownDescription: "Whether the series is currently active.",
			},
			"closed": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the series has closed.",
				MarkdownDescription: "Whether the series has closed.",
			},
			"archived": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the series has been archived and hidden from the default UI.",
				MarkdownDescription: "Whether the series has been archived and hidden from the default UI.",
			},
			"featured": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the series is featured on Polymarket.",
				MarkdownDescription: "Whether the series is featured on Polymarket.",
			},
			"restricted": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the series is geo-restricted or otherwise access-limited.",
				MarkdownDescription: "Whether the series is geo-restricted or otherwise access-limited.",
			},
			"comments_enabled": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether user comments are enabled on the series.",
				MarkdownDescription: "Whether user comments are enabled on the series.",
			},
			"competitive": schema.StringAttribute{
				Computed:            true,
				Description:         "Polymarket's competitiveness score for the series, as a decimal string.",
				MarkdownDescription: "Polymarket's competitiveness score for the series, as a decimal string.",
			},
			"volume_24hr": schema.Float64Attribute{
				Computed:            true,
				Description:         "Combined trading volume across the series over the trailing 24 hours, in USDC.",
				MarkdownDescription: "Combined trading volume across the series over the trailing 24 hours, in USDC.",
			},
			"volume": schema.Float64Attribute{
				Computed:            true,
				Description:         "Combined cumulative trading volume across the series, in USDC.",
				MarkdownDescription: "Combined cumulative trading volume across the series, in USDC.",
			},
			"liquidity": schema.Float64Attribute{
				Computed:            true,
				Description:         "Combined liquidity available across the series, in USDC.",
				MarkdownDescription: "Combined liquidity available across the series, in USDC.",
			},
			"start_date": schema.StringAttribute{
				Computed:            true,
				Description:         "ISO-8601 timestamp at which the series began.",
				MarkdownDescription: "ISO-8601 timestamp at which the series began.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "ISO-8601 timestamp at which the series record was created.",
				MarkdownDescription: "ISO-8601 timestamp at which the series record was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "ISO-8601 timestamp at which the series record was last updated.",
				MarkdownDescription: "ISO-8601 timestamp at which the series record was last updated.",
			},
			"comment_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "Number of user comments posted across the series.",
				MarkdownDescription: "Number of user comments posted across the series.",
			},
			"events": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Events that make up the series, each embedding its markets.",
				MarkdownDescription: "Events that make up the series, each embedding its markets.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: eventAttributes(false),
				},
			},
		},
	}
}

// Configure receives the shared API client from the provider.
func (d *seriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches the series and maps it onto Terraform state.
func (d *seriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data seriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	series, err := d.client.GetSeries(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Polymarket series",
			fmt.Sprintf("Could not fetch series %q: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(flattenSeries(ctx, *series, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// flattenSeries maps an API series onto a seriesDataSourceModel, including its
// nested events.
func flattenSeries(ctx context.Context, s client.Series, out *seriesDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	out.ID = types.StringValue(s.ID)
	out.Ticker = types.StringValue(s.Ticker)
	out.Slug = types.StringValue(s.Slug)
	out.Title = types.StringValue(s.Title)
	out.SeriesType = types.StringValue(s.SeriesType)
	out.Recurrence = types.StringValue(s.Recurrence)
	out.Image = types.StringValue(s.Image)
	out.Icon = types.StringValue(s.Icon)
	out.Active = types.BoolValue(s.Active)
	out.Closed = types.BoolValue(s.Closed)
	out.Archived = types.BoolValue(s.Archived)
	out.Featured = types.BoolValue(s.Featured)
	out.Restricted = types.BoolValue(s.Restricted)
	out.CommentsEnabled = types.BoolValue(s.CommentsEnabled)
	out.Competitive = types.StringValue(s.Competitive)
	out.Volume24hr = types.Float64Value(s.Volume24hr)
	out.Volume = types.Float64Value(s.Volume)
	out.Liquidity = types.Float64Value(s.Liquidity)
	out.StartDate = types.StringValue(s.StartDate)
	out.CreatedAt = types.StringValue(s.CreatedAt)
	out.UpdatedAt = types.StringValue(s.UpdatedAt)
	out.CommentCount = types.Int64Value(s.CommentCount)

	out.Events = make([]eventDataSourceModel, 0, len(s.Events))
	for _, e := range s.Events {
		var entry eventDataSourceModel
		diags.Append(flattenEvent(ctx, e, &entry)...)
		out.Events = append(out.Events, entry)
	}

	return diags
}
