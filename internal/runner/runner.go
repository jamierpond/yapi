package runner

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/filter"
	"yapi.run/cli/internal/output"
)

// Result holds the output of a yapi execution
type Result struct {
	Body        string
	ContentType string
	StatusCode  int
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

// Run executes a yapi request and returns the result.
func Run(ctx context.Context, exec executor.Executor, req *domain.Request, warnings []string, opts Options) (*Result, error) {
	// Apply URL override
	if opts.URLOverride != "" {
		req.URL = opts.URLOverride
	}

	// Execute the request
	resp, err := exec.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	body := string(bodyBytes)

	// Apply JQ filter if specified
	if jqFilter, ok := req.Metadata["jq_filter"]; ok && jqFilter != "" {
		body, err = filter.ApplyJQ(body, jqFilter)
		if err != nil {
			return nil, fmt.Errorf("jq filter failed: %w", err)
		}
		resp.Headers["Content-Type"] = "application/json"
	}

	bodyLines := strings.Count(body, "\n") + 1
	bodyChars := len(body)
	bodyBytesLen := len(bodyBytes)

	return &Result{
		Body:        body,
		ContentType: resp.Headers["Content-Type"],
		StatusCode:  resp.StatusCode,
		Warnings:    warnings,
		RequestURL:  req.URL,
		Duration:    resp.Duration,
		BodyLines:   bodyLines,
		BodyChars:   bodyChars,
		BodyBytes:   bodyBytesLen,
	}, nil
}

// RunAndFormat executes and returns highlighted output plus Result metadata
func RunAndFormat(ctx context.Context, exec executor.Executor, req *domain.Request, warnings []string, opts Options) (string, *Result, error) {
	result, err := Run(ctx, exec, req, warnings, opts)
	if err != nil {
		return "", nil, err
	}

	out := output.Highlight(result.Body, result.ContentType, opts.NoColor)

	return out, result, nil
}
