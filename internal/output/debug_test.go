package output

import (
	"testing"
)

func TestMultiJSONPrettyPrint(t *testing.T) {
	raw := `{"name":"foo"}
{"name":"bar"}
{"name":"baz"}`
	result := prettyPrintJSON(raw)
	t.Logf("Input:\n%s", raw)
	t.Logf("Output:\n%s", result)
}
