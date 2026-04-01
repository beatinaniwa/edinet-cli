package financial

// Summary holds key financial figures for the current period.
// Keys include both extracted values (e.g., "revenue", "total_assets") from financial
// statements and derived metrics (e.g., "roe", "operating_margin") calculated from them.
// Use schema derived-metrics to discover which keys are derived and their formulas.
// nil values indicate the item was not found or not applicable.
type Summary map[string]*float64

// LineItem is a single row in a financial statement.
type LineItem struct {
	ElementID  string   `json:"element_id"`
	Label      string   `json:"label"`
	Category   string   `json:"category,omitempty"`
	Value      *float64 `json:"value"`
	RawValue   string   `json:"raw_value,omitempty"`
	Unit       string   `json:"unit,omitempty"`
	UnitID     string   `json:"unit_id,omitempty"`
	SummaryKey string   `json:"summary_key,omitempty"`
	IsTotal    bool     `json:"is_total,omitempty"`
}

// PeriodData holds line items for a specific period.
type PeriodData struct {
	Period string     `json:"period"`
	Items  []LineItem `json:"items"`
}

// FinancialStatement is a structured BS, PL, or CF statement.
type FinancialStatement struct {
	Type          string       `json:"type"`
	Consolidated  bool         `json:"consolidated"`
	AccountingStd string       `json:"accounting_standard"`
	PeriodEnd     string       `json:"period_end,omitempty"`
	PeriodStart   string       `json:"period_start,omitempty"`
	Periods       []PeriodData `json:"periods"`
}

// ParseResult is the output of the CSV parser before service-layer metadata is added.
type ParseResult struct {
	Summary       Summary              `json:"summary"`
	Statements    []FinancialStatement `json:"statements"`
	AccountingStd string               `json:"accounting_standard"`
	Consolidated  bool                 `json:"consolidated"`
	Warnings      []string             `json:"warnings,omitempty"`
}

// CompanyFinancialsResult is the output of company financials command.
type CompanyFinancialsResult struct {
	Company  CompanyInfo     `json:"company"`
	Periods  []FinancialData `json:"periods"`
	Warnings []string        `json:"warnings,omitempty"`
}

// CompanyInfo identifies a company in the output.
type CompanyInfo struct {
	EdinetCode string `json:"edinet_code"`
	SecCode    string `json:"sec_code,omitempty"`
	Name       string `json:"name,omitempty"`
}

// StripStatements removes detailed statement data, keeping only summary and metadata.
// After stripping, JSON output will contain "statements": null.
func (d *FinancialData) StripStatements() {
	d.Statements = nil
}

// FinancialData is the final output of the service layer with document metadata.
type FinancialData struct {
	DocID         string               `json:"doc_id"`
	CompanyName   string               `json:"company_name,omitempty"`
	EdinetCode    string               `json:"edinet_code,omitempty"`
	SecCode       string               `json:"sec_code,omitempty"`
	FiscalYear    string               `json:"fiscal_year,omitempty"`
	AccountingStd string               `json:"accounting_standard"`
	Consolidated  bool                 `json:"consolidated"`
	Summary       Summary              `json:"summary"`
	Statements    []FinancialStatement `json:"statements"`
	Warnings      []string             `json:"warnings,omitempty"`
}
