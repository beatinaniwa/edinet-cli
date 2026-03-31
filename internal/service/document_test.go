package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/beatinaniwa/edinet-cli/internal/api"
	"github.com/beatinaniwa/edinet-cli/internal/cache"
)

func setupMockServer(t *testing.T, handler http.HandlerFunc) (*api.Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := api.NewClient("test-key", server.URL, false)
	return client, server
}

func TestDocumentService_List_SingleDate(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/documents_response.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(fixture)
	})

	svc := NewDocumentService(client, cache.NoopCache{}, nil)
	result, err := svc.List(context.Background(), ListOptions{Date: "2025-06-20"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Metadata.Date != "2025-06-20" {
		t.Errorf("Date = %q, want %q", result.Metadata.Date, "2025-06-20")
	}
	if result.Metadata.TotalResults != 2 {
		t.Errorf("TotalResults = %d, want %d", result.Metadata.TotalResults, 2)
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Results) = %d, want %d", len(result.Results), 2)
	}
	if result.Results[0].DocID != "S100ABCD" {
		t.Errorf("Results[0].DocID = %q, want %q", result.Results[0].DocID, "S100ABCD")
	}
}

func TestDocumentService_List_DateRange(t *testing.T) {
	callCount := 0
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		date := r.URL.Query().Get("date")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"status":"200","message":"OK","parameter":{"date":"` + date + `","type":"2"},"resultset":{"count":1},"processDateTime":"2025-06-20 13:01"},"results":[{"seqNumber":1,"docID":"S` + date + `","edinetCode":"E00001","secCode":null,"JCN":null,"filerName":"Test","fundCode":null,"ordinanceCode":null,"formCode":null,"docTypeCode":"120","periodStart":null,"periodEnd":null,"submitDateTime":null,"docDescription":null,"issuerEdinetCode":null,"subjectEdinetCode":null,"subsidiaryEdinetCode":null,"currentReportReason":null,"parentDocID":null,"opeDateTime":null,"withdrawalStatus":"0","docInfoEditStatus":"0","disclosureStatus":"0","xbrlFlag":"0","pdfFlag":"0","attachDocFlag":"0","englishDocFlag":"0","csvFlag":"0","legalStatus":"1"}]}`))
	})

	svc := NewDocumentService(client, cache.NoopCache{}, nil)
	result, err := svc.List(context.Background(), ListOptions{
		From:      "2025-06-18",
		To:        "2025-06-20",
		RateLimit: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if callCount != 3 {
		t.Errorf("API called %d times, want 3 (one per day)", callCount)
	}
	if result.Metadata.TotalResults != 3 {
		t.Errorf("TotalResults = %d, want 3", result.Metadata.TotalResults)
	}
	if result.Metadata.DateRange == nil {
		t.Fatal("DateRange is nil")
	}
	if result.Metadata.DateRange.From != "2025-06-18" {
		t.Errorf("From = %q, want %q", result.Metadata.DateRange.From, "2025-06-18")
	}
}

func TestDocumentService_List_FilterByDocType(t *testing.T) {
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"status":"200","message":"OK","parameter":{"date":"2025-06-20","type":"2"},"resultset":{"count":2},"processDateTime":"2025-06-20 13:01"},"results":[{"seqNumber":1,"docID":"S1","edinetCode":null,"secCode":null,"JCN":null,"filerName":null,"fundCode":null,"ordinanceCode":null,"formCode":null,"docTypeCode":"120","periodStart":null,"periodEnd":null,"submitDateTime":null,"docDescription":null,"issuerEdinetCode":null,"subjectEdinetCode":null,"subsidiaryEdinetCode":null,"currentReportReason":null,"parentDocID":null,"opeDateTime":null,"withdrawalStatus":"0","docInfoEditStatus":"0","disclosureStatus":"0","xbrlFlag":"0","pdfFlag":"0","attachDocFlag":"0","englishDocFlag":"0","csvFlag":"0","legalStatus":"1"},{"seqNumber":2,"docID":"S2","edinetCode":null,"secCode":null,"JCN":null,"filerName":null,"fundCode":null,"ordinanceCode":null,"formCode":null,"docTypeCode":"140","periodStart":null,"periodEnd":null,"submitDateTime":null,"docDescription":null,"issuerEdinetCode":null,"subjectEdinetCode":null,"subsidiaryEdinetCode":null,"currentReportReason":null,"parentDocID":null,"opeDateTime":null,"withdrawalStatus":"0","docInfoEditStatus":"0","disclosureStatus":"0","xbrlFlag":"0","pdfFlag":"0","attachDocFlag":"0","englishDocFlag":"0","csvFlag":"0","legalStatus":"1"}]}`))
	})

	svc := NewDocumentService(client, cache.NoopCache{}, nil)
	result, err := svc.List(context.Background(), ListOptions{Date: "2025-06-20", DocType: "120"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1 (filtered)", len(result.Results))
	}
	if result.Results[0].DocID != "S1" {
		t.Errorf("DocID = %q, want %q", result.Results[0].DocID, "S1")
	}
}

func TestDocumentService_List_PartialFailure(t *testing.T) {
	callCount := 0
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if callCount == 2 {
			// Second day fails
			_, _ = w.Write([]byte(`{"metadata":{"title":"API","status":"500","message":"Internal Server Error"}}`))
			return
		}
		date := r.URL.Query().Get("date")
		_, _ = w.Write([]byte(`{"metadata":{"status":"200","message":"OK","parameter":{"date":"` + date + `","type":"2"},"resultset":{"count":1},"processDateTime":"2025-06-20 13:01"},"results":[{"seqNumber":1,"docID":"S` + date + `","edinetCode":null,"secCode":null,"JCN":null,"filerName":null,"fundCode":null,"ordinanceCode":null,"formCode":null,"docTypeCode":"120","periodStart":null,"periodEnd":null,"submitDateTime":null,"docDescription":null,"issuerEdinetCode":null,"subjectEdinetCode":null,"subsidiaryEdinetCode":null,"currentReportReason":null,"parentDocID":null,"opeDateTime":null,"withdrawalStatus":"0","docInfoEditStatus":"0","disclosureStatus":"0","xbrlFlag":"0","pdfFlag":"0","attachDocFlag":"0","englishDocFlag":"0","csvFlag":"0","legalStatus":"1"}]}`))
	})

	svc := NewDocumentService(client, cache.NoopCache{}, nil)
	result, err := svc.List(context.Background(), ListOptions{
		From:      "2025-06-18",
		To:        "2025-06-20",
		RateLimit: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("List() should succeed on partial failure, got %v", err)
	}
	if len(result.Results) != 2 {
		t.Errorf("len(Results) = %d, want 2 (2 successful days)", len(result.Results))
	}
	if len(result.Metadata.Warnings) != 1 {
		t.Fatalf("len(Warnings) = %d, want 1", len(result.Metadata.Warnings))
	}
}

func TestDocumentService_List_AllDaysFail(t *testing.T) {
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"title":"API","status":"500","message":"Internal Server Error"}}`))
	})

	svc := NewDocumentService(client, cache.NoopCache{}, nil)
	_, err := svc.List(context.Background(), ListOptions{
		From:      "2025-06-18",
		To:        "2025-06-20",
		RateLimit: time.Millisecond,
	})
	if err == nil {
		t.Fatal("List() should return error when all days fail")
	}
}

func TestDocumentService_List_Limit(t *testing.T) {
	callCount := 0
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		date := r.URL.Query().Get("date")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// Return 2 results per day
		_, _ = w.Write([]byte(`{"metadata":{"status":"200","message":"OK","parameter":{"date":"` + date + `","type":"2"},"resultset":{"count":2},"processDateTime":"2025-06-20 13:01"},"results":[{"seqNumber":1,"docID":"SA` + date + `","edinetCode":null,"secCode":null,"JCN":null,"filerName":null,"fundCode":null,"ordinanceCode":null,"formCode":null,"docTypeCode":"120","periodStart":null,"periodEnd":null,"submitDateTime":null,"docDescription":null,"issuerEdinetCode":null,"subjectEdinetCode":null,"subsidiaryEdinetCode":null,"currentReportReason":null,"parentDocID":null,"opeDateTime":null,"withdrawalStatus":"0","docInfoEditStatus":"0","disclosureStatus":"0","xbrlFlag":"0","pdfFlag":"0","attachDocFlag":"0","englishDocFlag":"0","csvFlag":"0","legalStatus":"1"},{"seqNumber":2,"docID":"SB` + date + `","edinetCode":null,"secCode":null,"JCN":null,"filerName":null,"fundCode":null,"ordinanceCode":null,"formCode":null,"docTypeCode":"120","periodStart":null,"periodEnd":null,"submitDateTime":null,"docDescription":null,"issuerEdinetCode":null,"subjectEdinetCode":null,"subsidiaryEdinetCode":null,"currentReportReason":null,"parentDocID":null,"opeDateTime":null,"withdrawalStatus":"0","docInfoEditStatus":"0","disclosureStatus":"0","xbrlFlag":"0","pdfFlag":"0","attachDocFlag":"0","englishDocFlag":"0","csvFlag":"0","legalStatus":"1"}]}`))
	})

	svc := NewDocumentService(client, cache.NoopCache{}, nil)
	result, err := svc.List(context.Background(), ListOptions{
		From:      "2025-06-18",
		To:        "2025-06-20",
		RateLimit: time.Millisecond,
		Limit:     3,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result.Results) != 3 {
		t.Errorf("len(Results) = %d, want 3 (limited)", len(result.Results))
	}
	// Should stop early - not call all 3 days
	if callCount > 2 {
		t.Errorf("API called %d times, want <=2 (should stop early with limit)", callCount)
	}
}

func TestDocumentService_List_ProgressOutput(t *testing.T) {
	client, _ := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"status":"200","message":"OK","parameter":{"date":"` + date + `","type":"2"},"resultset":{"count":0},"processDateTime":"2025-06-20 13:01"},"results":[]}`))
	})

	var stderr bytes.Buffer
	svc := NewDocumentService(client, cache.NoopCache{}, &stderr)
	_, err := svc.List(context.Background(), ListOptions{
		From:      "2025-06-18",
		To:        "2025-06-19",
		RateLimit: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if stderr.Len() == 0 {
		t.Error("expected progress output on stderr, got none")
	}
}
