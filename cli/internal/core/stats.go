package core

import (
	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/validation"
	"yapi.run/cli/internal/vars"
)

// ExtractConfigStats extracts feature usage statistics from an analysis result.
// This is used by observability hooks to gather request metadata.
func ExtractConfigStats(analysis *validation.Analysis) map[string]any {
	stats := make(map[string]any)

	if analysis == nil || analysis.Base == nil {
		return stats
	}

	base := analysis.Base

	// Transport detection
	stats["transport"] = detectTransport(base)

	// Chain info
	isChain := len(analysis.Chain) > 0
	stats["is_chain"] = isChain
	stats["chain_step_count"] = len(analysis.Chain)

	// Expectations
	hasExpectations := analysis.Expect.Status != nil || len(analysis.Expect.Assert.Body) > 0 || len(analysis.Expect.Assert.Headers) > 0
	assertionCount := len(analysis.Expect.Assert.Body) + len(analysis.Expect.Assert.Headers)
	hasStatusCheck := analysis.Expect.Status != nil

	// Count expectations across chain steps too
	for _, step := range analysis.Chain {
		if step.Expect.Status != nil || len(step.Expect.Assert.Body) > 0 || len(step.Expect.Assert.Headers) > 0 {
			hasExpectations = true
		}
		assertionCount += len(step.Expect.Assert.Body) + len(step.Expect.Assert.Headers)
		if step.Expect.Status != nil {
			hasStatusCheck = true
		}
	}

	stats["has_expectations"] = hasExpectations
	stats["assertion_count"] = assertionCount
	stats["has_status_check"] = hasStatusCheck

	// Variable usage detection
	usesChainVars := false
	usesEnvVars := false

	for _, s := range collectStrings(base, analysis.Chain) {
		if vars.HasChainVars(s) {
			usesChainVars = true
		}
		if vars.HasEnvVars(s) {
			usesEnvVars = true
		}
	}

	stats["uses_chain_vars"] = usesChainVars
	stats["uses_env_vars"] = usesEnvVars

	return stats
}

// detectTransport determines the transport type from URL scheme
func detectTransport(c *config.ConfigV1) string {
	if c == nil || c.URL == "" {
		return "http"
	}

	url := c.URL
	if len(url) >= 7 && (url[:7] == "grpc://" || (len(url) >= 8 && url[:8] == "grpcs://")) {
		return "grpc"
	}
	if len(url) >= 6 && url[:6] == "tcp://" {
		return "tcp"
	}
	if c.Graphql != "" {
		return "graphql"
	}
	return "http"
}

// collectStrings gathers all string values from the config for variable detection.
func collectStrings(base *config.ConfigV1, chain []config.ChainStep) []string {
	if base == nil {
		return nil
	}

	strs := []string{
		base.URL, base.Path, base.Method, base.ContentType,
		base.JSON, base.Graphql, base.Service, base.RPC,
		base.Proto, base.ProtoPath, base.Data, base.Encoding, base.JQFilter,
		base.Delay,
	}

	for _, v := range base.Headers {
		strs = append(strs, v)
	}
	for _, v := range base.Query {
		strs = append(strs, v)
	}

	strs = append(strs, collectMapStrings(base.Body)...)
	strs = append(strs, collectMapStrings(base.Variables)...)

	// Collect from chain steps
	for _, step := range chain {
		strs = append(strs,
			step.URL, step.Path, step.Method, step.ContentType,
			step.JSON, step.Graphql, step.Service, step.RPC,
			step.Proto, step.ProtoPath, step.Data, step.Encoding, step.JQFilter,
			step.Delay,
		)
		for _, v := range step.Headers {
			strs = append(strs, v)
		}
		for _, v := range step.Query {
			strs = append(strs, v)
		}
		strs = append(strs, collectMapStrings(step.Body)...)
		strs = append(strs, collectMapStrings(step.Variables)...)
	}

	return strs
}

// collectMapStrings recursively extracts string values from a map.
func collectMapStrings(m map[string]any) []string {
	var strs []string
	for _, v := range m {
		switch val := v.(type) {
		case string:
			strs = append(strs, val)
		case map[string]any:
			strs = append(strs, collectMapStrings(val)...)
		case []any:
			for _, elem := range val {
				if s, ok := elem.(string); ok {
					strs = append(strs, s)
				} else if m, ok := elem.(map[string]any); ok {
					strs = append(strs, collectMapStrings(m)...)
				}
			}
		}
	}
	return strs
}
