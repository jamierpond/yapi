package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"yapi.run/cli/internal/constants"
	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/output"
	"yapi.run/cli/internal/runner"
)

func (app *rootCommand) sendE(cmd *cobra.Command, args []string) error {
	url := args[0]

	var body string
	if len(args) > 1 {
		body = args[1]
	}

	method, _ := cmd.Flags().GetString("method")
	headers, _ := cmd.Flags().GetStringSlice("header")
	verbose, _ := cmd.Flags().GetBool("verbose")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	jqFilter, _ := cmd.Flags().GetString("jq")

	log := NewLogger(verbose)

	// Detect transport from URL scheme
	transport := domain.DetectTransport(url, false)

	// Default method: POST if body provided, GET otherwise (HTTP only)
	if method == "" {
		if transport == constants.TransportHTTP {
			if body != "" {
				method = constants.MethodPOST
			} else {
				method = constants.MethodGET
			}
		}
	} else {
		method = constants.CanonicalizeMethod(method)
	}

	// Build request
	req := &domain.Request{
		URL:      url,
		Method:   method,
		Headers:  make(map[string]string),
		Metadata: make(map[string]string),
	}

	// Parse headers from -H flags
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid header format %q: expected 'Key: Value'", h)
		}
		req.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	// Set body
	if body != "" {
		req.Body = strings.NewReader(body)

		// For HTTP, default to JSON content type if body looks like JSON and no content-type set
		if transport == constants.TransportHTTP {
			if _, hasContentType := req.Headers["Content-Type"]; !hasContentType {
				trimmed := strings.TrimSpace(body)
				if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
					(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
					req.Headers["Content-Type"] = "application/json"
				}
			}
		}
	}

	// Set transport metadata
	req.Metadata["transport"] = transport
	if app.insecure {
		req.Metadata["insecure"] = "true"
	}

	// TCP-specific metadata defaults
	if transport == constants.TransportTCP {
		req.Metadata["data"] = body
		req.Metadata["encoding"] = "text"
		req.Metadata["read_timeout"] = "5"
		req.Metadata["idle_timeout"] = "500"
		req.Metadata["close_after_send"] = "true"
	}

	if jqFilter != "" {
		req.Metadata["jq_filter"] = jqFilter
	}

	log.Verbose("Transport: %s", transport)
	log.Request(method, url, req.Headers, body)

	// Get executor
	exec, err := executor.GetTransport(transport, app.httpClient)
	if err != nil {
		return err
	}

	// Execute
	opts := runner.Options{
		NoColor:      app.noColor,
		BinaryOutput: app.binaryOutput,
		Insecure:     app.insecure,
	}

	result, err := runner.Run(context.Background(), exec, req, nil, opts)
	if err != nil {
		return err
	}

	log.Response(result.StatusCode, result.Headers, result.Duration, result.BodyBytes)

	if jsonOutput {
		return output.PrintJSON(output.JSONParams{
			Result: result,
		})
	}

	return app.printResult(result, nil, "send", printResultOptions{skipMeta: verbose})
}
