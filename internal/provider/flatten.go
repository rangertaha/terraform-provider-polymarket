// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/Rangertaha/terraform-provider-polymarket/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// flattenMarket maps an API market onto a marketDataSourceModel, decoding the
// JSON-encoded outcome arrays into Terraform lists.
func flattenMarket(ctx context.Context, m client.Market, out *marketDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	out.ID = types.StringValue(m.ID)
	out.Question = types.StringValue(m.Question)
	out.Slug = types.StringValue(m.Slug)
	out.Description = types.StringValue(m.Description)
	out.ResolutionSource = types.StringValue(m.ResolutionSource)
	out.Active = types.BoolValue(m.Active)
	out.Closed = types.BoolValue(m.Closed)
	out.Archived = types.BoolValue(m.Archived)
	out.Liquidity = types.StringValue(m.Liquidity)
	out.Volume = types.StringValue(m.Volume)
	out.StartDate = types.StringValue(m.StartDate)
	out.EndDate = types.StringValue(m.EndDate)
	out.CreatedAt = types.StringValue(m.CreatedAt)
	out.UpdatedAt = types.StringValue(m.UpdatedAt)
	out.ConditionID = types.StringValue(m.ConditionID)
	out.QuestionID = types.StringValue(m.QuestionID)
	out.GroupItemTitle = types.StringValue(m.GroupItemTitle)
	out.Image = types.StringValue(m.Image)
	out.Icon = types.StringValue(m.Icon)

	out.Volume24hr = types.Float64Value(m.Volume24hr)
	out.Volume1wk = types.Float64Value(m.Volume1wk)
	out.Volume1mo = types.Float64Value(m.Volume1mo)
	out.Volume1yr = types.Float64Value(m.Volume1yr)
	out.Spread = types.Float64Value(m.Spread)
	out.BestBid = types.Float64Value(m.BestBid)
	out.BestAsk = types.Float64Value(m.BestAsk)
	out.LastTradePrice = types.Float64Value(m.LastTradePrice)
	out.Competitive = types.Float64Value(m.Competitive)
	out.OrderMinSize = types.Float64Value(m.OrderMinSize)
	out.OrderPriceMinTickSize = types.Float64Value(m.OrderPriceMinTickSize)

	out.EnableOrderBook = types.BoolValue(m.EnableOrderBook)
	out.AcceptingOrders = types.BoolValue(m.AcceptingOrders)

	outcomes, d := types.ListValueFrom(ctx, types.StringType, decodeStringArray(m.Outcomes))
	diags.Append(d...)
	out.Outcomes = outcomes

	prices, d := types.ListValueFrom(ctx, types.StringType, decodeStringArray(m.OutcomePrices))
	diags.Append(d...)
	out.OutcomePrices = prices

	tokenIDs, d := types.ListValueFrom(ctx, types.StringType, decodeStringArray(m.ClobTokenIDs))
	diags.Append(d...)
	out.ClobTokenIDs = tokenIDs

	return diags
}
