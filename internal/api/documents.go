package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// GetDocumentList retrieves the document list for a single date.
// listType: 1=metadata only, 2=document list + metadata. 0 defaults to 2.
func (c *Client) GetDocumentList(ctx context.Context, date string, listType int) (*DocumentListResponse, error) {
	if listType == 0 {
		listType = 2
	}

	params := url.Values{}
	params.Set("date", date)
	params.Set("type", strconv.Itoa(listType))

	body, err := c.Get(ctx, "/api/v2/documents.json", params)
	if err != nil {
		return nil, err
	}

	var resp DocumentListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &EDINETError{
			Code:    ErrInternal,
			Message: fmt.Sprintf("failed to parse document list response: %v", err),
			Raw:     string(body),
		}
	}

	return &resp, nil
}

// GetDocumentListRaw retrieves the document list and also returns the raw JSON bytes.
// This allows callers to cache the raw response without re-marshaling.
func (c *Client) GetDocumentListRaw(ctx context.Context, date string, listType int) (*DocumentListResponse, []byte, error) {
	if listType == 0 {
		listType = 2
	}

	params := url.Values{}
	params.Set("date", date)
	params.Set("type", strconv.Itoa(listType))

	body, err := c.Get(ctx, "/api/v2/documents.json", params)
	if err != nil {
		return nil, nil, err
	}

	var resp DocumentListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil, &EDINETError{
			Code:    ErrInternal,
			Message: fmt.Sprintf("failed to parse document list response: %v", err),
			Raw:     string(body),
		}
	}

	return &resp, body, nil
}
