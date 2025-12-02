package config

import (
	"encoding/json"
	"fmt"
)

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

// ToInternal converts V1 YAML to the Canonical Config
func (c *ConfigV1) ToInternal() (*Config, error) {
	// Map V1 fields to Internal fields
	cfg := &Config{
		Version:        "v1",
		URL:            c.URL,
		Method:         c.Method,
		Headers:        c.Headers,
		Query:          c.Query,
		Path:           c.Path,
		ContentType:    c.ContentType,
		Graphql:        c.Graphql,
		Variables:      c.Variables,
		Service:        c.Service,
		RPC:            c.RPC,
		Proto:          c.Proto,
		ProtoPath:      c.ProtoPath,
		Data:           c.Data,
		Encoding:       c.Encoding,
		JQFilter:       c.JQFilter,
		Insecure:       c.Insecure,
		Plaintext:      c.Plaintext,
		ReadTimeout:    c.ReadTimeout,
		IdleTimeout:    c.IdleTimeout,
		CloseAfterSend: c.CloseAfterSend,
	}

	// Example: V1 handles mutually exclusive body/json here
	// so the Runner receives a unified Body
	if c.JSON != "" {
		var bodyData interface{}
		// check if json is valid
		err := json.Unmarshal([]byte(c.JSON), &bodyData)
		if err != nil {
			return nil, fmt.Errorf("invalid json in 'json' field: %w", err)
		}
		cfg.Body = bodyData
	} else {
		cfg.Body = c.Body
	}

	return cfg, nil
}
