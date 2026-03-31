package company

import (
	"regexp"
	"strings"
)

var (
	edinetCodePattern = regexp.MustCompile(`^E\d{5}$`)
	secCodePattern    = regexp.MustCompile(`^\d{4,5}$`)
)

// Search finds entries matching the query.
// - EDINET code pattern (E\d{5}) → exact match on EdinetCode
// - Numeric (4-5 digits) → exact match on SecCode (4-digit matches prefix of 5-digit)
// - Otherwise → substring match on SubmitterName, SubmitterNameEN, SubmitterNameKana
func (r *Registry) Search(query string) []CompanyEntry {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	// EDINET code exact match
	if edinetCodePattern.MatchString(query) {
		for _, e := range r.Entries {
			if e.EdinetCode == query {
				return []CompanyEntry{e}
			}
		}
		return nil
	}

	// Securities code exact match
	if secCodePattern.MatchString(query) {
		for _, e := range r.Entries {
			if e.SecCode == query {
				return []CompanyEntry{e}
			}
			if len(query) == 4 && len(e.SecCode) == 5 && strings.HasPrefix(e.SecCode, query) {
				return []CompanyEntry{e}
			}
		}
		return nil
	}

	// Name substring match (case-insensitive for English)
	queryUpper := strings.ToUpper(query)
	var results []CompanyEntry
	for _, e := range r.Entries {
		if strings.Contains(e.SubmitterName, query) ||
			strings.Contains(strings.ToUpper(e.SubmitterNameEN), queryUpper) ||
			strings.Contains(e.SubmitterNameKana, query) {
			results = append(results, e)
		}
	}
	return results
}

// SearchByIndustry finds entries matching the industry name (substring match).
func (r *Registry) SearchByIndustry(industry string) []CompanyEntry {
	industry = strings.TrimSpace(industry)
	if industry == "" {
		return nil
	}
	var results []CompanyEntry
	for _, e := range r.Entries {
		if strings.Contains(e.IndustryName, industry) {
			results = append(results, e)
		}
	}
	return results
}
