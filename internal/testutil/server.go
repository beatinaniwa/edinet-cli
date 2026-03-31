package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// NewMockServer creates an httptest server with the given handler map.
// Routes are matched by path prefix. The server is automatically closed when the test ends.
func NewMockServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.HandleFunc(pattern, handler)
	}
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// JSONHandler returns an http.HandlerFunc that responds with the given status code and JSON body.
func JSONHandler(statusCode int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}
}

// BinaryHandler returns an http.HandlerFunc that responds with binary data.
func BinaryHandler(contentType string, data []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
