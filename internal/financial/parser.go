package financial

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/beatinaniwa/edinet-cli/internal/extract"
)

// ParseOpts configures the parser behavior.
type ParseOpts struct {
	Consolidated *bool // nil=auto per-statement, true/false=explicit
}

// requiredHeaders are the columns the parser needs to process a file.
var requiredHeaders = []string{"要素ID", "項目名", "コンテキストID", "連結・個別", "ユニットID", "単位", "値"}

// parsedRow is an intermediate representation of a CSV row after initial parsing.
type parsedRow struct {
	elementID      string
	label          string
	contextInfo    ContextInfo
	rawContextID   string
	unitID         string
	unit           string
	rawValue       string
	value          *float64
	classification ElementClassification
}

// consolidationGroup holds rows grouped by consolidation type for a statement.
type consolidationGroup struct {
	consolidated    []parsedRow
	nonConsolidated []parsedRow
	other           []parsedRow
}

// dedupeKey uniquely identifies an element within a statement and period.
type dedupeKey struct {
	stmtType  StatementType
	period    string
	elementID string
}

// Parse converts raw CSVDataResult into structured ParseResult with summary.
func Parse(csvResult *extract.CSVDataResult, opts ParseOpts) (*ParseResult, error) {
	if csvResult == nil {
		return nil, fmt.Errorf("csv result is nil")
	}
	if len(csvResult.Files) == 0 {
		return nil, fmt.Errorf("no CSV files to parse")
	}

	// Sort files: main file (001) first, then others in order
	files := sortFiles(csvResult.Files)

	// Track warnings
	var warnings []string

	// Check annual report marker
	if !hasASRMarker(files) {
		warnings = append(warnings, "no annual report (asr) marker found in filenames; data may be from a quarterly or other report")
	}

	// Parse all rows from eligible files
	var allRows []parsedRow
	processedAny := false

	for _, file := range files {
		// File selection: only jpcrp prefix files, skip jpaud
		if !isEligibleFile(file.Filename) {
			continue
		}

		// Header validation
		colMap, err := buildColumnMap(file.Headers)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skipped file %s: %s", file.Filename, err.Error()))
			continue
		}

		processedAny = true

		// Parse rows from this file
		for _, row := range file.Rows {
			parsed := parseRow(row, colMap)
			if parsed == nil {
				continue
			}
			allRows = append(allRows, *parsed)
		}
	}

	if !processedAny {
		return nil, fmt.Errorf("no eligible CSV files found (all files were skipped)")
	}

	// Build result from parsed rows
	return buildResult(allRows, opts, warnings)
}

// sortFiles returns files sorted with main file (001) first.
func sortFiles(files []extract.CSVFile) []extract.CSVFile {
	sorted := make([]extract.CSVFile, len(files))
	copy(sorted, files)
	sort.SliceStable(sorted, func(i, j int) bool {
		iMain := isMainFile(sorted[i].Filename)
		jMain := isMainFile(sorted[j].Filename)
		if iMain != jMain {
			return iMain
		}
		return false
	})
	return sorted
}

// isMainFile returns true if the filename matches the main file pattern (contains "-001").
func isMainFile(filename string) bool {
	return strings.Contains(filename, "-001")
}

// isEligibleFile returns true if the file should be processed.
func isEligibleFile(filename string) bool {
	return strings.HasPrefix(strings.ToLower(filename), "jpcrp")
}

// hasASRMarker checks if any eligible file has "asr" in its filename.
func hasASRMarker(files []extract.CSVFile) bool {
	for _, f := range files {
		lower := strings.ToLower(f.Filename)
		if strings.HasPrefix(lower, "jpcrp") && strings.Contains(lower, "-asr-") {
			return true
		}
	}
	return false
}

// columnMap maps header names to column indices.
type columnMap struct {
	elementID    int
	label        int
	contextID    int
	consolidated int
	pointType    int
	unitID       int
	unit         int
	value        int
}

// buildColumnMap validates headers and returns a column mapping.
func buildColumnMap(headers []string) (*columnMap, error) {
	idx := make(map[string]int)
	for i, h := range headers {
		idx[h] = i
	}

	for _, req := range requiredHeaders {
		if _, ok := idx[req]; !ok {
			return nil, fmt.Errorf("missing required header %q", req)
		}
	}

	cm := &columnMap{
		elementID:    idx["要素ID"],
		label:        idx["項目名"],
		contextID:    idx["コンテキストID"],
		consolidated: idx["連結・個別"],
		unitID:       idx["ユニットID"],
		unit:         idx["単位"],
		value:        idx["値"],
	}

	if i, ok := idx["期間・時点"]; ok {
		cm.pointType = i
	} else {
		cm.pointType = -1
	}

	return cm, nil
}

// safeGet returns the value at index i in row, or "" if out of bounds.
func safeGet(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}

// parseRow parses a single CSV row into a parsedRow, or nil if the row should be skipped.
func parseRow(row []string, cm *columnMap) *parsedRow {
	elementID := safeGet(row, cm.elementID)
	if elementID == "" {
		return nil
	}

	// Skip TextBlock elements
	if IsTextBlock(elementID) {
		return nil
	}

	contextID := safeGet(row, cm.contextID)
	consolidatedCol := safeGet(row, cm.consolidated)

	// Parse context
	ctxInfo := ParseContextID(contextID, consolidatedCol)

	// Skip segment members
	if ctxInfo.Member != "" {
		return nil
	}

	// Parse value
	rawValue := safeGet(row, cm.value)
	var value *float64
	if rawValue != "" && rawValue != "－" {
		if v, err := strconv.ParseFloat(rawValue, 64); err == nil {
			value = &v
		}
	}

	// Classify element
	classification := Classify(elementID, ctxInfo.PointType)

	return &parsedRow{
		elementID:      elementID,
		label:          safeGet(row, cm.label),
		contextInfo:    ctxInfo,
		rawContextID:   contextID,
		unitID:         safeGet(row, cm.unitID),
		unit:           safeGet(row, cm.unit),
		rawValue:       rawValue,
		value:          value,
		classification: classification,
	}
}

// buildResult constructs the ParseResult from parsed rows.
func buildResult(rows []parsedRow, opts ParseOpts, warnings []string) (*ParseResult, error) {
	// Group rows by statement type and consolidation
	byStmt := make(map[StatementType]*consolidationGroup)
	for _, st := range []StatementType{StmtBS, StmtPL, StmtCF} {
		byStmt[st] = &consolidationGroup{}
	}

	for _, r := range rows {
		if r.classification.Statement == StmtUnknown {
			continue
		}
		sr, ok := byStmt[r.classification.Statement]
		if !ok {
			continue
		}
		switch r.contextInfo.Consolidated {
		case "consolidated":
			sr.consolidated = append(sr.consolidated, r)
		case "non_consolidated":
			sr.nonConsolidated = append(sr.nonConsolidated, r)
		default:
			sr.other = append(sr.other, r)
		}
	}

	// Select consolidation per statement type
	type stmtSelection struct {
		stmtType     StatementType
		rows         []parsedRow
		consolidated bool
	}

	var selections []stmtSelection
	hasConsolidatedStmt := false

	for _, st := range []StatementType{StmtBS, StmtPL, StmtCF} {
		sr := byStmt[st]
		selected, isCons := selectConsolidation(sr, st, opts, &warnings)
		if len(selected) > 0 {
			selections = append(selections, stmtSelection{
				stmtType:     st,
				rows:         selected,
				consolidated: isCons,
			})
			if isCons {
				hasConsolidatedStmt = true
			}
		}
	}

	// Detect accounting standard from selected rows only
	var selectedRows []parsedRow
	for _, sel := range selections {
		selectedRows = append(selectedRows, sel.rows...)
	}
	acctStd := detectAccountingStandard(selectedRows)

	// Build statements
	var statements []FinancialStatement
	seen := make(map[dedupeKey]bool)

	for _, sel := range selections {
		stmt := buildStatement(sel.stmtType, sel.rows, sel.consolidated, acctStd, seen)
		if stmt != nil {
			statements = append(statements, *stmt)
		}
	}

	// Build summary from current period items and derive metrics
	return &ParseResult{
		Summary:       BuildAndDeriveSummary(statements),
		Statements:    statements,
		AccountingStd: acctStd,
		Consolidated:  hasConsolidatedStmt,
		Warnings:      warnings,
	}, nil
}

// selectConsolidation chooses which set of rows to use for a statement type.
func selectConsolidation(sr *consolidationGroup, st StatementType, opts ParseOpts, warnings *[]string) ([]parsedRow, bool) {
	// Split "other" rows into neutral (jpcrp_cor/jpdei_cor — applies to both modes)
	// and IFRS-consolidated (jpigp_cor, company-specific — consolidated only).
	var neutralOther, ifrsOther []parsedRow
	for _, r := range sr.other {
		prefix := elementPrefix(r.elementID)
		if prefix == "jpcrp_cor" || prefix == "jpdei_cor" {
			neutralOther = append(neutralOther, r)
		} else {
			ifrsOther = append(ifrsOther, r)
		}
	}

	consRows := append(append(sr.consolidated, ifrsOther...), neutralOther...)
	nonConsRows := append(sr.nonConsolidated, neutralOther...) // include neutral, exclude IFRS consolidated

	hasCons := len(sr.consolidated) > 0 || len(ifrsOther) > 0
	hasNonCons := len(sr.nonConsolidated) > 0

	if opts.Consolidated != nil {
		if *opts.Consolidated {
			if hasCons {
				return consRows, true
			}
			if hasNonCons {
				*warnings = append(*warnings, fmt.Sprintf("statement %s: consolidated data requested but not available, using non_consolidated as fallback", st))
				return nonConsRows, false
			}
			if len(sr.other) > 0 {
				return sr.other, true
			}
			return nil, false
		}
		// Explicit non-consolidated — do not include "other" (IFRS consolidated)
		if hasNonCons {
			return nonConsRows, false
		}
		if hasCons {
			*warnings = append(*warnings, fmt.Sprintf("statement %s: non_consolidated data requested but not available, using consolidated as fallback", st))
			return consRows, true
		}
		// Neither consolidated nor non-consolidated, but neutral rows exist
		if len(neutralOther) > 0 {
			*warnings = append(*warnings, fmt.Sprintf("statement %s: non-consolidated data not available; summary values populated from SummaryOfBusinessResults (neutral) rows", st))
			return nonConsRows, false
		}
		return nil, false
	}

	// Auto mode: prefer consolidated, fallback to non-consolidated
	if hasCons {
		return consRows, true
	}
	if hasNonCons {
		*warnings = append(*warnings, fmt.Sprintf("statement %s: no consolidated data available, using non_consolidated as fallback", st))
		return nonConsRows, false
	}
	if len(sr.other) > 0 {
		// "other" (その他) is typically IFRS consolidated data
		return sr.other, true
	}
	return nil, false
}

// detectAccountingStandard determines the accounting standard from element ID prefixes.
func detectAccountingStandard(rows []parsedRow) string {
	ifrsCount := 0
	jpgaapCount := 0

	for _, r := range rows {
		if r.classification.Statement == StmtUnknown {
			continue
		}
		prefix := elementPrefix(r.elementID)
		switch prefix {
		case "jpigp_cor":
			ifrsCount++
		case "jppfs_cor":
			jpgaapCount++
		case "jpcrp_cor":
			// Only count SummaryOfBusinessResults/KeyFinancialData elements
			// with core financial summaryKeys as accounting standard indicators.
			// Cross-standard items (dividend, shares, eps) do not imply a standard.
			localName := elementLocalName(r.elementID)
			if !strings.HasSuffix(localName, "SummaryOfBusinessResults") && !strings.HasSuffix(localName, "KeyFinancialData") {
				continue
			}
			// Only count if the element maps to a core financial metric
			key := r.classification.SummaryKey
			if key != "revenue" && key != "operating_income" && key != "ordinary_income" &&
				key != "net_income" && key != "total_assets" && key != "net_assets" {
				continue
			}
			if strings.Contains(localName, "IFRS") {
				ifrsCount++
			} else if strings.Contains(localName, "USGAAP") {
				// US GAAP detection: currently not needed as a separate standard
			} else {
				// Non-IFRS/USGAAP jpcrp_cor rows with core financial summaryKeys
				// imply JP-GAAP (e.g. NetSalesSummaryOfBusinessResults).
				jpgaapCount++
			}
		default:
			// Company-specific elements (jpcrp030000-asr_*)
			if strings.HasPrefix(prefix, "jpcrp030000-asr_") && r.classification.SummaryKey != "" {
				localName := elementLocalName(r.elementID)
				if strings.Contains(localName, "IFRS") {
					ifrsCount++
				}
				// Non-IFRS company-specific rows (e.g. NetSalesKeyFinancialData)
				// are standard-neutral and should not influence detection.
			}
		}
	}

	if ifrsCount > jpgaapCount {
		return "ifrs"
	}
	if jpgaapCount > 0 {
		return "jpgaap"
	}
	if ifrsCount > 0 {
		return "ifrs"
	}
	return "unknown"
}

// elementPrefix extracts the namespace prefix from an element ID (before the colon).
func elementPrefix(elementID string) string {
	if idx := strings.Index(elementID, ":"); idx >= 0 {
		return elementID[:idx]
	}
	return ""
}

// elementLocalName extracts the local name from an element ID (after the colon).
func elementLocalName(elementID string) string {
	if idx := strings.Index(elementID, ":"); idx >= 0 {
		return elementID[idx+1:]
	}
	return elementID
}

// buildStatement constructs a FinancialStatement from rows.
func buildStatement(stmtType StatementType, rows []parsedRow, consolidated bool, acctStd string, seen map[dedupeKey]bool) *FinancialStatement {
	if len(rows) == 0 {
		return nil
	}

	type itemWithOrder struct {
		item      LineItem
		sortOrder int
	}
	periodItems := make(map[string][]itemWithOrder)
	for _, r := range rows {
		if r.contextInfo.Period == "" {
			continue
		}

		key := dedupeKey{stmtType, r.contextInfo.Period, r.elementID}

		if seen[key] {
			continue
		}
		seen[key] = true

		iwo := itemWithOrder{
			item: LineItem{
				ElementID:  r.elementID,
				Label:      r.label,
				Category:   r.classification.Category,
				Value:      r.value,
				RawValue:   r.rawValue,
				Unit:       r.unit,
				UnitID:     r.unitID,
				SummaryKey: r.classification.SummaryKey,
				IsTotal:    r.classification.IsTotal,
			},
			sortOrder: r.classification.SortOrder,
		}
		periodItems[r.contextInfo.Period] = append(periodItems[r.contextInfo.Period], iwo)
	}

	if len(periodItems) == 0 {
		return nil
	}

	var periods []PeriodData
	for period, items := range periodItems {
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].sortOrder < items[j].sortOrder
		})
		sortedItems := make([]LineItem, len(items))
		for i, iwo := range items {
			sortedItems[i] = iwo.item
		}

		periods = append(periods, PeriodData{
			Period: period,
			Items:  sortedItems,
		})
	}

	sort.SliceStable(periods, func(i, j int) bool {
		return periodOrder(periods[i].Period) < periodOrder(periods[j].Period)
	})

	return &FinancialStatement{
		Type:          string(stmtType),
		Consolidated:  consolidated,
		AccountingStd: acctStd,
		Periods:       periods,
	}
}

// periodOrder returns a sort key for period names.
// current < current_quarter < current_ytd < current_interim < filing_date < prior1 < prior1_quarter < ...
func periodOrder(period string) int {
	// Suffix weights for sub-period types
	suffixWeight := func(s string) int {
		switch {
		case strings.HasSuffix(s, "_quarter"):
			return 1
		case strings.HasSuffix(s, "_ytd"):
			return 2
		case strings.HasSuffix(s, "_interim"):
			return 3
		default:
			return 0 // annual (no suffix)
		}
	}

	switch {
	case period == "current":
		return 0
	case strings.HasPrefix(period, "current_"):
		return 2 + suffixWeight(period)
	case period == "filing_date":
		return 6
	case strings.HasPrefix(period, "prior"):
		rest := period[5:]
		// Extract the numeric part
		numEnd := 0
		for numEnd < len(rest) && rest[numEnd] >= '0' && rest[numEnd] <= '9' {
			numEnd++
		}
		n := 0
		if numEnd > 0 {
			n, _ = strconv.Atoi(rest[:numEnd])
		}
		return 10 + n*10 + suffixWeight(period)
	default:
		return 200
	}
}

