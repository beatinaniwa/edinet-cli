package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beatinaniwa/edinet-cli/internal/api"
	"github.com/beatinaniwa/edinet-cli/internal/cache"
	"github.com/beatinaniwa/edinet-cli/internal/company"
)

// createCSVZip creates a ZIP archive containing a single CSV file.
func createCSVZip(t *testing.T, filename string, headers []string, rows [][]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	fw, err := zw.Create(filename)
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}

	cw := csv.NewWriter(fw)
	if err := cw.Write(headers); err != nil {
		t.Fatalf("failed to write CSV headers: %v", err)
	}
	for _, row := range rows {
		if err := cw.Write(row); err != nil {
			t.Fatalf("failed to write CSV row: %v", err)
		}
	}
	cw.Flush()

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	return buf.Bytes()
}

// makeCSVZip creates a minimal ZIP containing a CSV file in the XBRL_TO_CSV directory.
func makeCSVZip(t *testing.T) []byte {
	t.Helper()
	return createCSVZip(t, "XBRL_TO_CSV/jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		[]string{"要素ID", "項目名", "コンテキストID", "連結・個別", "期間・時点", "ユニットID", "単位", "値"},
		[][]string{
			{"jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "連結", "期間", "JPY", "円", "1000000"},
			{"jppfs_cor:OperatingIncome", "営業利益", "CurrentYearDuration", "連結", "期間", "JPY", "円", "200000"},
			{"jppfs_cor:TotalAssets", "総資産", "CurrentYearInstant", "連結", "時点", "JPY", "円", "5000000"},
			{"jppfs_cor:NetAssets", "純資産", "CurrentYearInstant", "連結", "時点", "JPY", "円", "3000000"},
			{"jppfs_cor:NetCashProvidedByUsedInOperatingActivities", "営業活動によるCF", "CurrentYearDuration", "連結", "期間", "JPY", "円", "300000"},
		},
	)
}

// mapCache is a simple in-memory cache for testing.
type mapCache struct {
	data map[string][]byte
}

func (m *mapCache) Get(key string, _ time.Duration) ([]byte, error) {
	if data, ok := m.data[key]; ok {
		return data, nil
	}
	return nil, cache.ErrCacheMiss
}

func (m *mapCache) Set(key string, data []byte) error {
	m.data[key] = data
	return nil
}

func TestFinancialService_GetStatements_CacheMiss(t *testing.T) {
	zipData := makeCSVZip(t)
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a CSV download (type=5)
		if r.URL.Query().Get("type") != "5" {
			t.Errorf("expected type=5, got %q", r.URL.Query().Get("type"))
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	svc := NewFinancialService(client, cache.NoopCache{})
	result, err := svc.GetStatements(context.Background(), "S100ABCD", StatementOpts{})
	if err != nil {
		t.Fatalf("GetStatements() error = %v", err)
	}
	if result.DocID != "S100ABCD" {
		t.Errorf("DocID = %q, want %q", result.DocID, "S100ABCD")
	}
	if len(result.Statements) == 0 {
		t.Error("expected at least one statement")
	}
	if result.Summary == nil {
		t.Error("expected non-nil summary")
	}
	// Check that revenue was parsed
	if rev, ok := result.Summary["revenue"]; !ok || rev == nil || *rev != 1000000 {
		t.Errorf("expected revenue=1000000, got %v", result.Summary["revenue"])
	}
}

func TestFinancialService_GetStatements_CacheHit(t *testing.T) {
	zipData := makeCSVZip(t)

	// Server should NOT be called if cache hits
	callCount := 0
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	// Pre-populate cache
	mc := &mapCache{data: map[string][]byte{"files/S100ABCD/5": zipData}}

	svc := NewFinancialService(client, mc)
	result, err := svc.GetStatements(context.Background(), "S100ABCD", StatementOpts{})
	if err != nil {
		t.Fatalf("GetStatements() error = %v", err)
	}
	if callCount != 0 {
		t.Errorf("API called %d times, expected 0 (cache hit)", callCount)
	}
	if result.DocID != "S100ABCD" {
		t.Errorf("DocID = %q, want %q", result.DocID, "S100ABCD")
	}
}

func TestFinancialService_GetStatements_CacheCorruptionRecovery(t *testing.T) {
	zipData := makeCSVZip(t)

	callCount := 0
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	// Pre-populate cache with corrupt data
	mc := &mapCache{data: map[string][]byte{"files/S100ABCD/5": []byte("corrupt zip data")}}

	svc := NewFinancialService(client, mc)
	result, err := svc.GetStatements(context.Background(), "S100ABCD", StatementOpts{})
	if err != nil {
		t.Fatalf("GetStatements() error = %v", err)
	}
	if callCount != 1 {
		t.Errorf("API called %d times, expected 1 (re-download after corruption)", callCount)
	}
	if result.DocID != "S100ABCD" {
		t.Errorf("DocID = %q, want %q", result.DocID, "S100ABCD")
	}
}

func TestFinancialService_GetStatements_DownloadFailure(t *testing.T) {
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"title":"API","status":"404","message":"Not Found"}}`))
	})

	svc := NewFinancialService(client, cache.NoopCache{})
	_, err := svc.GetStatements(context.Background(), "S100ABCD", StatementOpts{})
	if err == nil {
		t.Fatal("expected error for download failure")
	}
	edinetErr, ok := err.(*api.EDINETError)
	if !ok {
		t.Fatalf("expected *api.EDINETError, got %T", err)
	}
	if edinetErr.Code != api.ErrNotFound {
		t.Errorf("error code = %q, want %q", edinetErr.Code, api.ErrNotFound)
	}
}

func TestFinancialService_GetStatements_FilterByStatement(t *testing.T) {
	zipData := makeCSVZip(t)
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	svc := NewFinancialService(client, cache.NoopCache{})
	result, err := svc.GetStatements(context.Background(), "S100ABCD", StatementOpts{Statement: "pl"})
	if err != nil {
		t.Fatalf("GetStatements() error = %v", err)
	}
	for _, stmt := range result.Statements {
		if stmt.Type != "pl" {
			t.Errorf("expected only pl statements, got %q", stmt.Type)
		}
	}
}

func TestFinancialService_GetStatements_StatementNotFound(t *testing.T) {
	// Create CSV with only PL data (no CF data at all)
	zipData := createCSVZip(t, "XBRL_TO_CSV/jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		[]string{"要素ID", "項目名", "コンテキストID", "連結・個別", "期間・時点", "ユニットID", "単位", "値"},
		[][]string{
			{"jppfs_cor:NetSales", "売上高", "CurrentYearDuration", "連結", "期間", "JPY", "円", "1000000"},
		},
	)
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	svc := NewFinancialService(client, cache.NoopCache{})
	_, err := svc.GetStatements(context.Background(), "S100ABCD", StatementOpts{Statement: "cf"})
	if err == nil {
		t.Fatal("expected error when requested statement type not found")
	}
	edinetErr, ok := err.(*api.EDINETError)
	if !ok {
		t.Fatalf("expected *api.EDINETError, got %T", err)
	}
	if edinetErr.Code != api.ErrNotFound {
		t.Errorf("error code = %q, want %q", edinetErr.Code, api.ErrNotFound)
	}
}

func TestFinancialService_GetStatements_EmptyCSV(t *testing.T) {
	// Create ZIP with an empty CSV (no data rows)
	zipData := createCSVZip(t, "XBRL_TO_CSV/jpcrp030000-asr-001_E00001-000_2025-03-31_01_2025-06-20.csv",
		[]string{"要素ID", "項目名", "コンテキストID", "連結・個別", "期間・時点", "ユニットID", "単位", "値"},
		[][]string{},
	)
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	svc := NewFinancialService(client, cache.NoopCache{})
	_, err := svc.GetStatements(context.Background(), "S100ABCD", StatementOpts{})
	if err == nil {
		t.Fatal("expected error for empty CSV result")
	}
	edinetErr, ok := err.(*api.EDINETError)
	if !ok {
		t.Fatalf("expected *api.EDINETError, got %T", err)
	}
	if edinetErr.Code != api.ErrBadRequest {
		t.Errorf("error code = %q, want %q", edinetErr.Code, api.ErrBadRequest)
	}
}

func TestFinancialService_GetStatements_NonConsolidated(t *testing.T) {
	zipData := makeCSVZip(t) // has consolidated data
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	nonCons := false
	svc := NewFinancialService(client, cache.NoopCache{})
	result, err := svc.GetStatements(context.Background(), "S100ABCD", StatementOpts{Consolidated: &nonCons})
	if err != nil {
		t.Fatalf("GetStatements() error = %v", err)
	}
	// The result should still work (fallback with warning), or have data
	if result.DocID != "S100ABCD" {
		t.Errorf("DocID = %q, want %q", result.DocID, "S100ABCD")
	}
}

// makeDocListResponse creates a JSON document list response with the given documents.
func makeDocListResponse(date string, docs []map[string]string) string {
	var results []string
	for i, doc := range docs {
		docID := doc["docID"]
		edinetCode := doc["edinetCode"]
		secCode := "null"
		if v, ok := doc["secCode"]; ok {
			secCode = fmt.Sprintf("%q", v)
		}
		filerName := "null"
		if v, ok := doc["filerName"]; ok {
			filerName = fmt.Sprintf("%q", v)
		}
		docTypeCode := "120"
		if v, ok := doc["docTypeCode"]; ok {
			docTypeCode = v
		}
		periodEnd := "null"
		if v, ok := doc["periodEnd"]; ok {
			periodEnd = fmt.Sprintf("%q", v)
		}
		results = append(results, fmt.Sprintf(`{"seqNumber":%d,"docID":%q,"edinetCode":%q,"secCode":%s,"JCN":null,"filerName":%s,"fundCode":null,"ordinanceCode":null,"formCode":null,"docTypeCode":%q,"periodStart":null,"periodEnd":%s,"submitDateTime":null,"docDescription":null,"issuerEdinetCode":null,"subjectEdinetCode":null,"subsidiaryEdinetCode":null,"currentReportReason":null,"parentDocID":null,"opeDateTime":null,"withdrawalStatus":"0","docInfoEditStatus":"0","disclosureStatus":"0","xbrlFlag":"1","pdfFlag":"1","attachDocFlag":"0","englishDocFlag":"0","csvFlag":"1","legalStatus":"1"}`, i+1, docID, edinetCode, secCode, filerName, docTypeCode, periodEnd))
	}
	return fmt.Sprintf(`{"metadata":{"status":"200","message":"OK","parameter":{"date":"%s","type":"2"},"resultset":{"count":%d},"processDateTime":"2025-06-20 13:01"},"results":[%s]}`, date, len(results), strings.Join(results, ","))
}

func TestGetCompanyFinancials_MultiplePeriods(t *testing.T) {
	zipData := makeCSVZip(t)

	// Track requests by path to route doc list vs download
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/documents.json", func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// Return a filing for E02144 on specific dates
		resp := makeDocListResponse(date, []map[string]string{
			{
				"docID":       "S" + strings.ReplaceAll(date, "-", ""),
				"edinetCode":  "E02144",
				"secCode":     "72030",
				"filerName":   "トヨタ自動車株式会社",
				"docTypeCode": "120",
				"periodEnd":   date,
			},
		})
		_, _ = w.Write([]byte(resp))
	})
	mux.HandleFunc("/api/v2/documents/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	server := newTestServer(t, mux)
	client := api.NewClient("test-key", server.URL, false)

	mc := &mapCache{data: map[string][]byte{}}
	finSvc := NewFinancialService(client, mc)
	docSvc := NewDocumentService(client, mc, nil)

	reg := loadTestRegistry(t)
	companySvc := NewCompanyService(reg, docSvc)

	result, err := finSvc.GetCompanyFinancials(context.Background(), companySvc, "E02144", CompanyFinancialsOpts{
		Periods:   2,
		RateLimit: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("GetCompanyFinancials() error = %v", err)
	}

	if result.Company.EdinetCode != "E02144" {
		t.Errorf("Company.EdinetCode = %q, want %q", result.Company.EdinetCode, "E02144")
	}
	if result.Company.Name != "トヨタ自動車株式会社" {
		t.Errorf("Company.Name = %q, want %q", result.Company.Name, "トヨタ自動車株式会社")
	}
	if len(result.Periods) == 0 {
		t.Fatal("expected at least one period")
	}
	// Each period should have financial data
	for i, p := range result.Periods {
		if p.DocID == "" {
			t.Errorf("Periods[%d].DocID is empty", i)
		}
		if len(p.Statements) == 0 {
			t.Errorf("Periods[%d] has no statements", i)
		}
	}
}

func TestGetCompanyFinancials_PartialFailure(t *testing.T) {
	zipData := makeCSVZip(t)

	downloadCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/documents.json", func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		resp := makeDocListResponse(date, []map[string]string{
			{
				"docID":       "S" + strings.ReplaceAll(date, "-", ""),
				"edinetCode":  "E02144",
				"secCode":     "72030",
				"filerName":   "トヨタ自動車株式会社",
				"docTypeCode": "120",
				"periodEnd":   date,
			},
		})
		_, _ = w.Write([]byte(resp))
	})
	mux.HandleFunc("/api/v2/documents/", func(w http.ResponseWriter, r *http.Request) {
		downloadCount++
		if downloadCount == 1 {
			// First download fails
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, _ = w.Write([]byte(`{"metadata":{"title":"API","status":"404","message":"Not Found"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	server := newTestServer(t, mux)
	client := api.NewClient("test-key", server.URL, false)

	mc := &mapCache{data: map[string][]byte{}}
	finSvc := NewFinancialService(client, mc)
	docSvc := NewDocumentService(client, mc, nil)

	reg := loadTestRegistry(t)
	companySvc := NewCompanyService(reg, docSvc)

	result, err := finSvc.GetCompanyFinancials(context.Background(), companySvc, "E02144", CompanyFinancialsOpts{
		Periods:   2,
		RateLimit: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("GetCompanyFinancials() error = %v (should succeed with partial results)", err)
	}

	// Should have at least one successful period
	if len(result.Periods) == 0 {
		t.Fatal("expected at least one successful period")
	}
	// Should have warnings for the failed period
	if len(result.Warnings) == 0 {
		t.Error("expected warnings for failed period")
	}
}

// loadTestRegistry loads a test registry for company resolution.
func loadTestRegistry(t *testing.T) *company.Registry {
	t.Helper()
	// Minimal CSV data for E02144 (Toyota)
	csvData := []byte("EDINETコードリスト\nEDINETコード,提出者種別,上場区分,連結の有無,資本金,決算日,提出者名,提出者名（英字）,提出者名（ヨミ）,所在地,提出者業種,証券コード,提出者法人番号\nE02144,内国法人・組合,上場,あり,635402000000,3月,トヨタ自動車株式会社,TOYOTA MOTOR CORPORATION,トヨタジドウシャ,愛知県豊田市トヨタ町１番地,輸送用機器,72030,1180301018771\n")
	r := &company.Registry{}
	if err := r.LoadFromCSV(csvData); err != nil {
		t.Fatalf("LoadFromCSV() error = %v", err)
	}
	return r
}

// newTestServer creates an httptest.Server with automatic cleanup.
func newTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// makeFilingStyleCSVZip creates a ZIP with filing-style CSV (Prior1Year contexts only).
func makeFilingStyleCSVZip(t *testing.T) []byte {
	t.Helper()
	return createCSVZip(t, "XBRL_TO_CSV/jpcrp020400-srs-001_E41257-000_2025-04-30_01_2026-01-09.csv",
		[]string{"要素ID", "項目名", "コンテキストID", "連結・個別", "期間・時点", "ユニットID", "単位", "値"},
		[][]string{
			{"jpcrp_cor:NetSalesSummaryOfBusinessResults", "売上高、経営指標等", "Prior1YearDuration", "個別", "期間", "JPY", "円", "9426601000"},
			{"jpcrp_cor:OrdinaryIncomeSummaryOfBusinessResults", "経常利益、経営指標等", "Prior1YearDuration", "個別", "期間", "JPY", "円", "1145214000"},
			{"jpcrp_cor:TotalAssetsSummaryOfBusinessResults", "総資産額、経営指標等", "Prior1YearInstant", "個別", "時点", "JPY", "円", "6160640000"},
			{"jpcrp_cor:NetAssetsSummaryOfBusinessResults", "純資産額、経営指標等", "Prior1YearInstant", "個別", "時点", "JPY", "円", "4261992000"},
		},
	)
}

func TestFinancialService_GetStatements_FilingStyleCSV(t *testing.T) {
	zipData := makeFilingStyleCSVZip(t)
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	svc := NewFinancialService(client, cache.NoopCache{})
	result, err := svc.GetStatements(context.Background(), "S100XF4F", StatementOpts{})
	if err != nil {
		t.Fatalf("GetStatements() error = %v", err)
	}
	if result.DocID != "S100XF4F" {
		t.Errorf("DocID = %q, want %q", result.DocID, "S100XF4F")
	}
	if result.SummaryPeriod != "prior1" {
		t.Errorf("SummaryPeriod = %q, want %q", result.SummaryPeriod, "prior1")
	}
	if result.Summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if rev := result.Summary["revenue"]; rev == nil || *rev != 9426601000 {
		t.Errorf("expected revenue=9426601000, got %v", result.Summary["revenue"])
	}
	if ta := result.Summary["total_assets"]; ta == nil || *ta != 6160640000 {
		t.Errorf("expected total_assets=6160640000, got %v", result.Summary["total_assets"])
	}
}

func TestFinancialService_GetStatements_FilterRecomputesSummaryPeriod(t *testing.T) {
	zipData := makeCSVZip(t) // has current period data
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	})

	svc := NewFinancialService(client, cache.NoopCache{})
	result, err := svc.GetStatements(context.Background(), "S100ABCD", StatementOpts{Statement: "pl"})
	if err != nil {
		t.Fatalf("GetStatements() error = %v", err)
	}
	// After filtering to PL-only, summary should be recomputed with SummaryPeriod
	if result.SummaryPeriod != "current" {
		t.Errorf("SummaryPeriod = %q, want %q after statement filter recomputation", result.SummaryPeriod, "current")
	}
}
