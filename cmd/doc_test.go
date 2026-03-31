package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func executeCommand(args ...string) (stdout, stderr string, exitCode int) {
	var outBuf, errBuf bytes.Buffer

	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	if err != nil {
		exitCode = exitError(&errBuf, err)
	}

	return outBuf.String(), errBuf.String(), exitCode
}

func TestDocList_NoDateFlag(t *testing.T) {
	_, stderr, code := executeCommand("doc", "list")
	if code == 0 {
		t.Error("expected non-zero exit code when no date flags provided")
	}
	if stderr == "" {
		t.Error("expected error message on stderr")
	}
}

func TestDocList_DateAndFromExclusive(t *testing.T) {
	_, stderr, code := executeCommand("doc", "list", "--date", "2025-06-20", "--from", "2025-06-01")
	if code == 0 {
		t.Error("expected non-zero exit code for --date + --from")
	}
	// Should contain validation error
	var errResp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stderr), &errResp); err == nil {
		if errResp.Error.Code != "VALIDATION_ERROR" {
			t.Errorf("error code = %q, want VALIDATION_ERROR", errResp.Error.Code)
		}
	}
}

func TestDocList_FromWithoutTo(t *testing.T) {
	_, _, code := executeCommand("doc", "list", "--from", "2025-06-01")
	if code == 0 {
		t.Error("expected non-zero exit code for --from without --to")
	}
}

func TestDocList_ToWithoutFrom(t *testing.T) {
	_, _, code := executeCommand("doc", "list", "--to", "2025-06-30")
	if code == 0 {
		t.Error("expected non-zero exit code for --to without --from")
	}
}

func TestDocGet_NoDocID(t *testing.T) {
	_, _, code := executeCommand("doc", "get")
	if code == 0 {
		t.Error("expected non-zero exit code when no docID provided")
	}
}

func TestDocGet_NoType(t *testing.T) {
	_, _, code := executeCommand("doc", "get", "S100ABCD")
	if code == 0 {
		t.Error("expected non-zero exit code when no --type provided")
	}
}

func TestDocGet_InvalidType(t *testing.T) {
	_, stderr, code := executeCommand("doc", "get", "S100ABCD", "--type", "invalid")
	if code == 0 {
		t.Error("expected non-zero exit code for invalid --type")
	}
	if stderr == "" {
		t.Error("expected error message on stderr")
	}
}

func TestDocData_NoDocID(t *testing.T) {
	_, _, code := executeCommand("doc", "data")
	if code == 0 {
		t.Error("expected non-zero exit code when no docID provided")
	}
}

func TestDocText_NoDocID(t *testing.T) {
	_, _, code := executeCommand("doc", "text")
	if code == 0 {
		t.Error("expected non-zero exit code when no docID provided")
	}
}

func TestDocText_ListSectionsOutput(t *testing.T) {
	stdout, _, code := executeCommand("doc", "text", "--list-sections")
	if code != 0 {
		t.Fatalf("expected exit code 0 for --list-sections, got %d", code)
	}
	if !json.Valid([]byte(stdout)) {
		t.Errorf("--list-sections output is not valid JSON: %q", stdout)
	}
}
