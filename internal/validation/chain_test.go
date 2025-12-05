package validation

import (
	"strings"
	"testing"
)

func TestAnalyzeConfig_ValidChain(t *testing.T) {
	yaml := `yapi: v1
chain:
  - name: get_todo
    url: https://jsonplaceholder.typicode.com/todos/1
    method: GET
    expect:
      status: 200

  - name: get_user
    url: https://jsonplaceholder.typicode.com/users/${get_todo.userId}
    method: GET
    expect:
      status: 200`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if a.HasErrors() {
		t.Errorf("expected no errors for valid chain, got %+v", a.Diagnostics)
	}

	if len(a.Chain) != 2 {
		t.Errorf("expected 2 chain steps, got %d", len(a.Chain))
	}

	if a.Request != nil {
		t.Error("expected Request to be nil for chain config")
	}
}

func TestAnalyzeConfig_ChainMissingName(t *testing.T) {
	yaml := `yapi: v1
chain:
  - url: https://example.com
    method: GET`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for chain step missing name")
	}

	found := false
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "missing 'name'") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected error about missing name, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_ChainMissingURL(t *testing.T) {
	yaml := `yapi: v1
chain:
  - name: step1
    method: GET`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for chain step missing url")
	}

	found := false
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "missing 'url'") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected error about missing url, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_ChainDuplicateName(t *testing.T) {
	yaml := `yapi: v1
chain:
  - name: step1
    url: https://example.com/1
  - name: step1
    url: https://example.com/2`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for duplicate step name")
	}

	found := false
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "duplicate step name") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected error about duplicate step name, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_ChainForwardReference(t *testing.T) {
	yaml := `yapi: v1
chain:
  - name: step1
    url: https://example.com/${step2.token}
  - name: step2
    url: https://example.com/auth`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for forward reference")
	}

	found := false
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "references 'step2' before it is defined") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected error about forward reference, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_ChainSelfReference(t *testing.T) {
	yaml := `yapi: v1
chain:
  - name: step1
    url: https://example.com/${step1.token}`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for self reference")
	}

	found := false
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "cannot reference itself") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected error about self reference, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_ChainValidBackReference(t *testing.T) {
	yaml := `yapi: v1
chain:
  - name: auth
    url: https://example.com/login
    method: POST
  - name: getData
    url: https://example.com/data
    headers:
      Authorization: Bearer ${auth.access_token}`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	// Should not have errors for valid back reference
	if a.HasErrors() {
		t.Errorf("expected no errors for valid back reference, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_ChainWithEnvVarRef(t *testing.T) {
	yaml := `yapi: v1
chain:
  - name: step1
    url: https://example.com/$API_KEY
  - name: step2
    url: https://example.com/${ENV_VAR}`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	// Env vars (no dot) should not be flagged as undefined chain refs
	if a.HasErrors() {
		t.Errorf("expected no errors for env var references, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_ChainHeaderForwardRef(t *testing.T) {
	yaml := `yapi: v1
chain:
  - name: step1
    url: https://example.com
    headers:
      Authorization: Bearer ${step2.token}
  - name: step2
    url: https://example.com/auth`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for forward reference in headers")
	}
}

func TestAnalyzeConfig_ChainExpectations(t *testing.T) {
	yaml := `yapi: v1
chain:
  - name: step1
    url: https://example.com
    expect:
      status: 200
      body_contains: "success"`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	if a.HasErrors() {
		t.Errorf("expected no errors for chain with expectations, got %+v", a.Diagnostics)
	}

	if len(a.Chain) != 1 {
		t.Errorf("expected 1 chain step, got %d", len(a.Chain))
	}
}
