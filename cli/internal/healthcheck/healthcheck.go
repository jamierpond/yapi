// Package healthcheck provides health check polling for HTTP, gRPC, and TCP endpoints.
package healthcheck

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// newHTTPClient creates a reusable HTTP client for health checks.
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // #nosec G402 -- health checks often use self-signed certs
			},
		},
	}
}

// WaitForHealth polls all URLs until they're healthy or timeout.
// URLs can be http://, https://, grpc://, grpcs://, or tcp://.
// Returns nil when all URLs are healthy, or an error on timeout/failure.
func WaitForHealth(ctx context.Context, urls []string, timeout time.Duration) error {
	if len(urls) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create reusable HTTP client
	httpClient := newHTTPClient()

	// Track health status for each URL
	healthy := make(map[string]bool)
	for _, u := range urls {
		healthy[u] = false
	}

	// Exponential backoff: 100ms -> 200ms -> 400ms -> 800ms -> 1600ms -> capped at 2s
	backoff := 100 * time.Millisecond
	maxBackoff := 2 * time.Second

	for {
		allHealthy := true
		var lastErr error

		for _, u := range urls {
			if healthy[u] {
				continue
			}

			err := checkHealth(ctx, u, httpClient)
			if err == nil {
				healthy[u] = true
			} else {
				allHealthy = false
				lastErr = err
			}
		}

		if allHealthy {
			return nil
		}

		// Check if context is done
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return fmt.Errorf("health check timeout: %w", lastErr)
			}
			return fmt.Errorf("health check timeout")
		default:
		}

		// Wait with backoff
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return fmt.Errorf("health check timeout: %w", lastErr)
			}
			return fmt.Errorf("health check timeout")
		case <-time.After(backoff):
		}

		// Increase backoff
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// checkHealth checks a single URL based on its scheme.
func checkHealth(ctx context.Context, rawURL string, httpClient *http.Client) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return checkHTTP(ctx, rawURL, httpClient)
	case "grpc", "grpcs":
		return checkGRPC(ctx, parsed)
	case "tcp":
		return checkTCP(ctx, parsed.Host)
	default:
		return fmt.Errorf("unsupported health check scheme: %s", parsed.Scheme)
	}
}

// checkHTTP performs an HTTP GET request and expects a 2xx response.
func checkHTTP(ctx context.Context, rawURL string, client *http.Client) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unhealthy status: %d", resp.StatusCode)
	}

	return nil
}

// checkGRPC performs a gRPC health check using the standard health protocol.
func checkGRPC(ctx context.Context, parsed *url.URL) error {
	var opts []grpc.DialOption

	if strings.ToLower(parsed.Scheme) == "grpcs" {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true, // #nosec G402 -- health checks often use self-signed certs
		})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(parsed.Host, opts...)
	if err != nil {
		return fmt.Errorf("grpc dial failed: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client := healthpb.NewHealthClient(conn)

	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{
		Service: "", // Empty string checks overall server health
	})
	if err != nil {
		return fmt.Errorf("grpc health check failed: %w", err)
	}

	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("grpc unhealthy: %s", resp.Status.String())
	}

	return nil
}

// checkTCP attempts to establish a TCP connection.
func checkTCP(ctx context.Context, host string) error {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
