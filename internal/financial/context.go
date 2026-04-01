package financial

import "strings"

// ContextInfo holds parsed information from an EDINET CSV context ID.
type ContextInfo struct {
	Period       string // "current", "prior1", "prior2", "prior3", "prior4", "filing_date", etc.
	PointType    string // "instant", "duration"
	Consolidated string // "consolidated", "non_consolidated", "other", "unknown"
	Member       string // segment/equity member name if any, "" otherwise
}

// ParseContextID parses an EDINET CSV context ID string and consolidated column value
// into a structured ContextInfo.
//
// Context ID examples:
//   - "CurrentYearInstant" → current/instant
//   - "Prior1YearDuration" → prior1/duration
//   - "CurrentYearInstant_NonConsolidatedMember" → current/instant/non_consolidated
//   - "CurrentYearDuration_SomeSegmentMember" → current/duration with member
//   - "FilingDateInstant" → filing_date/instant
func ParseContextID(contextID string, consolidatedCol string) ContextInfo {
	if contextID == "" {
		return ContextInfo{Consolidated: parseConsolidatedCol(consolidatedCol)}
	}

	var info ContextInfo

	// Split on first underscore to separate period+type from member
	base := contextID
	member := ""
	if idx := strings.Index(contextID, "_"); idx >= 0 {
		base = contextID[:idx]
		member = contextID[idx+1:]
	}

	// Parse period and point type from base
	info.Period, info.PointType = parsePeriodAndType(base)

	// Parse member
	if member == "NonConsolidatedMember" {
		// NonConsolidatedMember overrides the column — it is always non-consolidated,
		// even when 連結・個別 says その他 (common in IFRS filings).
		info.Member = ""
		info.Consolidated = "non_consolidated"
	} else {
		if member != "" {
			info.Member = member
		}
		info.Consolidated = parseConsolidatedCol(consolidatedCol)
	}

	return info
}

func parsePeriodAndType(base string) (period, pointType string) {
	switch {
	case strings.HasPrefix(base, "CurrentYear"):
		period = "current"
		pointType = parsePointType(base[len("CurrentYear"):])
	case strings.HasPrefix(base, "CurrentQuarter"):
		period = "current_quarter"
		pointType = parsePointType(base[len("CurrentQuarter"):])
	case strings.HasPrefix(base, "CurrentInterim"):
		period = "current_interim"
		pointType = parsePointType(base[len("CurrentInterim"):])
	case strings.HasPrefix(base, "CurrentYTD"):
		period = "current_ytd"
		pointType = parsePointType(base[len("CurrentYTD"):])
	case strings.HasPrefix(base, "Prior") && len(base) > 5:
		// Prior1YearInstant, Prior2YearDuration, Prior1QuarterDuration, etc.
		numEnd := 5 // after "Prior"
		for numEnd < len(base) && base[numEnd] >= '0' && base[numEnd] <= '9' {
			numEnd++
		}
		if numEnd > 5 {
			num := base[5:numEnd]
			rest := base[numEnd:]
			switch {
			case strings.HasPrefix(rest, "Year"):
				period = "prior" + num
				pointType = parsePointType(rest[len("Year"):])
			case strings.HasPrefix(rest, "Quarter"):
				period = "prior" + num + "_quarter"
				pointType = parsePointType(rest[len("Quarter"):])
			case strings.HasPrefix(rest, "Interim"):
				period = "prior" + num + "_interim"
				pointType = parsePointType(rest[len("Interim"):])
			case strings.HasPrefix(rest, "YTD"):
				period = "prior" + num + "_ytd"
				pointType = parsePointType(rest[len("YTD"):])
			}
		}
	case strings.HasPrefix(base, "FilingDate"):
		period = "filing_date"
		pointType = parsePointType(base[len("FilingDate"):])
	}
	return period, pointType
}

func parsePointType(suffix string) string {
	switch suffix {
	case "Instant":
		return "instant"
	case "Duration":
		return "duration"
	default:
		return ""
	}
}

func parseConsolidatedCol(col string) string {
	switch col {
	case "連結":
		return "consolidated"
	case "個別":
		return "non_consolidated"
	case "その他":
		return "other"
	default:
		return "unknown"
	}
}
