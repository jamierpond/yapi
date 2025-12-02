package runner

import (
	"fmt"
	"strings"
	"time"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/filter"
	"yapi.run/cli/internal/output"
	"yapi.run/cli/internal/validation"
)

// Result holds the output of a yapi execution
type Result struct {
	Body        string
	ContentType string
	Warnings    []string
	RequestURL  string        // The full constructed URL (HTTP/GraphQL only)
	Duration    time.Duration // Time taken for the request
	BodyLines   int
	BodyChars   int
	BodyBytes   int
}

// Options for execution
type Options struct {
	URLOverride string
	NoColor     bool
}

// Run executes a yapi config and returns the result.
// This is the single source of truth for config execution.
func Run(cfg *config.YapiConfig, opts Options) (*Result, error) {
	// Validate
	issues := validation.ValidateConfig(cfg)
	var warnings []string
	for _, issue := range issues {
		if issue.Severity == validation.SeverityError {
			if issue.Field != "" {
				return nil, fmt.Errorf("%s: %s", issue.Field, issue.Message)
			}
			return nil, fmt.Errorf("%s", issue.Message)
		}
		if issue.Severity == validation.SeverityWarning {
			if issue.Field != "" {
				warnings = append(warnings, fmt.Sprintf("[WARN] %s: %s", issue.Field, issue.Message))
			} else {
				warnings = append(warnings, fmt.Sprintf("[WARN] %s", issue.Message))
			}
		}
	}

	// Apply URL override
	if opts.URLOverride != "" {
		cfg.URL = opts.URLOverride
	}

	// Substitute environment variables
	cfg.SubstituteEnvVars()

	// Execute based on transport
	body, ctype, requestURL, duration, err := execute(cfg)
	if err != nil {
		return nil, err
	}

	// Apply JQ filter if specified
	if cfg.JQFilter != "" {
		body, err = filter.ApplyJQ(body, cfg.JQFilter)
		if err != nil {
			return nil, fmt.Errorf("jq filter failed: %w", err)
		}
		ctype = "application/json"
	}

	bodyLines := strings.Count(body, "\n") + 1
	bodyChars := len(body)
	bodyBytes := len([]byte(body))

	return &Result{
		Body:        body,
		ContentType: ctype,
		Warnings:    warnings,
		RequestURL:  requestURL,
		Duration:    duration,
		BodyLines:   bodyLines,
		BodyChars:   bodyChars,
		BodyBytes:   bodyBytes,
	}, nil
}

// RunAndFormat executes and returns highlighted output plus Result metadata
func RunAndFormat(cfg *config.YapiConfig, opts Options) (string, *Result, error) {
	result, err := Run(cfg, opts)
	if err != nil {
		return "", nil, err
	}

	out := output.Highlight(result.Body, result.ContentType, opts.NoColor)
	if len(result.Warnings) > 0 {
		out = strings.Join(result.Warnings, "\n") + "\n\n" + out
	}

	return out, result, nil
}

// execute dispatches to the appropriate executor based on config
func execute(cfg *config.YapiConfig) (body, ctype, requestURL string, duration time.Duration, err error) {
	transport := detectTransport(cfg)

	start := time.Now()

	switch transport {
	case "graphql":
		var resp *executor.HTTPResponse
		resp, err = executor.NewGraphQLExecutor().Execute(cfg)
		if err == nil {
			body, ctype, requestURL = resp.Body, resp.ContentType, resp.RequestURL
		}
	case "grpc":
		body, err = executor.NewGRPCExecutor().Execute(cfg)
		ctype = "application/json"
	case "tcp":
		body, err = executor.NewTCPExecutor().Execute(cfg)
		ctype = "text/plain"
	case "http":
		if cfg.Method == "" {
			cfg.Method = "GET"
		}
		var resp *executor.HTTPResponse
		resp, err = executor.NewHTTPExecutor().Execute(cfg)
		if err == nil {
			body, ctype, requestURL = resp.Body, resp.ContentType, resp.RequestURL
		}
	default:
		err = fmt.Errorf("unsupported transport: %s", transport)
	}

	duration = time.Since(start)
	return
}

// detectTransport determines the transport type from URL scheme or config fields
func detectTransport(cfg *config.YapiConfig) string {
	urlLower := strings.ToLower(cfg.URL)

	// Check URL scheme first
	if strings.HasPrefix(urlLower, "grpc://") || strings.HasPrefix(urlLower, "grpcs://") {
		return "grpc"
	}
	if strings.HasPrefix(urlLower, "tcp://") {
		return "tcp"
	}

	// Check if graphql field is populated
	if cfg.Graphql != "" {
		return "graphql"
	}

	// Fall back to method field (deprecated but still supported)
	switch cfg.Method {
	case "grpc":
		return "grpc"
	case "tcp":
		return "tcp"
	default:
		return "http"
	}
}
