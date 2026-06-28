// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// eventDataSourceModel maps every event attribute to a Go type. Shared by the
// single-event data source and as the element type of the events list.
type eventDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Ticker           types.String `tfsdk:"ticker"`
	Slug             types.String `tfsdk:"slug"`
	Title            types.String `tfsdk:"title"`
	Description      types.String `tfsdk:"description"`
	ResolutionSource types.String `tfsdk:"resolution_source"`
	StartDate        types.String `tfsdk:"start_date"`
	CreationDate     types.String `tfsdk:"creation_date"`
	EndDate          types.String `tfsdk:"end_date"`
	Image            types.String `tfsdk:"image"`
	Icon             types.String `tfsdk:"icon"`
	Active           types.Bool   `tfsdk:"active"`
	Closed           types.Bool   `tfsdk:"closed"`
	Archived         types.Bool   `tfsdk:"archived"`
	New              types.Bool   `tfsdk:"new"`
	Featured         types.Bool   `tfsdk:"featured"`
	Restricted       types.Bool   `tfsdk:"restricted"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	EnableOrderBook  types.Bool   `tfsdk:"enable_order_book"`
	NegRisk          types.Bool   `tfsdk:"neg_risk"`
	CommentCount     types.Int64  `tfsdk:"comment_count"`
	SeriesSlug       types.String `tfsdk:"series_slug"`

	Tags    []tagModel              `tfsdk:"tags"`
	Markets []marketDataSourceModel `tfsdk:"markets"`
}

// tagModel maps a category tag to Go types.
type tagModel struct {
	ID    types.String `tfsdk:"id"`
	Label types.String `tfsdk:"label"`
	Slug  types.String `tfsdk:"slug"`
}

// eventAttributes returns the schema for an event. When idRequired is true the
// "id" attribute is a required input (single-event lookup); otherwise every
// attribute is computed (an event nested inside a list result).
func eventAttributes(idRequired bool) map[string]schema.Attribute {
	idAttr := schema.StringAttribute{
		Description:         "Numeric identifier of the event, as assigned by Polymarket.",
		MarkdownDescription: "Numeric identifier of the event, as assigned by Polymarket.",
	}
	if idRequired {
		idAttr.Required = true
		idAttr.Description = "Numeric identifier of the event to fetch, as assigned by " +
			"Polymarket. This is the only required input."
		idAttr.MarkdownDescription = "Numeric identifier of the event to fetch, as assigned by " +
			"Polymarket. This is the only required input."
	} else {
		idAttr.Computed = true
	}

	return map[string]schema.Attribute{
		"id": idAttr,
		"ticker": schema.StringAttribute{
			Computed:            true,
			Description:         "Short ticker symbol identifying the event.",
			MarkdownDescription: "Short ticker symbol identifying the event.",
		},
		"slug": schema.StringAttribute{
			Computed:            true,
			Description:         "URL-friendly slug identifying the event on polymarket.com.",
			MarkdownDescription: "URL-friendly slug identifying the event on polymarket.com.",
		},
		"title": schema.StringAttribute{
			Computed:            true,
			Description:         "Human-readable title of the event.",
			MarkdownDescription: "Human-readable title of the event.",
		},
		"description": schema.StringAttribute{
			Computed:            true,
			Description:         "Long-form description of the event and how its markets resolve.",
			MarkdownDescription: "Long-form description of the event and how its markets resolve.",
		},
		"resolution_source": schema.StringAttribute{
			Computed:            true,
			Description:         "Source of truth consulted to resolve the event's markets, when specified.",
			MarkdownDescription: "Source of truth consulted to resolve the event's markets, when specified.",
		},
		"start_date": schema.StringAttribute{
			Computed:            true,
			Description:         "ISO-8601 timestamp at which the event opened.",
			MarkdownDescription: "ISO-8601 timestamp at which the event opened.",
		},
		"creation_date": schema.StringAttribute{
			Computed:            true,
			Description:         "ISO-8601 timestamp at which the event was created.",
			MarkdownDescription: "ISO-8601 timestamp at which the event was created.",
		},
		"end_date": schema.StringAttribute{
			Computed:            true,
			Description:         "ISO-8601 timestamp at which the event is scheduled to end.",
			MarkdownDescription: "ISO-8601 timestamp at which the event is scheduled to end.",
		},
		"image": schema.StringAttribute{
			Computed:            true,
			Description:         "URL of the event's banner image.",
			MarkdownDescription: "URL of the event's banner image.",
		},
		"icon": schema.StringAttribute{
			Computed:            true,
			Description:         "URL of the event's icon image.",
			MarkdownDescription: "URL of the event's icon image.",
		},
		"active": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the event is currently active.",
			MarkdownDescription: "Whether the event is currently active.",
		},
		"closed": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the event has closed.",
			MarkdownDescription: "Whether the event has closed.",
		},
		"archived": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the event has been archived and hidden from the default UI.",
			MarkdownDescription: "Whether the event has been archived and hidden from the default UI.",
		},
		"new": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the event is flagged as newly listed.",
			MarkdownDescription: "Whether the event is flagged as newly listed.",
		},
		"featured": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the event is featured on Polymarket.",
			MarkdownDescription: "Whether the event is featured on Polymarket.",
		},
		"restricted": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the event is geo-restricted or otherwise access-limited.",
			MarkdownDescription: "Whether the event is geo-restricted or otherwise access-limited.",
		},
		"created_at": schema.StringAttribute{
			Computed:            true,
			Description:         "ISO-8601 timestamp at which the event record was created.",
			MarkdownDescription: "ISO-8601 timestamp at which the event record was created.",
		},
		"updated_at": schema.StringAttribute{
			Computed:            true,
			Description:         "ISO-8601 timestamp at which the event record was last updated.",
			MarkdownDescription: "ISO-8601 timestamp at which the event record was last updated.",
		},
		"enable_order_book": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether a CLOB order book is enabled for the event's markets.",
			MarkdownDescription: "Whether a CLOB order book is enabled for the event's markets.",
		},
		"neg_risk": schema.BoolAttribute{
			Computed: true,
			Description: "Whether the event uses negative-risk (mutually exclusive multi-outcome) " +
				"settlement across its markets.",
			MarkdownDescription: "Whether the event uses negative-risk (mutually exclusive multi-outcome) " +
				"settlement across its markets.",
		},
		"comment_count": schema.Int64Attribute{
			Computed:            true,
			Description:         "Number of user comments posted on the event.",
			MarkdownDescription: "Number of user comments posted on the event.",
		},
		"series_slug": schema.StringAttribute{
			Computed:            true,
			Description:         "Slug of the recurring series this event belongs to, when part of one.",
			MarkdownDescription: "Slug of the recurring series this event belongs to, when part of one.",
		},
		"tags": schema.ListNestedAttribute{
			Computed:            true,
			Description:         "Category tags attached to the event.",
			MarkdownDescription: "Category tags attached to the event.",
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
				},
			},
		},
		"markets": schema.ListNestedAttribute{
			Computed:            true,
			Description:         "Markets that make up the event.",
			MarkdownDescription: "Markets that make up the event.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: marketAttributes(false),
			},
		},
	}
}

// flattenEvent maps an API event onto an eventDataSourceModel, including its
// nested markets and tags.
func flattenEvent(ctx context.Context, e client.Event, out *eventDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	out.ID = types.StringValue(e.ID)
	out.Ticker = types.StringValue(e.Ticker)
	out.Slug = types.StringValue(e.Slug)
	out.Title = types.StringValue(e.Title)
	out.Description = types.StringValue(e.Description)
	out.ResolutionSource = types.StringValue(e.ResolutionSource)
	out.StartDate = types.StringValue(e.StartDate)
	out.CreationDate = types.StringValue(e.CreationDate)
	out.EndDate = types.StringValue(e.EndDate)
	out.Image = types.StringValue(e.Image)
	out.Icon = types.StringValue(e.Icon)
	out.Active = types.BoolValue(e.Active)
	out.Closed = types.BoolValue(e.Closed)
	out.Archived = types.BoolValue(e.Archived)
	out.New = types.BoolValue(e.New)
	out.Featured = types.BoolValue(e.Featured)
	out.Restricted = types.BoolValue(e.Restricted)
	out.CreatedAt = types.StringValue(e.CreatedAt)
	out.UpdatedAt = types.StringValue(e.UpdatedAt)
	out.EnableOrderBook = types.BoolValue(e.EnableOrderBook)
	out.NegRisk = types.BoolValue(e.NegRisk)
	out.CommentCount = types.Int64Value(e.CommentCount)
	out.SeriesSlug = types.StringValue(e.SeriesSlug)

	out.Tags = make([]tagModel, 0, len(e.Tags))
	for _, t := range e.Tags {
		out.Tags = append(out.Tags, tagModel{
			ID:    types.StringValue(t.ID),
			Label: types.StringValue(t.Label),
			Slug:  types.StringValue(t.Slug),
		})
	}

	out.Markets = make([]marketDataSourceModel, 0, len(e.Markets))
	for _, m := range e.Markets {
		var entry marketDataSourceModel
		diags.Append(flattenMarket(ctx, m, &entry)...)
		out.Markets = append(out.Markets, entry)
	}

	return diags
}
