// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*apiKeyResource)(nil)
	_ resource.ResourceWithConfigure = (*apiKeyResource)(nil)
)

// apiKeyResource provisions an L2 API key set for the configured wallet. Create
// calls POST /auth/api-key; Delete revokes it. The key set is non-updatable, so
// there is no Update behavior beyond persisting state.
type apiKeyResource struct {
	client *client.Client
}

type apiKeyResourceModel struct {
	APIKey     types.String `tfsdk:"api_key"`
	Secret     types.String `tfsdk:"secret"`
	Passphrase types.String `tfsdk:"passphrase"`
}

// NewAPIKeyResource is the resource constructor registered on the provider.
func NewAPIKeyResource() resource.Resource { return &apiKeyResource{} }

func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provisions an L2 API key set for the wallet configured via the provider's " +
			"private_key. Create issues a fresh key; destroying the resource revokes it.",
		MarkdownDescription: "Provisions an L2 API key set for the wallet configured via the " +
			"provider's `private_key`. Create issues a fresh key; destroying the resource " +
			"revokes it.\n\n~> **Note:** the `secret` and `passphrase` are stored in Terraform " +
			"state. Protect your state accordingly.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Computed:            true,
				Description:         "Issued API key (a UUID) identifying the credential set.",
				MarkdownDescription: "Issued API key (a UUID) identifying the credential set.",
			},
			"secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				Description:         "Issued API secret used to HMAC-sign L2 requests. Sensitive.",
				MarkdownDescription: "Issued API secret used to HMAC-sign L2 requests. Sensitive.",
			},
			"passphrase": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				Description:         "Issued API passphrase sent with L2 requests. Sensitive.",
				MarkdownDescription: "Issued API passphrase sent with L2 requests. Sensitive.",
			},
		},
	}
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *apiKeyResource) Create(ctx context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.client.HasSigner() {
		resp.Diagnostics.AddError("Missing private_key",
			"polymarket_api_key requires the provider's private_key (or POLYMARKET_PRIVATE_KEY).")
		return
	}

	creds, err := r.client.CreateAPIKey(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Polymarket API key", err.Error())
		return
	}

	state := apiKeyResourceModel{
		APIKey:     types.StringValue(creds.APIKey),
		Secret:     types.StringValue(creds.Secret),
		Passphrase: types.StringValue(creds.Passphrase),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read is a no-op: the CLOB has no fetch-by-key endpoint, and the credentials
// are already fully captured in state.
func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, req.State.Raw)...)
}

// Update never runs: the resource has no updatable attributes.
func (r *apiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, req.Plan.Raw)...)
}

func (r *apiKeyResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.DeleteAPIKey(ctx); err != nil {
		resp.Diagnostics.AddWarning(
			"API key could not be revoked",
			fmt.Sprintf("Revoking the API key failed (it may already be gone): %s", err),
		)
	}
}
