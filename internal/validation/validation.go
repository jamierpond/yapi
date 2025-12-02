package validation

import (
	"fmt"
	"strings"

	"yapi.run/cli/internal/domain"
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

// isGRPCRequest returns true if this is a gRPC request
func isGRPCRequest(req *domain.Request) bool {
	return req.Metadata["transport"] == "grpc"
}

// isTCPRequest returns true if this is a TCP request
func isTCPRequest(req *domain.Request) bool {
	return req.Metadata["transport"] == "tcp"
}

// isHTTPRequest returns true if this is an HTTP request
func isHTTPRequest(req *domain.Request) bool {
	return req.Metadata["transport"] == "http" || req.Metadata["transport"] == "graphql"
}

// ValidateRequest performs semantic validation on a domain.Request.
func ValidateRequest(req *domain.Request) []Issue {
	var issues []Issue

	// Rule 1: url is required
	if req.URL == "" {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Field:    "url",
			Message:  "missing required field `url`",
		})
	}

	// Rule 2: method validation
	method := strings.ToUpper(req.Method)
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

	if !validHTTPMethods[method] && isHTTPRequest(req) {
		issues = append(issues, Issue{
			Severity: SeverityWarning,
			Field:    "method",
			Message:  fmt.Sprintf("unknown HTTP method `%s`", req.Method),
		})
	}

	// Rule 3: gRPC requires service and rpc
	if isGRPCRequest(req) {
		if req.Metadata["service"] == "" {
			issues = append(issues, Issue{
				Severity: SeverityError,
				Field:    "service",
				Message:  "gRPC config requires `service`",
			})
		}
		if req.Metadata["rpc"] == "" {
			issues = append(issues, Issue{
				Severity: SeverityError,
				Field:    "rpc",
				Message:  "gRPC config requires `rpc`",
			})
		}
	}

	// Rule 4: TCP encoding validation
	if isTCPRequest(req) && req.Metadata["encoding"] != "" {
		validEncodings := map[string]bool{
			"text":   true,
			"hex":    true,
			"base64": true,
		}
		if !validEncodings[req.Metadata["encoding"]] {
			issues = append(issues, Issue{
				Severity: SeverityError,
				Field:    "encoding",
				Message:  fmt.Sprintf("unsupported TCP encoding `%s` (allowed: text, hex, base64)", req.Metadata["encoding"]),
			})
		}
	}

	hasBody := req.Body != nil
	hasJSON := req.Metadata["body_source"] == "json"
	hasGraphql := req.Metadata["graphql_query"] != ""

	// Rule 6: graphql is mutually exclusive with body/json
	if hasGraphql && hasBody {
		field := "body"
		if hasJSON {
			field = "json"
		}
		issues = append(issues, Issue{
			Severity: SeverityError,
			Field:    field,
			Message:  "`graphql` cannot be used with `body` or `json`",
		})
	}

	return issues
}
