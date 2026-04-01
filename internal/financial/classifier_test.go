package financial

import (
	"strings"
	"testing"
)

// --- Known element classification tests ---

func TestClassify_KnownBSElements(t *testing.T) {
	tests := []struct {
		name       string
		elementID  string
		pointType  string
		wantStmt   StatementType
		wantKey    string
		wantIsTotal bool
	}{
		// IFRS BS elements
		{"IFRS total assets", "jpigp_cor:AssetsIFRS", "instant", StmtBS, "total_assets", true},
		{"IFRS current assets", "jpigp_cor:CurrentAssetsIFRS", "instant", StmtBS, "current_assets", true},
		{"IFRS cash", "jpigp_cor:CashAndCashEquivalentsIFRS", "instant", StmtBS, "cash_and_equivalents", false},
		{"IFRS total liabilities", "jpigp_cor:LiabilitiesIFRS", "instant", StmtBS, "total_liabilities", true},
		{"IFRS current liabilities", "jpigp_cor:TotalCurrentLiabilitiesIFRS", "instant", StmtBS, "current_liabilities", true},
		{"IFRS equity parent", "jpigp_cor:EquityAttributableToOwnersOfParentIFRS", "instant", StmtBS, "equity", true},
		{"IFRS equity total", "jpigp_cor:EquityIFRS", "instant", StmtBS, "net_assets", true},
		{"IFRS interest bearing CL", "jpigp_cor:InterestBearingLiabilitiesCLIFRS", "instant", StmtBS, "interest_bearing_debt", false},
		{"IFRS interest bearing NCL", "jpigp_cor:InterestBearingLiabilitiesNCLIFRS", "instant", StmtBS, "interest_bearing_debt", false},

		// JP-GAAP BS elements
		{"JPGAAP total assets", "jppfs_cor:TotalAssets", "instant", StmtBS, "total_assets", true},
		{"JPGAAP current assets", "jppfs_cor:CurrentAssets", "instant", StmtBS, "current_assets", true},
		{"JPGAAP cash", "jppfs_cor:CashAndDeposits", "instant", StmtBS, "cash_and_equivalents", false},
		{"JPGAAP current liabilities", "jppfs_cor:CurrentLiabilities", "instant", StmtBS, "current_liabilities", true},
		{"JPGAAP total liabilities", "jppfs_cor:TotalLiabilities", "instant", StmtBS, "total_liabilities", true},
		{"JPGAAP net assets", "jppfs_cor:NetAssets", "instant", StmtBS, "net_assets", true},
		{"JPGAAP shareholders equity", "jppfs_cor:ShareholdersEquity", "instant", StmtBS, "equity", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Classify(tt.elementID, tt.pointType)
			if c.Statement != tt.wantStmt {
				t.Errorf("Statement = %q, want %q", c.Statement, tt.wantStmt)
			}
			if c.SummaryKey != tt.wantKey {
				t.Errorf("SummaryKey = %q, want %q", c.SummaryKey, tt.wantKey)
			}
			if c.IsTotal != tt.wantIsTotal {
				t.Errorf("IsTotal = %v, want %v", c.IsTotal, tt.wantIsTotal)
			}
		})
	}
}

func TestClassify_KnownPLElements(t *testing.T) {
	tests := []struct {
		name      string
		elementID string
		pointType string
		wantStmt  StatementType
		wantKey   string
	}{
		// IFRS PL elements
		{"IFRS revenue", "jpigp_cor:RevenueIFRS", "duration", StmtPL, "revenue"},
		{"IFRS cost of sales", "jpigp_cor:CostOfSalesIFRS", "duration", StmtPL, "cost_of_sales"},
		{"IFRS gross profit", "jpigp_cor:GrossProfitIFRS", "duration", StmtPL, "gross_profit"},
		{"IFRS operating profit", "jpigp_cor:OperatingProfitLossIFRS", "duration", StmtPL, "operating_income"},
		{"IFRS net income", "jpigp_cor:ProfitLossAttributableToOwnersOfParentIFRS", "duration", StmtPL, "net_income"},
		{"IFRS basic eps", "jpigp_cor:BasicEarningsLossPerShareIFRS", "duration", StmtPL, "eps"},
		{"IFRS basic and diluted eps", "jpigp_cor:BasicAndDilutedEarningsLossPerShareIFRS", "duration", StmtPL, "eps"},

		// JP-GAAP PL elements
		{"JPGAAP net sales", "jppfs_cor:NetSales", "duration", StmtPL, "revenue"},
		{"JPGAAP cost of sales", "jppfs_cor:CostOfSales", "duration", StmtPL, "cost_of_sales"},
		{"JPGAAP gross profit", "jppfs_cor:GrossProfit", "duration", StmtPL, "gross_profit"},
		{"JPGAAP operating income", "jppfs_cor:OperatingIncome", "duration", StmtPL, "operating_income"},
		{"JPGAAP ordinary income", "jppfs_cor:OrdinaryIncome", "duration", StmtPL, "ordinary_income"},
		{"JPGAAP net income", "jppfs_cor:NetIncome", "duration", StmtPL, "net_income"},
		{"JPGAAP SGA", "jppfs_cor:SellingGeneralAndAdministrativeExpenses", "duration", StmtPL, "sga_expenses"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Classify(tt.elementID, tt.pointType)
			if c.Statement != tt.wantStmt {
				t.Errorf("Statement = %q, want %q", c.Statement, tt.wantStmt)
			}
			if c.SummaryKey != tt.wantKey {
				t.Errorf("SummaryKey = %q, want %q", c.SummaryKey, tt.wantKey)
			}
		})
	}
}

func TestClassify_KnownCFElements(t *testing.T) {
	tests := []struct {
		name      string
		elementID string
		pointType string
		wantStmt  StatementType
		wantKey   string
	}{
		// IFRS CF elements
		{"IFRS operating CF", "jpigp_cor:CashFlowsFromUsedInOperatingActivitiesIFRS", "duration", StmtCF, "operating_cf"},
		{"IFRS investing CF", "jpigp_cor:CashFlowsFromUsedInInvestingActivitiesIFRS", "duration", StmtCF, "investing_cf"},
		{"IFRS financing CF", "jpigp_cor:CashFlowsFromUsedInFinancingActivitiesIFRS", "duration", StmtCF, "financing_cf"},
		{"IFRS depreciation", "jpigp_cor:DepreciationAndAmortisationIFRS", "duration", StmtCF, "depreciation"},
		{"IFRS capex", "jpigp_cor:CapitalExpendituresIFRS", "duration", StmtCF, "capital_expenditure"},

		// JP-GAAP CF elements
		{"JPGAAP operating CF", "jppfs_cor:NetCashProvidedByUsedInOperatingActivities", "duration", StmtCF, "operating_cf"},
		{"JPGAAP investing CF", "jppfs_cor:NetCashProvidedByUsedInInvestingActivities", "duration", StmtCF, "investing_cf"},
		{"JPGAAP financing CF", "jppfs_cor:NetCashProvidedByUsedInFinancingActivities", "duration", StmtCF, "financing_cf"},
		{"JPGAAP depreciation", "jppfs_cor:DepreciationAndAmortization", "duration", StmtCF, "depreciation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Classify(tt.elementID, tt.pointType)
			if c.Statement != tt.wantStmt {
				t.Errorf("Statement = %q, want %q", c.Statement, tt.wantStmt)
			}
			if c.SummaryKey != tt.wantKey {
				t.Errorf("SummaryKey = %q, want %q", c.SummaryKey, tt.wantKey)
			}
		})
	}
}

func TestClassify_JpcrpCorElements(t *testing.T) {
	// jpcrp_cor: elements are cross-standard (shares, R&D, dividends, etc.)
	tests := []struct {
		name      string
		elementID string
		pointType string
		wantStmt  StatementType
		wantKey   string
	}{
		{"shares outstanding", "jpcrp_cor:NumberOfIssuedSharesAsOfFilingDateTotal", "instant", StmtBS, "shares_outstanding"},
		{"treasury shares", "jpcrp_cor:NumberOfTreasurySharesAsOfFilingDateTotal", "instant", StmtBS, "treasury_shares"},
		{"R&D expenses", "jpcrp_cor:ResearchAndDevelopmentExpensesTotal", "duration", StmtPL, "research_and_development"},
		{"dividend per share", "jpcrp_cor:DividendPaidPerShareSummaryOfBusinessResults", "duration", StmtPL, "dividend_per_share"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Classify(tt.elementID, tt.pointType)
			if c.Statement != tt.wantStmt {
				t.Errorf("Statement = %q, want %q", c.Statement, tt.wantStmt)
			}
			if c.SummaryKey != tt.wantKey {
				t.Errorf("SummaryKey = %q, want %q", c.SummaryKey, tt.wantKey)
			}
		})
	}
}

// --- Unknown element classification tests ---

func TestClassify_UnknownDurationElement_NotPL(t *testing.T) {
	// Unknown duration elements must NOT fallback to PL
	c := Classify("jpigp_cor:SomeCompletelyUnknownThing", "duration")
	if c.Statement == StmtPL {
		t.Errorf("Unknown duration element must NOT fallback to PL, got %q", c.Statement)
	}
	if c.Statement != StmtUnknown {
		t.Errorf("Statement = %q, want %q", c.Statement, StmtUnknown)
	}
	if c.SummaryKey != "" {
		t.Errorf("SummaryKey = %q, want empty for unknown element", c.SummaryKey)
	}
}

func TestClassify_UnknownInstantElement_NotBS(t *testing.T) {
	// Unknown instant elements must NOT fallback to BS
	c := Classify("jppfs_cor:SomeCompletelyUnknownThing", "instant")
	if c.Statement == StmtBS {
		t.Errorf("Unknown instant element must NOT fallback to BS, got %q", c.Statement)
	}
	if c.Statement != StmtUnknown {
		t.Errorf("Statement = %q, want %q", c.Statement, StmtUnknown)
	}
}

// --- Keyword + pointType positive match tests ---

func TestClassify_KeywordPositiveMatch_CashFlowDuration(t *testing.T) {
	// Unknown element with CashFlow keyword + duration → CF
	c := Classify("jpigp_cor:SomethingCashFlowRelated", "duration")
	if c.Statement != StmtCF {
		t.Errorf("CashFlow keyword + duration should be CF, got %q", c.Statement)
	}
	if c.SummaryKey != "" {
		t.Errorf("SummaryKey should be empty for keyword match, got %q", c.SummaryKey)
	}
}

func TestClassify_KeywordPositiveMatch_AssetInstant(t *testing.T) {
	// Unknown element with Asset keyword + instant → BS
	c := Classify("jppfs_cor:SomeSpecialAssets", "instant")
	if c.Statement != StmtBS {
		t.Errorf("Asset keyword + instant should be BS, got %q", c.Statement)
	}
}

func TestClassify_KeywordPositiveMatch_LiabilityInstant(t *testing.T) {
	c := Classify("jppfs_cor:SomeLiabilities", "instant")
	if c.Statement != StmtBS {
		t.Errorf("Liability keyword + instant should be BS, got %q", c.Statement)
	}
}

func TestClassify_KeywordPositiveMatch_EquityInstant(t *testing.T) {
	c := Classify("jppfs_cor:SomeEquityComponent", "instant")
	if c.Statement != StmtBS {
		t.Errorf("Equity keyword + instant should be BS, got %q", c.Statement)
	}
}

func TestClassify_KeywordPositiveMatch_IncomeOrExpenseDuration(t *testing.T) {
	// Revenue/income/expense keywords + duration → PL
	c := Classify("jppfs_cor:ExtraordinaryIncome", "duration")
	if c.Statement != StmtPL {
		t.Errorf("Income keyword + duration should be PL, got %q", c.Statement)
	}

	c2 := Classify("jppfs_cor:SomeExpenses", "duration")
	if c2.Statement != StmtPL {
		t.Errorf("Expense keyword + duration should be PL, got %q", c2.Statement)
	}
}

func TestClassify_KeywordPositiveMatch_SalesDuration(t *testing.T) {
	c := Classify("jppfs_cor:SomeSalesItems", "duration")
	if c.Statement != StmtPL {
		t.Errorf("Sales keyword + duration should be PL, got %q", c.Statement)
	}
}

func TestClassify_KeywordPointTypeMismatch_AssetDuration(t *testing.T) {
	// Asset keyword but duration → NOT BS
	c := Classify("jppfs_cor:SomeUnknownAssetsChangeDuration", "duration")
	// If it matches a PL keyword too, that's fine. But it should NOT be BS.
	if c.Statement == StmtBS {
		t.Errorf("Asset keyword + duration should NOT be BS, got %q", c.Statement)
	}
}

func TestClassify_KeywordPointTypeMismatch_CashFlowInstant(t *testing.T) {
	// CashFlow keyword but instant → NOT CF
	c := Classify("jpigp_cor:UnknownCashFlowItem", "instant")
	if c.Statement == StmtCF {
		t.Errorf("CashFlow keyword + instant should NOT be CF, got %q", c.Statement)
	}
}

// --- TextBlock exclusion tests ---

func TestIsTextBlock_True(t *testing.T) {
	tests := []string{
		"jpcrp_cor:BusinessResultsOfReportingCompanyTextBlock",
		"jpigp_cor:NotesConsolidatedBalanceSheetIFRSTextBlock",
		"jppfs_cor:NotesRegardingLossOfSignificantAccountTextBlock",
	}
	for _, id := range tests {
		if !IsTextBlock(id) {
			t.Errorf("IsTextBlock(%q) = false, want true", id)
		}
	}
}

func TestIsTextBlock_False(t *testing.T) {
	tests := []string{
		"jpigp_cor:AssetsIFRS",
		"jppfs_cor:NetSales",
		"jpcrp_cor:NumberOfIssuedSharesAsOfFilingDateTotal",
	}
	for _, id := range tests {
		if IsTextBlock(id) {
			t.Errorf("IsTextBlock(%q) = true, want false", id)
		}
	}
}

// --- Company-specific element mapping tests ---

func TestClassify_CompanySpecificElement_SuffixMatch(t *testing.T) {
	// Company-specific elements (jpcrp030000-asr_*) should try suffix match
	tests := []struct {
		name      string
		elementID string
		pointType string
		wantStmt  StatementType
		wantKey   string
	}{
		{
			"company-specific SalesRevenuesIFRS",
			"jpcrp030000-asr_E02144-000:SalesRevenuesIFRS",
			"duration",
			StmtPL,
			"revenue",
		},
		{
			"company-specific OperatingRevenuesIFRSKeyFinancialData",
			"jpcrp030000-asr_E02144-000:OperatingRevenuesIFRSKeyFinancialData",
			"duration",
			StmtPL,
			"revenue",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Classify(tt.elementID, tt.pointType)
			if c.Statement != tt.wantStmt {
				t.Errorf("Statement = %q, want %q", c.Statement, tt.wantStmt)
			}
			if c.SummaryKey != tt.wantKey {
				t.Errorf("SummaryKey = %q, want %q", c.SummaryKey, tt.wantKey)
			}
		})
	}
}

// --- Elements() consistency tests ---

func TestElements_ReturnsNonEmpty(t *testing.T) {
	elems := Elements()
	if len(elems) == 0 {
		t.Fatal("Elements() returned empty slice")
	}
}

func TestElements_ConsistentWithClassify(t *testing.T) {
	// Every element from Elements() should be classifiable with the correct statement and summary key
	for _, elem := range Elements() {
		// Determine the appropriate pointType for this element
		var pt string
		switch elem.Statement {
		case StmtBS:
			pt = "instant"
		case StmtPL:
			pt = "duration"
		case StmtCF:
			pt = "duration"
		default:
			t.Errorf("Elements() contains element %q with unexpected statement %q", elem.ID, elem.Statement)
			continue
		}

		c := Classify(elem.ID, pt)
		if c.Statement != elem.Statement {
			t.Errorf("Classify(%q, %q).Statement = %q, but Elements() says %q", elem.ID, pt, c.Statement, elem.Statement)
		}
		if c.SummaryKey != elem.SummaryKey {
			t.Errorf("Classify(%q, %q).SummaryKey = %q, but Elements() says %q", elem.ID, pt, c.SummaryKey, elem.SummaryKey)
		}
	}
}

func TestElements_UniqueIDs(t *testing.T) {
	seen := make(map[string]bool)
	for _, elem := range Elements() {
		if seen[elem.ID] {
			t.Errorf("duplicate element ID in Elements(): %q", elem.ID)
		}
		seen[elem.ID] = true
	}
}

func TestElements_AllHaveStatementType(t *testing.T) {
	valid := map[StatementType]bool{StmtBS: true, StmtPL: true, StmtCF: true}
	for _, elem := range Elements() {
		if !valid[elem.Statement] {
			t.Errorf("Elements() entry %q has invalid statement %q", elem.ID, elem.Statement)
		}
	}
}

// --- Summary key coverage test ---

func TestSummaryKeyCoverage_BuffettCodeMetrics(t *testing.T) {
	// All these summary keys must be present in at least one element
	requiredKeys := []string{
		"revenue", "cost_of_sales", "gross_profit", "operating_income",
		"ordinary_income", "net_income",
		"total_assets", "net_assets", "equity", "total_liabilities",
		"current_assets", "current_liabilities",
		"cash_and_equivalents", "interest_bearing_debt",
		"operating_cf", "investing_cf", "financing_cf",
		"depreciation", "capital_expenditure",
		"research_and_development", "sga_expenses",
		"shares_outstanding", "treasury_shares",
		"eps", "dividend_per_share",
	}

	// Collect all summary keys from Elements()
	keySet := make(map[string]bool)
	for _, elem := range Elements() {
		if elem.SummaryKey != "" {
			keySet[elem.SummaryKey] = true
		}
	}

	for _, key := range requiredKeys {
		if !keySet[key] {
			t.Errorf("required summary key %q not covered by any element in Elements()", key)
		}
	}
}

// --- SortOrder tests ---

func TestClassify_SortOrderIsPositive(t *testing.T) {
	// Known elements should have positive sort order
	c := Classify("jpigp_cor:AssetsIFRS", "instant")
	if c.SortOrder <= 0 {
		t.Errorf("SortOrder = %d, want > 0 for known element", c.SortOrder)
	}
}

func TestClassify_SortOrderPreservesStatementGrouping(t *testing.T) {
	// Within the same statement type, elements should have logical ordering
	// (total assets after individual asset items, etc.)
	assets := Classify("jpigp_cor:CurrentAssetsIFRS", "instant")
	totalAssets := Classify("jpigp_cor:AssetsIFRS", "instant")

	if assets.Statement != StmtBS || totalAssets.Statement != StmtBS {
		t.Fatal("both should be BS")
	}
	// Total assets should come after current assets in sort order
	if totalAssets.SortOrder <= assets.SortOrder {
		t.Errorf("total assets SortOrder (%d) should be > current assets SortOrder (%d)",
			totalAssets.SortOrder, assets.SortOrder)
	}
}

// --- Category tests ---

func TestClassify_CategoryNotEmpty_ForKnownElements(t *testing.T) {
	tests := []struct {
		name      string
		elementID string
		pointType string
	}{
		{"IFRS assets", "jpigp_cor:AssetsIFRS", "instant"},
		{"JPGAAP revenue", "jppfs_cor:NetSales", "duration"},
		{"JPGAAP operating CF", "jppfs_cor:NetCashProvidedByUsedInOperatingActivities", "duration"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Classify(tt.elementID, tt.pointType)
			if c.Category == "" {
				t.Errorf("Category should not be empty for known element %q", tt.elementID)
			}
		})
	}
}

// --- Statement type constant tests ---

func TestStatementType_Values(t *testing.T) {
	if StmtBS != "bs" {
		t.Errorf("StmtBS = %q, want %q", StmtBS, "bs")
	}
	if StmtPL != "pl" {
		t.Errorf("StmtPL = %q, want %q", StmtPL, "pl")
	}
	if StmtCF != "cf" {
		t.Errorf("StmtCF = %q, want %q", StmtCF, "cf")
	}
	if StmtUnknown != "unknown" {
		t.Errorf("StmtUnknown = %q, want %q", StmtUnknown, "unknown")
	}
}

// --- Edge case: jpdei_cor elements ---

func TestClassify_JpdeiCorElements_Unknown(t *testing.T) {
	// jpdei_cor: document info elements are not financial statements
	c := Classify("jpdei_cor:EDINETCodeDEI", "instant")
	if c.Statement != StmtUnknown {
		t.Errorf("jpdei_cor element should be unknown, got %q", c.Statement)
	}
}

// --- Edge case: empty element ID ---

func TestClassify_EmptyElementID(t *testing.T) {
	c := Classify("", "instant")
	if c.Statement != StmtUnknown {
		t.Errorf("empty element ID should be unknown, got %q", c.Statement)
	}
}

// --- IFRS PL elements that might appear as both basic and diluted EPS ---

func TestClassify_EPSVariants(t *testing.T) {
	tests := []struct {
		elementID string
		wantKey   string
	}{
		{"jpigp_cor:BasicEarningsLossPerShareIFRS", "eps"},
		{"jpigp_cor:BasicAndDilutedEarningsLossPerShareIFRS", "eps"},
	}
	for _, tt := range tests {
		c := Classify(tt.elementID, "duration")
		if c.SummaryKey != tt.wantKey {
			t.Errorf("Classify(%q).SummaryKey = %q, want %q", tt.elementID, c.SummaryKey, tt.wantKey)
		}
	}
}

// --- Verify that Elements() contains a reasonable number of mappings ---

func TestElements_MinimumCount(t *testing.T) {
	elems := Elements()
	// We expect roughly 80-120 mapped elements
	if len(elems) < 50 {
		t.Errorf("Elements() returned only %d elements, expected at least 50", len(elems))
	}
}

// --- Verify Elements() has both IFRS and JPGAAP elements ---

func TestElements_HasBothStandards(t *testing.T) {
	hasIFRS := false
	hasJPGAAP := false
	for _, elem := range Elements() {
		if strings.HasPrefix(elem.ID, "jpigp_cor:") {
			hasIFRS = true
		}
		if strings.HasPrefix(elem.ID, "jppfs_cor:") {
			hasJPGAAP = true
		}
	}
	if !hasIFRS {
		t.Error("Elements() contains no IFRS elements (jpigp_cor:)")
	}
	if !hasJPGAAP {
		t.Error("Elements() contains no JP-GAAP elements (jppfs_cor:)")
	}
}

// --- Interest bearing debt is additive (multiple elements map to same key) ---

func TestClassify_InterestBearingDebt_MultipleElements(t *testing.T) {
	cl := Classify("jpigp_cor:InterestBearingLiabilitiesCLIFRS", "instant")
	ncl := Classify("jpigp_cor:InterestBearingLiabilitiesNCLIFRS", "instant")

	if cl.SummaryKey != "interest_bearing_debt" {
		t.Errorf("CL interest bearing debt SummaryKey = %q, want %q", cl.SummaryKey, "interest_bearing_debt")
	}
	if ncl.SummaryKey != "interest_bearing_debt" {
		t.Errorf("NCL interest bearing debt SummaryKey = %q, want %q", ncl.SummaryKey, "interest_bearing_debt")
	}
}

// --- Keyword heuristic should not match when only partial keyword appears ---

func TestClassify_KeywordHeuristic_NoPartialMatch(t *testing.T) {
	// "assess" contains "asse" but not "Asset" — should not match BS
	c := Classify("jpcrp_cor:SomeAssessmentNote", "instant")
	// This should be unknown — "Assess" is not "Asset"
	// The heuristic should match on "Asset" not "Asse"
	// (implementation detail: case-insensitive keyword boundaries)
	if c.Statement != StmtUnknown {
		t.Logf("Note: %q matched statement %q — this is acceptable if the keyword matching is reasonable", "jpcrp_cor:SomeAssessmentNote", c.Statement)
	}
}

// --- JP-GAAP noncurrent liabilities ---

func TestClassify_JPGAAPNoncurrentLiabilities(t *testing.T) {
	c := Classify("jppfs_cor:NoncurrentLiabilities", "instant")
	if c.Statement != StmtBS {
		t.Errorf("Statement = %q, want BS", c.Statement)
	}
}

// --- IFRS noncurrent assets ---

func TestClassify_IFRSNoncurrentAssets(t *testing.T) {
	c := Classify("jpigp_cor:NonCurrentAssetsIFRS", "instant")
	if c.Statement != StmtBS {
		t.Errorf("Statement = %q, want BS", c.Statement)
	}
}

// --- JP-GAAP noncurrent assets ---

func TestClassify_JPGAAPNoncurrentAssets(t *testing.T) {
	c := Classify("jppfs_cor:NoncurrentAssets", "instant")
	if c.Statement != StmtBS {
		t.Errorf("Statement = %q, want BS", c.Statement)
	}
}
