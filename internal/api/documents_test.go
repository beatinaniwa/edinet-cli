package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetDocumentList_Success(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/documents_response.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request params
		if r.URL.Query().Get("date") != "2025-06-20" {
			t.Errorf("date param = %q, want %q", r.URL.Query().Get("date"), "2025-06-20")
		}
		if r.URL.Query().Get("type") != "2" {
			t.Errorf("type param = %q, want %q", r.URL.Query().Get("type"), "2")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	c := NewClient("key", server.URL, false)
	resp, err := c.GetDocumentList(context.Background(), "2025-06-20", 2)
	if err != nil {
		t.Fatalf("GetDocumentList() error = %v", err)
	}
	if resp.Metadata.Status != "200" {
		t.Errorf("Status = %q, want %q", resp.Metadata.Status, "200")
	}
	if resp.Metadata.ResultSet.Count != 2 {
		t.Errorf("Count = %d, want %d", resp.Metadata.ResultSet.Count, 2)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("len(Results) = %d, want %d", len(resp.Results), 2)
	}

	doc := resp.Results[0]
	if doc.DocID != "S100ABCD" {
		t.Errorf("DocID = %q, want %q", doc.DocID, "S100ABCD")
	}
	if doc.EdinetCode == nil || *doc.EdinetCode != "E02144" {
		t.Errorf("EdinetCode = %v, want E02144", doc.EdinetCode)
	}
	if doc.FilerName == nil || *doc.FilerName != "トヨタ自動車株式会社" {
		t.Errorf("FilerName = %v, want トヨタ自動車株式会社", doc.FilerName)
	}
	if doc.FundCode != nil {
		t.Errorf("FundCode = %v, want nil", doc.FundCode)
	}
}

func TestGetDocumentList_Error400(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/error_400.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	c := NewClient("key", server.URL, false)
	_, err = c.GetDocumentList(context.Background(), "invalid-date", 2)
	if err == nil {
		t.Fatal("GetDocumentList() should return error for 400")
	}
	edinetErr, ok := err.(*EDINETError)
	if !ok {
		t.Fatalf("error type = %T, want *EDINETError", err)
	}
	if edinetErr.Code != ErrBadRequest {
		t.Errorf("Code = %q, want %q", edinetErr.Code, ErrBadRequest)
	}
}

func TestGetDocumentList_DefaultType(t *testing.T) {
	var gotType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotType = r.URL.Query().Get("type")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"status":"200","message":"OK","parameter":{"date":"2025-06-20","type":"2"},"resultset":{"count":0}},"results":[]}`))
	}))
	defer server.Close()

	c := NewClient("key", server.URL, false)
	_, err := c.GetDocumentList(context.Background(), "2025-06-20", 0)
	if err != nil {
		t.Fatalf("GetDocumentList() error = %v", err)
	}
	if gotType != "2" {
		t.Errorf("type param = %q, want %q (default)", gotType, "2")
	}
}
