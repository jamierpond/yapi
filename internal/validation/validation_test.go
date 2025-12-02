package validation

import (
	"fmt"
	"strings"
	"testing"

	"yapi.run/cli/internal/config"
)

func TestValidateRequest_MissingURL(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

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

func TestValidateRequest_ValidHTTPMethods(t *testing.T) {
	validMethods := []string{"", "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range validMethods {
		yaml := fmt.Sprintf(`yapi: v1
url: http://example.com
method: %s`, method)
		res, err := config.LoadFromString(yaml)
		if err != nil {
			t.Fatalf("unexpected error loading config for method %s: %v", method, err)
		}
		issues := ValidateRequest(res.Request)

		for _, issue := range issues {
			if issue.Field == "method" {
				t.Errorf("unexpected method issue for %q: %s", method, issue.Message)
			}
		}
	}
}

func TestValidateRequest_UnknownHTTPMethod(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1
url: http://example.com
method: FOOBAR`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

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

func TestValidateRequest_GRPCMissingService(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1
url: grpc://localhost:50051
method: grpc
rpc: GetData`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

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

func TestValidateRequest_GRPCMissingRPC(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1
url: grpc://localhost:50051
method: grpc
service: example.Service`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

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

func TestValidateRequest_TCPValidEncodings(t *testing.T) {
	validEncodings := []string{"text", "hex", "base64"}

	for _, enc := range validEncodings {
		yaml := fmt.Sprintf(`yapi: v1
url: tcp://localhost:9000
method: tcp
data: hello
encoding: %s`, enc)
		res, err := config.LoadFromString(yaml)
		if err != nil {
			t.Fatalf("unexpected error loading config for encoding %s: %v", enc, err)
		}
		issues := ValidateRequest(res.Request)

		for _, issue := range issues {
			if issue.Field == "encoding" {
				t.Errorf("unexpected encoding issue for %q: %s", enc, issue.Message)
			}
		}
	}
}

func TestValidateRequest_TCPInvalidEncoding(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1
url: tcp://localhost:9000
method: tcp
data: hello
encoding: invalid`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

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

func TestValidateRequest_ValidConfig(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1
url: http://example.com/api
method: POST
content_type: application/json
body:
  key: value`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

	if len(issues) != 0 {
		t.Errorf("expected no issues for valid config, got %d: %+v", len(issues), issues)
	}
}

func TestValidateRequest_GRPCByURLSchemeValid(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1
url: grpc://localhost:50051
service: example.Service
rpc: GetData`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

	for _, issue := range issues {
		if issue.Field == "service" || issue.Field == "rpc" {
			t.Errorf("unexpected issue for valid gRPC config: %s", issue.Message)
		}
	}
}

func TestValidateRequest_NoIssuesForMinimalValidHTTP(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1
url: http://example.com
method: GET`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

	if len(issues) != 0 {
		t.Errorf("expected no issues for minimal valid HTTP config, got %d: %+v", len(issues), issues)
	}
}

func TestValidateRequest_NoIssuesForMinimalValidGRPC(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1
url: grpc://localhost:50051
service: example.Service
rpc: GetData`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

	if len(issues) != 0 {
		t.Errorf("expected no issues for minimal valid gRPC config, got %d: %+v", len(issues), issues)
	}
}

func TestValidateRequest_GraphQLOnly(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1
url: http://example.com/graphql
graphql: 'query { foo }'`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

	if len(issues) != 0 {
		t.Errorf("expected no issues for graphql-only config, got %d: %+v", len(issues), issues)
	}
}

func TestValidateRequest_NoIssuesForMinimalValidTCP(t *testing.T) {
	res, err := config.LoadFromString(`yapi: v1
url: tcp://localhost:9000
data: hello`)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	issues := ValidateRequest(res.Request)

	if len(issues) != 0 {
		t.Errorf("expected no issues for minimal valid TCP config, got %d: %+v", len(issues), issues)
	}
}
