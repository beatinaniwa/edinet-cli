package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient_DefaultTimeout(t *testing.T) {
	c := NewClient("test-key", "http://localhost", false)
	if c.httpClient.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", c.httpClient.Timeout)
	}
}

func TestClient_Get_SubscriptionKeyInQuery(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"status":"200","message":"OK"}}`))
	}))
	defer server.Close()

	c := NewClient("my-api-key", server.URL, false)
	_, err := c.Get(context.Background(), "/api/v2/documents.json", nil)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !strings.Contains(gotQuery, "Subscription-Key=my-api-key") {
		t.Errorf("query = %q, missing Subscription-Key", gotQuery)
	}
}

func TestClient_Get_MergesExistingParams(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"metadata":{"status":"200","message":"OK"}}`))
	}))
	defer server.Close()

	c := NewClient("key123", server.URL, false)
	params := map[string]string{"date": "2025-06-20", "type": "2"}
	_, err := c.GetWithParams(context.Background(), "/api/v2/documents.json", params)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	for _, want := range []string{"Subscription-Key=key123", "date=2025-06-20", "type=2"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query = %q, missing %q", gotQuery, want)
		}
	}
}

func TestClient_Get_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer server.Close()

	c := NewClient("key", server.URL, false)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Get(ctx, "/api/v2/documents.json", nil)
	if err == nil {
		t.Fatal("Get() should fail with cancelled context")
	}
}

func TestClient_Get_NetworkError(t *testing.T) {
	// Connect to a server that doesn't exist
	c := NewClient("key", "http://127.0.0.1:1", false)
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Get() should fail with network error")
	}
	edinetErr, ok := err.(*EDINETError)
	if !ok {
		t.Fatalf("error type = %T, want *EDINETError", err)
	}
	if edinetErr.Code != ErrNetwork {
		t.Errorf("Code = %q, want %q", edinetErr.Code, ErrNetwork)
	}
}

func TestClient_Get_APIError400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200) // EDINET always returns 200
		_, _ = w.Write([]byte(`{"metadata":{"title":"API","status":"400","message":"Bad Request"}}`))
	}))
	defer server.Close()

	c := NewClient("key", server.URL, false)
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Get() should return error for status 400")
	}
	edinetErr, ok := err.(*EDINETError)
	if !ok {
		t.Fatalf("error type = %T, want *EDINETError", err)
	}
	if edinetErr.Code != ErrBadRequest {
		t.Errorf("Code = %q, want %q", edinetErr.Code, ErrBadRequest)
	}
}

func TestClient_Get_APIError401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"StatusCode":401,"message":"Access denied"}`))
	}))
	defer server.Close()

	c := NewClient("bad-key", server.URL, false)
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Get() should return error for 401")
	}
	edinetErr, ok := err.(*EDINETError)
	if !ok {
		t.Fatalf("error type = %T, want *EDINETError", err)
	}
	if edinetErr.Code != ErrAuthFailed {
		t.Errorf("Code = %q, want %q", edinetErr.Code, ErrAuthFailed)
	}
}

func TestClient_Get_NonHTTP200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		_, _ = w.Write([]byte("Service Unavailable"))
	}))
	defer server.Close()

	c := NewClient("key", server.URL, false)
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("Get() should return error for HTTP 503")
	}
}
