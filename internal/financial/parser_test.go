package financial

import (
	"math"
	"strings"
	"testing"

	"github.com/beatinaniwa/edinet-cli/internal/extract"
)

// --- Helper functions ---

// makeCSVFile creates an extract.CSVFile from headers and rows for testing.
func makeCSVFile(filename string, headers []string, rows [][]string) extract.CSVFile {
	return extract.CSVFile{
		Filename: filename,
		Headers:  headers,
		Rows:     rows,
	}
}

// standardHeaders returns the standard EDINET CSV headers.
func standardHeaders() []string {
	return []string{"要素ID", "項目名", "コンテキストID", "相対年度", "連結・個別", "期間・時点", "ユニットID", "単位", "値"}
}

// makeRow creates a row with standard column count, filling specified values.
func makeRow(elementID, label, contextID, relYear, consolidated, pointType, unitID, unit, value string) []string {
	return []string{elementID, label, contextID, relYear, consolidated, pointType, unitID, unit, value}
}

func ptrFloat(v float64) *float64 { return &v }

// findSummaryKey returns the summary value for a key, or nil if not present.
func findSummaryKey(s Summary, key string) *float64 {
	if s == nil {
		return nil
	}
	return s[key]
}

// --- IFRS Consolidated CSV test ---

func TestParse_IFRSConsolidated(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E02144-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			// BS items (current period, consolidated, instant)
			makeRow("jpigp_cor:AssetsIFRS", "資産合計", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "48036704000000"),
			makeRow("jpigp_cor:LiabilitiesIFRS", "負債合計", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "30000000000000"),
			makeRow("jpigp_cor:EquityIFRS", "資本合計", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "18036704000000"),
			makeRow("jpigp_cor:CashAndCashEquivalentsIFRS", "現金", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "5000000000000"),
			makeRow("jpigp_cor:CurrentAssetsIFRS", "流動資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "20000000000000"),
			makeRow("jpigp_cor:TotalCurrentLiabilitiesIFRS", "流動負債", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "15000000000000"),
			makeRow("jpigp_cor:InterestBearingLiabilitiesCLIFRS", "有利子負債CL", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "3000000000000"),
			makeRow("jpigp_cor:InterestBearingLiabilitiesNCLIFRS", "有利子負債NCL", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "7000000000000"),
			makeRow("jpigp_cor:EquityAttributableToOwnersOfParentIFRS", "親会社所有者帰属持分", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "17000000000000"),
			// Prior period BS
			makeRow("jpigp_cor:AssetsIFRS", "資産合計", "Prior1YearInstant", "前期", "連結", "時点", "JPY", "円", "45000000000000"),
			// PL items (current period, consolidated, duration)
			makeRow("jpigp_cor:RevenueIFRS", "売上収益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "37154298000000"),
			makeRow("jpigp_cor:CostOfSalesIFRS", "売上原価", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "30000000000000"),
			makeRow("jpigp_cor:GrossProfitIFRS", "売上総利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "7154298000000"),
			makeRow("jpigp_cor:OperatingProfitLossIFRS", "営業利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "5352934000000"),
			makeRow("jpigp_cor:ProfitLossAttributableToOwnersOfParentIFRS", "親会社帰属当期利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "4944898000000"),
			makeRow("jpigp_cor:BasicEarningsLossPerShareIFRS", "基本EPS", "CurrentYearDuration", "当期", "連結", "期間", "JPYPerShares", "円/株", "359.56"),
			// CF items
			makeRow("jpigp_cor:CashFlowsFromUsedInOperatingActivitiesIFRS", "営業CF", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "4300000000000"),
			makeRow("jpigp_cor:CashFlowsFromUsedInInvestingActivitiesIFRS", "投資CF", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "-2100000000000"),
			makeRow("jpigp_cor:CashFlowsFromUsedInFinancingActivitiesIFRS", "財務CF", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "-1500000000000"),
			makeRow("jpigp_cor:DepreciationAndAmortisationIFRS", "減価償却", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1200000000000"),
			// Prior period PL
			makeRow("jpigp_cor:RevenueIFRS", "売上収益", "Prior1YearDuration", "前期", "連結", "期間", "JPY", "円", "31000000000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check accounting standard
	if result.AccountingStd != "ifrs" {
		t.Errorf("AccountingStd = %q, want %q", result.AccountingStd, "ifrs")
	}
	// Check consolidated
	if !result.Consolidated {
		t.Error("Consolidated = false, want true")
	}

	// Check summary values
	assertSummaryValue(t, result.Summary, "total_assets", 48036704000000)
	assertSummaryValue(t, result.Summary, "total_liabilities", 30000000000000)
	assertSummaryValue(t, result.Summary, "net_assets", 18036704000000)
	assertSummaryValue(t, result.Summary, "revenue", 37154298000000)
	assertSummaryValue(t, result.Summary, "operating_income", 5352934000000)
	assertSummaryValue(t, result.Summary, "net_income", 4944898000000)
	assertSummaryValue(t, result.Summary, "cash_and_equivalents", 5000000000000)
	assertSummaryValue(t, result.Summary, "current_assets", 20000000000000)
	assertSummaryValue(t, result.Summary, "current_liabilities", 15000000000000)
	assertSummaryValue(t, result.Summary, "equity", 17000000000000)
	assertSummaryValue(t, result.Summary, "operating_cf", 4300000000000)
	assertSummaryValue(t, result.Summary, "investing_cf", -2100000000000)
	assertSummaryValue(t, result.Summary, "financing_cf", -1500000000000)
	assertSummaryValue(t, result.Summary, "depreciation", 1200000000000)
	assertSummaryValue(t, result.Summary, "eps", 359.56)

	// Interest bearing debt is additive: CL + NCL = 3T + 7T = 10T
	assertSummaryValue(t, result.Summary, "interest_bearing_debt", 10000000000000)

	// Check statements exist
	if len(result.Statements) == 0 {
		t.Fatal("Statements is empty")
	}

	// Check we have BS, PL, CF statements
	stmtTypes := make(map[string]bool)
	for _, stmt := range result.Statements {
		stmtTypes[stmt.Type] = true
	}
	for _, st := range []string{"bs", "pl", "cf"} {
		if !stmtTypes[st] {
			t.Errorf("missing statement type %q", st)
		}
	}

	// Check that statements have periods
	for _, stmt := range result.Statements {
		if len(stmt.Periods) == 0 {
			t.Errorf("statement %q has no periods", stmt.Type)
		}
	}
}

// --- JP-GAAP Consolidated CSV test ---

func TestParse_JPGAAPConsolidated(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			// BS
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "10000000000"),
			makeRow("jppfs_cor:TotalLiabilities", "負債合計", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "6000000000"),
			makeRow("jppfs_cor:NetAssets", "純資産合計", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "4000000000"),
			makeRow("jppfs_cor:CurrentAssets", "流動資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "5000000000"),
			makeRow("jppfs_cor:CurrentLiabilities", "流動負債", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "3000000000"),
			makeRow("jppfs_cor:CashAndDeposits", "現金預金", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "2000000000"),
			makeRow("jppfs_cor:ShareholdersEquity", "株主資本", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "3500000000"),
			// PL
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "8000000000"),
			makeRow("jppfs_cor:CostOfSales", "売上原価", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "5000000000"),
			makeRow("jppfs_cor:GrossProfit", "売上総利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "3000000000"),
			makeRow("jppfs_cor:SellingGeneralAndAdministrativeExpenses", "販管費", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "2000000000"),
			makeRow("jppfs_cor:OperatingIncome", "営業利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000000"),
			makeRow("jppfs_cor:OrdinaryIncome", "経常利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1100000000"),
			makeRow("jppfs_cor:NetIncome", "当期純利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "700000000"),
			// CF
			makeRow("jppfs_cor:NetCashProvidedByUsedInOperatingActivities", "営業CF", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1500000000"),
			makeRow("jppfs_cor:NetCashProvidedByUsedInInvestingActivities", "投資CF", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "-800000000"),
			makeRow("jppfs_cor:NetCashProvidedByUsedInFinancingActivities", "財務CF", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "-300000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.AccountingStd != "jpgaap" {
		t.Errorf("AccountingStd = %q, want %q", result.AccountingStd, "jpgaap")
	}
	if !result.Consolidated {
		t.Error("Consolidated = false, want true")
	}

	assertSummaryValue(t, result.Summary, "total_assets", 10000000000)
	assertSummaryValue(t, result.Summary, "revenue", 8000000000)
	assertSummaryValue(t, result.Summary, "operating_income", 1000000000)
	assertSummaryValue(t, result.Summary, "ordinary_income", 1100000000)
	assertSummaryValue(t, result.Summary, "net_income", 700000000)
	assertSummaryValue(t, result.Summary, "sga_expenses", 2000000000)
	assertSummaryValue(t, result.Summary, "operating_cf", 1500000000)
}

// --- Non-consolidated only company test ---

func TestParse_NonConsolidatedOnly(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E99999-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "個別", "時点", "JPY", "円", "5000000000"),
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "個別", "期間", "JPY", "円", "3000000000"),
			makeRow("jppfs_cor:NetCashProvidedByUsedInOperatingActivities", "営業CF", "CurrentYearDuration", "当期", "個別", "期間", "JPY", "円", "800000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Consolidated {
		t.Error("Consolidated = true, want false for non-consolidated only company")
	}
	assertSummaryValue(t, result.Summary, "total_assets", 5000000000)
	assertSummaryValue(t, result.Summary, "revenue", 3000000000)
}

// --- Mixed consolidation: BS consolidated, CF non-consolidated only ---

func TestParse_MixedConsolidation(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E88888-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			// BS has both consolidated and non-consolidated → auto picks consolidated
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "10000000000"),
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "個別", "時点", "JPY", "円", "8000000000"),
			// PL has consolidated
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "5000000000"),
			// CF only has non-consolidated (no consolidated CF data)
			makeRow("jppfs_cor:NetCashProvidedByUsedInOperatingActivities", "営業CF", "CurrentYearDuration", "当期", "個別", "期間", "JPY", "円", "600000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// BS should use consolidated
	assertSummaryValue(t, result.Summary, "total_assets", 10000000000)
	// CF should fallback to non-consolidated (with warning)
	assertSummaryValue(t, result.Summary, "operating_cf", 600000000)

	// Should have a warning about mixed consolidation
	hasWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "non_consolidated") || strings.Contains(w, "fallback") {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected warning about CF fallback to non-consolidated, got none")
	}
}

// --- Explicit Consolidated=true option ---

func TestParse_ExplicitConsolidated(t *testing.T) {
	consolidated := true
	file := makeCSVFile(
		"jpcrp030000-asr-001_E88888-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "10000000000"),
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "個別", "時点", "JPY", "円", "8000000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{Consolidated: &consolidated})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertSummaryValue(t, result.Summary, "total_assets", 10000000000)
	if !result.Consolidated {
		t.Error("Consolidated = false, want true")
	}
}

// --- Explicit Consolidated=false option ---

func TestParse_ExplicitNonConsolidated(t *testing.T) {
	nonConsolidated := false
	file := makeCSVFile(
		"jpcrp030000-asr-001_E88888-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "10000000000"),
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "個別", "時点", "JPY", "円", "8000000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{Consolidated: &nonConsolidated})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertSummaryValue(t, result.Summary, "total_assets", 8000000000)
	if result.Consolidated {
		t.Error("Consolidated = true, want false for explicit non-consolidated")
	}
}

// --- Multiple CSV files: jpaud excluded, header-missing file skipped ---

func TestParse_MultipleFiles_JpaudExcludedAndHeaderMissing(t *testing.T) {
	mainFile := makeCSVFile(
		"jpcrp030000-asr-001_E02144-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "10000000000"),
		},
	)

	// jpaud file should be skipped
	jpaudFile := makeCSVFile(
		"jpaud-asr-001_E02144-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "99999999"),
		},
	)

	// File with missing required headers should be skipped with warning
	badHeaderFile := makeCSVFile(
		"jpcrp030000-asr-002_E02144-000_2025-03-31_01_2025-06-20.csv",
		[]string{"col1", "col2", "col3"},
		[][]string{
			{"a", "b", "c"},
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{mainFile, jpaudFile, badHeaderFile}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should use main file's value, not jpaud's 99999999
	assertSummaryValue(t, result.Summary, "total_assets", 10000000000)

	// Should have warning about skipped file with bad headers
	hasHeaderWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "header") || strings.Contains(w, "jpcrp030000-asr-002") {
			hasHeaderWarning = true
			break
		}
	}
	if !hasHeaderWarning {
		t.Error("expected warning about skipped file with missing headers, got none")
	}
}

// --- Value "－" → Value=nil, RawValue preserved ---

func TestParse_DashValue_NilWithRawValue(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:OrdinaryIncome", "経常利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "－"),
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Find the dash item in statements
	found := false
	for _, stmt := range result.Statements {
		for _, pd := range stmt.Periods {
			for _, item := range pd.Items {
				if item.ElementID == "jppfs_cor:OrdinaryIncome" {
					found = true
					if item.Value != nil {
						t.Errorf("Value for '－' should be nil, got %v", *item.Value)
					}
					if item.RawValue != "－" {
						t.Errorf("RawValue = %q, want %q", item.RawValue, "－")
					}
				}
			}
		}
	}
	if !found {
		t.Error("could not find OrdinaryIncome item in statements")
	}

	// Summary should not have this key (it's nil)
	if v := findSummaryKey(result.Summary, "ordinary_income"); v != nil {
		t.Errorf("Summary ordinary_income should be nil for '－' value, got %v", *v)
	}
}

// --- Empty value → Value=nil ---

func TestParse_EmptyValue_Nil(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:OrdinaryIncome", "経常利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", ""),
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Find the empty value item in statements
	found := false
	for _, stmt := range result.Statements {
		for _, pd := range stmt.Periods {
			for _, item := range pd.Items {
				if item.ElementID == "jppfs_cor:OrdinaryIncome" {
					found = true
					if item.Value != nil {
						t.Errorf("Value for empty string should be nil, got %v", *item.Value)
					}
				}
			}
		}
	}
	if !found {
		t.Error("could not find OrdinaryIncome item in statements")
	}
}

// --- Decimal values (EPS 359.56) → float64 ---

func TestParse_DecimalValue(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E02144-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jpigp_cor:BasicEarningsLossPerShareIFRS", "基本EPS", "CurrentYearDuration", "当期", "連結", "期間", "JPYPerShares", "円/株", "359.56"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertSummaryValue(t, result.Summary, "eps", 359.56)
}

// --- Duplicate element resolution: first occurrence (main file) wins ---

func TestParse_DuplicateResolution_FirstWins(t *testing.T) {
	// Main file (001) has priority over supplementary file (002)
	mainFile := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "10000000000"),
		},
	)

	suppFile := makeCSVFile(
		"jpcrp030000-asr-002_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "99999999"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{mainFile, suppFile}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should use main file's value
	assertSummaryValue(t, result.Summary, "total_assets", 10000000000)
}

// --- Interest bearing debt additive (CL + NCL) ---

func TestParse_InterestBearingDebt_Additive(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			// JP-GAAP interest-bearing debt components
			makeRow("jppfs_cor:ShortTermLoansPayable", "短期借入金", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "1000000000"),
			makeRow("jppfs_cor:CurrentPortionOfLongTermLoansPayable", "一年以内返済長期借入金", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "500000000"),
			makeRow("jppfs_cor:BondsPayable", "社債", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "2000000000"),
			makeRow("jppfs_cor:LongTermLoansPayable", "長期借入金", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "3000000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// 1B + 500M + 2B + 3B = 6.5B
	assertSummaryValue(t, result.Summary, "interest_bearing_debt", 6500000000)
}

// --- Annual report check: warn if filename lacks "asr" marker ---

func TestParse_NonAnnualReport_Warning(t *testing.T) {
	// Quarterly report (not annual "asr")
	file := makeCSVFile(
		"jpcrp030000-q1r-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "5000000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	hasWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "annual") || strings.Contains(w, "asr") {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected warning about non-annual report, got none")
	}
}

// --- Empty CSVDataResult → error ---

func TestParse_EmptyCSVResult_Error(t *testing.T) {
	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{}}
	_, err := Parse(csvResult, ParseOpts{})
	if err == nil {
		t.Fatal("Parse() should return error for empty CSV result")
	}
}

// --- All files skipped → error ---

func TestParse_AllFilesSkipped_Error(t *testing.T) {
	// Only jpaud files and files with bad headers
	jpaudFile := makeCSVFile(
		"jpaud-asr-001_E02144-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "10000000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{jpaudFile}}
	_, err := Parse(csvResult, ParseOpts{})
	if err == nil {
		t.Fatal("Parse() should return error when all files are skipped")
	}
}

// --- TextBlock rows are skipped ---

func TestParse_TextBlockSkipped(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jpcrp_cor:BusinessResultsOfReportingCompanyTextBlock", "事業の内容", "CurrentYearDuration", "当期", "連結", "期間", "", "", "<html>long text...</html>"),
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// TextBlock should not appear in any statement
	for _, stmt := range result.Statements {
		for _, pd := range stmt.Periods {
			for _, item := range pd.Items {
				if strings.HasSuffix(item.ElementID, "TextBlock") {
					t.Errorf("TextBlock item %q should not be in statements", item.ElementID)
				}
			}
		}
	}
}

// --- Segment member rows are skipped ---

func TestParse_SegmentMemberSkipped(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E02144-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			// Normal row
			makeRow("jpigp_cor:RevenueIFRS", "売上収益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "37000000000000"),
			// Segment member row should be skipped
			makeRow("jpigp_cor:RevenueIFRS", "売上収益", "CurrentYearDuration_jpcrp030000-asr_E02144-000AutomotiveReportableSegmentMember", "当期", "その他", "期間", "JPY", "円", "20000000000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Revenue should be the non-segment value
	assertSummaryValue(t, result.Summary, "revenue", 37000000000000)
}

// --- File selection: only jpcrp prefix files ---

func TestParse_NonJpcrpFilesSkipped(t *testing.T) {
	// Non-jpcrp file (e.g., jpdei) should be skipped
	nonJpcrpFile := makeCSVFile(
		"jpdei030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jpdei_cor:EDINETCodeDEI", "EDINETコード", "FilingDateInstant", "", "その他", "時点", "", "", "E00001"),
		},
	)

	jpcrpFile := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "5000000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{nonJpcrpFile, jpcrpFile}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertSummaryValue(t, result.Summary, "total_assets", 5000000000)
}

// --- Accounting standard detection: per-statement prefix majority ---

func TestParse_AccountingStdDetection_IFRS(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E02144-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jpigp_cor:RevenueIFRS", "売上収益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000"),
			makeRow("jpigp_cor:OperatingProfitLossIFRS", "営業利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "500000"),
			makeRow("jpigp_cor:AssetsIFRS", "資産合計", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "2000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.AccountingStd != "ifrs" {
		t.Errorf("AccountingStd = %q, want %q", result.AccountingStd, "ifrs")
	}
}

func TestParse_AccountingStdDetection_JPGAAP(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000"),
			makeRow("jppfs_cor:OperatingIncome", "営業利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "500000"),
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "2000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.AccountingStd != "jpgaap" {
		t.Errorf("AccountingStd = %q, want %q", result.AccountingStd, "jpgaap")
	}
}

// --- Main file (001) has priority in ordering ---

func TestParse_MainFilePriority(t *testing.T) {
	// Supplementary file comes first in the slice, but main file (001) should have priority
	suppFile := makeCSVFile(
		"jpcrp030000-asr-002_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "999"),
		},
	)

	mainFile := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "5000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{suppFile, mainFile}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertSummaryValue(t, result.Summary, "revenue", 5000000)
}

// --- Nil CSVDataResult → error ---

func TestParse_NilCSVResult_Error(t *testing.T) {
	_, err := Parse(nil, ParseOpts{})
	if err == nil {
		t.Fatal("Parse() should return error for nil CSV result")
	}
}

// --- Shares and cross-standard elements ---

func TestParse_CrossStandardElements(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jpcrp_cor:NumberOfIssuedSharesAsOfFilingDateTotal", "発行済株式数", "FilingDateInstant", "", "その他", "時点", "shares", "株", "1000000000"),
			makeRow("jpcrp_cor:NumberOfTreasurySharesAsOfFilingDateTotal", "自己株式数", "FilingDateInstant", "", "その他", "時点", "shares", "株", "50000000"),
			makeRow("jpcrp_cor:DividendPaidPerShareSummaryOfBusinessResults", "一株配当", "CurrentYearDuration", "当期", "その他", "期間", "JPYPerShares", "円/株", "50.00"),
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertSummaryValue(t, result.Summary, "shares_outstanding", 1000000000)
	assertSummaryValue(t, result.Summary, "treasury_shares", 50000000)
	assertSummaryValue(t, result.Summary, "dividend_per_share", 50.00)
}

// --- Negative values ---

func TestParse_NegativeValues(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:NetCashProvidedByUsedInInvestingActivities", "投資CF", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "-800000000"),
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertSummaryValue(t, result.Summary, "investing_cf", -800000000)
}

// --- Statement accounting standard field ---

func TestParse_StatementAccountingStd(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E02144-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jpigp_cor:RevenueIFRS", "売上収益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000"),
			makeRow("jpigp_cor:AssetsIFRS", "資産合計", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "2000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	for _, stmt := range result.Statements {
		if stmt.AccountingStd != "ifrs" {
			t.Errorf("statement %q AccountingStd = %q, want %q", stmt.Type, stmt.AccountingStd, "ifrs")
		}
	}
}

// --- LineItem fields populated correctly ---

func TestParse_LineItemFields(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "8000000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Find the item
	found := false
	for _, stmt := range result.Statements {
		if stmt.Type != "pl" {
			continue
		}
		for _, pd := range stmt.Periods {
			for _, item := range pd.Items {
				if item.ElementID == "jppfs_cor:NetSales" {
					found = true
					if item.Label != "売上高" {
						t.Errorf("Label = %q, want %q", item.Label, "売上高")
					}
					if item.Value == nil || *item.Value != 8000000000 {
						t.Errorf("Value = %v, want 8000000000", item.Value)
					}
					if item.RawValue != "8000000000" {
						t.Errorf("RawValue = %q, want %q", item.RawValue, "8000000000")
					}
					if item.Unit != "円" {
						t.Errorf("Unit = %q, want %q", item.Unit, "円")
					}
					if item.UnitID != "JPY" {
						t.Errorf("UnitID = %q, want %q", item.UnitID, "JPY")
					}
					if item.SummaryKey != "revenue" {
						t.Errorf("SummaryKey = %q, want %q", item.SummaryKey, "revenue")
					}
					if !item.IsTotal {
						t.Error("IsTotal = false, want true for NetSales")
					}
				}
			}
		}
	}
	if !found {
		t.Error("could not find NetSales item in PL statement")
	}
}

// --- Prior period data is included ---

func TestParse_PriorPeriodIncluded(t *testing.T) {
	file := makeCSVFile(
		"jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "10000000000"),
			makeRow("jppfs_cor:TotalAssets", "総資産", "Prior1YearInstant", "前期", "連結", "時点", "JPY", "円", "9000000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Find BS statement
	for _, stmt := range result.Statements {
		if stmt.Type != "bs" {
			continue
		}
		periods := make(map[string]bool)
		for _, pd := range stmt.Periods {
			periods[pd.Period] = true
		}
		if !periods["current"] {
			t.Error("missing current period in BS statement")
		}
		if !periods["prior1"] {
			t.Error("missing prior1 period in BS statement")
		}
	}
}

// --- AccountingStd computed from selected rows only ---

func TestParse_AccountingStdFromSelectedRows(t *testing.T) {
	// Mixed filing: consolidated IFRS rows + non-consolidated JP-GAAP rows.
	// When requesting non-consolidated, AccountingStd should be "jpgaap"
	// because only JP-GAAP rows are selected.
	nonCons := false
	file := makeCSVFile(
		"jpcrp030000-asr-001_E99999-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			// Consolidated IFRS rows (should be excluded when non-consolidated is requested)
			makeRow("jpigp_cor:RevenueIFRS", "売上収益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "5000000"),
			makeRow("jpigp_cor:AssetsIFRS", "資産合計", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "10000000"),
			makeRow("jpigp_cor:OperatingProfitLossIFRS", "営業利益", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000"),
			// Non-consolidated JP-GAAP rows
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "個別", "期間", "JPY", "円", "3000000"),
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "個別", "時点", "JPY", "円", "8000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{Consolidated: &nonCons})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Result-level AccountingStd should reflect the selected (non-consolidated) rows
	if result.AccountingStd != "jpgaap" {
		t.Errorf("AccountingStd = %q, want %q", result.AccountingStd, "jpgaap")
	}

	// Each statement should also have jpgaap
	for _, stmt := range result.Statements {
		if stmt.AccountingStd != "jpgaap" {
			t.Errorf("statement %q AccountingStd = %q, want %q", stmt.Type, stmt.AccountingStd, "jpgaap")
		}
	}
}

// --- Neutral-only rows preserved in explicit non-consolidated mode ---

func TestParse_NeutralOnlyRows_NonConsolidated(t *testing.T) {
	// When only neutral "other" rows exist (jpcrp_cor) with no explicit
	// non-consolidated rows, explicit non-consolidated mode should still
	// return those neutral rows rather than dropping the statement.
	nonCons := false
	file := makeCSVFile(
		"jpcrp030000-asr-001_E99999-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			// Neutral rows — tagged as "その他" in 連結・個別 column
			makeRow("jpcrp_cor:NumberOfIssuedSharesAsOfFilingDateTotal", "発行済株式総数", "FilingDateInstant", "当期", "その他", "時点", "shares", "株", "1000000"),
			makeRow("jpcrp_cor:DividendPerShareSummary", "1株当たり配当額", "CurrentYearDuration", "当期", "その他", "期間", "JPYPerShares", "円", "50"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{Consolidated: &nonCons})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should have at least one statement with the neutral rows
	if len(result.Statements) == 0 {
		t.Fatal("expected at least one statement with neutral rows, got 0")
	}
}

// --- Consolidated + neutral rows: non-consolidated fallback uses consolidated ---

func TestParse_ConsolidatedPlusNeutral_NonConsolidatedFallback(t *testing.T) {
	// When consolidated rows + neutral rows exist but NO non-consolidated rows,
	// and non-consolidated is explicitly requested, it should fallback to
	// consolidated data (which includes neutral) with a warning.
	// It should NOT return only neutral rows.
	nonCons := false
	file := makeCSVFile(
		"jpcrp030000-asr-001_E99999-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			// Consolidated rows
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "10000000"),
			makeRow("jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "当期", "連結", "時点", "JPY", "円", "50000000"),
			// Neutral rows
			makeRow("jpcrp_cor:NumberOfIssuedSharesAsOfFilingDateTotal", "発行済株式総数", "FilingDateInstant", "当期", "その他", "時点", "shares", "株", "1000000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{Consolidated: &nonCons})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should have consolidated fallback with summary keys
	if result.Summary["revenue"] == nil {
		t.Error("Summary[revenue] should exist (consolidated fallback), got nil")
	}
	if result.Summary["total_assets"] == nil {
		t.Error("Summary[total_assets] should exist (consolidated fallback), got nil")
	}

	// Should have a warning about fallback
	hasWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "non_consolidated") && strings.Contains(w, "fallback") {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected warning about non_consolidated fallback, got none")
	}
}

// --- Helper assertion ---

func assertSummaryValue(t *testing.T, s Summary, key string, want float64) {
	t.Helper()
	v := s[key]
	if v == nil {
		t.Errorf("Summary[%q] is nil, want %v", key, want)
		return
	}
	if math.Abs(*v-want) > 0.001 {
		t.Errorf("Summary[%q] = %v, want %v", key, *v, want)
	}
}

// --- SummaryOfBusinessResults precedence tests ---

func TestParse_SummaryPrecedence_DetailedWinsOverFallback(t *testing.T) {
	// When both detailed (jppfs_cor) and SummaryOfBusinessResults (jpcrp_cor) rows exist,
	// the detailed value should win because it has lower SortOrder.
	file := makeCSVFile(
		"jpcrp030000-asr-001_E02367-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			// Detailed revenue (SortOrder 1000, should win)
			makeRow("jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "当期", "連結", "期間", "JPY", "円", "1000000000000"),
			// SummaryOfBusinessResults revenue (SortOrder 1008, should be ignored)
			makeRow("jpcrp_cor:NetSalesSummaryOfBusinessResults", "売上高、経営指標等", "CurrentYearDuration", "当期", "", "期間", "JPY", "円", "9999999999999"),
			// Operating income (only from summary — should be used as fallback)
			makeRow("jpcrp_cor:OperatingIncomeSummaryOfBusinessResults", "営業利益、経営指標等", "CurrentYearDuration", "当期", "", "期間", "JPY", "円", "200000000000"),
		},
	)

	result, err := Parse(&extract.CSVDataResult{Files: []extract.CSVFile{file}}, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Detailed value wins for revenue
	assertSummaryValue(t, result.Summary, "revenue", 1000000000000)
	// Fallback used for operating income (no detailed row)
	assertSummaryValue(t, result.Summary, "operating_income", 200000000000)
}

func TestParse_SummaryFallback_UsedWhenDetailedMissing(t *testing.T) {
	// When only SummaryOfBusinessResults rows exist, they should populate summary.
	file := makeCSVFile(
		"jpcrp030000-asr-001_E02367-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			makeRow("jpcrp_cor:TotalAssetsSummaryOfBusinessResults", "総資産額、経営指標等", "CurrentYearInstant", "当期", "", "時点", "JPY", "円", "3000000000000"),
			makeRow("jpcrp_cor:NetAssetsSummaryOfBusinessResults", "純資産額、経営指標等", "CurrentYearInstant", "当期", "", "時点", "JPY", "円", "2000000000000"),
			makeRow("jpcrp_cor:NetSalesSummaryOfBusinessResults", "売上高、経営指標等", "CurrentYearDuration", "当期", "", "期間", "JPY", "円", "1500000000000"),
			makeRow("jpcrp_cor:NetIncomeSummaryOfBusinessResults", "当期純利益、経営指標等", "CurrentYearDuration", "当期", "", "期間", "JPY", "円", "400000000000"),
		},
	)

	result, err := Parse(&extract.CSVDataResult{Files: []extract.CSVFile{file}}, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertSummaryValue(t, result.Summary, "total_assets", 3000000000000)
	assertSummaryValue(t, result.Summary, "net_assets", 2000000000000)
	assertSummaryValue(t, result.Summary, "revenue", 1500000000000)
	assertSummaryValue(t, result.Summary, "net_income", 400000000000)
}

func TestParse_NonConsolidated_NeutralFallback_EmitsWarning(t *testing.T) {
	// When explicit non-consolidated is requested but only neutral rows exist,
	// a warning should be emitted.
	file := makeCSVFile(
		"jpcrp030000-asr-001_E02367-000_2025-03-31_01_2025-06-20.csv",
		standardHeaders(),
		[][]string{
			// Only neutral (jpcrp_cor) rows, no actual non-consolidated rows
			makeRow("jpcrp_cor:NetSalesSummaryOfBusinessResults", "売上高、経営指標等", "CurrentYearDuration", "当期", "", "期間", "JPY", "円", "1000000000000"),
			makeRow("jpcrp_cor:TotalAssetsSummaryOfBusinessResults", "総資産額、経営指標等", "CurrentYearInstant", "当期", "", "時点", "JPY", "円", "3000000000000"),
		},
	)

	nonCons := false
	result, err := Parse(&extract.CSVDataResult{Files: []extract.CSVFile{file}}, ParseOpts{Consolidated: &nonCons})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should still produce summary values
	assertSummaryValue(t, result.Summary, "revenue", 1000000000000)

	// Should have a warning about neutral fallback
	hasWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "neutral") || strings.Contains(w, "SummaryOfBusinessResults") {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Errorf("expected warning about neutral fallback in non-consolidated mode, got warnings: %v", result.Warnings)
	}
}

// --- detectAccountingStandard with SummaryOfBusinessResults rows ---

func TestDetectAccountingStandard_IFRSSummaryRowsOnly(t *testing.T) {
	rows := []parsedRow{
		{elementID: "jpcrp_cor:RevenueIFRSSummaryOfBusinessResults", classification: ElementClassification{Statement: StmtPL, SummaryKey: "revenue"}},
		{elementID: "jpcrp_cor:ProfitLossAttributableToOwnersOfParentIFRSSummaryOfBusinessResults", classification: ElementClassification{Statement: StmtPL, SummaryKey: "net_income"}},
	}
	std := detectAccountingStandard(rows)
	if std != "ifrs" {
		t.Errorf("detectAccountingStandard = %q, want %q", std, "ifrs")
	}
}

func TestDetectAccountingStandard_JPGAAPSummaryRowsOnly(t *testing.T) {
	rows := []parsedRow{
		{elementID: "jpcrp_cor:NetSalesSummaryOfBusinessResults", classification: ElementClassification{Statement: StmtPL, SummaryKey: "revenue"}},
		{elementID: "jpcrp_cor:TotalAssetsSummaryOfBusinessResults", classification: ElementClassification{Statement: StmtBS, SummaryKey: "total_assets"}},
	}
	std := detectAccountingStandard(rows)
	if std != "jpgaap" {
		t.Errorf("detectAccountingStandard = %q, want %q", std, "jpgaap")
	}
}

func TestDetectAccountingStandard_NeutralOnlyRows_StaysUnknown(t *testing.T) {
	// Neutral rows like dividend/shares should NOT determine accounting standard
	rows := []parsedRow{
		{elementID: "jpcrp_cor:DividendPaidPerShareSummaryOfBusinessResults", classification: ElementClassification{Statement: StmtPL, SummaryKey: "dividend_per_share"}},
		{elementID: "jpcrp_cor:NumberOfIssuedSharesAsOfFilingDateTotal", classification: ElementClassification{Statement: StmtBS, SummaryKey: "shares_outstanding"}},
	}
	std := detectAccountingStandard(rows)
	if std != "unknown" {
		t.Errorf("detectAccountingStandard = %q, want %q", std, "unknown")
	}
}

// --- Filing-style CSV (IPO prospectus) with Prior1Year contexts ---

func TestParse_FilingStyleCSV_FallbackPeriod(t *testing.T) {
	file := makeCSVFile(
		"jpcrp020400-srs-001_E41257-000_2025-04-30_01_2026-01-09.csv",
		standardHeaders(),
		[][]string{
			makeRow("jpcrp_cor:NetSalesSummaryOfBusinessResults", "売上高、経営指標等", "Prior1YearDuration", "前期", "個別", "期間", "JPY", "円", "9426601000"),
			makeRow("jpcrp_cor:OrdinaryIncomeSummaryOfBusinessResults", "経常利益、経営指標等", "Prior1YearDuration", "前期", "個別", "期間", "JPY", "円", "1145214000"),
			makeRow("jpcrp_cor:TotalAssetsSummaryOfBusinessResults", "総資産額、経営指標等", "Prior1YearInstant", "前期末", "個別", "時点", "JPY", "円", "6160640000"),
			makeRow("jpcrp_cor:NetAssetsSummaryOfBusinessResults", "純資産額、経営指標等", "Prior1YearInstant", "前期末", "個別", "時点", "JPY", "円", "4261992000"),
			makeRow("jpcrp_cor:NetSalesSummaryOfBusinessResults", "売上高、経営指標等", "Prior2YearDuration", "前々期", "個別", "期間", "JPY", "円", "8735439000"),
		},
	)

	csvResult := &extract.CSVDataResult{Files: []extract.CSVFile{file}}
	result, err := Parse(csvResult, ParseOpts{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.SummaryPeriod != "prior1" {
		t.Errorf("SummaryPeriod = %q, want %q", result.SummaryPeriod, "prior1")
	}
	assertSummaryValue(t, result.Summary, "revenue", 9426601000)
	assertSummaryValue(t, result.Summary, "ordinary_income", 1145214000)
	assertSummaryValue(t, result.Summary, "total_assets", 6160640000)
	assertSummaryValue(t, result.Summary, "net_assets", 4261992000)
}
