// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/chain"
	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*allowanceResource)(nil)
	_ resource.ResourceWithConfigure = (*allowanceResource)(nil)
)

// allowanceResource manages a single on-chain token approval: either an ERC-20
// spend allowance (USDC) or an ERC-1155 operator approval (CTF outcome tokens).
// These are the approvals the CTF Exchange needs before it can fill orders.
type allowanceResource struct {
	client *client.Client
}

type allowanceResourceModel struct {
	Token           types.String `tfsdk:"token"`
	Spender         types.String `tfsdk:"spender"`
	ERC1155         types.Bool   `tfsdk:"erc1155"`
	Amount          types.String `tfsdk:"amount"`
	Approved        types.Bool   `tfsdk:"approved"`
	TransactionHash types.String `tfsdk:"transaction_hash"`
}

// NewAllowanceResource is the resource constructor registered on the provider.
func NewAllowanceResource() resource.Resource { return &allowanceResource{} }

func (r *allowanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allowance"
}

func (r *allowanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an on-chain token approval the Polymarket exchange needs to fill " +
			"orders: an ERC-20 USDC spend allowance or an ERC-1155 CTF operator approval. " +
			"Requires the provider's private_key and rpc_endpoint.",
		MarkdownDescription: "Manages an on-chain token approval the Polymarket exchange needs " +
			"to fill orders: an ERC-20 USDC spend allowance or an ERC-1155 CTF operator " +
			"approval. Requires the provider's `private_key` and `rpc_endpoint`.\n\n~> This " +
			"resource submits real Polygon transactions that cost gas. Destroying it revokes " +
			"the approval (sets the allowance to 0 / operator approval to false).",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{ethAddress()},
				Description: "Address of the token contract to approve. Typically the USDC token " +
					"(for ERC-20) or the Conditional Tokens contract (for ERC-1155).",
				MarkdownDescription: "Address of the token contract to approve. Typically the USDC " +
					"token (for ERC-20) or the Conditional Tokens contract (for ERC-1155).",
			},
			"spender": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{ethAddress()},
				Description: "Address granted the approval, e.g. the CTF Exchange or NegRisk " +
					"Exchange contract.",
				MarkdownDescription: "Address granted the approval, e.g. the CTF Exchange or " +
					"NegRisk Exchange contract.",
			},
			"erc1155": schema.BoolAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
				Description: "Set to true to manage an ERC-1155 operator approval " +
					"(setApprovalForAll); leave false (default) for an ERC-20 spend allowance.",
				MarkdownDescription: "Set to `true` to manage an ERC-1155 operator approval " +
					"(`setApprovalForAll`); leave `false` (default) for an ERC-20 spend allowance.",
			},
			"amount": schema.StringAttribute{
				Optional: true,
				Description: "ERC-20 approval amount in base units (USDC has 6 decimals). When " +
					"omitted, an unlimited (max-uint256) approval is granted. Ignored when " +
					"erc1155 is true.",
				MarkdownDescription: "ERC-20 approval amount in base units (USDC has 6 decimals). " +
					"When omitted, an unlimited (max-uint256) approval is granted. Ignored when " +
					"`erc1155` is true.",
			},
			"approved": schema.BoolAttribute{
				Computed: true,
				Description: "Whether the approval is currently in place on-chain, as last " +
					"refreshed. For ERC-20 this is true when the allowance covers the requested amount.",
				MarkdownDescription: "Whether the approval is currently in place on-chain, as last " +
					"refreshed. For ERC-20 this is true when the allowance covers the requested amount.",
			},
			"transaction_hash": schema.StringAttribute{
				Computed:            true,
				Description:         "Hash of the most recent approval transaction submitted for this resource.",
				MarkdownDescription: "Hash of the most recent approval transaction submitted for this resource.",
			},
		},
	}
}

func (r *allowanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// approvalAmount resolves the ERC-20 amount, defaulting to unlimited.
func approvalAmount(model allowanceResourceModel) (*big.Int, error) {
	if model.Amount.IsNull() || model.Amount.ValueString() == "" {
		return chain.MaxUint256, nil
	}
	amount, ok := new(big.Int).SetString(model.Amount.ValueString(), 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount %q", model.Amount.ValueString())
	}
	return amount, nil
}

// requireChain validates that the on-chain client is configured.
func (r *allowanceResource) requireChain(resp interface{ AddError(string, string) }) *chain.Client {
	ch := r.client.Chain()
	if ch == nil {
		resp.AddError("Missing rpc_endpoint",
			"polymarket_allowance requires the provider's private_key and rpc_endpoint "+
				"(or POLYMARKET_PRIVATE_KEY and POLYMARKET_RPC_ENDPOINT).")
		return nil
	}
	return ch
}

func (r *allowanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan allowanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ch := r.requireChain(&resp.Diagnostics)
	if ch == nil {
		return
	}

	token := common.HexToAddress(plan.Token.ValueString())
	spender := common.HexToAddress(plan.Spender.ValueString())

	var txHash string
	var err error
	if plan.ERC1155.ValueBool() {
		txHash, err = ch.SetApprovalForAll(ctx, token, spender, true)
	} else {
		var amount *big.Int
		amount, err = approvalAmount(plan)
		if err == nil {
			txHash, err = ch.ApproveERC20(ctx, token, spender, amount)
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to submit approval", err.Error())
		return
	}

	plan.TransactionHash = types.StringValue(txHash)
	plan.Approved = types.BoolValue(true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *allowanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state allowanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ch := r.requireChain(&resp.Diagnostics)
	if ch == nil {
		return
	}

	token := common.HexToAddress(state.Token.ValueString())
	spender := common.HexToAddress(state.Spender.ValueString())
	owner := ch.From()

	var approved bool
	if state.ERC1155.ValueBool() {
		var err error
		approved, err = ch.IsApprovedForAll(ctx, token, owner, spender)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read approval", err.Error())
			return
		}
	} else {
		current, err := ch.AllowanceERC20(ctx, token, owner, spender)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read allowance", err.Error())
			return
		}
		want, err := approvalAmount(state)
		if err != nil {
			resp.Diagnostics.AddError("Invalid amount in state", err.Error())
			return
		}
		approved = current.Cmp(want) >= 0
	}

	state.Approved = types.BoolValue(approved)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *allowanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan allowanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ch := r.requireChain(&resp.Diagnostics)
	if ch == nil {
		return
	}

	// Only the ERC-20 amount is updatable in place; re-approve to the new amount.
	token := common.HexToAddress(plan.Token.ValueString())
	spender := common.HexToAddress(plan.Spender.ValueString())
	amount, err := approvalAmount(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid amount", err.Error())
		return
	}
	txHash, err := ch.ApproveERC20(ctx, token, spender, amount)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update allowance", err.Error())
		return
	}

	plan.TransactionHash = types.StringValue(txHash)
	plan.Approved = types.BoolValue(true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *allowanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state allowanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ch := r.requireChain(&resp.Diagnostics)
	if ch == nil {
		return
	}

	token := common.HexToAddress(state.Token.ValueString())
	spender := common.HexToAddress(state.Spender.ValueString())

	var err error
	if state.ERC1155.ValueBool() {
		_, err = ch.SetApprovalForAll(ctx, token, spender, false)
	} else {
		_, err = ch.ApproveERC20(ctx, token, spender, big.NewInt(0))
	}
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Approval could not be revoked",
			fmt.Sprintf("Revoking the approval failed; you may need to revoke it manually: %s", err),
		)
	}
}
