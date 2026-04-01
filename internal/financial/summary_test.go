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

	s, period := BuildAndDeriveSummary(stmts)

	if period != "current" {
		t.Errorf("summaryPeriod = %q, want %q", period, "current")
	}

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

	s, period := BuildAndDeriveSummary(plOnly)

	if period != "current" {
		t.Errorf("summaryPeriod = %q, want %q", period, "current")
	}

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

func TestBuildAndDeriveSummary_FallbackToNearest_Prior1(t *testing.T) {
	stmts := []FinancialStatement{
		{
			Type: "pl", Consolidated: false, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "prior1", Items: []LineItem{
					{SummaryKey: "revenue", Value: ptrFloat(5000)},
					{SummaryKey: "net_income", Value: ptrFloat(500)},
				}},
				{Period: "prior2", Items: []LineItem{
					{SummaryKey: "revenue", Value: ptrFloat(4000)},
					{SummaryKey: "net_income", Value: ptrFloat(400)},
				}},
			},
		},
	}

	s, period := BuildAndDeriveSummary(stmts)

	if period != "prior1" {
		t.Errorf("summaryPeriod = %q, want %q", period, "prior1")
	}
	assertSummaryValue(t, s, "revenue", 5000)
	assertSummaryValue(t, s, "net_income", 500)
	assertSummaryValue(t, s, "net_margin", 0.1)
}

func TestBuildAndDeriveSummary_CurrentPeriodTakesPrecedence(t *testing.T) {
	stmts := []FinancialStatement{
		{
			Type: "pl", Consolidated: true, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "current", Items: []LineItem{
					{SummaryKey: "revenue", Value: ptrFloat(10000)},
				}},
				{Period: "prior1", Items: []LineItem{
					{SummaryKey: "revenue", Value: ptrFloat(8000)},
				}},
			},
		},
	}

	s, period := BuildAndDeriveSummary(stmts)

	if period != "current" {
		t.Errorf("summaryPeriod = %q, want %q", period, "current")
	}
	assertSummaryValue(t, s, "revenue", 10000)
}

func TestBuildAndDeriveSummary_BestPeriodOverridesFilingDate(t *testing.T) {
	stmts := []FinancialStatement{
		{
			Type: "pl", Consolidated: false, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "prior1", Items: []LineItem{
					{SummaryKey: "revenue", Value: ptrFloat(5000)},
				}},
			},
		},
		{
			Type: "bs", Consolidated: false, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "filing_date", Items: []LineItem{
					{SummaryKey: "revenue", Value: ptrFloat(9999)},          // should be overridden by prior1
					{SummaryKey: "shares_outstanding", Value: ptrFloat(100)}, // filing_date-only key
				}},
			},
		},
	}

	s, period := BuildAndDeriveSummary(stmts)

	if period != "prior1" {
		t.Errorf("summaryPeriod = %q, want %q", period, "prior1")
	}
	assertSummaryValue(t, s, "revenue", 5000)            // bestPeriod wins
	assertSummaryValue(t, s, "shares_outstanding", 100)  // filing_date supplements
}

func TestBuildAndDeriveSummary_MixedPeriods_PLPrior1_BSPrior2(t *testing.T) {
	stmts := []FinancialStatement{
		{
			Type: "pl", Consolidated: false, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "prior1", Items: []LineItem{
					{SummaryKey: "revenue", Value: ptrFloat(5000)},
					{SummaryKey: "net_income", Value: ptrFloat(500)},
				}},
			},
		},
		{
			Type: "bs", Consolidated: false, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "prior2", Items: []LineItem{
					{SummaryKey: "total_assets", Value: ptrFloat(30000)},
				}},
			},
		},
	}

	s, period := BuildAndDeriveSummary(stmts)

	if period != "prior1" {
		t.Errorf("summaryPeriod = %q, want %q", period, "prior1")
	}
	assertSummaryValue(t, s, "revenue", 5000)
	// BS values from prior2 should NOT be included (period consistency)
	if s["total_assets"] != nil {
		t.Errorf("total_assets should be nil (from different period prior2), got %v", *s["total_assets"])
	}
}

func TestBuildAndDeriveSummary_FilingDateOnly(t *testing.T) {
	stmts := []FinancialStatement{
		{
			Type: "bs", Consolidated: false, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "filing_date", Items: []LineItem{
					{SummaryKey: "shares_outstanding", Value: ptrFloat(1000)},
				}},
			},
		},
	}

	s, period := BuildAndDeriveSummary(stmts)

	if period != "" {
		t.Errorf("summaryPeriod = %q, want empty string for filing_date-only", period)
	}
	assertSummaryValue(t, s, "shares_outstanding", 1000)
}

func TestBuildAndDeriveSummary_SupplementalOnlyPeriodSkipped(t *testing.T) {
	stmts := []FinancialStatement{
		{
			Type: "pl", Consolidated: false, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "prior3", Items: []LineItem{
					{SummaryKey: "dividend_per_share", Value: ptrFloat(50)}, // supplemental only
				}},
				{Period: "prior1", Items: []LineItem{
					{SummaryKey: "revenue", Value: ptrFloat(5000)},
					{SummaryKey: "net_income", Value: ptrFloat(500)},
				}},
			},
		},
		{
			Type: "bs", Consolidated: false, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "prior1", Items: []LineItem{
					{SummaryKey: "total_assets", Value: ptrFloat(30000)},
				}},
			},
		},
	}

	s, period := BuildAndDeriveSummary(stmts)

	if period != "prior1" {
		t.Errorf("summaryPeriod = %q, want %q (prior3 has only supplemental keys)", period, "prior1")
	}
	assertSummaryValue(t, s, "revenue", 5000)
	assertSummaryValue(t, s, "total_assets", 30000)
	// prior3's dividend should NOT be included (different period)
	if s["dividend_per_share"] != nil {
		t.Errorf("dividend_per_share should be nil (from skipped period prior3), got %v", *s["dividend_per_share"])
	}
}

func TestBuildAndDeriveSummary_AdditiveKeys_FallbackPeriod(t *testing.T) {
	stmts := []FinancialStatement{
		{
			Type: "bs", Consolidated: true, AccountingStd: "jpgaap",
			Periods: []PeriodData{
				{Period: "prior1", Items: []LineItem{
					{SummaryKey: "total_assets", Value: ptrFloat(50000)},
					{SummaryKey: "interest_bearing_debt", Value: ptrFloat(1000)}, // short-term
					{SummaryKey: "interest_bearing_debt", Value: ptrFloat(2000)}, // long-term
				}},
			},
		},
	}

	s, period := BuildAndDeriveSummary(stmts)

	if period != "prior1" {
		t.Errorf("summaryPeriod = %q, want %q", period, "prior1")
	}
	assertSummaryValue(t, s, "interest_bearing_debt", 3000) // additive: 1000 + 2000
}

