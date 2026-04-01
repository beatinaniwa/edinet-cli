package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/beatinaniwa/edinet-cli/internal/api"
	"github.com/beatinaniwa/edinet-cli/internal/cache"
)

// DocumentService provides document listing with date iteration, filtering, and caching.
type DocumentService struct {
	client   *api.Client
	cache    cache.Cache
	progress io.Writer
}

// ListOptions configures a document list query.
type ListOptions struct {
	Date       string
	From       string
	To         string
	DocType    string
	SecCode    string
	EdinetCode string
	FilerName      string
	DocDescription string
	RateLimit      time.Duration
	Limit      int
	Reverse    bool
}

// NewDocumentService creates a new DocumentService.
func NewDocumentService(client *api.Client, c cache.Cache, progress io.Writer) *DocumentService {
	return &DocumentService{client: client, cache: c, progress: progress}
}

// List retrieves documents for a single date or date range, applying client-side filters.
func (s *DocumentService) List(ctx context.Context, opts ListOptions) (*ListResult, error) {
	if opts.Date != "" {
		return s.listSingleDate(ctx, opts)
	}
	return s.listDateRange(ctx, opts)
}

func (s *DocumentService) listSingleDate(ctx context.Context, opts ListOptions) (*ListResult, error) {
	docs, _, err := s.fetchDate(ctx, opts.Date)
	if err != nil {
		return nil, err
	}

	filtered := filterDocuments(docs, opts)
	return &ListResult{
		Metadata: ListMetadata{
			Date:         opts.Date,
			TotalResults: len(filtered),
		},
		Results: filtered,
	}, nil
}

func (s *DocumentService) listDateRange(ctx context.Context, opts ListOptions) (*ListResult, error) {
	dates, err := dateRange(opts.From, opts.To)
	if err != nil {
		return nil, err
	}

	if opts.Reverse {
		for i, j := 0, len(dates)-1; i < j; i, j = i+1, j-1 {
			dates[i], dates[j] = dates[j], dates[i]
		}
	}

	var allResults []DocumentInfo
	var warnings []string
	var lastErr error
	rateLimit := opts.RateLimit
	if rateLimit == 0 {
		rateLimit = 100 * time.Millisecond
	}

	for i, date := range dates {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		s.reportProgress(date, i+1, len(dates))

		docs, fromCache, fetchErr := s.fetchDate(ctx, date)
		if fetchErr != nil {
			lastErr = fetchErr
			warnings = append(warnings, fmt.Sprintf("%s: %s", date, classifyWarning(fetchErr)))
			if i < len(dates)-1 && !fromCache {
				time.Sleep(rateLimit)
			}
			continue
		}

		filtered := filterDocuments(docs, opts)
		allResults = append(allResults, filtered...)

		if opts.Limit > 0 && len(allResults) >= opts.Limit {
			allResults = allResults[:opts.Limit]
			break
		}

		// Skip rate-limit delay when result came from cache (no API call made)
		if i < len(dates)-1 && !fromCache {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(rateLimit):
			}
		}
	}

	if len(allResults) == 0 && len(warnings) == len(dates) {
		// Propagate the error code from the last failure
		code := api.ErrServer
		if edinetErr, ok := lastErr.(*api.EDINETError); ok {
			code = edinetErr.Code
		}
		return nil, &api.EDINETError{
			Code:    code,
			Message: fmt.Sprintf("all %d dates failed: %s", len(dates), strings.Join(warnings, "; ")),
		}
	}

	return &ListResult{
		Metadata: ListMetadata{
			DateRange:    &DateRange{From: opts.From, To: opts.To},
			TotalResults: len(allResults),
			Warnings:     warnings,
		},
		Results: allResults,
	}, nil
}

// fetchDate returns documents for a single date. fromCache indicates whether
// the result came from cache (so callers can skip rate-limit delays).
func (s *DocumentService) fetchDate(ctx context.Context, date string) (docs []api.Document, fromCache bool, err error) {
	cacheKey := "doclist/" + date + ".json"
	maxAge := 24 * time.Hour
	jst := time.FixedZone("JST", 9*60*60)
	if date == time.Now().In(jst).Format("2006-01-02") {
		maxAge = 5 * time.Minute
	}

	// Try cache first
	if data, cacheErr := s.cache.Get(cacheKey, maxAge); cacheErr == nil {
		var resp api.DocumentListResponse
		if jsonErr := json.Unmarshal(data, &resp); jsonErr == nil {
			return resp.Results, true, nil
		}
	}

	// Cache miss — fetch via API and cache the raw response
	resp, rawBody, fetchErr := s.client.GetDocumentListRaw(ctx, date, 2)
	if fetchErr != nil {
		return nil, false, fetchErr
	}

	_ = s.cache.Set(cacheKey, rawBody)
	return resp.Results, false, nil
}

func (s *DocumentService) reportProgress(date string, index, total int) {
	if s.progress == nil {
		return
	}
	msg := struct {
		Progress struct {
			Date  string `json:"date"`
			Index int    `json:"index"`
			Total int    `json:"total"`
		} `json:"progress"`
	}{}
	msg.Progress.Date = date
	msg.Progress.Index = index
	msg.Progress.Total = total
	data, _ := json.Marshal(msg)
	_, _ = fmt.Fprintln(s.progress, string(data))
}

func filterDocuments(docs []api.Document, opts ListOptions) []DocumentInfo {
	var result []DocumentInfo
	for _, doc := range docs {
		if opts.DocType != "" && (doc.DocTypeCode == nil || *doc.DocTypeCode != opts.DocType) {
			continue
		}
		if opts.SecCode != "" && (doc.SecCode == nil || *doc.SecCode != opts.SecCode) {
			continue
		}
		if opts.EdinetCode != "" && (doc.EdinetCode == nil || *doc.EdinetCode != opts.EdinetCode) {
			continue
		}
		if opts.FilerName != "" && (doc.FilerName == nil || !strings.Contains(*doc.FilerName, opts.FilerName)) {
			continue
		}
		if opts.DocDescription != "" && (doc.DocDescription == nil || !strings.Contains(*doc.DocDescription, opts.DocDescription)) {
			continue
		}
		result = append(result, ToDocumentInfo(doc))
	}
	return result
}

func dateRange(from, to string) ([]string, error) {
	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, fmt.Errorf("invalid from date: %w", err)
	}
	toTime, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil, fmt.Errorf("invalid to date: %w", err)
	}
	var dates []string
	for d := fromTime; !d.After(toTime); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format("2006-01-02"))
	}
	return dates, nil
}

func classifyWarning(err error) string {
	if edinetErr, ok := err.(*api.EDINETError); ok {
		return fmt.Sprintf("%s: %s", edinetErr.Code, edinetErr.Message)
	}
	return fmt.Sprintf("INTERNAL_ERROR: %s", err.Error())
}
