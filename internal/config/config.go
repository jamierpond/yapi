package config

import (
	"os"

	"gopkg.in/yaml.v3"
	"yapi.run/internal/envsubst"
)

type YapiConfig struct {
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

func LoadConfig(path string) (*YapiConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg YapiConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SubstituteEnvVars replaces all ${VAR_NAME} patterns with environment variable values.
func (c *YapiConfig) SubstituteEnvVars() {
	c.URL = envsubst.Substitute(c.URL)
	c.Path = envsubst.Substitute(c.Path)
	c.ContentType = envsubst.Substitute(c.ContentType)
	c.JSON = envsubst.Substitute(c.JSON)
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
	substituteMapValues(c.Body)

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
