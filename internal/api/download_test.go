package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_DownloadDocument_ZipSuccess(t *testing.T) {
	zipData := []byte("PK\x03\x04fake-zip-data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "5" {
			t.Errorf("type param = %q, want 5", r.URL.Query().Get("type"))
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipData)
	}))
	defer server.Close()

	c := NewClient("key", server.URL, false)
	body, mediaType, err := c.DownloadDocument(context.Background(), "S100ABCD", 5)
	if err != nil {
		t.Fatalf("DownloadDocument() error = %v", err)
	}
	if mediaType != "application/octet-stream" {
		t.Errorf("mediaType = %q, want application/octet-stream", mediaType)
	}
	if len(body) != len(zipData) {
		t.Errorf("body len = %d, want %d", len(body), len(zipData))
	}
}

func TestClient_DownloadDocument_PdfSuccess(t *testing.T) {
	pdfData := []byte("%PDF-1.4 fake pdf")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(pdfData)
	}))
	defer server.Close()

	c := NewClient("key", server.URL, false)
	body, mediaType, err := c.DownloadDocument(context.Background(), "S100ABCD", 2)
	if err != nil {
		t.Fatalf("DownloadDocument() error = %v", err)
	}
	if mediaType != "application/pdf" {
		t.Errorf("mediaType = %q, want application/pdf", mediaType)
	}
	if len(body) != len(pdfData) {
		t.Errorf("body len = %d, want %d", len(body), len(pdfData))
	}
}

func TestClient_DownloadDocument_ErrorJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"title":"API","status":"404","message":"Not Found"}}`))
	}))
	defer server.Close()

	c := NewClient("key", server.URL, false)
	_, _, err := c.DownloadDocument(context.Background(), "S100XXXX", 5)
	if err == nil {
		t.Fatal("DownloadDocument() should return error for JSON error response")
	}
	edinetErr, ok := err.(*EDINETError)
	if !ok {
		t.Fatalf("error type = %T, want *EDINETError", err)
	}
	if edinetErr.Code != ErrNotFound {
		t.Errorf("Code = %q, want %q", edinetErr.Code, ErrNotFound)
	}
}

func TestClient_DownloadDocument_ContentTypeWithCharset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Error response with charset parameter
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"title":"API","status":"400","message":"Bad Request"}}`))
	}))
	defer server.Close()

	c := NewClient("key", server.URL, false)
	_, _, err := c.DownloadDocument(context.Background(), "S100XXXX", 1)
	if err == nil {
		t.Fatal("DownloadDocument() should detect error from Content-Type with charset")
	}
	edinetErr, ok := err.(*EDINETError)
	if !ok {
		t.Fatalf("error type = %T, want *EDINETError", err)
	}
	if edinetErr.Code != ErrBadRequest {
		t.Errorf("Code = %q, want %q", edinetErr.Code, ErrBadRequest)
	}
}
