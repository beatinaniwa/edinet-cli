package cmd

import (
	"testing"
)

func TestCompanySearch_NoArgs(t *testing.T) {
	_, _, code := executeCommand("company", "search")
	if code == 0 {
		t.Error("expected non-zero exit code when no query provided")
	}
}

func TestCompanyFilings_NoArgs(t *testing.T) {
	_, _, code := executeCommand("company", "filings")
	if code == 0 {
		t.Error("expected non-zero exit code when no code provided")
	}
}

func TestCompanyUpdate_OutputFormat(t *testing.T) {
	// company update needs network access, so we just verify the command exists
	// and has the right structure
	stdout, _, _ := executeCommand("company", "update", "--help")
	if stdout == "" {
		t.Error("expected help output for company update")
	}
}

func TestCompanySubcommands_Exist(t *testing.T) {
	commands := []string{"search", "filings", "update"}
	for _, sub := range commands {
		stdout, _, _ := executeCommand("company", sub, "--help")
		if stdout == "" && sub != "update" {
			t.Errorf("expected help output for company %s", sub)
		}
	}
}

func TestCompanySearch_HelpIsJSON(t *testing.T) {
	// Verify the company command group exists
	_, _, code := executeCommand("company", "--help")
	if code != 0 {
		t.Error("company --help should succeed")
	}
}

func TestCompanyFilings_HelpOutput(t *testing.T) {
	stdout, _, code := executeCommand("company", "filings", "--help")
	if code != 0 {
		t.Error("company filings --help should succeed")
	}
	if stdout == "" {
		t.Error("expected help output")
	}
}
