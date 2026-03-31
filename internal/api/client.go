package api

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"time"
)

const defaultTimeout = 30 * time.Second

// Client is the HTTP client for the EDINET API v2.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	debug      bool
}

// NewClient creates a new EDINET API client.
func NewClient(apiKey, baseURL string, debug bool) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    baseURL,
		apiKey:     apiKey,
		debug:      debug,
	}
}

// Get sends a GET request and returns the response body.
// It automatically adds the Subscription-Key query parameter.
// For JSON responses, it checks the EDINET status field and returns an EDINETError if not 200.
func (c *Client) Get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	body, _, err := c.doRequest(ctx, path, params)
	return body, err
}

// doRequest is the common HTTP request handler for all EDINET API calls.
// Returns (body, mediaType, error).
func (c *Client) doRequest(ctx context.Context, path string, params url.Values) ([]byte, string, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("Subscription-Key", c.apiKey)

	reqURL := c.baseURL + path + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, "", &EDINETError{Code: ErrInternal, Message: fmt.Sprintf("failed to create request: %v", err)}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, "", &EDINETError{Code: ErrTimeout, Message: fmt.Sprintf("request cancelled: %v", ctx.Err())}
		}
		return nil, "", &EDINETError{Code: ErrNetwork, Message: fmt.Sprintf("request failed: %v", err)}
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", &EDINETError{Code: ErrNetwork, Message: fmt.Sprintf("failed to read response: %v", err)}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", &EDINETError{
			Code:    ErrInternal,
			Status:  resp.StatusCode,
			Message: fmt.Sprintf("unexpected HTTP status: %d", resp.StatusCode),
			Raw:     string(body),
		}
	}

	mediaType := parseMediaType(resp.Header.Get("Content-Type"))

	// JSON responses may be EDINET errors — check
	if mediaType == "application/json" || mediaType == "" {
		if edinetErr := ParseErrorResponse(body); edinetErr != nil {
			return nil, "", edinetErr
		}
	}

	return body, mediaType, nil
}

// GetWithParams is a convenience wrapper that accepts params as map[string]string.
func (c *Client) GetWithParams(ctx context.Context, path string, params map[string]string) ([]byte, error) {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return c.Get(ctx, path, values)
}

func parseMediaType(contentType string) string {
	if contentType == "" {
		return ""
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return contentType
	}
	return mediaType
}
