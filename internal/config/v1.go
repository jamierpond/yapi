package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"yapi.run/cli/internal/domain"
)

// ConfigV1 represents the v1 YAML schema
type ConfigV1 struct {
	Yapi        string                 `yaml:"yapi"` // The version tag
	URL         string                 `yaml:"url"`
	Path        string                 `yaml:"path,omitempty"`
	Method      string                 `yaml:"method,omitempty"` // GET, POST, grpc, tcp
	ContentType string                 `yaml:"content_type,omitempty"`
	Headers     map[string]string      `yaml:"headers,omitempty"`
	Body        map[string]interface{} `yaml:"body,omitempty"`
	JSON        string                 `yaml:"json,omitempty"` // Raw JSON override
	Query       map[string]string      `yaml:"query,omitempty"`
	Graphql     string                 `yaml:"graphql,omitempty"`   // GraphQL query/mutation
	Variables   map[string]interface{} `yaml:"variables,omitempty"` // GraphQL variables
	Service     string                 `yaml:"service,omitempty"`   // gRPC
	RPC         string                 `yaml:"rpc,omitempty"`       // gRPC
	Proto       string                 `yaml:"proto,omitempty"`     // gRPC
	ProtoPath   string                 `yaml:"proto_path,omitempty"`
	Data        string                 `yaml:"data,omitempty"`     // TCP raw data
	Encoding    string                 `yaml:"encoding,omitempty"` // text, hex, base64
	JQFilter    string                 `yaml:"jq_filter,omitempty"`
	Insecure    bool                   `yaml:"insecure,omitempty"`     // For gRPC
	Plaintext   bool                   `yaml:"plaintext,omitempty"`    // For gRPC
	ReadTimeout int                    `yaml:"read_timeout,omitempty"` // TCP read timeout in seconds
	IdleTimeout int                    `yaml:"idle_timeout,omitempty"` // TCP idle timeout in milliseconds (default 500)
	CloseAfterSend bool                `yaml:"close_after_send,omitempty"`
}

// ToDomain converts V1 YAML to the Canonical Config
func (c *ConfigV1) ToDomain() (*domain.Request, error) {
	// 1. Expand environment variables
	c.URL = os.ExpandEnv(c.URL)
	c.Path = os.ExpandEnv(c.Path)
	for k, v := range c.Headers {
		c.Headers[k] = os.ExpandEnv(v)
	}
	for k, v := range c.Query {
		c.Query[k] = os.ExpandEnv(v)
	}

	// 2. Set defaults
	if c.Method == "" {
		c.Method = "GET"
	}
	c.Method = strings.ToUpper(c.Method)

	// 3. Process body
	var bodyReader io.Reader
	if c.JSON != "" && c.Body != nil && len(c.Body) > 0 {
		return nil, fmt.Errorf("`body` and `json` are mutually exclusive")
	}

	if c.JSON != "" {
		bodyReader = strings.NewReader(c.JSON)
		if c.ContentType == "" {
			c.ContentType = "application/json"
		}
	} else if c.Body != nil {
		bodyBytes, err := json.Marshal(c.Body)
		if err != nil {
			return nil, fmt.Errorf("invalid json in 'body' field: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
		if c.ContentType == "" {
			c.ContentType = "application/json"
		}
	}

	// 4. Construct final URL with path and query
	finalURL := c.URL
	if c.Path != "" {
		finalURL += c.Path
	}
	if len(c.Query) > 0 {
		queryParams := make([]string, 0, len(c.Query))
		for k, v := range c.Query {
			queryParams = append(queryParams, fmt.Sprintf("%s=%s", k, v))
		}
		finalURL += "?" + strings.Join(queryParams, "&")
	}

	// 5. Build domain.Request
	req := &domain.Request{
		URL:      finalURL,
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

	// Add transport-specific data to metadata
	var transport string
	urlLower := strings.ToLower(c.URL)
	methodLower := strings.ToLower(c.Method)

	if strings.HasPrefix(urlLower, "grpc://") || strings.HasPrefix(urlLower, "grpcs://") || methodLower == "grpc" {
		transport = "grpc"
	} else if strings.HasPrefix(urlLower, "tcp://") || methodLower == "tcp" {
		transport = "tcp"
	} else if c.Graphql != "" {
		transport = "graphql"
	} else {
		transport = "http"
	}
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
				return nil, fmt.Errorf("could not marshal graphql variables: %w", err)
			}
			req.Metadata["graphql_variables"] = string(vars)
		}
	}


	return req, nil
}
