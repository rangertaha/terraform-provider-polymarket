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
	_ datasource.DataSource              = (*tagsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*tagsDataSource)(nil)
)

// tagsDataSource lists Polymarket category tags.
type tagsDataSource struct {
	client *client.Client
}

// tagsDataSourceModel is the top-level model: filter inputs plus results.
type tagsDataSourceModel struct {
	Limit  types.Int64 `tfsdk:"limit"`
	Offset types.Int64 `tfsdk:"offset"`

	Tags []tagListItemModel `tfsdk:"tags"`
}

// tagListItemModel maps a tag (with timestamps) returned by the /tags endpoint.
type tagListItemModel struct {
	ID        types.String `tfsdk:"id"`
	Label     types.String `tfsdk:"label"`
	Slug      types.String `tfsdk:"slug"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// NewTagsDataSource is the data source constructor registered on the provider.
func NewTagsDataSource() datasource.DataSource {
	return &tagsDataSource{}
}

// Metadata sets the data source type name ("polymarket_tags").
func (d *tagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

// Schema describes the filter inputs and the nested list of returned tags.
func (d *tagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Polymarket category tags, which are used to classify markets and " +
			"events. Supports pagination.",
		MarkdownDescription: "Lists Polymarket category tags, which are used to classify markets " +
			"and events. Supports pagination.",
		Attributes: map[string]schema.Attribute{
			"limit": schema.Int64Attribute{
				Optional:            true,
				Description:         "Maximum number of tags to return. Combine with offset to paginate.",
				MarkdownDescription: "Maximum number of tags to return. Combine with `offset` to paginate.",
			},
			"offset": schema.Int64Attribute{
				Optional:            true,
				Description:         "Number of tags to skip before collecting results, for pagination.",
				MarkdownDescription: "Number of tags to skip before collecting results, for pagination.",
			},
			"tags": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "List of category tags.",
				MarkdownDescription: "List of category tags.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "Numeric identifier of the tag.",
							MarkdownDescription: "Numeric identifier of the tag.",
						},
						"label": schema.StringAttribute{
							Computed:            true,
							Description:         "Human-readable label of the tag.",
							MarkdownDescription: "Human-readable label of the tag.",
						},
						"slug": schema.StringAttribute{
							Computed:            true,
							Description:         "URL-friendly slug of the tag.",
							MarkdownDescription: "URL-friendly slug of the tag.",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							Description:         "ISO-8601 timestamp at which the tag was created.",
							MarkdownDescription: "ISO-8601 timestamp at which the tag was created.",
						},
						"updated_at": schema.StringAttribute{
							Computed:            true,
							Description:         "ISO-8601 timestamp at which the tag was last updated.",
							MarkdownDescription: "ISO-8601 timestamp at which the tag was last updated.",
						},
					},
				},
			},
		},
	}
}

// Configure receives the shared API client from the provider.
func (d *tagsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read applies the filters, fetches matching tags, and maps them to state.
func (d *tagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tagsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, err := d.client.ListTags(ctx, client.TagFilter{
		Limit:  data.Limit.ValueInt64(),
		Offset: data.Offset.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Polymarket tags", err.Error())
		return
	}

	data.Tags = make([]tagListItemModel, 0, len(tags))
	for _, t := range tags {
		data.Tags = append(data.Tags, tagListItemModel{
			ID:        types.StringValue(t.ID),
			Label:     types.StringValue(t.Label),
			Slug:      types.StringValue(t.Slug),
			CreatedAt: types.StringValue(t.CreatedAt),
			UpdatedAt: types.StringValue(t.UpdatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
