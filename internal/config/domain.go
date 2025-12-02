package config

import "yapi.run/cli/internal/envsubst"

// Config is the Internal Representation (IR).
// The Runner consumes THIS, not the YAML struct directly.
type Config struct {
	// Meta
	Version string

	// Request
	URL         string
	Method      string
	Headers     map[string]string
	Body        interface{} // Processed body
	Query       map[string]string
	Path        string
	ContentType string
	Graphql     string                 // GraphQL query/mutation
	Variables   map[string]interface{} // GraphQL variables

	// Transport specific
	Transport      string // http, grpc, tcp
	Timeout        int
	Service        string // gRPC
	RPC            string // gRPC
	Proto          string // gRPC
	ProtoPath      string
	Data           string // TCP raw data
	Encoding       string // text, hex, base64
	JQFilter       string
	Insecure       bool // For gRPC
	Plaintext      bool // For gRPC
	ReadTimeout    int  // TCP read timeout in seconds
	IdleTimeout    int  // TCP idle timeout in milliseconds (default 500)
	CloseAfterSend bool
}

// SubstituteEnvVars replaces all ${VAR_NAME} patterns with environment variable values.
func (c *Config) SubstituteEnvVars() {
	c.URL = envsubst.Substitute(c.URL)
	c.Path = envsubst.Substitute(c.Path)
	c.ContentType = envsubst.Substitute(c.ContentType)
	c.Graphql = envsubst.Substitute(c.Graphql)
	c.Service = envsubst.Substitute(c.Service)
	c.RPC = envsubst.Substitute(c.RPC)
	c.Proto = envsubst.Substitute(c.Proto)
	c.ProtoPath = envsubst.Substitute(c.ProtoPath)
	c.Data = envsubst.Substitute(c.Data)
	c.JQFilter = envsubst.Substitute(c.JQFilter)

	// Substitute in headers map
	for k, v := range c.Headers {
		c.Headers[k] = envsubst.Substitute(v)
	}

	// Substitute in query map
	for k, v := range c.Query {
		c.Query[k] = envsubst.Substitute(v)
	}

	// Substitute string values in body map (recursive)
	if bodyMap, ok := c.Body.(map[string]interface{}); ok {
		substituteMapValues(bodyMap)
	}

	// Substitute string values in variables map (recursive)
	substituteMapValues(c.Variables)
}

// substituteMapValues recursively substitutes env vars in string values within a map
func substituteMapValues(m map[string]interface{}) {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			m[k] = envsubst.Substitute(val)
		case map[string]interface{}:
			substituteMapValues(val)
		case []interface{}:
			substituteSliceValues(val)
		}
	}
}

// substituteSliceValues recursively substitutes env vars in string values within a slice
func substituteSliceValues(s []interface{}) {
	for i, v := range s {
		switch val := v.(type) {
		case string:
			s[i] = envsubst.Substitute(val)
		case map[string]interface{}:
			substituteMapValues(val)
		case []interface{}:
			substituteSliceValues(val)
		}
	}
}
