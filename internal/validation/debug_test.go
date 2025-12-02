package validation

import (
	"strings"
	"testing"
)

func TestDiagnosticLineNumbers(t *testing.T) {
	yaml := `yapi: v1
# Comment
url: http://example.com/graphql

graphql: query { foo }
body:
  key: value`

	a, err := AnalyzeConfigString(yaml)
	if err != nil {
		t.Fatalf("AnalyzeConfigString error: %v", err)
	}

	// Find the graphql+body conflict diagnostic
	var graphqlDiag *Diagnostic
	for i := range a.Diagnostics {
		if strings.Contains(a.Diagnostics[i].Message, "cannot be used with") {
			graphqlDiag = &a.Diagnostics[i]
			break
		}
	}

	if graphqlDiag == nil {
		t.Fatal("expected graphql+body conflict diagnostic")
	}

	// body: is on line 5 (0-indexed)
	if graphqlDiag.Line != 5 {
		t.Errorf("expected body diagnostic on line 5, got %d", graphqlDiag.Line)
	}
	if graphqlDiag.Field != "body" {
		t.Errorf("expected field 'body', got %q", graphqlDiag.Field)
	}
}
