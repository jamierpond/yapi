package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"yapi/internal/config"
)

// HTTPExecutor handles HTTP requests.
type HTTPExecutor struct{}

// NewHTTPExecutor creates a new HTTPExecutor.
func NewHTTPExecutor() *HTTPExecutor {
	return &HTTPExecutor{}
}

// Execute performs an HTTP request based on the provided YapiConfig.
func (e *HTTPExecutor) Execute(cfg *config.YapiConfig) (string, error) {
	var reqBody io.Reader

	if cfg.Body != nil {
		b, err := json.Marshal(cfg.Body)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(b)
		// Default content type to application/json if body is present and not explicitly set
		if cfg.ContentType == "" {
			cfg.ContentType = "application/json"
		}
	}

	parsedURL, err := url.Parse(cfg.URL)
		if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	if len(cfg.Query) > 0 {
		query := parsedURL.Query()
		for k, v := range cfg.Query {
			query.Set(k, v)
		}
		parsedURL.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(cfg.Method, parsedURL.String(), reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if cfg.ContentType != "" {
		req.Header.Set("Content-Type", cfg.ContentType)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
