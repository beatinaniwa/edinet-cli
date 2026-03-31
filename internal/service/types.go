package service

import "github.com/beatinaniwa/edinet-cli/internal/api"

// DocumentInfo is the CLI output DTO for a document.
// It maps from the API wire format (api.Document) to a cleaner representation.
type DocumentInfo struct {
	DocID          string `json:"doc_id"`
	EdinetCode     string `json:"edinet_code,omitempty"`
	SecCode        string `json:"sec_code,omitempty"`
	FilerName      string `json:"filer_name,omitempty"`
	DocTypeCode    string `json:"doc_type_code,omitempty"`
	DocDescription string `json:"doc_description,omitempty"`
	PeriodStart    string `json:"period_start,omitempty"`
	PeriodEnd      string `json:"period_end,omitempty"`
	SubmitDateTime string `json:"submit_datetime,omitempty"`
	HasXbrl        bool   `json:"has_xbrl"`
	HasPdf         bool   `json:"has_pdf"`
	HasCsv         bool   `json:"has_csv"`
	LegalStatus    string `json:"legal_status,omitempty"`
}

// ListResult is the CLI output structure for document list commands.
type ListResult struct {
	Metadata ListMetadata   `json:"metadata"`
	Results  []DocumentInfo `json:"results"`
}

// ListMetadata contains metadata about a list query.
type ListMetadata struct {
	Date         string     `json:"date,omitempty"`
	DateRange    *DateRange `json:"dateRange,omitempty"`
	TotalResults int        `json:"totalResults"`
	Warnings     []string   `json:"warnings,omitempty"`
}

// DateRange represents a from/to date range.
type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ToDocumentInfo maps an API wire-format Document to a CLI output DocumentInfo.
func ToDocumentInfo(d api.Document) DocumentInfo {
	return DocumentInfo{
		DocID:          d.DocID,
		EdinetCode:     derefStr(d.EdinetCode),
		SecCode:        derefStr(d.SecCode),
		FilerName:      derefStr(d.FilerName),
		DocTypeCode:    derefStr(d.DocTypeCode),
		DocDescription: derefStr(d.DocDescription),
		PeriodStart:    derefStr(d.PeriodStart),
		PeriodEnd:      derefStr(d.PeriodEnd),
		SubmitDateTime: derefStr(d.SubmitDateTime),
		HasXbrl:        flagToBool(d.XbrlFlag),
		HasPdf:         flagToBool(d.PdfFlag),
		HasCsv:         flagToBool(d.CsvFlag),
		LegalStatus:    mapLegalStatus(d.LegalStatus),
	}
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func flagToBool(p *string) bool {
	return p != nil && *p == "1"
}

func mapLegalStatus(p *string) string {
	if p == nil {
		return ""
	}
	switch *p {
	case "1":
		return "active"
	case "2":
		return "extended"
	case "0":
		return "expired"
	default:
		return *p
	}
}
