package financial

import (
	"encoding/json"
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
