package cmd

import (
	"encoding/json"
	"testing"
)

func TestSchemaCommands_OutputIsJSON(t *testing.T) {
	stdout, _, code := executeCommand("schema", "commands")
	if code != 0 {
		t.Fatalf("schema commands exit code = %d, want 0", code)
	}
	if !json.Valid([]byte(stdout)) {
		t.Errorf("output is not valid JSON: %q", stdout[:min(100, len(stdout))])
	}
}

func TestSchemaDocTypes_OutputIsJSON(t *testing.T) {
	stdout, _, code := executeCommand("schema", "doc-types")
	if code != 0 {
		t.Fatalf("schema doc-types exit code = %d, want 0", code)
	}
	if !json.Valid([]byte(stdout)) {
		t.Errorf("output is not valid JSON")
	}
}

func TestSchemaSections_OutputIsJSON(t *testing.T) {
	stdout, _, code := executeCommand("schema", "sections")
	if code != 0 {
		t.Fatalf("schema sections exit code = %d, want 0", code)
	}
	if !json.Valid([]byte(stdout)) {
		t.Errorf("output is not valid JSON")
	}
}

func TestSchemaFinancialElements_OutputIsJSON(t *testing.T) {
	stdout, _, code := executeCommand("schema", "financial-elements")
	if code != 0 {
		t.Fatalf("schema financial-elements exit code = %d, want 0", code)
	}
	if !json.Valid([]byte(stdout)) {
		t.Errorf("output is not valid JSON: %q", stdout[:min(100, len(stdout))])
	}
	// Verify it's an array
	var elems []map[string]any
	if err := json.Unmarshal([]byte(stdout), &elems); err != nil {
		t.Fatalf("failed to parse output as array: %v", err)
	}
	if len(elems) == 0 {
		t.Error("expected non-empty elements array")
	}
}

func TestSchemaCommands_ContainsDocFinancial(t *testing.T) {
	stdout, _, _ := executeCommand("schema", "commands")
	var cmds []map[string]any
	if err := json.Unmarshal([]byte(stdout), &cmds); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}
	found := false
	for _, c := range cmds {
		if c["name"] == "doc financial" {
			found = true
			break
		}
	}
	if !found {
		t.Error("schema commands output missing 'doc financial'")
	}
}

func TestSchemaCommands_ContainsSchemaFinancialElements(t *testing.T) {
	stdout, _, _ := executeCommand("schema", "commands")
	var cmds []map[string]any
	if err := json.Unmarshal([]byte(stdout), &cmds); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}
	found := false
	for _, c := range cmds {
		if c["name"] == "schema financial-elements" {
			found = true
			break
		}
	}
	if !found {
		t.Error("schema commands output missing 'schema financial-elements'")
	}
}

func TestSchemaCommands_ContainsDocList(t *testing.T) {
	stdout, _, _ := executeCommand("schema", "commands")
	var cmds []map[string]any
	if err := json.Unmarshal([]byte(stdout), &cmds); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}
	found := false
	for _, c := range cmds {
		if c["name"] == "doc list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("schema commands output missing 'doc list'")
	}
}

func TestSchemaDocTypes_Contains120(t *testing.T) {
	stdout, _, _ := executeCommand("schema", "doc-types")
	var types []map[string]string
	if err := json.Unmarshal([]byte(stdout), &types); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}
	found := false
	for _, dt := range types {
		if dt["code"] == "120" {
			found = true
			if dt["name"] != "有価証券報告書" {
				t.Errorf("code 120 name = %q, want 有価証券報告書", dt["name"])
			}
		}
	}
	if !found {
		t.Error("doc-types missing code 120")
	}
}

func TestSchemaDerivedMetrics_OutputIsJSON(t *testing.T) {
	stdout, _, code := executeCommand("schema", "derived-metrics")
	if code != 0 {
		t.Fatalf("schema derived-metrics exit code = %d, want 0", code)
	}
	if !json.Valid([]byte(stdout)) {
		t.Errorf("output is not valid JSON: %q", stdout[:min(100, len(stdout))])
	}
}

func TestSchemaDerivedMetrics_AllKeysPresent(t *testing.T) {
	stdout, _, _ := executeCommand("schema", "derived-metrics")
	var metrics []map[string]any
	if err := json.Unmarshal([]byte(stdout), &metrics); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}
	expectedKeys := []string{"gross_margin", "operating_margin", "net_margin", "roe", "roa", "equity_ratio", "current_ratio", "fcf", "debt_to_equity"}
	keySet := make(map[string]bool)
	for _, m := range metrics {
		if k, ok := m["key"].(string); ok {
			keySet[k] = true
		}
	}
	for _, k := range expectedKeys {
		if !keySet[k] {
			t.Errorf("derived-metrics missing key %q", k)
		}
	}
}

func TestSchemaCommands_IncludesSummaryOnlyFlag(t *testing.T) {
	stdout, _, _ := executeCommand("schema", "commands")
	var cmds []map[string]any
	if err := json.Unmarshal([]byte(stdout), &cmds); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}
	for _, c := range cmds {
		name, _ := c["name"].(string)
		if name == "company financials" || name == "doc financial" {
			flags, _ := c["flags"].([]any)
			found := false
			for _, f := range flags {
				fm, _ := f.(map[string]any)
				if fm["name"] == "--summary-only" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("schema commands %q missing --summary-only flag", name)
			}
		}
	}
}

func TestSchemaCommands_IncludesDerivedMetrics(t *testing.T) {
	stdout, _, _ := executeCommand("schema", "commands")
	var cmds []map[string]any
	if err := json.Unmarshal([]byte(stdout), &cmds); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}
	found := false
	for _, c := range cmds {
		if c["name"] == "schema derived-metrics" {
			found = true
			break
		}
	}
	if !found {
		t.Error("schema commands output missing 'schema derived-metrics'")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
