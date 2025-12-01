package validation

import (
	"strings"
	"testing"

	"cli/internal/config"
)

func TestValidateConfig_MissingURL(t *testing.T) {
	cfg := &config.YapiConfig{}
	issues := ValidateConfig(cfg)

	if len(issues) == 0 {
		t.Fatal("expected at least one issue for missing URL")
	}

	found := false
	for _, issue := range issues {
		if issue.Field == "url" && issue.Severity == SeverityError {
			found = true
			if !strings.Contains(issue.Message, "missing required field") {
				t.Errorf("expected message about missing url, got: %s", issue.Message)
			}
		}
	}
	if !found {
		t.Error("expected error for missing url field")
	}
}

func TestValidateConfig_ValidHTTPMethods(t *testing.T) {
	validMethods := []string{"", "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range validMethods {
		cfg := &config.YapiConfig{
			URL:    "http://example.com",
			Method: method,
		}
		issues := ValidateConfig(cfg)

		for _, issue := range issues {
			if issue.Field == "method" {
				t.Errorf("unexpected method issue for %q: %s", method, issue.Message)
			}
		}
	}
}

func TestValidateConfig_UnknownHTTPMethod(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:    "http://example.com",
		Method: "FOOBAR",
	}
	issues := ValidateConfig(cfg)

	found := false
	for _, issue := range issues {
		if issue.Field == "method" && issue.Severity == SeverityWarning {
			found = true
			if !strings.Contains(issue.Message, "unknown HTTP method") {
				t.Errorf("expected unknown method message, got: %s", issue.Message)
			}
		}
	}
	if !found {
		t.Error("expected warning for unknown HTTP method")
	}
}

func TestValidateConfig_GRPCDeprecated(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:     "localhost:50051",
		Method:  "grpc",
		Service: "example.Service",
		RPC:     "GetData",
	}
	issues := ValidateConfig(cfg)

	foundDeprecation := false
	for _, issue := range issues {
		if issue.Field == "method" && issue.Severity == SeverityWarning {
			foundDeprecation = true
			if !strings.Contains(issue.Message, "deprecated") {
				t.Errorf("expected deprecation warning, got: %s", issue.Message)
			}
		}
	}
	if !foundDeprecation {
		t.Error("expected deprecation warning for grpc method")
	}
}

func TestValidateConfig_GRPCMissingService(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:    "localhost:50051",
		Method: "grpc",
		RPC:    "GetData",
	}
	issues := ValidateConfig(cfg)

	found := false
	for _, issue := range issues {
		if issue.Field == "service" && issue.Severity == SeverityError {
			found = true
			if !strings.Contains(issue.Message, "gRPC config requires `service`") {
				t.Errorf("expected service required message, got: %s", issue.Message)
			}
		}
	}
	if !found {
		t.Error("expected error for missing service in gRPC config")
	}
}

func TestValidateConfig_GRPCMissingRPC(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:     "localhost:50051",
		Method:  "grpc",
		Service: "example.Service",
	}
	issues := ValidateConfig(cfg)

	found := false
	for _, issue := range issues {
		if issue.Field == "rpc" && issue.Severity == SeverityError {
			found = true
			if !strings.Contains(issue.Message, "gRPC config requires `rpc`") {
				t.Errorf("expected rpc required message, got: %s", issue.Message)
			}
		}
	}
	if !found {
		t.Error("expected error for missing rpc in gRPC config")
	}
}

func TestValidateConfig_TCPDeprecated(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:    "localhost:9000",
		Method: "tcp",
		Data:   "hello",
	}
	issues := ValidateConfig(cfg)

	foundDeprecation := false
	for _, issue := range issues {
		if issue.Field == "method" && issue.Severity == SeverityWarning {
			foundDeprecation = true
			if !strings.Contains(issue.Message, "deprecated") {
				t.Errorf("expected deprecation warning, got: %s", issue.Message)
			}
		}
	}
	if !foundDeprecation {
		t.Error("expected deprecation warning for tcp method")
	}
}

func TestValidateConfig_TCPValidEncodings(t *testing.T) {
	validEncodings := []string{"text", "hex", "base64"}

	for _, enc := range validEncodings {
		cfg := &config.YapiConfig{
			URL:      "localhost:9000",
			Method:   "tcp",
			Data:     "hello",
			Encoding: enc,
		}
		issues := ValidateConfig(cfg)

		for _, issue := range issues {
			if issue.Field == "encoding" {
				t.Errorf("unexpected encoding issue for %q: %s", enc, issue.Message)
			}
		}
	}
}

func TestValidateConfig_TCPInvalidEncoding(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:      "localhost:9000",
		Method:   "tcp",
		Data:     "hello",
		Encoding: "invalid",
	}
	issues := ValidateConfig(cfg)

	found := false
	for _, issue := range issues {
		if issue.Field == "encoding" && issue.Severity == SeverityError {
			found = true
			if !strings.Contains(issue.Message, "unsupported TCP encoding") {
				t.Errorf("expected unsupported encoding message, got: %s", issue.Message)
			}
		}
	}
	if !found {
		t.Error("expected error for invalid TCP encoding")
	}
}

func TestValidateConfig_ValidConfig(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:         "http://example.com/api",
		Method:      "POST",
		ContentType: "application/json",
		Body:        map[string]interface{}{"key": "value"},
	}
	issues := ValidateConfig(cfg)

	if len(issues) != 0 {
		t.Errorf("expected no issues for valid config, got %d: %+v", len(issues), issues)
	}
}

func TestValidateConfig_BodyAndJSONMutuallyExclusive(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:         "http://example.com/api",
		Method:      "POST",
		ContentType: "application/json",
		Body:        map[string]interface{}{"key": "value"},
		JSON:        `{"key": "value"}`,
	}
	issues := ValidateConfig(cfg)

	found := false
	for _, issue := range issues {
		if issue.Field == "body" && issue.Severity == SeverityError {
			found = true
			if !strings.Contains(issue.Message, "mutually exclusive") {
				t.Errorf("expected mutually exclusive message, got: %s", issue.Message)
			}
		}
	}
	if !found {
		t.Error("expected error for body and json both present")
	}
}

func TestValidateConfig_ContentTypeRequiredWithBody(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:    "http://example.com/api",
		Method: "POST",
		Body:   map[string]interface{}{"key": "value"},
	}
	issues := ValidateConfig(cfg)

	found := false
	for _, issue := range issues {
		if issue.Field == "content_type" && issue.Severity == SeverityError {
			found = true
			if !strings.Contains(issue.Message, "content_type") {
				t.Errorf("expected content_type required message, got: %s", issue.Message)
			}
		}
	}
	if !found {
		t.Error("expected error for missing content_type with body")
	}
}

func TestValidateConfig_ContentTypeRequiredWithJSON(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:    "http://example.com/api",
		Method: "POST",
		JSON:   `{"key": "value"}`,
	}
	issues := ValidateConfig(cfg)

	found := false
	for _, issue := range issues {
		if issue.Field == "content_type" && issue.Severity == SeverityError {
			found = true
		}
	}
	if !found {
		t.Error("expected error for missing content_type with json")
	}
}

func TestValidateConfig_ContentTypeNotRequiredForGRPC(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:     "grpc://localhost:50051",
		Service: "example.Service",
		RPC:     "GetData",
		Body:    map[string]interface{}{"key": "value"},
	}
	issues := ValidateConfig(cfg)

	for _, issue := range issues {
		if issue.Field == "content_type" {
			t.Errorf("unexpected content_type issue for gRPC: %s", issue.Message)
		}
	}
}

func TestValidateConfig_ContentTypeNotRequiredForTCP(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:  "tcp://localhost:9000",
		Body: map[string]interface{}{"key": "value"},
	}
	issues := ValidateConfig(cfg)

	for _, issue := range issues {
		if issue.Field == "content_type" {
			t.Errorf("unexpected content_type issue for TCP: %s", issue.Message)
		}
	}
}

func TestValidateConfig_GRPCByURLScheme(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"grpc scheme", "grpc://localhost:50051"},
		{"grpcs scheme", "grpcs://api.example.com:443"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.YapiConfig{
				URL: tt.url,
			}
			issues := ValidateConfig(cfg)

			foundService := false
			foundRPC := false
			for _, issue := range issues {
				if issue.Field == "service" && issue.Severity == SeverityError {
					foundService = true
				}
				if issue.Field == "rpc" && issue.Severity == SeverityError {
					foundRPC = true
				}
			}
			if !foundService {
				t.Error("expected error for missing service with gRPC URL scheme")
			}
			if !foundRPC {
				t.Error("expected error for missing rpc with gRPC URL scheme")
			}
		})
	}
}

func TestValidateConfig_GRPCByURLSchemeValid(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:     "grpc://localhost:50051",
		Service: "example.Service",
		RPC:     "GetData",
	}
	issues := ValidateConfig(cfg)

	for _, issue := range issues {
		if issue.Field == "service" || issue.Field == "rpc" {
			t.Errorf("unexpected issue for valid gRPC config: %s", issue.Message)
		}
	}
}

func TestValidateConfig_TCPByURLScheme(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:      "tcp://localhost:9000",
		Encoding: "invalid",
	}
	issues := ValidateConfig(cfg)

	found := false
	for _, issue := range issues {
		if issue.Field == "encoding" && issue.Severity == SeverityError {
			found = true
		}
	}
	if !found {
		t.Error("expected encoding validation for TCP URL scheme")
	}
}

func TestValidateConfig_TCPByURLSchemeValidEncoding(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:      "tcp://localhost:9000",
		Encoding: "hex",
	}
	issues := ValidateConfig(cfg)

	for _, issue := range issues {
		if issue.Field == "encoding" {
			t.Errorf("unexpected encoding issue: %s", issue.Message)
		}
	}
}

func TestValidateConfig_LowercaseHTTPMethods(t *testing.T) {
	methods := []string{"get", "post", "put", "delete", "patch", "head", "options"}

	for _, method := range methods {
		cfg := &config.YapiConfig{
			URL:    "http://example.com",
			Method: method,
		}
		issues := ValidateConfig(cfg)

		for _, issue := range issues {
			if issue.Field == "method" {
				t.Errorf("unexpected method issue for lowercase %q: %s", method, issue.Message)
			}
		}
	}
}

func TestValidateConfig_EmptyBodyNotTriggersContentTypeRule(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:    "http://example.com",
		Method: "GET",
		Body:   map[string]interface{}{},
	}
	issues := ValidateConfig(cfg)

	for _, issue := range issues {
		if issue.Field == "content_type" {
			t.Errorf("unexpected content_type issue for empty body: %s", issue.Message)
		}
	}
}

func TestValidateConfig_NoIssuesForMinimalValidHTTP(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:    "http://example.com",
		Method: "GET",
	}
	issues := ValidateConfig(cfg)

	if len(issues) != 0 {
		t.Errorf("expected no issues for minimal valid HTTP config, got %d: %+v", len(issues), issues)
	}
}

func TestValidateConfig_NoIssuesForMinimalValidGRPC(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:     "grpc://localhost:50051",
		Service: "example.Service",
		RPC:     "GetData",
	}
	issues := ValidateConfig(cfg)

	if len(issues) != 0 {
		t.Errorf("expected no issues for minimal valid gRPC config, got %d: %+v", len(issues), issues)
	}
}

func TestValidateConfig_NoIssuesForMinimalValidTCP(t *testing.T) {
	cfg := &config.YapiConfig{
		URL:  "tcp://localhost:9000",
		Data: "hello",
	}
	issues := ValidateConfig(cfg)

	if len(issues) != 0 {
		t.Errorf("expected no issues for minimal valid TCP config, got %d: %+v", len(issues), issues)
	}
}
