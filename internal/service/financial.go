package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/beatinaniwa/edinet-cli/internal/api"
	"github.com/beatinaniwa/edinet-cli/internal/cache"
	"github.com/beatinaniwa/edinet-cli/internal/extract"
	"github.com/beatinaniwa/edinet-cli/internal/financial"
)

// CompanyFinancialsOpts configures a company financials query.
type CompanyFinancialsOpts struct {
	StatementOpts
	Periods   int           // number of fiscal years (default 3)
	RateLimit time.Duration // delay between API calls (default 100ms)
}

// permanentCacheTTL is used for downloaded documents that never change.
// Securities reports are immutable once filed, so we use a very large TTL.
const permanentCacheTTL = 100 * 365 * 24 * time.Hour // ~100 years

// FinancialService provides structured financial statement extraction from EDINET CSV data.
type FinancialService struct {
	client *api.Client
	cache  cache.Cache
}

// StatementOpts configures the financial statement extraction.
type StatementOpts struct {
	Statement    string // "bs", "pl", "cf", "all" (default "all")
	Consolidated *bool  // nil=auto, true=consolidated, false=non-consolidated
}

// NewFinancialService creates a new FinancialService.
func NewFinancialService(client *api.Client, c cache.Cache) *FinancialService {
	return &FinancialService{client: client, cache: c}
}

// GetStatements retrieves and parses financial statements for a document.
func (s *FinancialService) GetStatements(ctx context.Context, docID string, opts StatementOpts) (*financial.FinancialData, error) {
	cacheKey := fmt.Sprintf("files/%s/5", docID)

	body, fromCache, err := s.fetchCSV(ctx, docID, cacheKey)
	if err != nil {
		return nil, err
	}

	// Extract CSV from ZIP — retry on cache corruption (extraction failure only)
	csvResult, err := extract.ExtractCSVData(body)
	if err != nil && fromCache {
		freshBody, dlErr := s.downloadAndCache(ctx, docID, cacheKey)
		if dlErr != nil {
			return nil, &api.EDINETError{
				Code:    api.ErrInternal,
				Message: fmt.Sprintf("cache corruption recovery failed: %v (original: %v)", dlErr, err),
			}
		}
		csvResult, err = extract.ExtractCSVData(freshBody)
	}
	if err != nil {
		return nil, fmt.Errorf("csv extraction failed: %w", err)
	}

	// Parse (semantic errors are NOT retried — they are not cache corruption)
	data, err := s.parseAndBuild(csvResult, docID, opts)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// fetchCSV retrieves the CSV ZIP data, using cache.
// Returns (body, fromCache, error).
func (s *FinancialService) fetchCSV(ctx context.Context, docID, cacheKey string) ([]byte, bool, error) {
	// Try cache first
	if data, err := s.cache.Get(cacheKey, permanentCacheTTL); err == nil {
		return data, true, nil
	}

	// Cache miss — download from API
	body, err := s.downloadAndCache(ctx, docID, cacheKey)
	if err != nil {
		return nil, false, err
	}
	return body, false, nil
}

// downloadAndCache downloads CSV data from the API and stores it in cache.
func (s *FinancialService) downloadAndCache(ctx context.Context, docID, cacheKey string) ([]byte, error) {
	body, _, err := s.client.DownloadDocument(ctx, docID, 5)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(cacheKey, body)
	return body, nil
}

// GetCompanyFinancials retrieves financial statements for multiple fiscal periods of a company.
func (s *FinancialService) GetCompanyFinancials(ctx context.Context, companySvc *CompanyService, code string, opts CompanyFinancialsOpts) (*financial.CompanyFinancialsResult, error) {
	periods := opts.Periods
	if periods <= 0 {
		periods = 3
	}

	rateLimit := opts.RateLimit
	if rateLimit == 0 {
		rateLimit = 100 * time.Millisecond
	}

	// Find annual reports (doc type 120) going back enough days to cover the requested periods
	lookbackDays := periods * 400
	jst := time.FixedZone("JST", 9*60*60)
	nowJST := time.Now().In(jst)
	from := nowJST.AddDate(0, 0, -lookbackDays).Format("2006-01-02")
	to := nowJST.Format("2006-01-02")

	filingsResult, err := companySvc.Filings(ctx, code, FilingsOptions{
		DocType:   "120",
		From:      from,
		To:        to,
		RateLimit: rateLimit,
		Limit:     periods,
		Reverse:   true, // scan from newest to oldest, stop once enough found
	})
	if err != nil {
		return nil, err
	}

	// Reverse results back to chronological order (oldest first)
	for i, j := 0, len(filingsResult.Results)-1; i < j; i, j = i+1, j-1 {
		filingsResult.Results[i], filingsResult.Results[j] = filingsResult.Results[j], filingsResult.Results[i]
	}

	if len(filingsResult.Results) == 0 {
		return nil, &api.EDINETError{
			Code:    api.ErrNotFound,
			Message: fmt.Sprintf("no annual reports found for %s in the last %d days", code, lookbackDays),
		}
	}

	// Build company info from first filing
	first := filingsResult.Results[0]
	companyInfo := financial.CompanyInfo{
		EdinetCode: first.EdinetCode,
		SecCode:    first.SecCode,
		Name:       first.FilerName,
	}

	stmtOpts := opts.StatementOpts

	var financialPeriods []financial.FinancialData
	var warnings []string

	for i, filing := range filingsResult.Results {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Check if data is in cache before applying rate limit
		cacheKey := fmt.Sprintf("files/%s/5", filing.DocID)
		_, cacheErr := s.cache.Get(cacheKey, permanentCacheTTL)
		isCacheHit := cacheErr == nil

		// Rate limit only for API calls (not cache hits), and skip for first request
		if i > 0 && !isCacheHit {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(rateLimit):
			}
		}

		data, err := s.GetStatements(ctx, filing.DocID, stmtOpts)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", filing.DocID, err))
			continue
		}

		// Enrich with filing metadata
		data.EdinetCode = filing.EdinetCode
		data.SecCode = filing.SecCode
		data.CompanyName = filing.FilerName
		data.FiscalYear = filing.PeriodEnd

		financialPeriods = append(financialPeriods, *data)
	}

	if len(financialPeriods) == 0 {
		return nil, &api.EDINETError{
			Code:    api.ErrServer,
			Message: "all " + strconv.Itoa(len(filingsResult.Results)) + " filings failed: " + fmt.Sprintf("%v", warnings),
		}
	}

	// Add filing-level warnings
	if len(filingsResult.Metadata.Warnings) > 0 {
		warnings = append(warnings, filingsResult.Metadata.Warnings...)
	}

	result := &financial.CompanyFinancialsResult{
		Company: companyInfo,
		Periods: financialPeriods,
	}
	if len(warnings) > 0 {
		result.Warnings = warnings
	}

	return result, nil
}

// rebuildSummary builds a summary from the given statements' current period items.
func rebuildSummary(stmts []financial.FinancialStatement) financial.Summary {
	summary := make(financial.Summary)
	additiveKeys := map[string]bool{"interest_bearing_debt": true}

	for _, stmt := range stmts {
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
				} else if _, exists := summary[item.SummaryKey]; !exists {
					v := *item.Value
					summary[item.SummaryKey] = &v
				}
			}
		}
	}
	return summary
}

// parseAndBuild parses CSV data and builds the FinancialData output.
func (s *FinancialService) parseAndBuild(csvResult *extract.CSVDataResult, docID string, opts StatementOpts) (*financial.FinancialData, error) {
	parseOpts := financial.ParseOpts{
		Consolidated: opts.Consolidated,
	}

	parseResult, err := financial.Parse(csvResult, parseOpts)
	if err != nil {
		return nil, &api.EDINETError{
			Code:    api.ErrBadRequest,
			Message: fmt.Sprintf("failed to parse financial data: %v", err),
		}
	}

	// Build FinancialData from ParseResult
	data := &financial.FinancialData{
		DocID:         docID,
		AccountingStd: parseResult.AccountingStd,
		Consolidated:  parseResult.Consolidated,
		Summary:       parseResult.Summary,
		Statements:    parseResult.Statements,
		Warnings:      parseResult.Warnings,
	}

	// Filter by statement type if requested
	stmtFilter := opts.Statement
	if stmtFilter == "" {
		stmtFilter = "all"
	}

	if stmtFilter != "all" {
		var filtered []financial.FinancialStatement
		for _, stmt := range data.Statements {
			if stmt.Type == stmtFilter {
				filtered = append(filtered, stmt)
			}
		}
		if len(filtered) == 0 {
			return nil, &api.EDINETError{
				Code:    api.ErrNotFound,
				Message: fmt.Sprintf("statement type %q not found in document %s", stmtFilter, docID),
			}
		}
		data.Statements = filtered

		// Recompute top-level fields from filtered statements
		hasCons := false
		for _, stmt := range data.Statements {
			if stmt.Consolidated {
				hasCons = true
				break
			}
		}
		data.Consolidated = hasCons

		// Rebuild summary from only the filtered statements
		data.Summary = rebuildSummary(data.Statements)
	}

	// Empty result check
	if len(data.Statements) == 0 {
		return nil, &api.EDINETError{
			Code:    api.ErrBadRequest,
			Message: fmt.Sprintf("no financial statements found in document %s", docID),
		}
	}

	return data, nil
}
