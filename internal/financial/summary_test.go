package financial

import "testing"

func TestBuildAndDeriveSummary_IntegrationWithStatements(t *testing.T) {
	stmts := []FinancialStatement{
		{
			Type: "pl", Consolidated: true, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "current", Items: []LineItem{
					{ElementID: "jppfs_cor:NetSales", SummaryKey: "revenue", Value: ptrFloat(10000)},
					{ElementID: "jppfs_cor:OperatingIncome", SummaryKey: "operating_income", Value: ptrFloat(2000)},
					{ElementID: "jppfs_cor:NetIncome", SummaryKey: "net_income", Value: ptrFloat(1000)},
				}},
			},
		},
		{
			Type: "bs", Consolidated: true, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "current", Items: []LineItem{
					{ElementID: "jppfs_cor:TotalAssets", SummaryKey: "total_assets", Value: ptrFloat(50000)},
					{ElementID: "jppfs_cor:ShareholdersEquity", SummaryKey: "equity", Value: ptrFloat(20000)},
				}},
			},
		},
	}

	s := BuildAndDeriveSummary(stmts)

	// Extracted values
	assertSummaryValue(t, s, "revenue", 10000)
	assertSummaryValue(t, s, "operating_income", 2000)
	assertSummaryValue(t, s, "net_income", 1000)
	assertSummaryValue(t, s, "total_assets", 50000)
	assertSummaryValue(t, s, "equity", 20000)

	// Derived values
	assertSummaryValue(t, s, "operating_margin", 0.2)
	assertSummaryValue(t, s, "roe", 0.05)         // 1000/20000
	assertSummaryValue(t, s, "roa", 0.02)         // 1000/50000
	assertSummaryValue(t, s, "equity_ratio", 0.4) // 20000/50000
}

func TestBuildAndDeriveSummary_FilteredStatementsOmitIrrelevantMetrics(t *testing.T) {
	// PL-only input: BS-derived metrics should not appear
	plOnly := []FinancialStatement{
		{
			Type: "pl", Consolidated: true, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "current", Items: []LineItem{
					{ElementID: "jppfs_cor:NetSales", SummaryKey: "revenue", Value: ptrFloat(10000)},
					{ElementID: "jppfs_cor:NetIncome", SummaryKey: "net_income", Value: ptrFloat(1000)},
				}},
			},
		},
	}

	s := BuildAndDeriveSummary(plOnly)

	// PL-derived metrics should be present
	assertSummaryValue(t, s, "net_margin", 0.1)

	// BS-derived metrics should NOT be present (no BS data)
	if s["equity_ratio"] != nil {
		t.Errorf("equity_ratio should be nil without BS data, got %v", *s["equity_ratio"])
	}
	if s["roe"] != nil {
		t.Errorf("roe should be nil without BS data, got %v", *s["roe"])
	}
	if s["current_ratio"] != nil {
		t.Errorf("current_ratio should be nil without BS data, got %v", *s["current_ratio"])
	}
}

