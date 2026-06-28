// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// marketDataSourceModel maps every market attribute to a Go type. It is shared
// by the single-market data source and as the element type of the markets list,
// so the two surfaces never drift apart.
type marketDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Question         types.String `tfsdk:"question"`
	Slug             types.String `tfsdk:"slug"`
	Description      types.String `tfsdk:"description"`
	ResolutionSource types.String `tfsdk:"resolution_source"`
	Active           types.Bool   `tfsdk:"active"`
	Closed           types.Bool   `tfsdk:"closed"`
	Archived         types.Bool   `tfsdk:"archived"`
	Liquidity        types.String `tfsdk:"liquidity"`
	Volume           types.String `tfsdk:"volume"`
	StartDate        types.String `tfsdk:"start_date"`
	EndDate          types.String `tfsdk:"end_date"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	ConditionID      types.String `tfsdk:"condition_id"`
	QuestionID       types.String `tfsdk:"question_id"`
	GroupItemTitle   types.String `tfsdk:"group_item_title"`
	Image            types.String `tfsdk:"image"`
	Icon             types.String `tfsdk:"icon"`
	Outcomes         types.List   `tfsdk:"outcomes"`
	OutcomePrices    types.List   `tfsdk:"outcome_prices"`
	ClobTokenIDs     types.List   `tfsdk:"clob_token_ids"`

	Volume24hr            types.Float64 `tfsdk:"volume_24hr"`
	Volume1wk             types.Float64 `tfsdk:"volume_1wk"`
	Volume1mo             types.Float64 `tfsdk:"volume_1mo"`
	Volume1yr             types.Float64 `tfsdk:"volume_1yr"`
	Spread                types.Float64 `tfsdk:"spread"`
	BestBid               types.Float64 `tfsdk:"best_bid"`
	BestAsk               types.Float64 `tfsdk:"best_ask"`
	LastTradePrice        types.Float64 `tfsdk:"last_trade_price"`
	Competitive           types.Float64 `tfsdk:"competitive"`
	OrderMinSize          types.Float64 `tfsdk:"order_min_size"`
	OrderPriceMinTickSize types.Float64 `tfsdk:"order_price_min_tick_size"`

	EnableOrderBook types.Bool `tfsdk:"enable_order_book"`
	AcceptingOrders types.Bool `tfsdk:"accepting_orders"`
}

// marketAttributes returns the schema for a market. When idRequired is true the
// "id" attribute is a required input (single-market lookup); otherwise every
// attribute is computed (a market nested inside a list result).
func marketAttributes(idRequired bool) map[string]schema.Attribute {
	idAttr := schema.StringAttribute{
		Description: "Numeric identifier of the market, as assigned by Polymarket " +
			"(for example \"253123\").",
		MarkdownDescription: "Numeric identifier of the market, as assigned by Polymarket " +
			"(for example `253123`).",
	}
	if idRequired {
		idAttr.Required = true
		idAttr.Description = "Numeric identifier of the market to fetch, as assigned by " +
			"Polymarket (for example \"253123\"). This is the only required input."
		idAttr.MarkdownDescription = "Numeric identifier of the market to fetch, as assigned " +
			"by Polymarket (for example `253123`). This is the only required input."
	} else {
		idAttr.Computed = true
	}

	return map[string]schema.Attribute{
		"id": idAttr,
		"question": schema.StringAttribute{
			Computed:            true,
			Description:         "Human-readable question the market resolves, e.g. \"Will candidate X win?\".",
			MarkdownDescription: "Human-readable question the market resolves, e.g. `Will candidate X win?`.",
		},
		"slug": schema.StringAttribute{
			Computed:            true,
			Description:         "URL-friendly slug identifying the market on polymarket.com.",
			MarkdownDescription: "URL-friendly slug identifying the market on polymarket.com.",
		},
		"description": schema.StringAttribute{
			Computed:            true,
			Description:         "Long-form description and resolution criteria explaining how the market settles.",
			MarkdownDescription: "Long-form description and resolution criteria explaining how the market settles.",
		},
		"resolution_source": schema.StringAttribute{
			Computed:            true,
			Description:         "Source of truth consulted to resolve the market, when specified.",
			MarkdownDescription: "Source of truth consulted to resolve the market, when specified.",
		},
		"active": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the market is currently active and accepting trades.",
			MarkdownDescription: "Whether the market is currently active and accepting trades.",
		},
		"closed": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the market has closed and no longer accepts trades.",
			MarkdownDescription: "Whether the market has closed and no longer accepts trades.",
		},
		"archived": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the market has been archived and hidden from the default UI.",
			MarkdownDescription: "Whether the market has been archived and hidden from the default UI.",
		},
		"liquidity": schema.StringAttribute{
			Computed:            true,
			Description:         "Total liquidity available in the market, in USDC, as a decimal string.",
			MarkdownDescription: "Total liquidity available in the market, in USDC, as a decimal string.",
		},
		"volume": schema.StringAttribute{
			Computed:            true,
			Description:         "Cumulative trading volume of the market, in USDC, as a decimal string.",
			MarkdownDescription: "Cumulative trading volume of the market, in USDC, as a decimal string.",
		},
		"start_date": schema.StringAttribute{
			Computed:            true,
			Description:         "ISO-8601 timestamp at which the market opened for trading.",
			MarkdownDescription: "ISO-8601 timestamp at which the market opened for trading.",
		},
		"end_date": schema.StringAttribute{
			Computed:            true,
			Description:         "ISO-8601 timestamp at which the market is scheduled to close.",
			MarkdownDescription: "ISO-8601 timestamp at which the market is scheduled to close.",
		},
		"created_at": schema.StringAttribute{
			Computed:            true,
			Description:         "ISO-8601 timestamp at which the market record was created.",
			MarkdownDescription: "ISO-8601 timestamp at which the market record was created.",
		},
		"updated_at": schema.StringAttribute{
			Computed:            true,
			Description:         "ISO-8601 timestamp at which the market record was last updated.",
			MarkdownDescription: "ISO-8601 timestamp at which the market record was last updated.",
		},
		"condition_id": schema.StringAttribute{
			Computed:            true,
			Description:         "On-chain condition ID (a 0x-prefixed hash) linking the market to its CTF contract.",
			MarkdownDescription: "On-chain condition ID (a `0x`-prefixed hash) linking the market to its CTF contract.",
		},
		"question_id": schema.StringAttribute{
			Computed:            true,
			Description:         "On-chain question ID (a 0x-prefixed hash) used by the UMA oracle to resolve the market.",
			MarkdownDescription: "On-chain question ID (a `0x`-prefixed hash) used by the UMA oracle to resolve the market.",
		},
		"group_item_title": schema.StringAttribute{
			Computed:            true,
			Description:         "Short label distinguishing this market within its parent event group, when grouped.",
			MarkdownDescription: "Short label distinguishing this market within its parent event group, when grouped.",
		},
		"image": schema.StringAttribute{
			Computed:            true,
			Description:         "URL of the market's banner image.",
			MarkdownDescription: "URL of the market's banner image.",
		},
		"icon": schema.StringAttribute{
			Computed:            true,
			Description:         "URL of the market's icon image.",
			MarkdownDescription: "URL of the market's icon image.",
		},
		"outcomes": schema.ListAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			Description:         "Ordered list of outcome names, e.g. [\"Yes\", \"No\"]; aligns positionally with outcome_prices and clob_token_ids.",
			MarkdownDescription: "Ordered list of outcome names, e.g. `[\"Yes\", \"No\"]`; aligns positionally with `outcome_prices` and `clob_token_ids`.",
		},
		"outcome_prices": schema.ListAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			Description:         "Current price (implied probability in [0, 1]) of each outcome as a decimal string; aligns with outcomes.",
			MarkdownDescription: "Current price (implied probability in `[0, 1]`) of each outcome as a decimal string; aligns with `outcomes`.",
		},
		"clob_token_ids": schema.ListAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			Description:         "CLOB ERC-1155 token IDs for each outcome, used to query the order book; aligns with outcomes.",
			MarkdownDescription: "CLOB ERC-1155 token IDs for each outcome, used to query the order book; aligns with `outcomes`.",
		},
		"volume_24hr": schema.Float64Attribute{
			Computed:            true,
			Description:         "Trading volume over the trailing 24 hours, in USDC.",
			MarkdownDescription: "Trading volume over the trailing 24 hours, in USDC.",
		},
		"volume_1wk": schema.Float64Attribute{
			Computed:            true,
			Description:         "Trading volume over the trailing week, in USDC.",
			MarkdownDescription: "Trading volume over the trailing week, in USDC.",
		},
		"volume_1mo": schema.Float64Attribute{
			Computed:            true,
			Description:         "Trading volume over the trailing month, in USDC.",
			MarkdownDescription: "Trading volume over the trailing month, in USDC.",
		},
		"volume_1yr": schema.Float64Attribute{
			Computed:            true,
			Description:         "Trading volume over the trailing year, in USDC.",
			MarkdownDescription: "Trading volume over the trailing year, in USDC.",
		},
		"spread": schema.Float64Attribute{
			Computed:            true,
			Description:         "Current bid-ask spread of the market's primary outcome.",
			MarkdownDescription: "Current bid-ask spread of the market's primary outcome.",
		},
		"best_bid": schema.Float64Attribute{
			Computed:            true,
			Description:         "Current best (highest) bid price for the primary outcome, in [0, 1].",
			MarkdownDescription: "Current best (highest) bid price for the primary outcome, in `[0, 1]`.",
		},
		"best_ask": schema.Float64Attribute{
			Computed:            true,
			Description:         "Current best (lowest) ask price for the primary outcome, in [0, 1].",
			MarkdownDescription: "Current best (lowest) ask price for the primary outcome, in `[0, 1]`.",
		},
		"last_trade_price": schema.Float64Attribute{
			Computed:            true,
			Description:         "Price at which the primary outcome most recently traded, in [0, 1].",
			MarkdownDescription: "Price at which the primary outcome most recently traded, in `[0, 1]`.",
		},
		"competitive": schema.Float64Attribute{
			Computed:            true,
			Description:         "Polymarket's competitiveness score for the market, in [0, 1]; higher is more liquid/tighter.",
			MarkdownDescription: "Polymarket's competitiveness score for the market, in `[0, 1]`; higher is more liquid/tighter.",
		},
		"order_min_size": schema.Float64Attribute{
			Computed:            true,
			Description:         "Minimum order size accepted by the order book, in shares.",
			MarkdownDescription: "Minimum order size accepted by the order book, in shares.",
		},
		"order_price_min_tick_size": schema.Float64Attribute{
			Computed:            true,
			Description:         "Smallest price increment accepted by the order book (e.g. 0.01).",
			MarkdownDescription: "Smallest price increment accepted by the order book (e.g. `0.01`).",
		},
		"enable_order_book": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether a CLOB order book is enabled for this market.",
			MarkdownDescription: "Whether a CLOB order book is enabled for this market.",
		},
		"accepting_orders": schema.BoolAttribute{
			Computed:            true,
			Description:         "Whether the market is currently accepting new orders.",
			MarkdownDescription: "Whether the market is currently accepting new orders.",
		},
	}
}
