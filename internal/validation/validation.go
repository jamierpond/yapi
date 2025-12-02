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
	add := func(sev Severity, field, msg string) {
		issues = append(issues, Issue{Severity: sev, Field: field, Message: msg})
	}

	if req.URL == "" {
		add(SeverityError, "url", "missing required field `url`")
	}

	method := strings.ToUpper(req.Method)
	if isHTTPRequest(req) && !validHTTPMethod(method) {
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

func validHTTPMethod(m string) bool {
	switch m {
	case "", "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

func validEncoding(enc string) bool {
	switch enc {
	case "text", "hex", "base64":
		return true
	default:
		return false
	}
}
