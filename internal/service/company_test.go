package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/beatinaniwa/edinet-cli/internal/api"
	"github.com/beatinaniwa/edinet-cli/internal/cache"
	"github.com/beatinaniwa/edinet-cli/internal/company"
)

func loadRegistry(t *testing.T) *company.Registry {
	t.Helper()
	data, err := os.ReadFile("../../testdata/edinetcode_sample.csv")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	r := &company.Registry{}
	if err := r.LoadFromCSV(data); err != nil {
		t.Fatalf("LoadFromCSV() error = %v", err)
	}
	return r
}

func TestCompanyService_Search(t *testing.T) {
	reg := loadRegistry(t)
	svc := NewCompanyService(reg, nil)

	results := svc.Search("トヨタ")
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].EdinetCode != "E02144" {
		t.Errorf("EdinetCode = %q, want E02144", results[0].EdinetCode)
	}
}

func TestCompanyService_Filings(t *testing.T) {
	reg := loadRegistry(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// Return one document per day with EDINET code E02144
		_, _ = w.Write([]byte(`{"metadata":{"status":"200","message":"OK","parameter":{"date":"` + date + `","type":"2"},"resultset":{"count":1},"processDateTime":"2025-06-20 13:01"},"results":[{"seqNumber":1,"docID":"S` + date + `","edinetCode":"E02144","secCode":"72030","JCN":null,"filerName":"トヨタ自動車株式会社","fundCode":null,"ordinanceCode":null,"formCode":null,"docTypeCode":"120","periodStart":null,"periodEnd":null,"submitDateTime":null,"docDescription":null,"issuerEdinetCode":null,"subjectEdinetCode":null,"subsidiaryEdinetCode":null,"currentReportReason":null,"parentDocID":null,"opeDateTime":null,"withdrawalStatus":"0","docInfoEditStatus":"0","disclosureStatus":"0","xbrlFlag":"1","pdfFlag":"1","attachDocFlag":"0","englishDocFlag":"0","csvFlag":"1","legalStatus":"1"}]}`))
	}))
	defer server.Close()

	client := api.NewClient("key", server.URL, false)
	docSvc := NewDocumentService(client, cache.NoopCache{}, nil)
	svc := NewCompanyService(reg, docSvc)

	result, err := svc.Filings(context.Background(), "7203", FilingsOptions{
		From:      "2025-06-18",
		To:        "2025-06-19",
		RateLimit: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Filings() error = %v", err)
	}
	if len(result.Results) != 2 {
		t.Errorf("len(Results) = %d, want 2", len(result.Results))
	}
}

func TestCompanyService_Filings_WithLimit(t *testing.T) {
	reg := loadRegistry(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"status":"200","message":"OK","parameter":{"date":"` + date + `","type":"2"},"resultset":{"count":2},"processDateTime":"2025-06-20 13:01"},"results":[{"seqNumber":1,"docID":"SA` + date + `","edinetCode":"E02144","secCode":"72030","JCN":null,"filerName":"トヨタ自動車株式会社","fundCode":null,"ordinanceCode":null,"formCode":null,"docTypeCode":"120","periodStart":null,"periodEnd":null,"submitDateTime":null,"docDescription":null,"issuerEdinetCode":null,"subjectEdinetCode":null,"subsidiaryEdinetCode":null,"currentReportReason":null,"parentDocID":null,"opeDateTime":null,"withdrawalStatus":"0","docInfoEditStatus":"0","disclosureStatus":"0","xbrlFlag":"1","pdfFlag":"1","attachDocFlag":"0","englishDocFlag":"0","csvFlag":"1","legalStatus":"1"},{"seqNumber":2,"docID":"SB` + date + `","edinetCode":"E02144","secCode":"72030","JCN":null,"filerName":"トヨタ自動車株式会社","fundCode":null,"ordinanceCode":null,"formCode":null,"docTypeCode":"120","periodStart":null,"periodEnd":null,"submitDateTime":null,"docDescription":null,"issuerEdinetCode":null,"subjectEdinetCode":null,"subsidiaryEdinetCode":null,"currentReportReason":null,"parentDocID":null,"opeDateTime":null,"withdrawalStatus":"0","docInfoEditStatus":"0","disclosureStatus":"0","xbrlFlag":"1","pdfFlag":"1","attachDocFlag":"0","englishDocFlag":"0","csvFlag":"1","legalStatus":"1"}]}`))
	}))
	defer server.Close()

	client := api.NewClient("key", server.URL, false)
	docSvc := NewDocumentService(client, cache.NoopCache{}, nil)
	svc := NewCompanyService(reg, docSvc)

	result, err := svc.Filings(context.Background(), "E02144", FilingsOptions{
		From:      "2025-06-18",
		To:        "2025-06-20",
		RateLimit: time.Millisecond,
		Limit:     3,
	})
	if err != nil {
		t.Fatalf("Filings() error = %v", err)
	}
	if len(result.Results) != 3 {
		t.Errorf("len(Results) = %d, want 3 (limited)", len(result.Results))
	}
}

func TestCompanyService_Filings_UnknownName(t *testing.T) {
	reg := loadRegistry(t)
	svc := NewCompanyService(reg, nil)

	_, err := svc.Filings(context.Background(), "存在しない会社名", FilingsOptions{})
	if err == nil {
		t.Fatal("Filings() should fail for unknown company name")
	}
}
