package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"

	"yapi.run/cli/internal/domain"
)

// knownV1Keys is the set of valid keys for v1 config files.
// Must be kept in sync with ConfigV1 struct yaml tags.
var knownV1Keys = map[string]bool{
	"yapi":             true,
	"url":              true,
	"path":             true,
	"method":           true,
	"content_type":     true,
	"headers":          true,
	"body":             true,
	"json":             true,
	"query":            true,
	"graphql":          true,
	"variables":        true,
	"service":          true,
	"rpc":              true,
	"proto":            true,
	"proto_path":       true,
	"data":             true,
	"encoding":         true,
	"jq_filter":        true,
	"insecure":         true,
	"plaintext":        true,
	"read_timeout":     true,
	"idle_timeout":     true,
	"close_after_send": true,
}

// FindUnknownKeys checks a raw map for keys not in knownV1Keys.
// Returns a sorted slice of unknown key names.
func FindUnknownKeys(raw map[string]interface{}) []string {
	var unknown []string
	for key := range raw {
		if !knownV1Keys[key] {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)
	return unknown
}

// ConfigV1 represents the v1 YAML schema
type ConfigV1 struct {
	Yapi           string                 `yaml:"yapi"` // The version tag
	URL            string                 `yaml:"url"`
	Path           string                 `yaml:"path,omitempty"`
	Method         string                 `yaml:"method,omitempty"` // GET, POST, grpc, tcp
	ContentType    string                 `yaml:"content_type,omitempty"`
	Headers        map[string]string      `yaml:"headers,omitempty"`
	Body           map[string]interface{} `yaml:"body,omitempty"`
	JSON           string                 `yaml:"json,omitempty"` // Raw JSON override
	Query          map[string]string      `yaml:"query,omitempty"`
	Graphql        string                 `yaml:"graphql,omitempty"`   // GraphQL query/mutation
	Variables      map[string]interface{} `yaml:"variables,omitempty"` // GraphQL variables
	Service        string                 `yaml:"service,omitempty"`   // gRPC
	RPC            string                 `yaml:"rpc,omitempty"`       // gRPC
	Proto          string                 `yaml:"proto,omitempty"`     // gRPC
	ProtoPath      string                 `yaml:"proto_path,omitempty"`
	Data           string                 `yaml:"data,omitempty"`     // TCP raw data
	Encoding       string                 `yaml:"encoding,omitempty"` // text, hex, base64
	JQFilter       string                 `yaml:"jq_filter,omitempty"`
	Insecure       bool                   `yaml:"insecure,omitempty"`     // For gRPC
	Plaintext      bool                   `yaml:"plaintext,omitempty"`    // For gRPC
	ReadTimeout    int                    `yaml:"read_timeout,omitempty"` // TCP read timeout in seconds
	IdleTimeout    int                    `yaml:"idle_timeout,omitempty"` // TCP idle timeout in milliseconds (default 500)
	CloseAfterSend bool                   `yaml:"close_after_send,omitempty"`
}

// ToDomain converts V1 YAML to the Canonical Config
func (c *ConfigV1) ToDomain() (*domain.Request, error) {
	c.expandEnvVars()
	c.setDefaults()

	bodyReader, bodySource, err := c.prepareBody()
	if err != nil {
		return nil, err
	}

	req := &domain.Request{
		URL:      c.buildURL(),
		Method:   c.Method,
		Headers:  c.Headers,
		Body:     bodyReader,
		Metadata: make(map[string]string),
	}

	if c.ContentType != "" {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Content-Type"] = c.ContentType
	}

	if bodySource != "" {
		req.Metadata["body_source"] = bodySource
	}

	if err := c.enrichMetadata(req); err != nil {
		return nil, err
	}

	return req, nil
}

// expandEnvVars expands environment variables in URL, Path, Headers, and Query
func (c *ConfigV1) expandEnvVars() {
	c.URL = os.ExpandEnv(c.URL)
	c.Path = os.ExpandEnv(c.Path)
	for k, v := range c.Headers {
		c.Headers[k] = os.ExpandEnv(v)
	}
	for k, v := range c.Query {
		c.Query[k] = os.ExpandEnv(v)
	}
}

// setDefaults applies default values for Method
func (c *ConfigV1) setDefaults() {
	if c.Method == "" {
		c.Method = "GET"
	}
	c.Method = strings.ToUpper(c.Method)
}

// prepareBody processes the body/json fields and returns a reader, source identifier, and any error
func (c *ConfigV1) prepareBody() (io.Reader, string, error) {
	if c.JSON != "" && c.Body != nil && len(c.Body) > 0 {
		return nil, "", fmt.Errorf("`body` and `json` are mutually exclusive")
	}

	if c.JSON != "" {
		if c.ContentType == "" {
			c.ContentType = "application/json"
		}
		return strings.NewReader(c.JSON), "json", nil
	}

	if c.Body != nil {
		bodyBytes, err := json.Marshal(c.Body)
		if err != nil {
			return nil, "", fmt.Errorf("invalid json in 'body' field: %w", err)
		}
		if c.ContentType == "" {
			c.ContentType = "application/json"
		}
		return bytes.NewReader(bodyBytes), "", nil
	}

	return nil, "", nil
}

// buildURL constructs the final URL with path and query parameters
func (c *ConfigV1) buildURL() string {
	finalURL := c.URL
	if c.Path != "" {
		finalURL += c.Path
	}
	if len(c.Query) > 0 {
		q := url.Values{}
		for k, v := range c.Query {
			q.Set(k, v)
		}
		finalURL += "?" + q.Encode()
	}
	return finalURL
}

// detectTransport determines the transport type from URL and method
func (c *ConfigV1) detectTransport() string {
	urlLower := strings.ToLower(c.URL)
	methodLower := strings.ToLower(c.Method)

	if strings.HasPrefix(urlLower, "grpc://") || strings.HasPrefix(urlLower, "grpcs://") || methodLower == "grpc" {
		return "grpc"
	}
	if strings.HasPrefix(urlLower, "tcp://") || methodLower == "tcp" {
		return "tcp"
	}
	if c.Graphql != "" {
		return "graphql"
	}
	return "http"
}

// enrichMetadata adds transport-specific metadata to the request
func (c *ConfigV1) enrichMetadata(req *domain.Request) error {
	transport := c.detectTransport()
	req.Metadata["transport"] = transport

	switch transport {
	case "grpc":
		req.Metadata["service"] = c.Service
		req.Metadata["rpc"] = c.RPC
		req.Metadata["proto"] = c.Proto
		req.Metadata["proto_path"] = c.ProtoPath
		req.Metadata["insecure"] = fmt.Sprintf("%t", c.Insecure)
		req.Metadata["plaintext"] = fmt.Sprintf("%t", c.Plaintext)
	case "tcp":
		req.Metadata["data"] = c.Data
		req.Metadata["encoding"] = c.Encoding
		req.Metadata["read_timeout"] = fmt.Sprintf("%d", c.ReadTimeout)
		req.Metadata["idle_timeout"] = fmt.Sprintf("%d", c.IdleTimeout)
		req.Metadata["close_after_send"] = fmt.Sprintf("%t", c.CloseAfterSend)
	}

	if c.JQFilter != "" {
		req.Metadata["jq_filter"] = c.JQFilter
	}

	if c.Graphql != "" {
		req.Metadata["graphql_query"] = c.Graphql
		if c.Variables != nil {
			vars, err := json.Marshal(c.Variables)
			if err != nil {
				return fmt.Errorf("could not marshal graphql variables: %w", err)
			}
			req.Metadata["graphql_variables"] = string(vars)
		}
	}

	return nil
}
