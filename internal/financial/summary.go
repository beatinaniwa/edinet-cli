package financial

// BuildAndDeriveSummary builds the summary from statements and calculates derived metrics.
// This is the single entry point for summary construction, used by both the parser
// and the service layer's statement-filtering path.
func BuildAndDeriveSummary(statements []FinancialStatement) Summary {
	summary := make(Summary)
	populateSummary(summary, statements)
	DeriveMetrics(summary)
	return summary
}

// populateSummary extracts key financial figures from the current period items
// of the given statements and stores them in the summary map.
func populateSummary(summary Summary, statements []FinancialStatement) {
	additiveKeys := map[string]bool{
		"interest_bearing_debt": true,
	}

	for _, stmt := range statements {
		for _, pd := range stmt.Periods {
			if pd.Period != "current" && pd.Period != "filing_date" {
				continue
			}
			for _, item := range pd.Items {
				if item.SummaryKey == "" || item.Value == nil {
					continue
				}

				if additiveKeys[item.SummaryKey] {
					existing := summary[item.SummaryKey]
					if existing == nil {
						v := *item.Value
						summary[item.SummaryKey] = &v
					} else {
						v := *existing + *item.Value
						summary[item.SummaryKey] = &v
					}
				} else {
					if _, exists := summary[item.SummaryKey]; !exists {
						v := *item.Value
						summary[item.SummaryKey] = &v
					}
				}
			}
		}
	}
}
