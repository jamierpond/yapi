package config

import (
	"testing"

	"yapi.run/cli/internal/utils"
)

func TestMerge_MapIsolation(t *testing.T) {
	// This test ensures that modifying merged config doesn't pollute the base config
	base := &ConfigV1{
		URL: "http://example.com",
		Headers: map[string]string{
			"Authorization": "Bearer base-token",
		},
		Query: map[string]string{
			"page": "1",
		},
	}

	// Store original values
	originalAuthHeader := base.Headers["Authorization"]
	originalPage := base.Query["page"]

	// Step 1 adds headers
	step1 := ChainStep{
		Name: "step1",
		ConfigV1: ConfigV1{
			Headers: map[string]string{
				"X-Step": "1",
			},
		},
	}

	merged1 := base.Merge(step1)

	// Verify step1 merged correctly
	if merged1.Headers["X-Step"] != "1" {
		t.Error("step1 header not merged")
	}
	if merged1.Headers["Authorization"] != "Bearer base-token" {
		t.Error("base Authorization header not inherited")
	}

	// Modify the merged config (simulating what runner does)
	merged1.Headers["X-Modified"] = "yes"

	// Critical: base should NOT be affected
	if _, exists := base.Headers["X-Modified"]; exists {
		t.Error("base.Headers was polluted by modifying merged1.Headers")
	}
	if _, exists := base.Headers["X-Step"]; exists {
		t.Error("base.Headers was polluted by step1 merge")
	}
	if base.Headers["Authorization"] != originalAuthHeader {
		t.Error("base.Headers Authorization was modified")
	}

	// Step 2 should also get clean base headers
	step2 := ChainStep{
		Name: "step2",
		ConfigV1: ConfigV1{
			Headers: map[string]string{
				"X-Step": "2",
			},
		},
	}

	merged2 := base.Merge(step2)

	// merged2 should NOT have X-Modified from merged1
	if _, exists := merged2.Headers["X-Modified"]; exists {
		t.Error("merged2 inherited pollution from merged1")
	}
	if merged2.Headers["X-Step"] != "2" {
		t.Error("step2 header not merged correctly")
	}

	// Base should still be unchanged
	if base.Headers["Authorization"] != originalAuthHeader {
		t.Error("base.Headers Authorization was modified after step2")
	}
	if base.Query["page"] != originalPage {
		t.Error("base.Query was modified")
	}
}

func TestMerge_BodyDeepCopy(t *testing.T) {
	base := &ConfigV1{
		URL: "http://example.com",
		Body: map[string]any{
			"nested": map[string]any{
				"key": "value",
			},
		},
	}

	step := ChainStep{Name: "step1"}
	merged := base.Merge(step)

	// Modify the nested map in merged
	if nested, ok := merged.Body["nested"].(map[string]any); ok {
		nested["key"] = "modified"
	}

	// Base should NOT be affected
	if baseNested, ok := base.Body["nested"].(map[string]any); ok {
		if baseNested["key"] != "value" {
			t.Error("base.Body nested map was polluted")
		}
	}
}

func TestMerge_FlowControl(t *testing.T) {
	// Case 1: Base has delay, step inherits it
	base := &ConfigV1{
		URL:   "http://example.com",
		Delay: "1s",
	}
	step := ChainStep{
		Name: "inherit_delay",
	}
	merged := base.Merge(step)
	if merged.Delay != "1s" {
		t.Errorf("expected delay '1s', got '%s'", merged.Delay)
	}

	// Case 2: Step overrides delay
	stepOverride := ChainStep{
		Name: "override_delay",
		ConfigV1: ConfigV1{
			Delay: "5s",
		},
	}
	mergedOverride := base.Merge(stepOverride)
	if mergedOverride.Delay != "5s" {
		t.Errorf("expected delay '5s', got '%s'", mergedOverride.Delay)
	}

	// Case 3: Step adds delay
	baseNoDelay := &ConfigV1{
		URL: "http://example.com",
	}
	stepWithDelay := ChainStep{
		Name: "add_delay",
		ConfigV1: ConfigV1{
			Delay: "500ms",
		},
	}
	mergedAddDelay := baseNoDelay.Merge(stepWithDelay)
	if mergedAddDelay.Delay != "500ms" {
		t.Errorf("expected delay '500ms', got '%s'", mergedAddDelay.Delay)
	}
}

func TestDeepCloneMap(t *testing.T) {
	src := map[string]any{
		"string": "value",
		"number": 42,
		"nested": map[string]any{
			"inner": "data",
		},
		"array": []any{
			"a",
			map[string]any{"b": "c"},
		},
	}

	dst := utils.DeepCloneMap(src)

	// Modify dst
	dst["string"] = "changed"
	if nested, ok := dst["nested"].(map[string]any); ok {
		nested["inner"] = "changed"
	}
	if arr, ok := dst["array"].([]any); ok {
		arr[0] = "changed"
		if m, ok := arr[1].(map[string]any); ok {
			m["b"] = "changed"
		}
	}

	// Verify src is unchanged
	if src["string"] != "value" {
		t.Error("src string was modified")
	}
	if nested, ok := src["nested"].(map[string]any); ok {
		if nested["inner"] != "data" {
			t.Error("src nested map was modified")
		}
	}
	if arr, ok := src["array"].([]any); ok {
		if arr[0] != "a" {
			t.Error("src array element was modified")
		}
		if m, ok := arr[1].(map[string]any); ok {
			if m["b"] != "c" {
				t.Error("src array nested map was modified")
			}
		}
	}
}
