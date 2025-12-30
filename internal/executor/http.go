package executor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"yapi.run/cli/internal/domain"
)

// HTTPTransport returns a transport function for HTTP requests.
func HTTPTransport(client HTTPClient) TransportFunc {
	return func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		// Apply timeout if specified
		if timeoutStr, ok := req.Metadata["timeout"]; ok && timeoutStr != "" {
			timeout, err := time.ParseDuration(timeoutStr)
			if err != nil {
				return nil, fmt.Errorf("invalid timeout value %q: %w", timeoutStr, err)
			}
			// Create timeout context
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set custom headers
		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}

		clientToUse := client
		if insecureStr, ok := req.Metadata["insecure"]; ok && insecureStr != "" {
			insecure, err := strconv.ParseBool(insecureStr)
			if err == nil && insecure {
				clientToUse = insecureHTTPClient(client)
			}
		}

		res, err := clientToUse.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		// Convert http.Header to map[string]string
		headers := make(map[string]string)
		for k, v := range res.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		return &domain.Response{
			StatusCode: res.StatusCode,
			Headers:    headers,
			Body:       res.Body,
		}, nil
	}
}

func insecureHTTPClient(base HTTPClient) *http.Client {
	var baseClient *http.Client
	if client, ok := base.(*http.Client); ok {
		baseClient = client
	}

	transport := cloneTransport(baseClient)
	tlsConfig := &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-controlled insecure TLS option
	if transport.TLSClientConfig != nil {
		tlsConfig = transport.TLSClientConfig.Clone()
		tlsConfig.InsecureSkipVerify = true
	}
	transport.TLSClientConfig = tlsConfig

	client := &http.Client{
		Transport: transport,
	}
	if baseClient != nil {
		client.Timeout = baseClient.Timeout
		client.Jar = baseClient.Jar
		client.CheckRedirect = baseClient.CheckRedirect
	}
	return client
}

func cloneTransport(base *http.Client) *http.Transport {
	if base != nil && base.Transport != nil {
		if transport, ok := base.Transport.(*http.Transport); ok {
			return transport.Clone()
		}
	}
	return http.DefaultTransport.(*http.Transport).Clone()
}
