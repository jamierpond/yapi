package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"sort"
	"strings"

	"yapi.run/cli/internal/constants"
	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/utils"
	"yapi.run/cli/internal/vars"
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
	"form":             true,
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
	"delay":            true,
	"output_file":      true,
	"timeout":          true,
}

// FindUnknownKeys checks a raw map for keys not in knownV1Keys.
// Returns a sorted slice of unknown key names.
func FindUnknownKeys(raw map[string]any) []string {
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
	Yapi           string            `yaml:"yapi"` // The version tag
	URL            string            `yaml:"url"`
	Path           string            `yaml:"path,omitempty"`
	Method         string            `yaml:"method,omitempty"` // HTTP method (GET, POST, PUT, DELETE, etc.)
	ContentType    string            `yaml:"content_type,omitempty"`
	Headers        map[string]string `yaml:"headers,omitempty"`
	Body           map[string]any    `yaml:"body,omitempty"`
	JSON           string            `yaml:"json,omitempty"` // Raw JSON override
	Form           map[string]string `yaml:"form,omitempty"` // Form data (application/x-www-form-urlencoded or multipart/form-data)
	Query          map[string]string `yaml:"query,omitempty"`
	Graphql        string            `yaml:"graphql,omitempty"`   // GraphQL query/mutation
	Variables      map[string]any    `yaml:"variables,omitempty"` // GraphQL variables
	Service        string            `yaml:"service,omitempty"`   // gRPC
	RPC            string            `yaml:"rpc,omitempty"`       // gRPC
	Proto          string            `yaml:"proto,omitempty"`     // gRPC
	ProtoPath      string            `yaml:"proto_path,omitempty"`
	Data           string            `yaml:"data,omitempty"`     // TCP raw data
	Encoding       string            `yaml:"encoding,omitempty"` // text, hex, base64
	JQFilter       string            `yaml:"jq_filter,omitempty"`
	Insecure       bool              `yaml:"insecure,omitempty"`     // Skip TLS verification for HTTP/GraphQL; uses insecure transport for gRPC
	Plaintext      bool              `yaml:"plaintext,omitempty"`    // For gRPC
	ReadTimeout    int               `yaml:"read_timeout,omitempty"` // TCP read timeout in seconds
	IdleTimeout    int               `yaml:"idle_timeout,omitempty"` // TCP idle timeout in milliseconds (default 500)
	CloseAfterSend bool              `yaml:"close_after_send,omitempty"`

	// Flow control
	Delay   string `yaml:"delay,omitempty"`   // Wait before executing this step (e.g. "5s", "500ms")
	Timeout string `yaml:"timeout,omitempty"` // HTTP request timeout (e.g. "4s", "100ms", "1m")

	// Output
	OutputFile string `yaml:"output_file,omitempty"` // Save response to file (e.g. "./output.json", "./image.png")

	// Expect defines assertions to run after the request
	Expect Expectation `yaml:"expect,omitempty"`

	// Chain allows executing multiple dependent requests
	Chain []ChainStep `yaml:"chain,omitempty"`
}

// ChainStep represents a single step in a request chain.
// It embeds ConfigV1 so all config fields are available as overrides.
type ChainStep struct {
	Name     string           `yaml:"name"` // Required: unique step identifier
	ConfigV1 `yaml:",inline"` // All ConfigV1 fields available as overrides
}

// Merge creates a full ConfigV1 by applying step overrides to the base config.
// Maps are deep copied to avoid polluting the shared base config between steps.
func (c *ConfigV1) Merge(step ChainStep) ConfigV1 {
	m := *c
	m.Chain = nil
	m.Expect = step.Expect

	// Scalar overrides using Coalesce
	m.URL = utils.Coalesce(step.URL, c.URL)
	m.Path = utils.Coalesce(step.Path, c.Path)
	m.Method = utils.Coalesce(step.Method, c.Method)
	m.ContentType = utils.Coalesce(step.ContentType, c.ContentType)
	m.JSON = utils.Coalesce(step.JSON, c.JSON)
	m.Graphql = utils.Coalesce(step.Graphql, c.Graphql)
	m.Service = utils.Coalesce(step.Service, c.Service)
	m.RPC = utils.Coalesce(step.RPC, c.RPC)
	m.Proto = utils.Coalesce(step.Proto, c.Proto)
	m.ProtoPath = utils.Coalesce(step.ProtoPath, c.ProtoPath)
	m.Data = utils.Coalesce(step.Data, c.Data)
	m.Encoding = utils.Coalesce(step.Encoding, c.Encoding)
	m.JQFilter = utils.Coalesce(step.JQFilter, c.JQFilter)
	m.Delay = utils.Coalesce(step.Delay, c.Delay)
	m.Timeout = utils.Coalesce(step.Timeout, c.Timeout)
	m.OutputFile = utils.Coalesce(step.OutputFile, c.OutputFile)

	// Bool/Int overrides
	if step.Insecure {
		m.Insecure = true
	}
	if step.Plaintext {
		m.Plaintext = true
	}
	if step.CloseAfterSend {
		m.CloseAfterSend = true
	}
	if step.ReadTimeout != 0 {
		m.ReadTimeout = step.ReadTimeout
	}
	if step.IdleTimeout != 0 {
		m.IdleTimeout = step.IdleTimeout
	}

	// Generic map merging
	m.Headers = utils.MergeMaps(c.Headers, step.Headers)
	m.Query = utils.MergeMaps(c.Query, step.Query)

	// Deep clone Body/Variables from c, then override if step has values
	m.Body = utils.DeepCloneMap(c.Body)
	if step.Body != nil {
		m.Body = step.Body
	}

	m.Variables = utils.DeepCloneMap(c.Variables)
	if step.Variables != nil {
		m.Variables = step.Variables
	}

	return m
}

// MergeWithDefaults applies environment defaults to a file config.
// File values take precedence over environment defaults.
func (c *ConfigV1) MergeWithDefaults(defaults ConfigV1) ConfigV1 {
	m := defaults // Start with defaults

	// File values override defaults (reverse of Coalesce order)
	m.URL = utils.Coalesce(c.URL, defaults.URL)
	m.Path = utils.Coalesce(c.Path, defaults.Path)
	m.Method = utils.Coalesce(c.Method, defaults.Method)
	m.ContentType = utils.Coalesce(c.ContentType, defaults.ContentType)
	m.JSON = utils.Coalesce(c.JSON, defaults.JSON)
	m.Graphql = utils.Coalesce(c.Graphql, defaults.Graphql)
	m.Service = utils.Coalesce(c.Service, defaults.Service)
	m.RPC = utils.Coalesce(c.RPC, defaults.RPC)
	m.Proto = utils.Coalesce(c.Proto, defaults.Proto)
	m.ProtoPath = utils.Coalesce(c.ProtoPath, defaults.ProtoPath)
	m.Data = utils.Coalesce(c.Data, defaults.Data)
	m.Encoding = utils.Coalesce(c.Encoding, defaults.Encoding)
	m.JQFilter = utils.Coalesce(c.JQFilter, defaults.JQFilter)
	m.Delay = utils.Coalesce(c.Delay, defaults.Delay)
	m.Timeout = utils.Coalesce(c.Timeout, defaults.Timeout)
	m.OutputFile = utils.Coalesce(c.OutputFile, defaults.OutputFile)

	// Bool/Int overrides - file values take precedence
	if c.Insecure {
		m.Insecure = true
	}
	if c.Plaintext {
		m.Plaintext = true
	}
	if c.CloseAfterSend {
		m.CloseAfterSend = true
	}
	if c.ReadTimeout != 0 {
		m.ReadTimeout = c.ReadTimeout
	}
	if c.IdleTimeout != 0 {
		m.IdleTimeout = c.IdleTimeout
	}

	// Map merging - file values override defaults
	m.Headers = utils.MergeMaps(defaults.Headers, c.Headers)
	m.Query = utils.MergeMaps(defaults.Query, c.Query)

	// Body/Variables - file values override if present
	m.Body = utils.DeepCloneMap(defaults.Body)
	if c.Body != nil {
		m.Body = c.Body
	}

	m.Variables = utils.DeepCloneMap(defaults.Variables)
	if c.Variables != nil {
		m.Variables = c.Variables
	}

	// Preserve file-specific fields
	m.Expect = c.Expect
	m.Chain = c.Chain

	return m
}

// Expectation defines assertions for a chain step
type Expectation struct {
	Status any          `yaml:"status,omitempty"` // int or []int
	Assert AssertionSet `yaml:"assert,omitempty"` // JQ expressions that must evaluate to true
}

// AssertionSet represents assertions that can be either a flat list or grouped by context
type AssertionSet struct {
	Body    []string // Assertions on response body (default context)
	Headers []string // Assertions on response headers
}

// UnmarshalYAML implements custom unmarshaling for AssertionSet to support both:
// - Flat array: assert: [...]  (all treated as body assertions)
// - Grouped map: assert: { headers: [...], body: [...] }
func (a *AssertionSet) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try to unmarshal as array first (backward compatible)
	var flatList []string
	if err := unmarshal(&flatList); err == nil {
		a.Body = flatList
		return nil
	}

	// Try to unmarshal as map (grouped assertions)
	var grouped map[string][]string
	if err := unmarshal(&grouped); err != nil {
		return err
	}

	a.Headers = grouped["headers"]
	a.Body = grouped["body"]
	return nil
}

// ToDomain converts V1 YAML to the Canonical Config
func (c *ConfigV1) ToDomain() (*domain.Request, error) {
	c.expandEnvVars()
	return c.toDomainInternal()
}

// ToDomainWithResolver converts V1 YAML using a custom resolver for variable expansion
func (c *ConfigV1) ToDomainWithResolver(resolver vars.Resolver) (*domain.Request, error) {
	c.ExpandWithResolver(resolver)
	return c.toDomainInternal()
}

// toDomainInternal is the shared conversion logic (assumes variables are already expanded)
func (c *ConfigV1) toDomainInternal() (*domain.Request, error) {
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

// expandEnvVars expands environment variables in all string fields using reflection
func (c *ConfigV1) expandEnvVars() {
	vars.ExpandAll(c, vars.EnvResolver)
}

// ExpandWithResolver expands environment variables using a custom resolver
func (c *ConfigV1) ExpandWithResolver(resolver vars.Resolver) {
	vars.ExpandAll(c, resolver)
}

// setDefaults applies default values for Method
func (c *ConfigV1) setDefaults() {
	if c.Method == "" {
		c.Method = constants.MethodGET
	}
	c.Method = constants.CanonicalizeMethod(c.Method)
}

// prepareBody processes the body/json/form fields and returns a reader, source identifier, and any error
func (c *ConfigV1) prepareBody() (io.Reader, string, error) {
	// Check for mutually exclusive body fields
	bodyFieldCount := 0
	if c.JSON != "" {
		bodyFieldCount++
	}
	if len(c.Body) > 0 {
		bodyFieldCount++
	}
	if len(c.Form) > 0 {
		bodyFieldCount++
	}
	if bodyFieldCount > 1 {
		return nil, "", fmt.Errorf("`body`, `json`, and `form` are mutually exclusive")
	}

	// Handle JSON string
	if c.JSON != "" {
		if c.ContentType == "" {
			c.ContentType = "application/json"
		}
		return strings.NewReader(c.JSON), "json", nil
	}

	// Handle JSON object
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

	// Handle form data
	if len(c.Form) > 0 {
		// Default to urlencoded if no content type specified
		if c.ContentType == "" {
			c.ContentType = "application/x-www-form-urlencoded"
		}

		// Use multipart for multipart/form-data
		if strings.Contains(c.ContentType, "multipart/form-data") {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			for k, v := range c.Form {
				if err := writer.WriteField(k, v); err != nil {
					return nil, "", fmt.Errorf("failed to write form field %s: %w", k, err)
				}
			}

			if err := writer.Close(); err != nil {
				return nil, "", fmt.Errorf("failed to close multipart writer: %w", err)
			}

			// Update content type to include boundary
			c.ContentType = writer.FormDataContentType()
			return &buf, "form", nil
		}

		// Use URL encoding for urlencoded or unknown content types (fallback)
		formValues := url.Values{}
		for k, v := range c.Form {
			formValues.Set(k, v)
		}
		return strings.NewReader(formValues.Encode()), "form", nil
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

// enrichMetadata adds transport-specific metadata to the request
func (c *ConfigV1) enrichMetadata(req *domain.Request) error {
	transport := domain.DetectTransport(c.URL, c.Graphql != "")
	req.Metadata["transport"] = transport
	req.Metadata["insecure"] = fmt.Sprintf("%t", c.Insecure)

	switch transport {
	case constants.TransportGRPC:
		req.Metadata["service"] = c.Service
		req.Metadata["rpc"] = c.RPC
		req.Metadata["proto"] = c.Proto
		req.Metadata["proto_path"] = c.ProtoPath
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

	if c.OutputFile != "" {
		req.Metadata["output_file"] = c.OutputFile
	}

	if c.Timeout != "" {
		req.Metadata["timeout"] = c.Timeout
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
