package api

// DocumentListResponse is the wire format for the EDINET document list API (type=2).
type DocumentListResponse struct {
	Metadata Metadata   `json:"metadata"`
	Results  []Document `json:"results"`
}

// Metadata contains the response metadata from the document list API.
type Metadata struct {
	Title           string    `json:"title"`
	Parameter       Parameter `json:"parameter"`
	ResultSet       ResultSet `json:"resultset"`
	ProcessDateTime string    `json:"processDateTime"`
	Status          string    `json:"status"`
	Message         string    `json:"message"`
}

// Parameter represents the request parameters echoed in the response.
type Parameter struct {
	Date string `json:"date"`
	Type string `json:"type"`
}

// ResultSet contains the count of results.
type ResultSet struct {
	Count int `json:"count"`
}

// Document represents a single document entry from the EDINET API.
// All fields except SeqNumber and DocID use *string to handle null values
// (e.g., when a document's viewing period has expired, most fields become null).
type Document struct {
	SeqNumber            int     `json:"seqNumber"`
	DocID                string  `json:"docID"`
	EdinetCode           *string `json:"edinetCode"`
	SecCode              *string `json:"secCode"`
	JCN                  *string `json:"JCN"`
	FilerName            *string `json:"filerName"`
	FundCode             *string `json:"fundCode"`
	OrdinanceCode        *string `json:"ordinanceCode"`
	FormCode             *string `json:"formCode"`
	DocTypeCode          *string `json:"docTypeCode"`
	PeriodStart          *string `json:"periodStart"`
	PeriodEnd            *string `json:"periodEnd"`
	SubmitDateTime       *string `json:"submitDateTime"`
	DocDescription       *string `json:"docDescription"`
	IssuerEdinetCode     *string `json:"issuerEdinetCode"`
	SubjectEdinetCode    *string `json:"subjectEdinetCode"`
	SubsidiaryEdinetCode *string `json:"subsidiaryEdinetCode"`
	CurrentReportReason  *string `json:"currentReportReason"`
	ParentDocID          *string `json:"parentDocID"`
	OpeDateTime          *string `json:"opeDateTime"`
	WithdrawalStatus     *string `json:"withdrawalStatus"`
	DocInfoEditStatus    *string `json:"docInfoEditStatus"`
	DisclosureStatus     *string `json:"disclosureStatus"`
	XbrlFlag             *string `json:"xbrlFlag"`
	PdfFlag              *string `json:"pdfFlag"`
	AttachDocFlag        *string `json:"attachDocFlag"`
	EnglishDocFlag       *string `json:"englishDocFlag"`
	CsvFlag              *string `json:"csvFlag"`
	LegalStatus          *string `json:"legalStatus"`
}

// AuthErrorResponse is the wire format for EDINET 401 errors.
// Note: StatusCode is int (not string), with uppercase S and C.
type AuthErrorResponse struct {
	StatusCode int    `json:"StatusCode"`
	Message    string `json:"message"`
}

// ErrorResponse is the wire format for non-401 EDINET errors (400, 404, 500).
type ErrorResponse struct {
	Metadata ErrorMetadata `json:"metadata"`
}

// ErrorMetadata contains the error information in the standard error format.
type ErrorMetadata struct {
	Title   string `json:"title"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
