package validation

import (
	"fmt"
	"strings"

	"cli/internal/config"
)

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

type Issue struct {
	Severity Severity
	Field    string // e.g. "url", "method", "service"
	Message  string // human-readable
}

// ValidateConfig performs semantic validation on a parsed YapiConfig.
// It must not read files, print, or talk to LSP/CLI. Pure logic only.
func ValidateConfig(cfg *config.YapiConfig) []Issue {
	var issues []Issue

	// Rule 1: url is required
	if cfg.URL == "" {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Field:    "url",
			Message:  "missing required field `url`",
		})
	}

	// Rule 2: method validation
	method := strings.ToUpper(cfg.Method)
	validHTTPMethods := map[string]bool{
		"":        true, // empty is allowed, defaults to GET
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}

	if cfg.Method == "grpc" {
		issues = append(issues, Issue{
			Severity: SeverityWarning,
			Field:    "method",
			Message:  "`method: grpc` is deprecated, use a dedicated transport field instead",
		})
	} else if cfg.Method == "tcp" {
		issues = append(issues, Issue{
			Severity: SeverityWarning,
			Field:    "method",
			Message:  "`method: tcp` is deprecated, use a dedicated transport field instead",
		})
	} else if !validHTTPMethods[method] {
		issues = append(issues, Issue{
			Severity: SeverityWarning,
			Field:    "method",
			Message:  fmt.Sprintf("unknown HTTP method `%s`", cfg.Method),
		})
	}

	// Rule 3: gRPC requires service and rpc
	if cfg.Method == "grpc" {
		if cfg.Service == "" {
			issues = append(issues, Issue{
				Severity: SeverityError,
				Field:    "service",
				Message:  "gRPC config requires `service`",
			})
		}
		if cfg.RPC == "" {
			issues = append(issues, Issue{
				Severity: SeverityError,
				Field:    "rpc",
				Message:  "gRPC config requires `rpc`",
			})
		}
	}

	// Rule 4: TCP encoding validation
	if cfg.Method == "tcp" && cfg.Encoding != "" {
		validEncodings := map[string]bool{
			"text":   true,
			"hex":    true,
			"base64": true,
		}
		if !validEncodings[cfg.Encoding] {
			issues = append(issues, Issue{
				Severity: SeverityError,
				Field:    "encoding",
				Message:  fmt.Sprintf("unsupported TCP encoding `%s` (allowed: text, hex, base64)", cfg.Encoding),
			})
		}
	}

	return issues
}
