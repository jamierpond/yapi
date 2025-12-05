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

	"yapi.run/cli/internal/constants"
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
	"chain":            true,
	"expect":           true,
}

// knownChainStepKeys is the set of valid keys for chain step entries.
var knownChainStepKeys = map[string]bool{
	"name":      true,
	"url":       true,
	"method":    true,
	"headers":   true,
	"body":      true,
	"json":      true,
	"graphql":   true,
	"variables": true,
	"expect":    true,
}

// knownExpectKeys is the set of valid keys for expect blocks.
var knownExpectKeys = map[string]bool{
	"status": true,
	"assert": true,
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
	Method         string                 `yaml:"method,omitempty"` // HTTP method (GET, POST, PUT, DELETE, etc.)
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

	// Expect defines assertions to run after the request
	Expect Expectation `yaml:"expect,omitempty"`

	// Chain allows executing multiple dependent requests
	Chain []ChainStep `yaml:"chain,omitempty"`
}

// ChainStep represents a single step in a request chain.
// All fields are optional overrides on top of the base ConfigV1.
type ChainStep struct {
	Name string `yaml:"name"` // Required: unique step identifier

	// All fields below override the base config if set
	URL            string                 `yaml:"url,omitempty"`
	Path           string                 `yaml:"path,omitempty"`
	Method         string                 `yaml:"method,omitempty"`
	ContentType    string                 `yaml:"content_type,omitempty"`
	Headers        map[string]string      `yaml:"headers,omitempty"`
	Body           map[string]interface{} `yaml:"body,omitempty"`
	JSON           string                 `yaml:"json,omitempty"`
	Query          map[string]string      `yaml:"query,omitempty"`
	Graphql        string                 `yaml:"graphql,omitempty"`
	Variables      map[string]interface{} `yaml:"variables,omitempty"`
	Service        string                 `yaml:"service,omitempty"`
	RPC            string                 `yaml:"rpc,omitempty"`
	Proto          string                 `yaml:"proto,omitempty"`
	ProtoPath      string                 `yaml:"proto_path,omitempty"`
	Data           string                 `yaml:"data,omitempty"`
	Encoding       string                 `yaml:"encoding,omitempty"`
	JQFilter       string                 `yaml:"jq_filter,omitempty"`
	Insecure       *bool                  `yaml:"insecure,omitempty"`
	Plaintext      *bool                  `yaml:"plaintext,omitempty"`
	ReadTimeout    *int                   `yaml:"read_timeout,omitempty"`
	IdleTimeout    *int                   `yaml:"idle_timeout,omitempty"`
	CloseAfterSend *bool                  `yaml:"close_after_send,omitempty"`

	Expect Expectation `yaml:"expect,omitempty"`
}

// Merge creates a full ConfigV1 by applying step overrides to the base config.
// Maps are deep copied to avoid polluting the shared base config between steps.
func (base *ConfigV1) Merge(step ChainStep) ConfigV1 {
	merged := *base // Shallow copy of scalar fields
	merged.Chain = nil // Don't copy chain into merged step
	merged.Expect = step.Expect // Step expectations override base

	// Deep copy maps from base to avoid reference pollution between steps
	if base.Headers != nil {
		merged.Headers = make(map[string]string, len(base.Headers))
		for k, v := range base.Headers {
			merged.Headers[k] = v
		}
	}
	if base.Query != nil {
		merged.Query = make(map[string]string, len(base.Query))
		for k, v := range base.Query {
			merged.Query[k] = v
		}
	}
	if base.Body != nil {
		merged.Body = deepCopyMap(base.Body)
	}
	if base.Variables != nil {
		merged.Variables = deepCopyMap(base.Variables)
	}

	// Override string fields if step has them
	if step.URL != "" {
		merged.URL = step.URL
	}
	if step.Path != "" {
		merged.Path = step.Path
	}
	if step.Method != "" {
		merged.Method = step.Method
	}
	if step.ContentType != "" {
		merged.ContentType = step.ContentType
	}
	if step.JSON != "" {
		merged.JSON = step.JSON
	}
	if step.Graphql != "" {
		merged.Graphql = step.Graphql
	}
	if step.Service != "" {
		merged.Service = step.Service
	}
	if step.RPC != "" {
		merged.RPC = step.RPC
	}
	if step.Proto != "" {
		merged.Proto = step.Proto
	}
	if step.ProtoPath != "" {
		merged.ProtoPath = step.ProtoPath
	}
	if step.Data != "" {
		merged.Data = step.Data
	}
	if step.Encoding != "" {
		merged.Encoding = step.Encoding
	}
	if step.JQFilter != "" {
		merged.JQFilter = step.JQFilter
	}

	// Override pointer fields (bools/ints that need explicit false/0)
	if step.Insecure != nil {
		merged.Insecure = *step.Insecure
	}
	if step.Plaintext != nil {
		merged.Plaintext = *step.Plaintext
	}
	if step.ReadTimeout != nil {
		merged.ReadTimeout = *step.ReadTimeout
	}
	if step.IdleTimeout != nil {
		merged.IdleTimeout = *step.IdleTimeout
	}
	if step.CloseAfterSend != nil {
		merged.CloseAfterSend = *step.CloseAfterSend
	}

	// Merge maps (step headers/query add to or override base)
	if step.Headers != nil {
		if merged.Headers == nil {
			merged.Headers = make(map[string]string)
		}
		for k, v := range step.Headers {
			merged.Headers[k] = v
		}
	}
	if step.Query != nil {
		if merged.Query == nil {
			merged.Query = make(map[string]string)
		}
		for k, v := range step.Query {
			merged.Query[k] = v
		}
	}
	if step.Body != nil {
		merged.Body = step.Body
	}
	if step.Variables != nil {
		merged.Variables = step.Variables
	}

	return merged
}

// deepCopyMap creates a deep copy of a map[string]interface{}
func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		switch val := v.(type) {
		case map[string]interface{}:
			dst[k] = deepCopyMap(val)
		case []interface{}:
			dst[k] = deepCopySlice(val)
		default:
			dst[k] = v
		}
	}
	return dst
}

// deepCopySlice creates a deep copy of a []interface{}
func deepCopySlice(src []interface{}) []interface{} {
	if src == nil {
		return nil
	}
	dst := make([]interface{}, len(src))
	for i, v := range src {
		switch val := v.(type) {
		case map[string]interface{}:
			dst[i] = deepCopyMap(val)
		case []interface{}:
			dst[i] = deepCopySlice(val)
		default:
			dst[i] = v
		}
	}
	return dst
}

// Expectation defines assertions for a chain step
type Expectation struct {
	Status interface{} `yaml:"status,omitempty"` // int or []int
	Assert []string    `yaml:"assert,omitempty"` // JQ expressions that must evaluate to true
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
	c.Headers = expandMapEnv(c.Headers)
	c.Query = expandMapEnv(c.Query)
}

func expandMapEnv(m map[string]string) map[string]string {
	if len(m) == 0 {
		return m
	}
	for k, v := range m {
		m[k] = os.ExpandEnv(v)
	}
	return m
}

// setDefaults applies default values for Method
func (c *ConfigV1) setDefaults() {
	if c.Method == "" {
		c.Method = constants.MethodGET
	}
	c.Method = constants.CanonicalizeMethod(c.Method)
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

// detectTransport determines the transport type from URL scheme
func (c *ConfigV1) detectTransport() string {
	urlLower := strings.ToLower(c.URL)

	if strings.HasPrefix(urlLower, "grpc://") || strings.HasPrefix(urlLower, "grpcs://") {
		return constants.TransportGRPC
	}
	if strings.HasPrefix(urlLower, "tcp://") {
		return constants.TransportTCP
	}
	if c.Graphql != "" {
		return constants.TransportGraphQL
	}
	return constants.TransportHTTP
}

// enrichMetadata adds transport-specific metadata to the request
func (c *ConfigV1) enrichMetadata(req *domain.Request) error {
	transport := c.detectTransport()
	req.Metadata["transport"] = transport

	switch transport {
	case constants.TransportGRPC:
		req.Metadata["service"] = c.Service
		req.Metadata["rpc"] = c.RPC
		req.Metadata["proto"] = c.Proto
		req.Metadata["proto_path"] = c.ProtoPath
		req.Metadata["insecure"] = fmt.Sprintf("%t", c.Insecure)
		req.Metadata["plaintext"] = fmt.Sprintf("%t", c.Plaintext)
	case constants.TransportTCP:
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
