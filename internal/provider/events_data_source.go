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
	_ datasource.DataSource              = (*eventsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*eventsDataSource)(nil)
)

// eventsDataSource lists Polymarket events matching optional filters.
type eventsDataSource struct {
	client *client.Client
}

// eventsDataSourceModel is the top-level model: filter inputs plus results.
type eventsDataSourceModel struct {
	Limit  types.Int64  `tfsdk:"limit"`
	Offset types.Int64  `tfsdk:"offset"`
	Active types.Bool   `tfsdk:"active"`
	Closed types.Bool   `tfsdk:"closed"`
	Slug   types.String `tfsdk:"slug"`

	Events []eventDataSourceModel `tfsdk:"events"`
}

// NewEventsDataSource is the data source constructor registered on the provider.
func NewEventsDataSource() datasource.DataSource {
	return &eventsDataSource{}
}

// Metadata sets the data source type name ("polymarket_events").
func (d *eventsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_events"
}

// Schema describes the filter inputs and the nested list of returned events.
func (d *eventsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Polymarket events, optionally filtered by activity, closed state, " +
			"or slug, with pagination support. Each event embeds its markets.",
		MarkdownDescription: "Lists Polymarket events, optionally filtered by activity, closed " +
			"state, or slug, with pagination support. Each event embeds its markets.",
		Attributes: map[string]schema.Attribute{
			"limit": schema.Int64Attribute{
				Optional:            true,
				Description:         "Maximum number of events to return. Combine with offset to paginate.",
				MarkdownDescription: "Maximum number of events to return. Combine with `offset` to paginate.",
			},
			"offset": schema.Int64Attribute{
				Optional:            true,
				Description:         "Number of events to skip before collecting results, for pagination.",
				MarkdownDescription: "Number of events to skip before collecting results, for pagination.",
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Description:         "When set, returns only events whose active flag matches this value.",
				MarkdownDescription: "When set, returns only events whose `active` flag matches this value.",
			},
			"closed": schema.BoolAttribute{
				Optional:            true,
				Description:         "When set, returns only events whose closed flag matches this value.",
				MarkdownDescription: "When set, returns only events whose `closed` flag matches this value.",
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				Description:         "When set, returns only the event with this exact URL slug.",
				MarkdownDescription: "When set, returns only the event with this exact URL slug.",
			},
			"events": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "List of events matching the supplied filters.",
				MarkdownDescription: "List of events matching the supplied filters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: eventAttributes(false),
				},
			},
		},
	}
}

// Configure receives the shared API client from the provider.
func (d *eventsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read applies the filters, fetches matching events, and maps them to state.
func (d *eventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data eventsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter := client.EventFilter{
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

	events, err := d.client.ListEvents(ctx, filter)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Polymarket events", err.Error())
		return
	}

	data.Events = make([]eventDataSourceModel, 0, len(events))
	for _, e := range events {
		var entry eventDataSourceModel
		resp.Diagnostics.Append(flattenEvent(ctx, e, &entry)...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Events = append(data.Events, entry)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
