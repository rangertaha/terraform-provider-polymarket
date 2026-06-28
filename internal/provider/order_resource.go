// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*orderResource)(nil)
	_ resource.ResourceWithConfigure = (*orderResource)(nil)
)

// orderResource manages a single CLOB limit order. Because an order book moves
// out-of-band (orders fill, partially fill, or expire without Terraform's
// involvement), every input forces replacement: a change cancels the old order
// and places a new one. Create places, Read refreshes status, Delete cancels.
type orderResource struct {
	client *client.Client
}

type orderResourceModel struct {
	TokenID     types.String  `tfsdk:"token_id"`
	Side        types.String  `tfsdk:"side"`
	Price       types.Float64 `tfsdk:"price"`
	Size        types.Float64 `tfsdk:"size"`
	FeeRateBps  types.Int64   `tfsdk:"fee_rate_bps"`
	Expiration  types.Int64   `tfsdk:"expiration"`
	NegRisk     types.Bool    `tfsdk:"neg_risk"`
	OrderType   types.String  `tfsdk:"order_type"`
	ID          types.String  `tfsdk:"id"`
	Status      types.String  `tfsdk:"status"`
	SizeMatched types.String  `tfsdk:"size_matched"`
}

// NewOrderResource is the resource constructor registered on the provider.
func NewOrderResource() resource.Resource { return &orderResource{} }

func (r *orderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_order"
}

func (r *orderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Places a limit order on the Polymarket CLOB. Requires the provider's " +
			"private_key. Because orders fill out-of-band, any change replaces the order " +
			"(cancel + re-place); deleting the resource cancels the order.",
		MarkdownDescription: "Places a limit order on the Polymarket CLOB. Requires the " +
			"provider's `private_key`.\n\n~> **Orders are not idempotent infrastructure.** An " +
			"order can fill, partially fill, or expire without Terraform's knowledge. Every " +
			"input forces replacement: changing one cancels the old order and places a new " +
			"one. Deleting the resource cancels the order if it is still live.",
		Attributes: map[string]schema.Attribute{
			"token_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "CLOB ERC-1155 token ID of the outcome to trade. Obtain it from a " +
					"market's clob_token_ids attribute.",
				MarkdownDescription: "CLOB ERC-1155 token ID of the outcome to trade. Obtain it " +
					"from a market's `clob_token_ids` attribute.",
			},
			"side": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:          []validator.String{stringvalidator.OneOf("BUY", "SELL")},
				Description:         "Order side: \"BUY\" or \"SELL\".",
				MarkdownDescription: "Order side: `BUY` or `SELL`.",
			},
			"price": schema.Float64Attribute{
				Required:            true,
				PlanModifiers:       []planmodifier.Float64{float64planmodifier.RequiresReplace()},
				Validators:          []validator.Float64{openUnitInterval()},
				Description:         "Limit price per share, in the open interval (0, 1).",
				MarkdownDescription: "Limit price per share, in the open interval `(0, 1)`.",
			},
			"size": schema.Float64Attribute{
				Required:            true,
				PlanModifiers:       []planmodifier.Float64{float64planmodifier.RequiresReplace()},
				Description:         "Order size, in shares.",
				MarkdownDescription: "Order size, in shares.",
			},
			"fee_rate_bps": schema.Int64Attribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Description:         "Fee rate in basis points. Defaults to 0 when omitted.",
				MarkdownDescription: "Fee rate in basis points. Defaults to `0` when omitted.",
			},
			"expiration": schema.Int64Attribute{
				Optional:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Description: "Expiration as a Unix timestamp (seconds). 0 or omitted means no " +
					"expiry (good-till-cancelled).",
				MarkdownDescription: "Expiration as a Unix timestamp (seconds). `0` or omitted " +
					"means no expiry (good-till-cancelled).",
			},
			"neg_risk": schema.BoolAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
				Description: "Set to true when the market uses negative-risk settlement, so the " +
					"order is signed against the NegRisk exchange contract. Defaults to false.",
				MarkdownDescription: "Set to `true` when the market uses negative-risk settlement, " +
					"so the order is signed against the NegRisk exchange contract. Defaults to `false`.",
			},
			"order_type": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "Order type: \"GTC\" (good-till-cancelled, the default), \"GTD\" " +
					"(good-till-date, requires expiration), \"FOK\", or \"FAK\".",
				MarkdownDescription: "Order type: `GTC` (good-till-cancelled, the default), `GTD` " +
					"(good-till-date, requires `expiration`), `FOK`, or `FAK`.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Description:         "CLOB-assigned order ID.",
				MarkdownDescription: "CLOB-assigned order ID.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "Current order status as last refreshed (e.g. LIVE, MATCHED, CANCELED).",
				MarkdownDescription: "Current order status as last refreshed (e.g. `LIVE`, `MATCHED`, `CANCELED`).",
			},
			"size_matched": schema.StringAttribute{
				Computed:            true,
				Description:         "Size filled so far, in shares, as a decimal string.",
				MarkdownDescription: "Size filled so far, in shares, as a decimal string.",
			},
		},
	}
}

func (r *orderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *orderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan orderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !r.client.HasSigner() {
		resp.Diagnostics.AddError("Missing private_key",
			"polymarket_order requires the provider's private_key (or POLYMARKET_PRIVATE_KEY).")
		return
	}

	placed, err := r.client.PlaceOrder(ctx, client.OrderArgs{
		TokenID:    plan.TokenID.ValueString(),
		Side:       plan.Side.ValueString(),
		Price:      plan.Price.ValueFloat64(),
		Size:       plan.Size.ValueFloat64(),
		FeeRateBps: plan.FeeRateBps.ValueInt64(),
		Expiration: plan.Expiration.ValueInt64(),
		NegRisk:    plan.NegRisk.ValueBool(),
		OrderType:  plan.OrderType.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to place Polymarket order", err.Error())
		return
	}

	plan.ID = types.StringValue(placed.OrderID)
	plan.Status = types.StringValue(placed.Status)
	plan.SizeMatched = types.StringValue("0")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *orderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state orderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, ok, err := r.client.GetOrder(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to refresh Polymarket order", err.Error())
		return
	}
	if !ok {
		// The order is no longer known to the CLOB (filled, cancelled, or
		// expired). Drop it from state so Terraform plans a fresh order.
		resp.State.RemoveResource(ctx)
		return
	}

	state.Status = types.StringValue(status.Status)
	state.SizeMatched = types.StringValue(status.SizeMatched)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is effectively unreachable because every input forces replacement; it
// exists to satisfy the interface and simply persists the planned values.
func (r *orderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan orderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *orderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state orderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.CancelOrder(ctx, state.ID.ValueString()); err != nil {
		// A non-live order cannot be cancelled; treat that as already gone.
		resp.Diagnostics.AddWarning(
			"Order could not be cancelled",
			fmt.Sprintf("Cancelling order %s failed (it may already be filled or expired): %s",
				state.ID.ValueString(), err),
		)
	}
}
