package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type YapiConfig struct {
	URL            string                 `yaml:"url"`
	Path           string                 `yaml:"path,omitempty"`
	Method         string                 `yaml:"method,omitempty"` // GET, POST, grpc, tcp
	ContentType    string                 `yaml:"content_type,omitempty"`
	Body           map[string]interface{} `yaml:"body,omitempty"`
	JSON           string                 `yaml:"json,omitempty"` // Raw JSON override
	Query          map[string]string      `yaml:"query,omitempty"`
	Service        string                 `yaml:"service,omitempty"` // gRPC
	RPC            string                 `yaml:"rpc,omitempty"`     // gRPC
	Proto          string                 `yaml:"proto,omitempty"`   // gRPC
	ProtoPath      string                 `yaml:"proto_path,omitempty"`
	Data           string                 `yaml:"data,omitempty"`     // TCP raw data
	Encoding       string                 `yaml:"encoding,omitempty"` // text, hex, base64
	JQFilter       string                 `yaml:"jq_filter,omitempty"`
	Insecure       bool                   `yaml:"insecure,omitempty"`     // For gRPC
	Plaintext      bool                   `yaml:"plaintext,omitempty"`    // For gRPC
	ReadTimeout    int                    `yaml:"read_timeout,omitempty"` // TCP read timeout in seconds
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
