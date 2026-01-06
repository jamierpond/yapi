package executor

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"yapi.run/cli/internal/domain"
)

// cancelOnCloseBody wraps an io.ReadCloser and calls cancel when Close is called.
// This ensures the context isn't canceled until the body is fully read and closed.
type cancelOnCloseBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnCloseBody) Close() error {
	err := c.ReadCloser.Close()
	if c.cancel != nil {
		c.cancel()
	}
	return err
}

// HTTPTransport returns a transport function for HTTP requests.
func HTTPTransport(client HTTPClient) TransportFunc {
	return func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		// Apply timeout if specified
		var cancel context.CancelFunc
		if timeoutStr, ok := req.Metadata["timeout"]; ok && timeoutStr != "" {
			timeout, err := time.ParseDuration(timeoutStr)
			if err != nil {
				return nil, fmt.Errorf("invalid timeout value %q: %w", timeoutStr, err)
			}
			ctx, cancel = context.WithTimeout(ctx, timeout)
		}

		httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, req.Body)
		if err != nil {
			if cancel != nil {
				cancel()
			}
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

		res, err := clientToUse.Do(httpReq) //nolint:bodyclose // Body is returned to caller who closes it
		if err != nil {
			if cancel != nil {
				cancel()
			}
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		// Convert http.Header to map[string]string
		headers := make(map[string]string)
		for k, v := range res.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		// Wrap body so cancel is called when body is closed, not before
		body := res.Body
		if cancel != nil {
			body = &cancelOnCloseBody{ReadCloser: res.Body, cancel: cancel}
		}

		return &domain.Response{
			StatusCode: res.StatusCode,
			Headers:    headers,
			Body:       body,
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
