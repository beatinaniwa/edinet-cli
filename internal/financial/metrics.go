package financial

// DerivedMetricDef describes a derived financial metric for schema output.
type DerivedMetricDef struct {
	Key         string   `json:"key"`
	Formula     string   `json:"formula"`
	Description string   `json:"description"`
	Requires    []string `json:"requires"`
}

// DerivedMetricDefs returns metadata for all derived metrics.
// This is the single source of truth used by both DeriveMetrics() and schema output.
func DerivedMetricDefs() []DerivedMetricDef {
	return []DerivedMetricDef{
		{Key: "gross_margin", Formula: "gross_profit / revenue", Description: "Gross profit margin", Requires: []string{"gross_profit", "revenue"}},
		{Key: "operating_margin", Formula: "operating_income / revenue", Description: "Operating profit margin", Requires: []string{"operating_income", "revenue"}},
		{Key: "net_margin", Formula: "net_income / revenue", Description: "Net profit margin", Requires: []string{"net_income", "revenue"}},
		{Key: "roe", Formula: "net_income / equity (or net_assets)", Description: "Return on equity (ending-balance approximation)", Requires: []string{"net_income", "equity or net_assets"}},
		{Key: "roa", Formula: "net_income / total_assets", Description: "Return on assets (ending-balance approximation)", Requires: []string{"net_income", "total_assets"}},
		{Key: "equity_ratio", Formula: "equity (or net_assets) / total_assets", Description: "Equity ratio", Requires: []string{"equity or net_assets", "total_assets"}},
		{Key: "current_ratio", Formula: "current_assets / current_liabilities", Description: "Current ratio", Requires: []string{"current_assets", "current_liabilities"}},
		{Key: "fcf", Formula: "operating_cf + investing_cf", Description: "Free cash flow (simplified)", Requires: []string{"operating_cf", "investing_cf"}},
		{Key: "debt_to_equity", Formula: "interest_bearing_debt / equity (or net_assets)", Description: "Debt-to-equity ratio", Requires: []string{"interest_bearing_debt", "equity or net_assets"}},
	}
}

// DeriveMetrics calculates derived financial metrics and adds them to the summary.
// Metrics are skipped silently when prerequisite values are missing or denominators are zero.
// Existing keys are never overwritten.
func DeriveMetrics(summary Summary) {
	// Helper to get a value or nil
	get := func(key string) *float64 { return summary[key] }

	// Helper to set a value only if the key doesn't already exist
	set := func(key string, val float64) {
		if _, exists := summary[key]; !exists {
			v := val
			summary[key] = &v
		}
	}

	// Helper for safe division (returns nil if denominator is zero or nil)
	div := func(num, denom *float64) *float64 {
		if num == nil || denom == nil || *denom == 0 {
			return nil
		}
		v := *num / *denom
		return &v
	}

	// Equity with fallback to net_assets
	eq := equityValue(summary)

	// Margin metrics
	if v := div(get("gross_profit"), get("revenue")); v != nil {
		set("gross_margin", *v)
	}
	if v := div(get("operating_income"), get("revenue")); v != nil {
		set("operating_margin", *v)
	}
	if v := div(get("net_income"), get("revenue")); v != nil {
		set("net_margin", *v)
	}

	// Return metrics
	if v := div(get("net_income"), eq); v != nil {
		set("roe", *v)
	}
	if v := div(get("net_income"), get("total_assets")); v != nil {
		set("roa", *v)
	}

	// Balance sheet ratios
	if v := div(eq, get("total_assets")); v != nil {
		set("equity_ratio", *v)
	}
	if v := div(get("current_assets"), get("current_liabilities")); v != nil {
		set("current_ratio", *v)
	}
	if v := div(get("interest_bearing_debt"), eq); v != nil {
		set("debt_to_equity", *v)
	}

	// FCF (additive, not a ratio)
	ocf := get("operating_cf")
	icf := get("investing_cf")
	if ocf != nil && icf != nil {
		set("fcf", *ocf+*icf)
	}
}

// equityValue returns equity if present, otherwise falls back to net_assets.
func equityValue(summary Summary) *float64 {
	if v := summary["equity"]; v != nil {
		return v
	}
	return summary["net_assets"]
}
