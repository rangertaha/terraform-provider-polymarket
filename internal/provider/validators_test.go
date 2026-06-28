// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestOpenUnitIntervalValidator(t *testing.T) {
	cases := []struct {
		name    string
		value   types.Float64
		wantErr bool
	}{
		{"mid", types.Float64Value(0.5), false},
		{"near zero", types.Float64Value(0.01), false},
		{"near one", types.Float64Value(0.99), false},
		{"zero", types.Float64Value(0), true},
		{"one", types.Float64Value(1), true},
		{"negative", types.Float64Value(-0.1), true},
		{"above one", types.Float64Value(1.5), true},
		{"null skipped", types.Float64Null(), false},
		{"unknown skipped", types.Float64Unknown(), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validator.Float64Request{ConfigValue: tc.value}
			resp := &validator.Float64Response{}
			openUnitInterval().ValidateFloat64(context.Background(), req, resp)
			if got := resp.Diagnostics.HasError(); got != tc.wantErr {
				t.Errorf("HasError() = %v, want %v (diags: %v)", got, tc.wantErr, resp.Diagnostics)
			}
		})
	}
}

func TestPositiveFloatValidator(t *testing.T) {
	cases := []struct {
		name    string
		value   types.Float64
		wantErr bool
	}{
		{"positive", types.Float64Value(100), false},
		{"tiny positive", types.Float64Value(0.0001), false},
		{"zero", types.Float64Value(0), true},
		{"negative", types.Float64Value(-5), true},
		{"null skipped", types.Float64Null(), false},
		{"unknown skipped", types.Float64Unknown(), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validator.Float64Request{ConfigValue: tc.value}
			resp := &validator.Float64Response{}
			positiveFloat().ValidateFloat64(context.Background(), req, resp)
			if got := resp.Diagnostics.HasError(); got != tc.wantErr {
				t.Errorf("HasError() = %v, want %v (diags: %v)", got, tc.wantErr, resp.Diagnostics)
			}
		})
	}
}
