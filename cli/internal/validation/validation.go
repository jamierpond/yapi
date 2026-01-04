package validation

import (
	"fmt"

	"yapi.run/cli/internal/constants"
	"yapi.run/cli/internal/domain"
)

// Severity indicates the level of a validation issue.
type Severity int

// Severity levels for validation issues.
const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	default:
		return "info"
	}
}

// Issue represents a single validation problem.
type Issue struct {
	Severity Severity
	Field    string // e.g. "url", "method", "service"
	Message  string // human-readable
}

// isGRPCRequest returns true if this is a gRPC request
func isGRPCRequest(req *domain.Request) bool {
	return req.Metadata["transport"] == constants.TransportGRPC
}

// isTCPRequest returns true if this is a TCP request
func isTCPRequest(req *domain.Request) bool {
	return req.Metadata["transport"] == constants.TransportTCP
}

// isHTTPRequest returns true if this is an HTTP request
func isHTTPRequest(req *domain.Request) bool {
	t := req.Metadata["transport"]
	return t == constants.TransportHTTP || t == constants.TransportGraphQL
}

// ValidateRequest performs semantic validation on a domain.Request.
func ValidateRequest(req *domain.Request) []Issue {
	var issues []Issue
	add := func(sev Severity, field, msg string) {
		issues = append(issues, Issue{Severity: sev, Field: field, Message: msg})
	}

	if req.URL == "" {
		add(SeverityError, "url", "missing required field `url`")
	}

	method := constants.CanonicalizeMethod(req.Method)
	if isHTTPRequest(req) && method != "" && !constants.ValidHTTPMethods[method] {
		add(SeverityWarning, "method", fmt.Sprintf("unknown HTTP method `%s`", req.Method))
	}

	if isGRPCRequest(req) {
		if req.Metadata["service"] == "" {
			add(SeverityError, "service", "gRPC config requires `service`")
		}
		if req.Metadata["rpc"] == "" {
			add(SeverityError, "rpc", "gRPC config requires `rpc`")
		}
	}

	if isTCPRequest(req) && req.Metadata["encoding"] != "" && !validEncoding(req.Metadata["encoding"]) {
		add(SeverityError, "encoding",
			fmt.Sprintf("unsupported TCP encoding `%s` (allowed: text, hex, base64)", req.Metadata["encoding"]))
	}

	hasBody := req.Body != nil
	if req.Metadata["graphql_query"] != "" && hasBody {
		field := "body"
		if req.Metadata["body_source"] == "json" {
			field = "json"
		}
		add(SeverityError, field, "`graphql` cannot be used with `body` or `json`")
	}

	return issues
}

func validEncoding(enc string) bool {
	switch enc {
	case "text", "hex", "base64":
		return true
	default:
		return false
	}
}
