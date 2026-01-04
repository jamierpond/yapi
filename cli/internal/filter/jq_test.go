package filter

import (
	"strings"
	"testing"
)

func TestApplyJQ(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		filterExpr string
		want       string
		wantErr    bool
	}{
		{
			name:       "empty filter returns input unchanged",
			input:      `{"foo": "bar"}`,
			filterExpr: "",
			want:       `{"foo": "bar"}`,
		},
		{
			name:       "whitespace-only filter returns input unchanged",
			input:      `{"foo": "bar"}`,
			filterExpr: "   ",
			want:       `{"foo": "bar"}`,
		},
		{
			name:       "simple field access",
			input:      `{"foo": 128}`,
			filterExpr: ".foo",
			want:       "128",
		},
		{
			name:       "nested field access",
			input:      `{"a": {"b": 42}}`,
			filterExpr: ".a.b",
			want:       "42",
		},
		{
			name:       "string field access returns unquoted string",
			input:      `{"name": "hello"}`,
			filterExpr: ".name",
			want:       "hello",
		},
		{
			name:       "object construction",
			input:      `{"id": "sample", "10": {"b": 42}}`,
			filterExpr: `{(.id): .["10"].b}`,
			want:       "{\n  \"sample\": 42\n}",
		},
		{
			name:       "array iteration",
			input:      `[{"id":1},{"id":2},{"id":3}]`,
			filterExpr: ".[] | .id",
			want:       "1\n2\n3",
		},
		{
			name:       "arithmetic operations",
			input:      `{"a":1,"b":2}`,
			filterExpr: ".a += 1 | .b *= 2",
			want:       "{\n  \"a\": 2,\n  \"b\": 4\n}",
		},
		{
			name:       "filter with map",
			input:      `{"samples": [{"name": "a", "value": 1}, {"name": "b", "value": 2}]}`,
			filterExpr: ".samples | map({name})",
			want:       "[\n  {\n    \"name\": \"a\"\n  },\n  {\n    \"name\": \"b\"\n  }\n]",
		},
		{
			name:       "select filter",
			input:      `[1, 2, 3, 4, 5]`,
			filterExpr: ".[] | select(. > 2)",
			want:       "3\n4\n5",
		},
		{
			name:       "keys function",
			input:      `{"b": 1, "a": 2, "c": 3}`,
			filterExpr: "keys",
			want:       "[\n  \"a\",\n  \"b\",\n  \"c\"\n]",
		},
		{
			name:       "length function",
			input:      `[1, 2, 3, 4]`,
			filterExpr: "length",
			want:       "4",
		},
		{
			name:       "null value",
			input:      `{"foo": null}`,
			filterExpr: ".foo",
			want:       "null",
		},
		{
			name:       "boolean true",
			input:      `{"active": true}`,
			filterExpr: ".active",
			want:       "true",
		},
		{
			name:       "boolean false",
			input:      `{"active": false}`,
			filterExpr: ".active",
			want:       "false",
		},
		{
			name:       "invalid jq filter syntax",
			input:      `{"foo": 1}`,
			filterExpr: ".foo & .bar",
			wantErr:    true,
		},
		{
			name:       "invalid JSON input",
			input:      `{foo: bar}`,
			filterExpr: ".foo",
			wantErr:    true,
		},
		{
			name:       "access non-existent field returns null",
			input:      `{"foo": 1}`,
			filterExpr: ".bar",
			want:       "null",
		},
		{
			name:       "pipe chain",
			input:      `{"data": {"items": [{"x": 1}, {"x": 2}]}}`,
			filterExpr: ".data.items | map(.x) | add",
			want:       "3",
		},
		{
			name:       "type function",
			input:      `{"arr": [], "obj": {}, "num": 1, "str": "hi"}`,
			filterExpr: ".arr | type",
			want:       "array",
		},
		{
			name:       "update assignment",
			input:      `{"json": {"nested": {"foo": "bar"}}}`,
			filterExpr: ".json.nested",
			want:       "{\n  \"foo\": \"bar\"\n}",
		},
		{
			name:       "httpbin-style filter",
			input:      `{"json": {"samples": [{"name": "sample1", "value": 1}, {"name": "sample2", "value": 2}]}}`,
			filterExpr: ".json.samples |= map({name})",
			want:       "{\n  \"json\": {\n    \"samples\": [\n      {\n        \"name\": \"sample1\"\n      },\n      {\n        \"name\": \"sample2\"\n      }\n    ]\n  }\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyJQ(tt.input, tt.filterExpr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyJQ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ApplyJQ() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyJQ_LargeNumbers(t *testing.T) {
	// gojq supports arbitrary-precision integers
	input := `{"big": 4722366482869645213696}`
	got, err := ApplyJQ(input, ".big")
	if err != nil {
		t.Fatalf("ApplyJQ() error = %v", err)
	}
	if !strings.Contains(got, "4722366482869645213696") {
		t.Errorf("ApplyJQ() = %q, expected large number to be preserved", got)
	}
}

func TestApplyJQ_EmptyResult(t *testing.T) {
	// Empty filter result from select that matches nothing
	input := `[1, 2, 3]`
	got, err := ApplyJQ(input, ".[] | select(. > 10)")
	if err != nil {
		t.Fatalf("ApplyJQ() error = %v", err)
	}
	if got != "" {
		t.Errorf("ApplyJQ() = %q, want empty string", got)
	}
}

func TestApplyJQ_MultipleResults(t *testing.T) {
	input := `{"items":[{"name":"foo","stars":1},{"name":"bar","stars":2},{"name":"baz","stars":3}]}`
	filter := ".items[] | {name: .name, stars: .stars}"
	result, err := ApplyJQ(input, filter)
	if err != nil {
		t.Fatalf("ApplyJQ() error = %v", err)
	}
	expected := `{
  "name": "foo",
  "stars": 1
}
{
  "name": "bar",
  "stars": 2
}
{
  "name": "baz",
  "stars": 3
}`
	if result != expected {
		t.Errorf("ApplyJQ with multiple results:\ngot:\n%s\n\nexpected:\n%s", result, expected)
	}
}

func FuzzApplyJQ(f *testing.F) {
	// Seed with valid JSON + jq filter pairs
	f.Add(`{"foo": "bar"}`, ".foo")
	f.Add(`{"a": {"b": 42}}`, ".a.b")
	f.Add(`[1, 2, 3, 4, 5]`, ".[] | select(. > 2)")
	f.Add(`{"items": [{"x": 1}, {"x": 2}]}`, ".items | map(.x)")
	f.Add(`{"big": 4722366482869645213696}`, ".big")
	f.Add(`null`, ".")
	f.Add(`"hello"`, ".")
	f.Add(`123`, ". + 1")
	f.Add(`[{"id":1},{"id":2}]`, ".[] | .id")
	f.Add(`{}`, "keys")

	f.Fuzz(func(t *testing.T, input string, filter string) {
		// ApplyJQ should not panic on any input
		_, _ = ApplyJQ(input, filter)
	})
}

func FuzzEvalJQBool(f *testing.F) {
	// Seed with valid JSON + jq boolean expressions
	f.Add(`{"status": 200}`, ".status == 200")
	f.Add(`{"items": [1,2,3]}`, ".items | length > 0")
	f.Add(`{"active": true}`, ".active")
	f.Add(`{"value": 10}`, ".value >= 5 and .value <= 15")
	f.Add(`{"name": "test"}`, `.name == "test"`)

	f.Fuzz(func(t *testing.T, input string, expr string) {
		// EvalJQBoolWithDetail should not panic on any input
		_, _, _ = EvalJQBoolWithDetail(input, expr)
	})
}

func TestEvalJQBoolWithDetail(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expr              string
		wantPassed        bool
		wantErr           bool
		wantLeftSide      string
		wantOperator      string
		wantRightSide     string
		wantActualValue   string
		wantExpectedValue string
	}{
		{
			name:              "simple equality - pass",
			input:             `{"id": 1}`,
			expr:              ".id == 1",
			wantPassed:        true,
			wantLeftSide:      ".id",
			wantOperator:      "==",
			wantRightSide:     "1",
			wantActualValue:   "1",
			wantExpectedValue: "1",
		},
		{
			name:              "simple equality - fail",
			input:             `{"id": 1}`,
			expr:              ".id == 999",
			wantPassed:        false,
			wantLeftSide:      ".id",
			wantOperator:      "==",
			wantRightSide:     "999",
			wantActualValue:   "1",
			wantExpectedValue: "999",
		},
		{
			name:              "not equal - pass",
			input:             `{"userId": 1}`,
			expr:              ".userId != null",
			wantPassed:        true,
			wantLeftSide:      ".userId",
			wantOperator:      "!=",
			wantRightSide:     "null",
			wantActualValue:   "1",
			wantExpectedValue: "null",
		},
		{
			name:              "not equal - fail",
			input:             `{"userId": 1}`,
			expr:              ".userId != 1",
			wantPassed:        false,
			wantLeftSide:      ".userId",
			wantOperator:      "!=",
			wantRightSide:     "1",
			wantActualValue:   "1",
			wantExpectedValue: "1",
		},
		{
			name:              "greater than - pass",
			input:             `{"count": 10}`,
			expr:              ".count > 5",
			wantPassed:        true,
			wantLeftSide:      ".count",
			wantOperator:      ">",
			wantRightSide:     "5",
			wantActualValue:   "10",
			wantExpectedValue: "5",
		},
		{
			name:              "greater than - fail",
			input:             `{"id": 1}`,
			expr:              ".id > 100",
			wantPassed:        false,
			wantLeftSide:      ".id",
			wantOperator:      ">",
			wantRightSide:     "100",
			wantActualValue:   "1",
			wantExpectedValue: "100",
		},
		{
			name:              "greater than or equal - pass",
			input:             `{"score": 10}`,
			expr:              ".score >= 10",
			wantPassed:        true,
			wantLeftSide:      ".score",
			wantOperator:      ">=",
			wantRightSide:     "10",
			wantActualValue:   "10",
			wantExpectedValue: "10",
		},
		{
			name:              "greater than or equal - fail",
			input:             `{"id": 1}`,
			expr:              ".id >= 10",
			wantPassed:        false,
			wantLeftSide:      ".id",
			wantOperator:      ">=",
			wantRightSide:     "10",
			wantActualValue:   "1",
			wantExpectedValue: "10",
		},
		{
			name:              "less than - pass",
			input:             `{"value": 5}`,
			expr:              ".value < 10",
			wantPassed:        true,
			wantLeftSide:      ".value",
			wantOperator:      "<",
			wantRightSide:     "10",
			wantActualValue:   "5",
			wantExpectedValue: "10",
		},
		{
			name:              "less than - fail",
			input:             `{"userId": 1}`,
			expr:              ".userId < 1",
			wantPassed:        false,
			wantLeftSide:      ".userId",
			wantOperator:      "<",
			wantRightSide:     "1",
			wantActualValue:   "1",
			wantExpectedValue: "1",
		},
		{
			name:              "less than or equal - pass",
			input:             `{"value": 5}`,
			expr:              ".value <= 5",
			wantPassed:        true,
			wantLeftSide:      ".value",
			wantOperator:      "<=",
			wantRightSide:     "5",
			wantActualValue:   "5",
			wantExpectedValue: "5",
		},
		{
			name:              "less than or equal - fail",
			input:             `{"value": 10}`,
			expr:              ".value <= 5",
			wantPassed:        false,
			wantLeftSide:      ".value",
			wantOperator:      "<=",
			wantRightSide:     "5",
			wantActualValue:   "10",
			wantExpectedValue: "5",
		},
		{
			name:              "complex expression with pipe",
			input:             `{"title": "delectus aut autem"}`,
			expr:              ".title | length > 100",
			wantPassed:        false,
			wantLeftSide:      ".title | length",
			wantOperator:      ">",
			wantRightSide:     "100",
			wantActualValue:   "18",
			wantExpectedValue: "100",
		},
		{
			name:              "boolean comparison",
			input:             `{"completed": false}`,
			expr:              ".completed == false",
			wantPassed:        true,
			wantLeftSide:      ".completed",
			wantOperator:      "==",
			wantRightSide:     "false",
			wantActualValue:   "false",
			wantExpectedValue: "false",
		},
		{
			name:              "null comparison - equal",
			input:             `{"userId": 1}`,
			expr:              ".userId == null",
			wantPassed:        false,
			wantLeftSide:      ".userId",
			wantOperator:      "==",
			wantRightSide:     "null",
			wantActualValue:   "1",
			wantExpectedValue: "null",
		},
		{
			name:              "string comparison",
			input:             `{"name": "test"}`,
			expr:              `.name == "test"`,
			wantPassed:        true,
			wantLeftSide:      ".name",
			wantOperator:      "==",
			wantRightSide:     `"test"`,
			wantActualValue:   `"test"`,
			wantExpectedValue: `"test"`,
		},
		{
			name:       "empty expression",
			input:      `{"id": 1}`,
			expr:       "",
			wantPassed: false,
			wantErr:    true,
		},
		{
			name:       "invalid JSON input",
			input:      `{invalid}`,
			expr:       ".id == 1",
			wantPassed: false,
			wantErr:    true,
		},
		{
			name:       "non-boolean result",
			input:      `{"id": 1}`,
			expr:       ".id",
			wantPassed: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, detail, err := EvalJQBoolWithDetail(tt.input, tt.expr)

			if (err != nil) != tt.wantErr {
				t.Errorf("EvalJQBoolWithDetail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if passed != tt.wantPassed {
				t.Errorf("EvalJQBoolWithDetail() passed = %v, want %v", passed, tt.wantPassed)
			}

			if detail == nil {
				t.Fatal("EvalJQBoolWithDetail() detail is nil")
			}

			if detail.Expression != tt.expr {
				t.Errorf("detail.Expression = %q, want %q", detail.Expression, tt.expr)
			}

			if tt.wantLeftSide != "" && detail.LeftSide != tt.wantLeftSide {
				t.Errorf("detail.LeftSide = %q, want %q", detail.LeftSide, tt.wantLeftSide)
			}

			if tt.wantOperator != "" && detail.Operator != tt.wantOperator {
				t.Errorf("detail.Operator = %q, want %q", detail.Operator, tt.wantOperator)
			}

			if tt.wantRightSide != "" && detail.RightSide != tt.wantRightSide {
				t.Errorf("detail.RightSide = %q, want %q", detail.RightSide, tt.wantRightSide)
			}

			if tt.wantActualValue != "" && detail.ActualValue != tt.wantActualValue {
				t.Errorf("detail.ActualValue = %q, want %q", detail.ActualValue, tt.wantActualValue)
			}

			if tt.wantExpectedValue != "" && detail.ExpectedValue != tt.wantExpectedValue {
				t.Errorf("detail.ExpectedValue = %q, want %q", detail.ExpectedValue, tt.wantExpectedValue)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "nil value",
			value: nil,
			want:  "null",
		},
		{
			name:  "string value",
			value: "hello",
			want:  `"hello"`,
		},
		{
			name:  "boolean true",
			value: true,
			want:  "true",
		},
		{
			name:  "boolean false",
			value: false,
			want:  "false",
		},
		{
			name:  "int value",
			value: 42,
			want:  "42",
		},
		{
			name:  "int64 value",
			value: int64(123),
			want:  "123",
		},
		{
			name:  "float64 value",
			value: float64(3.14),
			want:  "3.14",
		},
		{
			name:  "array value",
			value: []any{1, 2, 3},
			want:  "[1,2,3]",
		},
		{
			name:  "object value",
			value: map[string]any{"key": "value"},
			want:  `{"key":"value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.value)
			if got != tt.want {
				t.Errorf("formatValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyJQWithVars(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		filterExpr string
		variables  map[string]any
		want       string
		wantErr    bool
	}{
		{
			name:       "simple variable usage",
			input:      `{"id": 1}`,
			filterExpr: `.id == $_headers["Content-Type"]`,
			variables: map[string]any{
				"_headers": map[string]any{"Content-Type": "application/json"},
			},
			want:    "false",
			wantErr: false,
		},
		{
			name:       "access variable in filter",
			input:      `{"data": "test"}`,
			filterExpr: `$_headers`,
			variables: map[string]any{
				"_headers": map[string]any{"Content-Type": "application/json", "X-Custom": "value"},
			},
			want:    "{\n  \"Content-Type\": \"application/json\",\n  \"X-Custom\": \"value\"\n}",
			wantErr: false,
		},
		{
			name:       "combine input and variable",
			input:      `{"name": "test"}`,
			filterExpr: `.name + " " + $_headers["Content-Type"]`,
			variables: map[string]any{
				"_headers": map[string]any{"Content-Type": "application/json"},
			},
			want:    "test application/json",
			wantErr: false,
		},
		{
			name:       "no variables provided (nil)",
			input:      `{"id": 1}`,
			filterExpr: `.id`,
			variables:  nil,
			want:       "1",
			wantErr:    false,
		},
		{
			name:       "empty variables map",
			input:      `{"id": 1}`,
			filterExpr: `.id`,
			variables:  map[string]any{},
			want:       "1",
			wantErr:    false,
		},
		{
			name:       "multiple variables",
			input:      `{"value": 10}`,
			filterExpr: `$_a + $_b + .value`,
			variables: map[string]any{
				"_a": 5,
				"_b": 3,
			},
			want:    "18",
			wantErr: false,
		},
		{
			name:       "invalid filter with variables",
			input:      `{"id": 1}`,
			filterExpr: `.id &`,
			variables: map[string]any{
				"_headers": map[string]any{},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:       "empty filter with variables",
			input:      `{"id": 1}`,
			filterExpr: "",
			variables: map[string]any{
				"_headers": map[string]any{"Content-Type": "application/json"},
			},
			want:    `{"id": 1}`,
			wantErr: false,
		},
		{
			name:       "empty result with variables",
			input:      `[1, 2, 3]`,
			filterExpr: `.[] | select(. > 10)`,
			variables: map[string]any{
				"_headers": map[string]any{},
			},
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyJQWithVars(tt.input, tt.filterExpr, tt.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyJQWithVars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ApplyJQWithVars() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEvalJQBoolWithDetailAndVars(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expr              string
		variables         map[string]any
		wantPassed        bool
		wantErr           bool
		wantLeftSide      string
		wantOperator      string
		wantRightSide     string
		wantActualValue   string
		wantExpectedValue string
	}{
		{
			name:  "simple variable comparison",
			input: `{"status": "ok"}`,
			expr:  `$_headers["Content-Type"] == "application/json"`,
			variables: map[string]any{
				"_headers": map[string]any{"Content-Type": "application/json"},
			},
			wantPassed:        true,
			wantLeftSide:      `$_headers["Content-Type"]`,
			wantOperator:      "==",
			wantRightSide:     `"application/json"`,
			wantActualValue:   `"application/json"`,
			wantExpectedValue: `"application/json"`,
		},
		{
			name:  "variable comparison fails",
			input: `{"status": "ok"}`,
			expr:  `$_headers["Content-Type"] == "text/html"`,
			variables: map[string]any{
				"_headers": map[string]any{"Content-Type": "application/json"},
			},
			wantPassed:        false,
			wantLeftSide:      `$_headers["Content-Type"]`,
			wantOperator:      "==",
			wantRightSide:     `"text/html"`,
			wantActualValue:   `"application/json"`,
			wantExpectedValue: `"text/html"`,
		},
		{
			name:  "variable null check passes",
			input: `{}`,
			expr:  `$_headers["X-Custom"] != null`,
			variables: map[string]any{
				"_headers": map[string]any{"X-Custom": "value123"},
			},
			wantPassed:        true,
			wantLeftSide:      `$_headers["X-Custom"]`,
			wantOperator:      "!=",
			wantRightSide:     "null",
			wantActualValue:   `"value123"`,
			wantExpectedValue: "null",
		},
		{
			name:  "variable null check fails",
			input: `{}`,
			expr:  `$_headers["Missing"] != null`,
			variables: map[string]any{
				"_headers": map[string]any{"Content-Type": "application/json"},
			},
			wantPassed:        false,
			wantLeftSide:      `$_headers["Missing"]`,
			wantOperator:      "!=",
			wantRightSide:     "null",
			wantActualValue:   "null",
			wantExpectedValue: "null",
		},
		{
			name:  "no variables - regular expression",
			input: `{"id": 1}`,
			expr:  `.id == 1`,
			variables: map[string]any{
				"_headers": map[string]any{},
			},
			wantPassed:        true,
			wantLeftSide:      ".id",
			wantOperator:      "==",
			wantRightSide:     "1",
			wantActualValue:   "1",
			wantExpectedValue: "1",
		},
		{
			name:       "nil variables - backward compatible",
			input:      `{"id": 1}`,
			expr:       `.id == 1`,
			variables:  nil,
			wantPassed: true,
		},
		{
			name:  "combine input and variable in expression",
			input: `{"value": 10}`,
			expr:  `.value > $_threshold`,
			variables: map[string]any{
				"_threshold": 5,
			},
			wantPassed:      true,
			wantLeftSide:    ".value",
			wantOperator:    ">",
			wantRightSide:   "$_threshold",
			wantActualValue: "10",
		},
		{
			name:  "invalid expression with variables",
			input: `{"id": 1}`,
			expr:  `$_bad &`,
			variables: map[string]any{
				"_bad": "value",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, detail, err := EvalJQBoolWithDetailAndVars(tt.input, tt.expr, tt.variables)

			if (err != nil) != tt.wantErr {
				t.Errorf("EvalJQBoolWithDetailAndVars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if passed != tt.wantPassed {
				t.Errorf("EvalJQBoolWithDetailAndVars() passed = %v, want %v", passed, tt.wantPassed)
			}

			if detail == nil {
				t.Fatal("EvalJQBoolWithDetailAndVars() detail is nil")
			}

			if detail.Expression != tt.expr {
				t.Errorf("detail.Expression = %q, want %q", detail.Expression, tt.expr)
			}

			if tt.wantLeftSide != "" && detail.LeftSide != tt.wantLeftSide {
				t.Errorf("detail.LeftSide = %q, want %q", detail.LeftSide, tt.wantLeftSide)
			}

			if tt.wantOperator != "" && detail.Operator != tt.wantOperator {
				t.Errorf("detail.Operator = %q, want %q", detail.Operator, tt.wantOperator)
			}

			if tt.wantRightSide != "" && detail.RightSide != tt.wantRightSide {
				t.Errorf("detail.RightSide = %q, want %q", detail.RightSide, tt.wantRightSide)
			}

			if tt.wantActualValue != "" && detail.ActualValue != tt.wantActualValue {
				t.Errorf("detail.ActualValue = %q, want %q", detail.ActualValue, tt.wantActualValue)
			}

			if tt.wantExpectedValue != "" && detail.ExpectedValue != tt.wantExpectedValue {
				t.Errorf("detail.ExpectedValue = %q, want %q", detail.ExpectedValue, tt.wantExpectedValue)
			}
		})
	}
}
