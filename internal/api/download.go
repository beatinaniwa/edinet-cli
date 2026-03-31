package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// DownloadDocument downloads a document from the EDINET document download API.
// Returns (body, mediaType, error). The mediaType is the normalized Content-Type.
// If the response is JSON (indicating an error), it parses and returns an EDINETError.
func (c *Client) DownloadDocument(ctx context.Context, docID string, docType int) ([]byte, string, error) {
	params := url.Values{}
	params.Set("type", strconv.Itoa(docType))

	path := fmt.Sprintf("/api/v2/documents/%s", docID)
	return c.doRequest(ctx, path, params)
}
