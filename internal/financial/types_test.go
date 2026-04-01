package financial

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSummary_JSONRoundtrip(t *testing.T) {
	rev := 48036704000000.0
	s := Summary{
		"revenue":      &rev,
		"net_income":   nil,
		"total_assets": nil,
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got Summary
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got["revenue"] == nil || *got["revenue"] != rev {
		t.Errorf("revenue = %v, want %v", got["revenue"], rev)
	}
	// nil values should roundtrip as JSON null
	if got["net_income"] != nil {
		t.Errorf("net_income = %v, want nil", got["net_income"])
	}
}

func TestFinancialData_JSONOutput(t *testing.T) {
	val := 1000.0
	fd := FinancialData{
		DocID:         "S100TEST",
		AccountingStd: "jpgaap",
		Consolidated:  true,
		Summary: Summary{
			"revenue": &val,
		},
		Statements: []FinancialStatement{
			{
				Type:          "pl",
				Consolidated:  true,
				AccountingStd: "jpgaap",
				Periods: []PeriodData{
					{
						Period: "current",
						Items: []LineItem{
							{
								ElementID:  "jppfs_cor:NetSales",
								Label:      "売上高",
								Value:      &val,
								SummaryKey: "revenue",
							},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(fd)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	if !json.Valid(data) {
		t.Error("output is not valid JSON")
	}
}

func TestFinancialData_StripStatements(t *testing.T) {
	val := 1000.0
	fd := FinancialData{
		DocID:         "S100TEST",
		FiscalYear:    "2025-03-31",
		AccountingStd: "jpgaap",
		Consolidated:  true,
		Summary: Summary{
			"revenue": &val,
		},
		Statements: []FinancialStatement{
			{Type: "pl", Consolidated: true},
		},
	}

	fd.StripStatements()

	// Statements should be nil
	if fd.Statements != nil {
		t.Errorf("Statements should be nil after StripStatements, got %v", fd.Statements)
	}

	// Metadata should be preserved
	if fd.DocID != "S100TEST" {
		t.Errorf("DocID = %q, want %q", fd.DocID, "S100TEST")
	}
	if fd.FiscalYear != "2025-03-31" {
		t.Errorf("FiscalYear = %q, want %q", fd.FiscalYear, "2025-03-31")
	}
	if fd.AccountingStd != "jpgaap" {
		t.Errorf("AccountingStd = %q, want %q", fd.AccountingStd, "jpgaap")
	}
	if fd.Summary["revenue"] == nil || *fd.Summary["revenue"] != val {
		t.Error("Summary should be preserved after StripStatements")
	}

	// JSON should contain "statements":null
	data, err := json.Marshal(fd)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"statements":null`) {
		t.Errorf("JSON should contain \"statements\":null, got: %s", jsonStr)
	}
}
