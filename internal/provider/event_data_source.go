// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = (*eventDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*eventDataSource)(nil)
)

// eventDataSource looks up a single Polymarket event (a group of markets) by ID.
type eventDataSource struct {
	client *client.Client
}

// NewEventDataSource is the data source constructor registered on the provider.
func NewEventDataSource() datasource.DataSource {
	return &eventDataSource{}
}

// Metadata sets the data source type name ("polymarket_event").
func (d *eventDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event"
}

// Schema describes a single event. The attribute set is shared with the events
// list data source via eventAttributes so the two never diverge.
func (d *eventDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Polymarket event by its numeric ID. An event groups " +
			"one or more related markets, e.g. all candidates in an election.",
		MarkdownDescription: "Fetches a single Polymarket event by its numeric `id`. An event " +
			"groups one or more related markets, e.g. all candidates in an election.",
		Attributes: eventAttributes(true),
	}
}

// Configure receives the shared API client from the provider.
func (d *eventDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches the event and maps it onto Terraform state.
func (d *eventDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data eventDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	event, err := d.client.GetEvent(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Polymarket event",
			fmt.Sprintf("Could not fetch event %q: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(flattenEvent(ctx, *event, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
