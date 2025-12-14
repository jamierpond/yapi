package filter

import (
	"testing"
)

func TestMultipleResults(t *testing.T) {
	input := `{"items":[{"name":"foo","stars":1},{"name":"bar","stars":2},{"name":"baz","stars":3}]}`
	filter := ".items[] | {name: .name, stars: .stars}"
	result, err := ApplyJQ(input, filter)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Result:\n%s", result)
	t.Logf("Result length: %d", len(result))
}
