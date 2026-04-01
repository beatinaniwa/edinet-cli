package financial

import "sort"

// supplementalKeys are summary keys that represent per-share or non-core items.
// A period containing only supplemental keys is not considered a meaningful financial snapshot.
var supplementalKeys = map[string]bool{
	"dividend_per_share":     true,
	"eps":                    true,
	"research_and_development": true,
	"shares_outstanding":     true,
	"treasury_shares":        true,
}

// BuildAndDeriveSummary builds the summary from statements and calculates derived metrics.
// Returns the summary and the period name that was used as the primary data source.
// For annual reports this is typically "current"; for filing documents it may be "prior1" etc.
// An empty string means only filing_date items were available (or no data at all).
func BuildAndDeriveSummary(statements []FinancialStatement) (Summary, string) {
	summary := make(Summary)
	period := populateSummary(summary, statements)
	DeriveMetrics(summary)
	return summary, period
}

// additiveKeys are summary keys whose values are summed across multiple line items
// (e.g., short-term + long-term debt), rather than using first-wins.
var additiveKeys = map[string]bool{
	"interest_bearing_debt": true,
}

// populateSummary selects the best available period across all statements and extracts
// summary items from it. Filing_date items are always included as supplemental data.
// For annual reports the best period is "current"; for filing documents (IPO prospectuses)
// it falls back to "prior1" etc. Returns the selected period name, or "" if only
// filing_date data was available.
func populateSummary(summary Summary, statements []FinancialStatement) string {
	// Phase 1: collect distinct periods and check which ones contain non-supplemental keys.
	const estPeriods = 10
	seen := make(map[string]bool, estPeriods)
	hasCore := make(map[string]bool, estPeriods)
	periods := make([]string, 0, estPeriods)

	for _, stmt := range statements {
		for _, pd := range stmt.Periods {
			if pd.Period == "filing_date" {
				continue
			}
			if !seen[pd.Period] {
				seen[pd.Period] = true
				periods = append(periods, pd.Period)
			}
			if hasCore[pd.Period] {
				continue // already confirmed core key for this period
			}
			for _, item := range pd.Items {
				if item.SummaryKey != "" && item.Value != nil && !supplementalKeys[item.SummaryKey] {
					hasCore[pd.Period] = true
					break
				}
			}
		}
	}

	// Phase 2: sort periods by priority and select bestPeriod.
	sort.Slice(periods, func(i, j int) bool {
		return periodOrder(periods[i]) < periodOrder(periods[j])
	})

	bestPeriod := ""
	for _, p := range periods {
		if hasCore[p] {
			bestPeriod = p
			break
		}
	}
	// If no period has core keys, fall back to the highest-priority period.
	// This handles edge cases like EPS-only or dividend-only statements.
	if bestPeriod == "" && len(periods) > 0 {
		bestPeriod = periods[0]
	}

	// Phase 3: extract summary items from bestPeriod, then supplement from filing_date.
	// Additive keys (e.g. interest_bearing_debt) are accumulated within each period
	// (CL + NCL debt), but once set by one period they are not overwritten by another.
	extractPeriod := func(target string) {
		// Track which additive keys were already set before this period.
		frozenAdditive := make(map[string]bool, len(additiveKeys))
		for k := range additiveKeys {
			if _, exists := summary[k]; exists {
				frozenAdditive[k] = true
			}
		}

		for _, stmt := range statements {
			for _, pd := range stmt.Periods {
				if pd.Period != target {
					continue
				}
				for _, item := range pd.Items {
					if item.SummaryKey == "" || item.Value == nil {
						continue
					}
					if additiveKeys[item.SummaryKey] {
						if frozenAdditive[item.SummaryKey] {
							continue // already set by a prior period
						}
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

	if bestPeriod != "" {
		extractPeriod(bestPeriod)
	}
	extractPeriod("filing_date")

	return bestPeriod
}
