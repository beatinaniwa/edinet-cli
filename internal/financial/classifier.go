package financial

import "strings"

// StatementType represents a financial statement type.
type StatementType string

const (
	StmtBS      StatementType = "bs"
	StmtPL      StatementType = "pl"
	StmtCF      StatementType = "cf"
	StmtUnknown StatementType = "unknown"
)

// ElementClassification holds the classification result for an XBRL element.
type ElementClassification struct {
	Statement  StatementType
	Category   string
	SortOrder  int
	IsTotal    bool
	SummaryKey string // maps to summary field, "" if not a summary item
}

// ElementInfo describes a known element mapping for schema output.
type ElementInfo struct {
	ID         string        `json:"id"`
	Statement  StatementType `json:"statement"`
	Category   string        `json:"category"`
	SummaryKey string        `json:"summary_key,omitempty"`
	LabelEN    string        `json:"label_en,omitempty"`
}

// elementDef is the internal definition for a known element.
type elementDef struct {
	statement  StatementType
	category   string
	sortOrder  int
	isTotal    bool
	summaryKey string
	labelEN    string
}

// knownElements maps full element IDs to their classification.
// This is the single source of truth for element classification and summary key mapping.
var knownElements map[string]elementDef

// companySuffixes maps element suffixes (after the colon in company-specific IDs) to their classification.
// Used for jpcrp030000-asr_* elements.
var companySuffixes map[string]elementDef

func init() {
	knownElements = map[string]elementDef{
		// ============================================================
		// IFRS Balance Sheet (jpigp_cor:)
		// ============================================================

		// Current Assets
		"jpigp_cor:CashAndCashEquivalentsIFRS":           {StmtBS, "current_assets", 100, false, "cash_and_equivalents", "Cash and cash equivalents"},
		"jpigp_cor:TradeAndOtherCurrentReceivablesIFRS":   {StmtBS, "current_assets", 110, false, "", "Trade and other receivables (current)"},
		"jpigp_cor:OtherCurrentFinancialAssetsIFRS":       {StmtBS, "current_assets", 120, false, "", "Other current financial assets"},
		"jpigp_cor:InventoriesIFRS":                       {StmtBS, "current_assets", 130, false, "", "Inventories"},
		"jpigp_cor:OtherCurrentAssetsIFRS":                {StmtBS, "current_assets", 140, false, "", "Other current assets"},
		"jpigp_cor:CurrentAssetsIFRS":                     {StmtBS, "current_assets", 199, true, "current_assets", "Total current assets"},

		// Non-current Assets
		"jpigp_cor:PropertyPlantAndEquipmentIFRS":         {StmtBS, "noncurrent_assets", 200, false, "", "Property, plant and equipment"},
		"jpigp_cor:RightOfUseAssetsIFRS":                  {StmtBS, "noncurrent_assets", 205, false, "", "Right-of-use assets"},
		"jpigp_cor:GoodwillIFRS":                          {StmtBS, "noncurrent_assets", 210, false, "", "Goodwill"},
		"jpigp_cor:IntangibleAssetsIFRS":                  {StmtBS, "noncurrent_assets", 220, false, "", "Intangible assets"},
		"jpigp_cor:InvestmentAccountedForUsingEquityMethodIFRS": {StmtBS, "noncurrent_assets", 230, false, "", "Investments using equity method"},
		"jpigp_cor:OtherNonCurrentFinancialAssetsIFRS":    {StmtBS, "noncurrent_assets", 240, false, "", "Other non-current financial assets"},
		"jpigp_cor:DeferredTaxAssetsIFRS":                 {StmtBS, "noncurrent_assets", 250, false, "", "Deferred tax assets"},
		"jpigp_cor:OtherNonCurrentAssetsIFRS":             {StmtBS, "noncurrent_assets", 260, false, "", "Other non-current assets"},
		"jpigp_cor:NonCurrentAssetsIFRS":                  {StmtBS, "noncurrent_assets", 299, true, "", "Total non-current assets"},
		"jpigp_cor:AssetsIFRS":                            {StmtBS, "total", 300, true, "total_assets", "Total assets"},

		// Current Liabilities
		"jpigp_cor:TradeAndOtherCurrentPayablesIFRS":      {StmtBS, "current_liabilities", 400, false, "", "Trade and other payables (current)"},
		"jpigp_cor:InterestBearingLiabilitiesCLIFRS":      {StmtBS, "current_liabilities", 410, false, "interest_bearing_debt", "Interest-bearing liabilities (current)"},
		"jpigp_cor:OtherCurrentFinancialLiabilitiesIFRS":  {StmtBS, "current_liabilities", 420, false, "", "Other current financial liabilities"},
		"jpigp_cor:IncomeTaxPayablesIFRS":                 {StmtBS, "current_liabilities", 430, false, "", "Income tax payables"},
		"jpigp_cor:ProvisionsCurrentIFRS":                 {StmtBS, "current_liabilities", 440, false, "", "Provisions (current)"},
		"jpigp_cor:OtherCurrentLiabilitiesIFRS":           {StmtBS, "current_liabilities", 450, false, "", "Other current liabilities"},
		"jpigp_cor:TotalCurrentLiabilitiesIFRS":           {StmtBS, "current_liabilities", 499, true, "current_liabilities", "Total current liabilities"},

		// Non-current Liabilities
		"jpigp_cor:InterestBearingLiabilitiesNCLIFRS":     {StmtBS, "noncurrent_liabilities", 500, false, "interest_bearing_debt", "Interest-bearing liabilities (non-current)"},
		"jpigp_cor:OtherNonCurrentFinancialLiabilitiesIFRS": {StmtBS, "noncurrent_liabilities", 510, false, "", "Other non-current financial liabilities"},
		"jpigp_cor:DeferredTaxLiabilitiesIFRS":            {StmtBS, "noncurrent_liabilities", 520, false, "", "Deferred tax liabilities"},
		"jpigp_cor:ProvisionsNonCurrentIFRS":              {StmtBS, "noncurrent_liabilities", 530, false, "", "Provisions (non-current)"},
		"jpigp_cor:RetirementBenefitLiabilityIFRS":        {StmtBS, "noncurrent_liabilities", 540, false, "", "Retirement benefit liability"},
		"jpigp_cor:OtherNonCurrentLiabilitiesIFRS":        {StmtBS, "noncurrent_liabilities", 550, false, "", "Other non-current liabilities"},
		"jpigp_cor:NonCurrentLiabilitiesIFRS":             {StmtBS, "noncurrent_liabilities", 599, true, "", "Total non-current liabilities"},
		"jpigp_cor:LiabilitiesIFRS":                       {StmtBS, "total", 600, true, "total_liabilities", "Total liabilities"},

		// Equity
		"jpigp_cor:ShareCapitalIFRS":                      {StmtBS, "equity", 700, false, "", "Share capital"},
		"jpigp_cor:CapitalSurplusIFRS":                    {StmtBS, "equity", 710, false, "", "Capital surplus"},
		"jpigp_cor:RetainedEarningsIFRS":                  {StmtBS, "equity", 720, false, "", "Retained earnings"},
		"jpigp_cor:TreasurySharesIFRS":                    {StmtBS, "equity", 730, false, "", "Treasury shares"},
		"jpigp_cor:OtherComponentsOfEquityIFRS":           {StmtBS, "equity", 740, false, "", "Other components of equity"},
		"jpigp_cor:EquityAttributableToOwnersOfParentIFRS": {StmtBS, "equity", 790, true, "equity", "Equity attributable to owners of parent"},
		"jpigp_cor:NonControllingInterestsIFRS":           {StmtBS, "equity", 795, false, "", "Non-controlling interests"},
		"jpigp_cor:EquityIFRS":                            {StmtBS, "equity", 799, true, "net_assets", "Total equity"},

		// ============================================================
		// IFRS Income Statement (jpigp_cor:)
		// ============================================================
		"jpigp_cor:RevenueIFRS":                           {StmtPL, "revenue", 1000, true, "revenue", "Revenue"},
		"jpigp_cor:CostOfSalesIFRS":                       {StmtPL, "cost_of_sales", 1010, false, "cost_of_sales", "Cost of sales"},
		"jpigp_cor:GrossProfitIFRS":                       {StmtPL, "gross_profit", 1020, true, "gross_profit", "Gross profit"},
		"jpigp_cor:SellingGeneralAndAdministrativeExpensesIFRS": {StmtPL, "operating", 1030, false, "sga_expenses", "SGA expenses"},
		"jpigp_cor:OtherIncomeIFRS":                       {StmtPL, "operating", 1040, false, "", "Other income"},
		"jpigp_cor:OtherExpensesIFRS":                     {StmtPL, "operating", 1050, false, "", "Other expenses"},
		"jpigp_cor:OperatingProfitLossIFRS":               {StmtPL, "operating_income", 1060, true, "operating_income", "Operating profit/loss"},
		"jpigp_cor:FinanceIncomeIFRS":                     {StmtPL, "finance", 1070, false, "", "Finance income"},
		"jpigp_cor:FinanceCostsIFRS":                      {StmtPL, "finance", 1080, false, "", "Finance costs"},
		"jpigp_cor:ShareOfProfitLossOfInvestmentsAccountedForUsingEquityMethodIFRS": {StmtPL, "finance", 1090, false, "", "Share of profit of equity method investments"},
		"jpigp_cor:ProfitLossBeforeTaxIFRS":               {StmtPL, "pretax", 1100, true, "", "Profit before tax"},
		"jpigp_cor:IncomeTaxExpenseIFRS":                   {StmtPL, "tax", 1110, false, "", "Income tax expense"},
		"jpigp_cor:ProfitLossIFRS":                        {StmtPL, "net_income", 1120, true, "", "Profit/loss"},
		"jpigp_cor:ProfitLossAttributableToOwnersOfParentIFRS": {StmtPL, "net_income", 1130, true, "net_income", "Profit attributable to owners of parent"},
		"jpigp_cor:ProfitLossAttributableToNonControllingInterestsIFRS": {StmtPL, "net_income", 1140, false, "", "Profit attributable to non-controlling interests"},

		// EPS
		"jpigp_cor:BasicEarningsLossPerShareIFRS":         {StmtPL, "eps", 1200, false, "eps", "Basic EPS"},
		"jpigp_cor:BasicAndDilutedEarningsLossPerShareIFRS": {StmtPL, "eps", 1201, false, "eps", "Basic and diluted EPS"},
		"jpigp_cor:DilutedEarningsLossPerShareIFRS":       {StmtPL, "eps", 1210, false, "", "Diluted EPS"},

		// ============================================================
		// IFRS Cash Flow Statement (jpigp_cor:)
		// ============================================================
		"jpigp_cor:DepreciationAndAmortisationIFRS":       {StmtCF, "operating_activities", 2000, false, "depreciation", "Depreciation and amortisation"},
		"jpigp_cor:ImpairmentLossReversalOfImpairmentLossRecognisedInProfitOrLossIFRS": {StmtCF, "operating_activities", 2010, false, "", "Impairment loss"},
		"jpigp_cor:CashFlowsFromUsedInOperatingActivitiesIFRS": {StmtCF, "operating_activities", 2099, true, "operating_cf", "Cash flows from operating activities"},
		"jpigp_cor:CashFlowsFromUsedInInvestingActivitiesIFRS": {StmtCF, "investing_activities", 2199, true, "investing_cf", "Cash flows from investing activities"},
		"jpigp_cor:CashFlowsFromUsedInFinancingActivitiesIFRS": {StmtCF, "financing_activities", 2299, true, "financing_cf", "Cash flows from financing activities"},
		"jpigp_cor:NetCashProvidedByUsedInOperatingActivitiesIFRS": {StmtCF, "operating_activities", 2099, true, "operating_cf", "Net cash from operating activities"},
		"jpigp_cor:NetCashProvidedByUsedInInvestingActivitiesIFRS": {StmtCF, "investing_activities", 2199, true, "investing_cf", "Net cash from investing activities"},
		"jpigp_cor:NetCashProvidedByUsedInFinancingActivitiesIFRS": {StmtCF, "financing_activities", 2299, true, "financing_cf", "Net cash from financing activities"},
		"jpigp_cor:CapitalExpendituresIFRS":               {StmtCF, "investing_activities", 2110, false, "capital_expenditure", "Capital expenditures"},

		// jpcrp_cor CF summary elements (key financial data / summary of business results)
		"jpcrp_cor:CashFlowsFromUsedInOperatingActivitiesIFRSSummaryOfBusinessResults": {StmtCF, "operating_activities", 2098, true, "operating_cf", "Operating CF (summary)"},
		"jpcrp_cor:CashFlowsFromUsedInInvestingActivitiesIFRSSummaryOfBusinessResults": {StmtCF, "investing_activities", 2198, true, "investing_cf", "Investing CF (summary)"},
		"jpcrp_cor:CashFlowsFromUsedInFinancingActivitiesIFRSSummaryOfBusinessResults": {StmtCF, "financing_activities", 2298, true, "financing_cf", "Financing CF (summary)"},

		// ============================================================
		// JP-GAAP Balance Sheet (jppfs_cor:)
		// ============================================================

		// Current Assets
		"jppfs_cor:CashAndDeposits":                       {StmtBS, "current_assets", 100, false, "cash_and_equivalents", "Cash and deposits"},
		"jppfs_cor:NotesAndAccountsReceivableTrade":       {StmtBS, "current_assets", 110, false, "", "Notes and accounts receivable - trade"},
		"jppfs_cor:NotesAndAccountsReceivableTradeAndContractAssets": {StmtBS, "current_assets", 111, false, "", "Notes and accounts receivable - trade, and contract assets"},
		"jppfs_cor:SecuritiesCurrent":                     {StmtBS, "current_assets", 115, false, "", "Securities (current)"},
		"jppfs_cor:MerchandiseAndFinishedGoods":           {StmtBS, "current_assets", 120, false, "", "Merchandise and finished goods"},
		"jppfs_cor:WorkInProcess":                         {StmtBS, "current_assets", 121, false, "", "Work in process"},
		"jppfs_cor:RawMaterialsAndSupplies":               {StmtBS, "current_assets", 122, false, "", "Raw materials and supplies"},
		"jppfs_cor:OtherCurrentAssets":                    {StmtBS, "current_assets", 140, false, "", "Other current assets"},
		"jppfs_cor:AllowanceForDoubtfulAccountsCurrentAssets": {StmtBS, "current_assets", 145, false, "", "Allowance for doubtful accounts (current)"},
		"jppfs_cor:CurrentAssets":                         {StmtBS, "current_assets", 199, true, "current_assets", "Total current assets"},

		// Non-current Assets
		"jppfs_cor:BuildingsAndStructuresNet":             {StmtBS, "noncurrent_assets", 200, false, "", "Buildings and structures (net)"},
		"jppfs_cor:MachineryEquipmentAndVehiclesNet":      {StmtBS, "noncurrent_assets", 201, false, "", "Machinery, equipment and vehicles (net)"},
		"jppfs_cor:LandNet":                               {StmtBS, "noncurrent_assets", 202, false, "", "Land"},
		"jppfs_cor:ConstructionInProgress":                {StmtBS, "noncurrent_assets", 203, false, "", "Construction in progress"},
		"jppfs_cor:PropertyPlantAndEquipment":             {StmtBS, "noncurrent_assets", 210, true, "", "Total property, plant and equipment"},
		"jppfs_cor:GoodwillNet":                           {StmtBS, "noncurrent_assets", 220, false, "", "Goodwill"},
		"jppfs_cor:IntangibleAssetsNet":                   {StmtBS, "noncurrent_assets", 229, true, "", "Total intangible assets"},
		"jppfs_cor:InvestmentSecurities":                  {StmtBS, "noncurrent_assets", 230, false, "", "Investment securities"},
		"jppfs_cor:InvestmentsAndOtherAssets":             {StmtBS, "noncurrent_assets", 249, true, "", "Total investments and other assets"},
		"jppfs_cor:NoncurrentAssets":                      {StmtBS, "noncurrent_assets", 299, true, "", "Total non-current assets"},
		"jppfs_cor:TotalAssets":                           {StmtBS, "total", 300, true, "total_assets", "Total assets"},

		// Current Liabilities
		"jppfs_cor:NotesAndAccountsPayableTrade":          {StmtBS, "current_liabilities", 400, false, "", "Notes and accounts payable - trade"},
		"jppfs_cor:ShortTermLoansPayable":                 {StmtBS, "current_liabilities", 410, false, "interest_bearing_debt", "Short-term loans payable"},
		"jppfs_cor:CurrentPortionOfLongTermLoansPayable":  {StmtBS, "current_liabilities", 411, false, "interest_bearing_debt", "Current portion of long-term loans payable"},
		"jppfs_cor:CurrentPortionOfBonds":                 {StmtBS, "current_liabilities", 412, false, "interest_bearing_debt", "Current portion of bonds"},
		"jppfs_cor:CommercialPapersLiabilities":           {StmtBS, "current_liabilities", 413, false, "interest_bearing_debt", "Commercial papers"},
		"jppfs_cor:AccruedExpenses":                       {StmtBS, "current_liabilities", 420, false, "", "Accrued expenses"},
		"jppfs_cor:IncomeTaxesPayable":                    {StmtBS, "current_liabilities", 430, false, "", "Income taxes payable"},
		"jppfs_cor:ProvisionForBonuses":                   {StmtBS, "current_liabilities", 435, false, "", "Provision for bonuses"},
		"jppfs_cor:OtherCurrentLiabilities":               {StmtBS, "current_liabilities", 450, false, "", "Other current liabilities"},
		"jppfs_cor:CurrentLiabilities":                    {StmtBS, "current_liabilities", 499, true, "current_liabilities", "Total current liabilities"},

		// Non-current Liabilities
		"jppfs_cor:BondsPayable":                          {StmtBS, "noncurrent_liabilities", 500, false, "interest_bearing_debt", "Bonds payable"},
		"jppfs_cor:LongTermLoansPayable":                  {StmtBS, "noncurrent_liabilities", 510, false, "interest_bearing_debt", "Long-term loans payable"},
		"jppfs_cor:DeferredTaxLiabilities":                {StmtBS, "noncurrent_liabilities", 520, false, "", "Deferred tax liabilities"},
		"jppfs_cor:RetirementBenefitLiability":            {StmtBS, "noncurrent_liabilities", 530, false, "", "Retirement benefit liability"},
		"jppfs_cor:OtherNoncurrentLiabilities":            {StmtBS, "noncurrent_liabilities", 550, false, "", "Other non-current liabilities"},
		"jppfs_cor:NoncurrentLiabilities":                 {StmtBS, "noncurrent_liabilities", 599, true, "", "Total non-current liabilities"},
		"jppfs_cor:TotalLiabilities":                      {StmtBS, "total", 600, true, "total_liabilities", "Total liabilities"},

		// Equity / Net Assets
		"jppfs_cor:CapitalStock":                          {StmtBS, "equity", 700, false, "", "Capital stock"},
		"jppfs_cor:CapitalSurplus":                        {StmtBS, "equity", 710, false, "", "Capital surplus"},
		"jppfs_cor:RetainedEarnings":                      {StmtBS, "equity", 720, false, "", "Retained earnings"},
		"jppfs_cor:TreasuryShares":                        {StmtBS, "equity", 730, false, "", "Treasury shares"},
		"jppfs_cor:ShareholdersEquity":                    {StmtBS, "equity", 790, true, "equity", "Total shareholders' equity"},
		"jppfs_cor:ValuationAndTranslationAdjustments":    {StmtBS, "equity", 792, false, "", "Valuation and translation adjustments"},
		"jppfs_cor:NonControllingInterests":               {StmtBS, "equity", 795, false, "", "Non-controlling interests"},
		"jppfs_cor:NetAssets":                             {StmtBS, "equity", 799, true, "net_assets", "Total net assets"},

		// ============================================================
		// JP-GAAP Income Statement (jppfs_cor:)
		// ============================================================
		"jppfs_cor:NetSales":                              {StmtPL, "revenue", 1000, true, "revenue", "Net sales"},
		"jppfs_cor:CostOfSales":                           {StmtPL, "cost_of_sales", 1010, false, "cost_of_sales", "Cost of sales"},
		"jppfs_cor:GrossProfit":                           {StmtPL, "gross_profit", 1020, true, "gross_profit", "Gross profit"},
		"jppfs_cor:SellingGeneralAndAdministrativeExpenses": {StmtPL, "operating", 1030, false, "sga_expenses", "SGA expenses"},
		"jppfs_cor:OperatingIncome":                       {StmtPL, "operating_income", 1060, true, "operating_income", "Operating income"},
		"jppfs_cor:NonOperatingIncome":                    {StmtPL, "non_operating", 1070, false, "", "Non-operating income"},
		"jppfs_cor:NonOperatingExpenses":                  {StmtPL, "non_operating", 1080, false, "", "Non-operating expenses"},
		"jppfs_cor:OrdinaryIncome":                        {StmtPL, "ordinary_income", 1090, true, "ordinary_income", "Ordinary income"},
		"jppfs_cor:ExtraordinaryIncome":                   {StmtPL, "extraordinary", 1100, false, "", "Extraordinary income"},
		"jppfs_cor:ExtraordinaryLoss":                     {StmtPL, "extraordinary", 1110, false, "", "Extraordinary loss"},
		"jppfs_cor:IncomeBeforeIncomeTaxes":               {StmtPL, "pretax", 1120, true, "", "Income before income taxes"},
		"jppfs_cor:IncomeTaxes":                           {StmtPL, "tax", 1130, false, "", "Income taxes - current"},
		"jppfs_cor:IncomeTaxesDeferred":                   {StmtPL, "tax", 1131, false, "", "Income taxes - deferred"},
		"jppfs_cor:NetIncome":                             {StmtPL, "net_income", 1150, true, "net_income", "Net income"},
		"jppfs_cor:ProfitLoss":                            {StmtPL, "net_income", 1150, true, "net_income", "Profit/loss"},
		"jppfs_cor:NetIncomeAttributableToOwnersOfParent": {StmtPL, "net_income", 1151, true, "net_income", "Net income attributable to owners of parent"},
		"jppfs_cor:ProfitLossAttributableToOwnersOfParent": {StmtPL, "net_income", 1151, true, "net_income", "Profit/loss attributable to owners of parent"},
		"jppfs_cor:NetIncomeAttributableToNonControllingInterests": {StmtPL, "net_income", 1152, false, "", "Net income attributable to non-controlling interests"},

		// ============================================================
		// JP-GAAP Cash Flow Statement (jppfs_cor:)
		// ============================================================
		"jppfs_cor:DepreciationAndAmortization":           {StmtCF, "operating_activities", 2000, false, "depreciation", "Depreciation and amortization"},
		"jppfs_cor:NetCashProvidedByUsedInOperatingActivities": {StmtCF, "operating_activities", 2099, true, "operating_cf", "Cash flows from operating activities"},
		"jppfs_cor:PurchaseOfPropertyPlantAndEquipmentAndIntangibleAssets": {StmtCF, "investing_activities", 2110, false, "capital_expenditure", "Purchase of property, plant and equipment"},
		"jppfs_cor:PurchaseOfPropertyPlantAndEquipment":   {StmtCF, "investing_activities", 2111, false, "capital_expenditure", "Purchase of property, plant and equipment"},
		"jppfs_cor:NetCashProvidedByUsedInInvestingActivities": {StmtCF, "investing_activities", 2199, true, "investing_cf", "Cash flows from investing activities"},
		"jppfs_cor:NetCashProvidedByUsedInFinancingActivities": {StmtCF, "financing_activities", 2299, true, "financing_cf", "Cash flows from financing activities"},
		"jppfs_cor:CashAndCashEquivalents":                {StmtCF, "cash_position", 2400, false, "", "Cash and cash equivalents at end of period"},

		// ============================================================
		// Cross-standard elements (jpcrp_cor:)
		// ============================================================
		"jpcrp_cor:NumberOfIssuedSharesAsOfFilingDateTotal":  {StmtBS, "shares", 3000, false, "shares_outstanding", "Shares outstanding"},
		"jpcrp_cor:NumberOfTreasurySharesAsOfFilingDateTotal": {StmtBS, "shares", 3010, false, "treasury_shares", "Treasury shares"},
		"jpcrp_cor:ResearchAndDevelopmentExpensesTotal":      {StmtPL, "other_pl", 3100, false, "research_and_development", "R&D expenses"},
		"jpcrp_cor:DividendPaidPerShareSummaryOfBusinessResults": {StmtPL, "dividends", 3200, false, "dividend_per_share", "Dividend per share"},
		"jpcrp_cor:BasicEarningsLossPerShare":                {StmtPL, "eps", 3210, false, "eps", "Basic EPS"},
		"jpcrp_cor:BasicEarningsLossPerShareSummaryOfBusinessResults": {StmtPL, "eps", 3211, false, "eps", "Basic EPS (summary)"},
		"jpcrp_cor:DilutedEarningsLossPerShare":              {StmtPL, "eps", 3220, false, "", "Diluted EPS"},

		// ============================================================
		// Additional IFRS elements commonly seen
		// ============================================================
		"jpigp_cor:ProfitLossBeforeTaxFromContinuingOperationsIFRS": {StmtPL, "pretax", 1095, true, "", "Profit before tax from continuing operations"},
		"jpigp_cor:ComprehensiveIncomeIFRS":                {StmtPL, "comprehensive_income", 1300, true, "", "Comprehensive income"},
		"jpigp_cor:ComprehensiveIncomeAttributableToOwnersOfParentIFRS": {StmtPL, "comprehensive_income", 1310, true, "", "Comprehensive income attributable to owners of parent"},

		// ============================================================
		// Additional JP-GAAP elements
		// ============================================================
		"jppfs_cor:DeferredTaxAssets":                      {StmtBS, "current_assets", 141, false, "", "Deferred tax assets (current)"},
		"jppfs_cor:DeferredTaxAssetsNCA":                   {StmtBS, "noncurrent_assets", 251, false, "", "Deferred tax assets (non-current)"},
		"jppfs_cor:ResearchAndDevelopmentExpenses":         {StmtPL, "other_pl", 3101, false, "research_and_development", "R&D expenses"},

		// ============================================================
		// SummaryOfBusinessResults fallback elements (jpcrp_cor:)
		// These have higher SortOrder than main statement elements so
		// populateSummary's "first wins" rule prefers detailed values.
		// ============================================================

		// JP-GAAP SummaryOfBusinessResults
		"jpcrp_cor:NetSalesSummaryOfBusinessResults":            {StmtPL, "revenue", 1008, true, "revenue", "Net sales (summary)"},
		"jpcrp_cor:OperatingIncomeSummaryOfBusinessResults":     {StmtPL, "operating_income", 1068, true, "operating_income", "Operating income (summary)"},
		"jpcrp_cor:OrdinaryIncomeSummaryOfBusinessResults":      {StmtPL, "ordinary_income", 1098, true, "ordinary_income", "Ordinary income (summary)"},
		"jpcrp_cor:NetIncomeSummaryOfBusinessResults":           {StmtPL, "net_income", 1158, true, "net_income", "Net income (summary)"},
		"jpcrp_cor:ProfitLossAttributableToOwnersOfParentSummaryOfBusinessResults": {StmtPL, "net_income", 1159, true, "net_income", "Net income attributable to parent (summary)"},
		"jpcrp_cor:TotalAssetsSummaryOfBusinessResults":         {StmtBS, "total", 308, true, "total_assets", "Total assets (summary)"},
		"jpcrp_cor:NetAssetsSummaryOfBusinessResults":           {StmtBS, "equity", 808, true, "net_assets", "Net assets (summary)"},

		// IFRS SummaryOfBusinessResults
		"jpcrp_cor:RevenueIFRSSummaryOfBusinessResults":         {StmtPL, "revenue", 1008, true, "revenue", "Revenue IFRS (summary)"},
		"jpcrp_cor:ProfitLossAttributableToOwnersOfParentIFRSSummaryOfBusinessResults": {StmtPL, "net_income", 1138, true, "net_income", "Net income IFRS (summary)"},

		// US GAAP SummaryOfBusinessResults
		"jpcrp_cor:RevenuesUSGAAPSummaryOfBusinessResults":      {StmtPL, "revenue", 1008, true, "revenue", "Revenue US GAAP (summary)"},
		"jpcrp_cor:NetIncomeLossAttributableToOwnersOfParentUSGAAPSummaryOfBusinessResults": {StmtPL, "net_income", 1158, true, "net_income", "Net income US GAAP (summary)"},
		"jpcrp_cor:TotalAssetsUSGAAPSummaryOfBusinessResults":   {StmtBS, "total", 308, true, "total_assets", "Total assets US GAAP (summary)"},
		"jpcrp_cor:EquityIncludingPortionAttributableToNonControllingInterestUSGAAPSummaryOfBusinessResults": {StmtBS, "equity", 808, true, "net_assets", "Equity US GAAP (summary)"},
	}

	// Company-specific suffix mappings for jpcrp030000-asr_* elements.
	companySuffixes = map[string]elementDef{
		"SalesRevenuesIFRS":                              {StmtPL, "revenue", 1001, true, "revenue", "Sales revenues (IFRS)"},
		"OperatingRevenuesIFRS":                          {StmtPL, "revenue", 1002, true, "revenue", "Operating revenues (IFRS)"},
		"OperatingRevenuesIFRSKeyFinancialData":          {StmtPL, "revenue", 1003, true, "revenue", "Operating revenues (IFRS, key financial data)"},
		"SalesRevenuesIFRSKeyFinancialData":              {StmtPL, "revenue", 1004, true, "revenue", "Sales revenues (IFRS, key financial data)"},
		"OperatingProfitLossIFRS":                        {StmtPL, "operating_income", 1061, true, "operating_income", "Operating profit/loss (company-specific IFRS)"},
		"RevenueIFRS":                                    {StmtPL, "revenue", 1005, true, "revenue", "Revenue (company-specific IFRS)"},
		"NetSales":                                       {StmtPL, "revenue", 1006, true, "revenue", "Net sales (company-specific)"},

		// Company-specific revenue variants (e.g., Sony's financial services revenue)
		// SortOrder: company-specific (1007) before SummaryOfBusinessResults (1008),
		// and KeyFinancialData variants at same level (1007) to match precedence.
		"SalesAndFinancialServicesRevenueIFRS":            {StmtPL, "revenue", 1007, true, "revenue", "Sales and financial services revenue (IFRS)"},
		"SalesAndFinancialServicesRevenueIFRSKeyFinancialData": {StmtPL, "revenue", 1007, true, "revenue", "Sales and financial services revenue (IFRS, key financial data)"},
		"OperatingProfitLossIFRSKeyFinancialData":         {StmtPL, "operating_income", 1067, true, "operating_income", "Operating profit/loss (IFRS, key financial data)"},
		"NetSalesKeyFinancialData":                        {StmtPL, "revenue", 1007, true, "revenue", "Net sales (key financial data)"},
	}
}

// Classify determines the financial statement type and metadata for an XBRL element.
// The pointType parameter ("instant" or "duration") is used for keyword-based heuristic
// matching of unknown elements.
func Classify(elementID string, pointType string) ElementClassification {
	if elementID == "" {
		return ElementClassification{Statement: StmtUnknown}
	}

	// 1. Check exact match in known elements
	if def, ok := knownElements[elementID]; ok {
		return ElementClassification{
			Statement:  def.statement,
			Category:   def.category,
			SortOrder:  def.sortOrder,
			IsTotal:    def.isTotal,
			SummaryKey: def.summaryKey,
		}
	}

	// 2. Check company-specific elements (jpcrp030000-asr_*)
	if strings.HasPrefix(elementID, "jpcrp030000-asr_") {
		if suffix := elementLocalName(elementID); suffix != elementID {
			if def, ok := companySuffixes[suffix]; ok {
				return ElementClassification{
					Statement:  def.statement,
					Category:   def.category,
					SortOrder:  def.sortOrder,
					IsTotal:    def.isTotal,
					SummaryKey: def.summaryKey,
				}
			}
		}
	}

	// 3. Keyword + pointType heuristic for unknown elements (no SummaryKey)
	return classifyByKeyword(elementID, pointType)
}

// classifyByKeyword uses keyword matching and pointType to classify unknown elements.
// Only positive matches are returned; no fallback to PL or BS.
func classifyByKeyword(elementID string, pointType string) ElementClassification {
	localName := elementLocalName(elementID)
	upper := strings.ToUpper(localName)

	// CF keywords — must be duration
	if pointType == "duration" {
		cfKeywords := []string{"CASHFLOW", "CASH_FLOW", "CASHPROVIDED", "CASHUSED",
			"OPERATINGACTIVIT", "INVESTINGACTIVIT", "FINANCINGACTIVIT"}
		for _, kw := range cfKeywords {
			if strings.Contains(upper, kw) {
				return ElementClassification{Statement: StmtCF, Category: "heuristic"}
			}
		}
	}

	// BS keywords — must be instant
	if pointType == "instant" {
		bsKeywords := []string{"ASSET", "LIABILIT", "EQUITY", "CAPITAL",
			"RECEIVABLE", "PAYABLE", "INVENTORY", "INVENTORIES",
			"GOODWILL", "INTANGIBLE", "SECURITIES", "DEPOSIT",
			"PROVISION", "RESERVE", "SURPLUS", "SHARES", "NETASSETS"}
		for _, kw := range bsKeywords {
			if strings.Contains(upper, kw) {
				return ElementClassification{Statement: StmtBS, Category: "heuristic"}
			}
		}
	}

	// PL keywords — must be duration
	if pointType == "duration" {
		plKeywords := []string{"REVENUE", "SALES", "INCOME", "LOSS", "PROFIT",
			"EXPENSE", "COST", "EARNING", "DEPRECIATION", "AMORTIZATION", "AMORTISATION"}
		for _, kw := range plKeywords {
			if strings.Contains(upper, kw) {
				return ElementClassification{Statement: StmtPL, Category: "heuristic"}
			}
		}
	}

	return ElementClassification{Statement: StmtUnknown}
}

// IsTextBlock returns true if the element ID represents a text block element
// that should be excluded from financial statement processing.
func IsTextBlock(elementID string) bool {
	return strings.HasSuffix(elementID, "TextBlock")
}

// Elements returns all known element mappings as a slice of ElementInfo.
// This is the single source of truth for the schema command.
func Elements() []ElementInfo {
	result := make([]ElementInfo, 0, len(knownElements))
	for id, def := range knownElements {
		result = append(result, ElementInfo{
			ID:         id,
			Statement:  def.statement,
			Category:   def.category,
			SummaryKey: def.summaryKey,
			LabelEN:    def.labelEN,
		})
	}
	// Also include company-specific suffixes with a prefix indicator
	for suffix, def := range companySuffixes {
		result = append(result, ElementInfo{
			ID:         "jpcrp030000-asr_*:" + suffix,
			Statement:  def.statement,
			Category:   def.category,
			SummaryKey: def.summaryKey,
			LabelEN:    def.labelEN,
		})
	}
	return result
}
