// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// hash32Pattern matches a 0x-prefixed 32-byte hex hash (a condition ID).
var hash32Pattern = regexp.MustCompile("^0x[0-9a-fA-F]{64}$")

// openUnitIntervalValidator requires a float in the open interval (0, 1) — the
// valid range for a Polymarket price (0 and 1 are excluded).
type openUnitIntervalValidator struct{}

func (v openUnitIntervalValidator) Description(_ context.Context) string {
	return "value must be greater than 0 and less than 1"
}

func (v openUnitIntervalValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v openUnitIntervalValidator) ValidateFloat64(_ context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueFloat64()
	if val <= 0 || val >= 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid price",
			fmt.Sprintf("price must be in the open interval (0, 1), got %v", val),
		)
	}
}

// openUnitInterval validates that a price lies strictly between 0 and 1.
func openUnitInterval() validator.Float64 { return openUnitIntervalValidator{} }

// positiveFloatValidator requires a strictly positive float (the built-in
// AtLeast validator is inclusive, so it cannot express "> 0").
type positiveFloatValidator struct{}

func (v positiveFloatValidator) Description(_ context.Context) string {
	return "value must be greater than 0"
}

func (v positiveFloatValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v positiveFloatValidator) ValidateFloat64(_ context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if val := req.ConfigValue.ValueFloat64(); val <= 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value",
			fmt.Sprintf("value must be greater than 0, got %v", val),
		)
	}
}

// positiveFloat validates that a float is strictly greater than 0.
func positiveFloat() validator.Float64 { return positiveFloatValidator{} }

// ethAddressValidator requires a valid 0x-prefixed 20-byte hex Ethereum address.
type ethAddressValidator struct{}

func (v ethAddressValidator) Description(_ context.Context) string {
	return "value must be a 0x-prefixed 20-byte hex Ethereum address"
}

func (v ethAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ethAddressValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	if val == "" {
		return
	}
	if !common.IsHexAddress(val) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Ethereum address",
			fmt.Sprintf("value must be a 0x-prefixed 20-byte hex Ethereum address, got %q", val),
		)
	}
}

// ethAddress validates that a string is a well-formed Ethereum address.
func ethAddress() validator.String { return ethAddressValidator{} }

// conditionIDValidator requires a 0x-prefixed 32-byte hex hash, the form of a
// Polymarket market condition ID.
type conditionIDValidator struct{}

func (v conditionIDValidator) Description(_ context.Context) string {
	return "value must be a 0x-prefixed 32-byte hex hash (66 characters)"
}

func (v conditionIDValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v conditionIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	if val == "" {
		return
	}
	if !hash32Pattern.MatchString(val) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid condition ID",
			fmt.Sprintf("value must be a 0x-prefixed 32-byte hex hash (66 characters), got %q", val),
		)
	}
}

// conditionID validates that a string is a well-formed market condition ID.
func conditionID() validator.String { return conditionIDValidator{} }
