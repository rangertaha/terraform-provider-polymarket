// Copyright (c) Rangertaha
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"os"
	"strconv"

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

// firstNonZeroInt resolves an int64 config value using the precedence
// explicit-config > environment-variable > fallback. A null or zero config
// value falls through; an unparseable env var is ignored.
func firstNonZeroInt(configValue types.Int64, envVar string, fallback int64) int64 {
	if !configValue.IsNull() && configValue.ValueInt64() != 0 {
		return configValue.ValueInt64()
	}
	if v := os.Getenv(envVar); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
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
