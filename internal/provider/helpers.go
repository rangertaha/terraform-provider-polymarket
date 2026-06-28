// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// firstNonEmpty resolves a configuration value using the precedence
// explicit-config > environment-variable > fallback.
func firstNonEmpty(configValue types.String, envVar, fallback string) string {
	if !configValue.IsNull() && configValue.ValueString() != "" {
		return configValue.ValueString()
	}
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	return fallback
}

// decodeStringArray parses a Polymarket JSON-encoded string array (e.g.
// "[\"Yes\",\"No\"]") into a slice. It returns nil for empty or invalid input
// so callers degrade gracefully rather than erroring on malformed upstream data.
func decodeStringArray(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}
