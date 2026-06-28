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
	_ datasource.DataSource              = (*apiCredentialsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*apiCredentialsDataSource)(nil)
)

// apiCredentialsDataSource derives the L2 API credentials for the wallet
// configured via the provider's private_key.
type apiCredentialsDataSource struct {
	client *client.Client
}

type apiCredentialsDataSourceModel struct {
	APIKey     types.String `tfsdk:"api_key"`
	Secret     types.String `tfsdk:"secret"`
	Passphrase types.String `tfsdk:"passphrase"`
}

// NewAPICredentialsDataSource is the data source constructor registered on the provider.
func NewAPICredentialsDataSource() datasource.DataSource { return &apiCredentialsDataSource{} }

func (d *apiCredentialsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_credentials"
}

func (d *apiCredentialsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Derives the L2 API credentials (API key, secret, passphrase) for the " +
			"wallet configured via the provider's private_key, using an EIP-712 L1 signature. " +
			"Requires private_key to be set. The credentials are deterministic for a wallet.",
		MarkdownDescription: "Derives the L2 API credentials (API key, secret, passphrase) for " +
			"the wallet configured via the provider's `private_key`, using an EIP-712 L1 " +
			"signature. Requires `private_key` to be set. The credentials are deterministic for " +
			"a wallet.\n\n~> **Note:** These are secrets and will be stored in Terraform state. " +
			"Protect your state accordingly.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Computed:            true,
				Description:         "Derived API key (a UUID) identifying the credential set.",
				MarkdownDescription: "Derived API key (a UUID) identifying the credential set.",
			},
			"secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				Description:         "Derived API secret used to HMAC-sign L2 requests. Sensitive.",
				MarkdownDescription: "Derived API secret used to HMAC-sign L2 requests. Sensitive.",
			},
			"passphrase": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				Description:         "Derived API passphrase sent with L2 requests. Sensitive.",
				MarkdownDescription: "Derived API passphrase sent with L2 requests. Sensitive.",
			},
		},
	}
}

func (d *apiCredentialsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req, resp)
}

func (d *apiCredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}
	if !d.client.HasSigner() {
		resp.Diagnostics.AddError(
			"Missing private_key",
			"The polymarket_api_credentials data source requires the provider's private_key "+
				"(or POLYMARKET_PRIVATE_KEY) to be configured.",
		)
		return
	}

	creds, err := d.client.DeriveAPIKey(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to derive Polymarket API credentials", err.Error())
		return
	}

	state := apiCredentialsDataSourceModel{
		APIKey:     types.StringValue(creds.APIKey),
		Secret:     types.StringValue(creds.Secret),
		Passphrase: types.StringValue(creds.Passphrase),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
