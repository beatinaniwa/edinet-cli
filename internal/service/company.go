package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/beatinaniwa/edinet-cli/internal/company"
)

var (
	edinetCodeRe = regexp.MustCompile(`^E\d{5}$`)
	secCodeRe    = regexp.MustCompile(`^\d{4,5}$`)
)

// CompanyService provides company search and filings lookup.
type CompanyService struct {
	registry *company.Registry
	docSvc   *DocumentService
}

// FilingsOptions configures a company filings query.
type FilingsOptions struct {
	DocType   string
	From      string
	To        string
	RateLimit time.Duration
	Limit     int
}

// NewCompanyService creates a new CompanyService.
func NewCompanyService(registry *company.Registry, docSvc *DocumentService) *CompanyService {
	return &CompanyService{registry: registry, docSvc: docSvc}
}

// Search finds companies matching the query.
func (s *CompanyService) Search(query string) []company.CompanyEntry {
	return s.registry.Search(query)
}

// SearchByIndustry finds companies by industry name.
func (s *CompanyService) SearchByIndustry(industry string) []company.CompanyEntry {
	return s.registry.SearchByIndustry(industry)
}

// Filings retrieves document filings for a company.
// The code can be an EDINET code (E\d{5}), a securities code (\d{4,5}), or a company name.
func (s *CompanyService) Filings(ctx context.Context, code string, opts FilingsOptions) (*ListResult, error) {
	edinetCode, err := s.resolveEdinetCode(code)
	if err != nil {
		return nil, err
	}

	jst := time.FixedZone("JST", 9*60*60)
	nowJST := time.Now().In(jst)

	from := opts.From
	to := opts.To
	if from == "" {
		from = nowJST.AddDate(0, 0, -365).Format("2006-01-02")
	}
	if to == "" {
		to = nowJST.Format("2006-01-02")
	}

	return s.docSvc.List(ctx, ListOptions{
		From:       from,
		To:         to,
		EdinetCode: edinetCode,
		DocType:    opts.DocType,
		RateLimit:  opts.RateLimit,
		Limit:      opts.Limit,
	})
}

func (s *CompanyService) resolveEdinetCode(code string) (string, error) {
	code = strings.TrimSpace(code)

	// Already an EDINET code — pass through even if not in local registry
	if edinetCodeRe.MatchString(code) {
		return code, nil
	}

	// Securities code
	if secCodeRe.MatchString(code) {
		entry, err := s.registry.FindBySecCode(code)
		if err != nil {
			return "", fmt.Errorf("unknown securities code %q", code)
		}
		return entry.EdinetCode, nil
	}

	// Name search — return first match
	results := s.registry.Search(code)
	if len(results) == 0 {
		return "", fmt.Errorf("no company found for %q", code)
	}
	return results[0].EdinetCode, nil
}
