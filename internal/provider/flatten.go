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
	out.Active = types.BoolValue(m.Active)
	out.Closed = types.BoolValue(m.Closed)
	out.Archived = types.BoolValue(m.Archived)
	out.Liquidity = types.StringValue(m.Liquidity)
	out.Volume = types.StringValue(m.Volume)
	out.StartDate = types.StringValue(m.StartDate)
	out.EndDate = types.StringValue(m.EndDate)
	out.ConditionID = types.StringValue(m.ConditionID)

	outcomes, d := types.ListValueFrom(ctx, types.StringType, decodeStringArray(m.Outcomes))
	diags.Append(d...)
	out.Outcomes = outcomes

	prices, d := types.ListValueFrom(ctx, types.StringType, decodeStringArray(m.OutcomePrices))
	diags.Append(d...)
	out.OutcomePrices = prices

	return diags
}
