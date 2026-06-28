// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

// Package provider implements the Polymarket Terraform provider.
package provider

import (
	"context"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/chain"
	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/Rangertaha/terraform-provider-polymarket/internal/sign"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure polymarketProvider satisfies the provider.Provider interface.
var _ provider.Provider = (*polymarketProvider)(nil)

// polymarketProvider is the provider implementation.
type polymarketProvider struct {
	// version is set to the provider release version, or "dev" for local builds.
	version string
}

// polymarketProviderModel maps provider schema attributes to Go types.
type polymarketProviderModel struct {
	Endpoint      types.String `tfsdk:"endpoint"`
	ClobEndpoint  types.String `tfsdk:"clob_endpoint"`
	DataEndpoint  types.String `tfsdk:"data_endpoint"`
	APIKey        types.String `tfsdk:"api_key"`
	PrivateKey    types.String `tfsdk:"private_key"`
	FunderAddress types.String `tfsdk:"funder_address"`
	SignatureType types.Int64  `tfsdk:"signature_type"`
	ChainID       types.Int64  `tfsdk:"chain_id"`
	RPCEndpoint   types.String `tfsdk:"rpc_endpoint"`
}

// New returns a function that constructs the provider, capturing the build
// version so it can be reported to Terraform.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &polymarketProvider{version: version}
	}
}

// Metadata sets the provider type name and version. The TypeName becomes the
// prefix for every resource and data source (e.g. "polymarket_market").
func (p *polymarketProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "polymarket"
	resp.Version = p.version
}

// Schema defines the provider-level configuration block.
func (p *polymarketProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Polymarket provider reads prediction-market data from the " +
			"public Polymarket Gamma Markets API. Configure the optional endpoint and " +
			"API key, or rely on the public defaults which require no credentials.",
		MarkdownDescription: "The Polymarket provider reads prediction-market data from the " +
			"public [Polymarket Gamma Markets API](https://docs.polymarket.com). Configure " +
			"the optional `endpoint` and `api_key`, or rely on the public defaults which " +
			"require no credentials.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the Polymarket Gamma API. Defaults to " +
					"\"https://gamma-api.polymarket.com\". May also be set with the " +
					"POLYMARKET_ENDPOINT environment variable. Override this to target a " +
					"proxy or a mock server in tests.",
				MarkdownDescription: "Base URL of the Polymarket Gamma API. Defaults to " +
					"`https://gamma-api.polymarket.com`. May also be set with the " +
					"`POLYMARKET_ENDPOINT` environment variable. Override this to target a " +
					"proxy or a mock server in tests.",
			},
			"clob_endpoint": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the Polymarket CLOB (order book) API. Defaults to " +
					"\"https://clob.polymarket.com\". May also be set with the " +
					"POLYMARKET_CLOB_ENDPOINT environment variable.",
				MarkdownDescription: "Base URL of the Polymarket CLOB (order book) API. Defaults " +
					"to `https://clob.polymarket.com`. May also be set with the " +
					"`POLYMARKET_CLOB_ENDPOINT` environment variable.",
			},
			"data_endpoint": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the Polymarket Data API (positions, trades, holders, " +
					"portfolio value). Defaults to \"https://data-api.polymarket.com\". May also " +
					"be set with the POLYMARKET_DATA_ENDPOINT environment variable.",
				MarkdownDescription: "Base URL of the Polymarket Data API (positions, trades, " +
					"holders, portfolio value). Defaults to `https://data-api.polymarket.com`. May " +
					"also be set with the `POLYMARKET_DATA_ENDPOINT` environment variable.",
			},
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "Optional API key forwarded as a bearer token. The public " +
					"Gamma endpoints do not require authentication, so this is only needed " +
					"for authenticated or rate-lifted access. May also be set with the " +
					"POLYMARKET_API_KEY environment variable.",
				MarkdownDescription: "Optional API key forwarded as a bearer token. The public " +
					"Gamma endpoints do not require authentication, so this is only needed " +
					"for authenticated or rate-lifted access. May also be set with the " +
					"`POLYMARKET_API_KEY` environment variable.",
			},
			"private_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "Polygon wallet private key (hex, with or without 0x) used to " +
					"EIP-712 sign authenticated CLOB requests and orders. Only required for " +
					"authenticated/trading use; read-only data sources work without it. May " +
					"also be set with the POLYMARKET_PRIVATE_KEY environment variable.",
				MarkdownDescription: "Polygon wallet private key (hex, with or without `0x`) used " +
					"to EIP-712 sign authenticated CLOB requests and orders. Only required for " +
					"authenticated/trading use; read-only data sources work without it. May also " +
					"be set with the `POLYMARKET_PRIVATE_KEY` environment variable.",
			},
			"funder_address": schema.StringAttribute{
				Optional: true,
				Description: "Address of the wallet that holds USDC and funds orders. For " +
					"email/magic or browser proxy accounts this differs from the signing key's " +
					"address; for a plain EOA it defaults to the signing address. May also be " +
					"set with the POLYMARKET_FUNDER environment variable.",
				MarkdownDescription: "Address of the wallet that holds USDC and funds orders. For " +
					"email/magic or browser proxy accounts this differs from the signing key's " +
					"address; for a plain EOA it defaults to the signing address. May also be set " +
					"with the `POLYMARKET_FUNDER` environment variable.",
			},
			"signature_type": schema.Int64Attribute{
				Optional: true,
				Description: "Polymarket signature type: 0 = EOA (default), 1 = email/magic " +
					"proxy, 2 = browser/Gnosis proxy. May also be set with the " +
					"POLYMARKET_SIGNATURE_TYPE environment variable.",
				MarkdownDescription: "Polymarket signature type: `0` = EOA (default), `1` = " +
					"email/magic proxy, `2` = browser/Gnosis proxy. May also be set with the " +
					"`POLYMARKET_SIGNATURE_TYPE` environment variable.",
			},
			"chain_id": schema.Int64Attribute{
				Optional: true,
				Description: "EVM chain ID used when signing. Defaults to 137 (Polygon mainnet); " +
					"use 80002 for the Amoy testnet. May also be set with the " +
					"POLYMARKET_CHAIN_ID environment variable.",
				MarkdownDescription: "EVM chain ID used when signing. Defaults to `137` (Polygon " +
					"mainnet); use `80002` for the Amoy testnet. May also be set with the " +
					"`POLYMARKET_CHAIN_ID` environment variable.",
			},
			"rpc_endpoint": schema.StringAttribute{
				Optional: true,
				Description: "Polygon JSON-RPC endpoint URL used by the polymarket_allowance " +
					"resource to read and submit on-chain approvals. Required only for managing " +
					"allowances. May also be set with the POLYMARKET_RPC_ENDPOINT environment variable.",
				MarkdownDescription: "Polygon JSON-RPC endpoint URL used by the " +
					"`polymarket_allowance` resource to read and submit on-chain approvals. " +
					"Required only for managing allowances. May also be set with the " +
					"`POLYMARKET_RPC_ENDPOINT` environment variable.",
			},
		},
	}
}

// Configure builds the shared API client from provider configuration and
// environment variables, then hands it to data sources and resources.
func (p *polymarketProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config polymarketProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration value precedence: explicit config > environment > default.
	endpoint := firstNonEmpty(config.Endpoint, "POLYMARKET_ENDPOINT", client.DefaultEndpoint)
	clobEndpoint := firstNonEmpty(config.ClobEndpoint, "POLYMARKET_CLOB_ENDPOINT", client.DefaultClobEndpoint)
	dataEndpoint := firstNonEmpty(config.DataEndpoint, "POLYMARKET_DATA_ENDPOINT", client.DefaultDataEndpoint)
	apiKey := firstNonEmpty(config.APIKey, "POLYMARKET_API_KEY", "")

	opts := []client.Option{
		client.WithEndpoint(endpoint),
		client.WithClobEndpoint(clobEndpoint),
		client.WithDataEndpoint(dataEndpoint),
		client.WithAPIKey(apiKey),
	}

	// Build a signer only when a private key is supplied; read-only data sources
	// work without one.
	if privKey := firstNonEmpty(config.PrivateKey, "POLYMARKET_PRIVATE_KEY", ""); privKey != "" {
		chainID := firstNonZeroInt(config.ChainID, "POLYMARKET_CHAIN_ID", 137)
		sigType := firstNonZeroInt(config.SignatureType, "POLYMARKET_SIGNATURE_TYPE", 0)
		funder := firstNonEmpty(config.FunderAddress, "POLYMARKET_FUNDER", "")

		signer, err := sign.NewSigner(privKey, chainID, uint8(sigType), funder)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("private_key"),
				"Invalid Polymarket signing configuration",
				err.Error(),
			)
			return
		}
		opts = append(opts, client.WithSigner(signer))

		// Build the on-chain client only when an RPC endpoint is also supplied;
		// it is needed solely for the polymarket_allowance resource.
		if rpcEndpoint := firstNonEmpty(config.RPCEndpoint, "POLYMARKET_RPC_ENDPOINT", ""); rpcEndpoint != "" {
			chainClient, err := chain.New(ctx, rpcEndpoint, privKey, chainID)
			if err != nil {
				resp.Diagnostics.AddAttributeError(
					path.Root("rpc_endpoint"),
					"Invalid Polymarket on-chain configuration",
					err.Error(),
				)
				return
			}
			opts = append(opts, client.WithChain(chainClient))
		}
	}

	c := client.New(opts...)

	// Make the client available to data sources and resources.
	resp.DataSourceData = c
	resp.ResourceData = c
}

// DataSources registers the provider's data sources.
func (p *polymarketProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMarketDataSource,
		NewMarketsDataSource,
		NewEventDataSource,
		NewEventsDataSource,
		NewSeriesDataSource,
		NewTagsDataSource,
		NewOrderBookDataSource,
		NewOrderBooksDataSource,
		NewPriceDataSource,
		NewPricesDataSource,
		NewMidpointDataSource,
		NewSpreadDataSource,
		NewPriceHistoryDataSource,
		NewPositionsDataSource,
		NewTradesDataSource,
		NewPortfolioValueDataSource,
		NewHoldersDataSource,
		NewAPICredentialsDataSource,
	}
}

// Resources registers the provider's managed resources. These require the
// provider's private_key to be configured.
func (p *polymarketProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrderResource,
		NewAPIKeyResource,
		NewAllowanceResource,
	}
}
