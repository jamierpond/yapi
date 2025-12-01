package validation

import (
	"fmt"
	"strings"

	"yapi.run/cli/internal/config"
)

// URL scheme constants
const (
	schemeHTTP  = "http://"
	schemeHTTPS = "https://"
	schemeGRPC  = "grpc://"
	schemeGRPCS = "grpcs://"
	schemeTCP   = "tcp://"
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

// getURLScheme returns the protocol scheme from a URL
func getURLScheme(url string) string {
	urlLower := strings.ToLower(url)
	switch {
	case strings.HasPrefix(urlLower, schemeGRPCS):
		return "grpcs"
	case strings.HasPrefix(urlLower, schemeGRPC):
		return "grpc"
	case strings.HasPrefix(urlLower, schemeTCP):
		return "tcp"
	case strings.HasPrefix(urlLower, schemeHTTPS):
		return "https"
	case strings.HasPrefix(urlLower, schemeHTTP):
		return "http"
	default:
		return ""
	}
}

// isGRPCRequest returns true if this is a gRPC request (by method or URL scheme)
func isGRPCRequest(cfg *config.YapiConfig) bool {
	scheme := getURLScheme(cfg.URL)
	return cfg.Method == "grpc" || scheme == "grpc" || scheme == "grpcs"
}

// isTCPRequest returns true if this is a TCP request (by method or URL scheme)
func isTCPRequest(cfg *config.YapiConfig) bool {
	scheme := getURLScheme(cfg.URL)
	return cfg.Method == "tcp" || scheme == "tcp"
}

// isHTTPRequest returns true if this is an HTTP request
func isHTTPRequest(cfg *config.YapiConfig) bool {
	return !isGRPCRequest(cfg) && !isTCPRequest(cfg)
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

	// Rule 3: gRPC requires service and rpc (by method or URL scheme)
	if isGRPCRequest(cfg) {
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
	if isTCPRequest(cfg) && cfg.Encoding != "" {
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

	// Rule 5: body and json are mutually exclusive
	hasBody := cfg.Body != nil && len(cfg.Body) > 0
	hasJSON := cfg.JSON != ""
	hasGraphql := cfg.Graphql != ""

	if hasBody && hasJSON {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Field:    "body",
			Message:  "`body` and `json` are mutually exclusive",
		})
	}

	// Rule 6: graphql is mutually exclusive with body and json
	if hasGraphql && (hasBody || hasJSON) {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Field:    "graphql",
			Message:  "`graphql` cannot be used with `body` or `json`",
		})
	}

	// Rule 7: content_type required when body or json is present (HTTP only, not GraphQL)
	if isHTTPRequest(cfg) && !hasGraphql && (hasBody || hasJSON) && cfg.ContentType == "" {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Field:    "content_type",
			Message:  "`content_type` is required when `body` or `json` is present",
		})
	}

	return issues
}
