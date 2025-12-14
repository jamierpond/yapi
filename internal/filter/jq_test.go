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
		// EvalJQBool should not panic on any input
		_, _ = EvalJQBool(input, expr)
	})
}
