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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
