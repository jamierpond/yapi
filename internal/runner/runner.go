package runner

import (
	"fmt"
	"strings"

	"yapi.run/internal/config"
	"yapi.run/internal/executor"
	"yapi.run/internal/filter"
	"yapi.run/internal/output"
	"yapi.run/internal/validation"
)

// Result holds the output of a yapi execution
type Result struct {
	Body        string
	ContentType string
	Warnings    []string
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
	body, ctype, err := execute(cfg)
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

	return &Result{
		Body:        body,
		ContentType: ctype,
		Warnings:    warnings,
	}, nil
}

// RunAndFormat executes and returns highlighted output
func RunAndFormat(cfg *config.YapiConfig, opts Options) (string, error) {
	result, err := Run(cfg, opts)
	if err != nil {
		return "", err
	}

	output := output.Highlight(result.Body, result.ContentType, opts.NoColor)
	if len(result.Warnings) > 0 {
		output = strings.Join(result.Warnings, "\n") + "\n\n" + output
	}

	return output, nil
}

// execute dispatches to the appropriate executor based on config
func execute(cfg *config.YapiConfig) (string, string, error) {
	transport := detectTransport(cfg)

	switch transport {
	case "graphql":
		resp, err := executor.NewGraphQLExecutor().Execute(cfg)
		if err != nil {
			return "", "", err
		}
		return resp.Body, resp.ContentType, nil
	case "grpc":
		body, err := executor.NewGRPCExecutor().Execute(cfg)
		return body, "application/json", err
	case "tcp":
		body, err := executor.NewTCPExecutor().Execute(cfg)
		return body, "text/plain", err
	case "http":
		if cfg.Method == "" {
			cfg.Method = "GET"
		}
		resp, err := executor.NewHTTPExecutor().Execute(cfg)
		if err != nil {
			return "", "", err
		}
		return resp.Body, resp.ContentType, nil
	default:
		return "", "", fmt.Errorf("unsupported transport: %s", transport)
	}
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
