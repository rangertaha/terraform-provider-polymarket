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
	_ datasource.DataSource              = (*priceHistoryDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*priceHistoryDataSource)(nil)
)

// priceHistoryDataSource reads the historical price series for an outcome token.
type priceHistoryDataSource struct {
	client *client.Client
}

type priceHistoryDataSourceModel struct {
	TokenID  types.String      `tfsdk:"token_id"`
	Interval types.String      `tfsdk:"interval"`
	Fidelity types.Int64       `tfsdk:"fidelity"`
	History  []pricePointModel `tfsdk:"history"`
}

type pricePointModel struct {
	Timestamp types.Int64   `tfsdk:"timestamp"`
	Price     types.Float64 `tfsdk:"price"`
}

// NewPriceHistoryDataSource is the data source constructor registered on the provider.
func NewPriceHistoryDataSource() datasource.DataSource { return &priceHistoryDataSource{} }

func (d *priceHistoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_price_history"
}

func (d *priceHistoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the historical price series for a CLOB outcome token, as a list of " +
			"timestamped price samples suitable for charting.",
		MarkdownDescription: "Reads the historical price series for a CLOB outcome token, as a " +
			"list of timestamped price samples suitable for charting.",
		Attributes: map[string]schema.Attribute{
			"token_id": tokenIDAttribute(),
			"interval": schema.StringAttribute{
				Optional: true,
				Description: "Time window to return, e.g. \"1h\", \"6h\", \"1d\", \"1w\", or " +
					"\"max\". Defaults to \"1d\" when omitted.",
				MarkdownDescription: "Time window to return, e.g. `1h`, `6h`, `1d`, `1w`, or " +
					"`max`. Defaults to `1d` when omitted.",
			},
			"fidelity": schema.Int64Attribute{
				Optional: true,
				Description: "Resolution of the series, in minutes. When omitted, the server " +
					"chooses an appropriate resolution for the interval.",
				MarkdownDescription: "Resolution of the series, in minutes. When omitted, the " +
					"server chooses an appropriate resolution for the interval.",
			},
			"history": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Price samples, oldest first.",
				MarkdownDescription: "Price samples, oldest first.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"timestamp": schema.Int64Attribute{
							Computed:            true,
							Description:         "Sample time, in Unix seconds.",
							MarkdownDescription: "Sample time, in Unix seconds.",
						},
						"price": schema.Float64Attribute{
							Computed:            true,
							Description:         "Price at the sample time, in [0, 1].",
							MarkdownDescription: "Price at the sample time, in `[0, 1]`.",
						},
					},
				},
			},
		},
	}
}

func (d *priceHistoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *priceHistoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data priceHistoryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	interval := data.Interval.ValueString()
	if interval == "" {
		interval = "1d"
	}

	points, err := d.client.GetPriceHistory(ctx, data.TokenID.ValueString(), interval, data.Fidelity.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Polymarket price history", err.Error())
		return
	}

	data.History = make([]pricePointModel, 0, len(points))
	for _, p := range points {
		data.History = append(data.History, pricePointModel{
			Timestamp: types.Int64Value(p.T),
			Price:     types.Float64Value(p.P),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
