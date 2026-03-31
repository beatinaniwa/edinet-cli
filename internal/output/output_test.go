package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestPrintJSON_SimpleStruct(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}
	if err := PrintJSON(&buf, data); err != nil {
		t.Fatalf("PrintJSON() error = %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("got[key] = %q, want %q", got["key"], "value")
	}
}

func TestPrintJSON_PrettyPrinted(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]int{"count": 42}
	if err := PrintJSON(&buf, data); err != nil {
		t.Fatalf("PrintJSON() error = %v", err)
	}
	output := buf.String()
	// Pretty-printed JSON should contain indentation
	if len(output) < 10 {
		t.Errorf("output seems too short for pretty-printed JSON: %q", output)
	}
	// Should end with newline
	if output[len(output)-1] != '\n' {
		t.Errorf("output should end with newline")
	}
}

func TestPrintError_Format(t *testing.T) {
	var buf bytes.Buffer
	PrintError(&buf, "NOT_FOUND", "document not found")
	var got struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if got.Error.Code != "NOT_FOUND" {
		t.Errorf("code = %q, want %q", got.Error.Code, "NOT_FOUND")
	}
	if got.Error.Message != "document not found" {
		t.Errorf("message = %q, want %q", got.Error.Message, "document not found")
	}
}

func TestPrintTable_BasicOutput(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"Name", "Code"}
	rows := [][]string{
		{"Toyota", "7203"},
		{"Sony", "6758"},
	}
	PrintTable(&buf, headers, rows)
	output := buf.String()
	if len(output) == 0 {
		t.Fatal("PrintTable produced no output")
	}
	// Should contain headers and data
	for _, s := range []string{"Name", "Code", "Toyota", "7203", "Sony", "6758"} {
		if !bytes.Contains([]byte(output), []byte(s)) {
			t.Errorf("output missing %q:\n%s", s, output)
		}
	}
}
