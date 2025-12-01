package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"cli/internal/config"
)

// HTTPResponse contains the response from an HTTP request.
type HTTPResponse struct {
	Body        string
	ContentType string
	StatusCode  int
}

// HTTPExecutor handles HTTP requests.
type HTTPExecutor struct{}

// NewHTTPExecutor creates a new HTTPExecutor.
func NewHTTPExecutor() *HTTPExecutor {
	return &HTTPExecutor{}
}

// Execute performs an HTTP request based on the provided YapiConfig.
func (e *HTTPExecutor) Execute(cfg *config.YapiConfig) (*HTTPResponse, error) {
	var reqBody io.Reader

	if cfg.Body != nil || cfg.JSON != "" {
		if cfg.Body != nil {
			b, err := json.Marshal(cfg.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(b)
		} else if cfg.JSON != "" {
			reqBody = bytes.NewBuffer([]byte(cfg.JSON))
		}

		// Default content type to application/json if body or json is present and not explicitly set
		if cfg.ContentType == "" {
			cfg.ContentType = "application/json"
		}
	}

	baseURL, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	if cfg.Path != "" {
		fullURL, err := baseURL.Parse(cfg.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse path: %w", err)
		}
		baseURL = fullURL
	}

	// Add query parameters
	if len(cfg.Query) > 0 {
		query := baseURL.Query()
		for k, v := range cfg.Query {
			query.Set(k, v)
		}
		baseURL.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(cfg.Method, baseURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if cfg.ContentType != "" {
		req.Header.Set("Content-Type", cfg.ContentType)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &HTTPResponse{
		Body:        string(body),
		ContentType: res.Header.Get("Content-Type"),
		StatusCode:  res.StatusCode,
	}, nil
}
