package company

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// CompanyEntry represents a single entry from the EDINET code list.
type CompanyEntry struct {
	EdinetCode      string `json:"edinet_code"`
	SubmitterType   string `json:"submitter_type"`
	ListingStatus   string `json:"listing_status"`
	Consolidated    string `json:"consolidated"`
	CapitalAmount   string `json:"capital_amount"`
	EndOfFiscalYear string `json:"end_of_fiscal_year"`
	SubmitterName   string `json:"submitter_name"`
	SubmitterNameEN string `json:"submitter_name_en"`
	SubmitterNameKana string `json:"submitter_name_kana"`
	Address         string `json:"address"`
	IndustryName    string `json:"industry_name"`
	SecCode         string `json:"sec_code"`
	JCN             string `json:"jcn"`
}

// Registry holds the EDINET code list entries.
type Registry struct {
	Entries []CompanyEntry
}

// expectedHeaders are the known column headers for the EDINET code list CSV.
var expectedHeaders = []string{
	"EDINETコード",
	"提出者種別",
	"上場区分",
	"連結の有無",
	"資本金",
	"決算日",
	"提出者名",
	"提出者名（英字）",
	"提出者名（ヨミ）",
	"所在地",
	"提出者業種",
	"証券コード",
	"提出者法人番号",
}

// LoadFromCSV parses the EDINET code list CSV data.
// The CSV format: row 1 = title (skipped), row 2 = headers (validated), row 3+ = data.
func (r *Registry) LoadFromCSV(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("empty CSV data")
	}

	// Strip BOM if present
	data = stripBOM(data)

	// Convert Shift-JIS to UTF-8 if needed
	data, err := ensureUTF8(data)
	if err != nil {
		return fmt.Errorf("failed to convert encoding: %w", err)
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	// Skip title row (row 1)
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("failed to read title row: %w", err)
	}

	// Read and validate header row (row 2)
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header row: %w", err)
	}
	if err := validateHeaders(headers); err != nil {
		return err
	}

	// Read data rows
	var entries []CompanyEntry
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV row: %w", err)
		}
		if len(record) < 13 {
			continue // Skip malformed rows
		}
		entries = append(entries, CompanyEntry{
			EdinetCode:      strings.TrimSpace(record[0]),
			SubmitterType:   strings.TrimSpace(record[1]),
			ListingStatus:   strings.TrimSpace(record[2]),
			Consolidated:    strings.TrimSpace(record[3]),
			CapitalAmount:   strings.TrimSpace(record[4]),
			EndOfFiscalYear: strings.TrimSpace(record[5]),
			SubmitterName:   strings.TrimSpace(record[6]),
			SubmitterNameEN: strings.TrimSpace(record[7]),
			SubmitterNameKana: strings.TrimSpace(record[8]),
			Address:         strings.TrimSpace(record[9]),
			IndustryName:    strings.TrimSpace(record[10]),
			SecCode:         strings.TrimSpace(record[11]),
			JCN:             strings.TrimSpace(record[12]),
		})
	}

	r.Entries = entries
	return nil
}

// FindByEdinetCode returns the entry matching the given EDINET code.
func (r *Registry) FindByEdinetCode(code string) (*CompanyEntry, error) {
	for i := range r.Entries {
		if r.Entries[i].EdinetCode == code {
			return &r.Entries[i], nil
		}
	}
	return nil, fmt.Errorf("EDINET code %q not found", code)
}

// FindBySecCode returns the entry matching the given securities code.
// Supports both 4-digit and 5-digit codes (4-digit matches prefix of 5-digit).
func (r *Registry) FindBySecCode(code string) (*CompanyEntry, error) {
	for i := range r.Entries {
		if r.Entries[i].SecCode == code {
			return &r.Entries[i], nil
		}
		// 4-digit code matches 5-digit code prefix
		if len(code) == 4 && len(r.Entries[i].SecCode) == 5 && strings.HasPrefix(r.Entries[i].SecCode, code) {
			return &r.Entries[i], nil
		}
	}
	return nil, fmt.Errorf("securities code %q not found", code)
}

func validateHeaders(headers []string) error {
	if len(headers) < len(expectedHeaders) {
		return fmt.Errorf("unexpected registry format: expected %d columns, got %d", len(expectedHeaders), len(headers))
	}
	// Check that the first expected header matches (normalize full-width ASCII)
	got := normalizeFullWidth(strings.TrimSpace(headers[0]))
	if got != expectedHeaders[0] {
		return fmt.Errorf("unexpected registry format: missing column %q, got %q", expectedHeaders[0], headers[0])
	}
	return nil
}

// normalizeFullWidth converts full-width ASCII characters (Ａ-Ｚ, ａ-ｚ, ０-９) to half-width.
func normalizeFullWidth(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 'Ａ' && r <= 'Ｚ' {
			b.WriteRune(r - 'Ａ' + 'A')
		} else if r >= 'ａ' && r <= 'ｚ' {
			b.WriteRune(r - 'ａ' + 'a')
		} else if r >= '０' && r <= '９' {
			b.WriteRune(r - '０' + '0')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ensureUTF8 converts Shift-JIS data to UTF-8 if the data is not valid UTF-8.
func ensureUTF8(data []byte) ([]byte, error) {
	if utf8.Valid(data) {
		return data, nil
	}
	decoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(data), japanese.ShiftJIS.NewDecoder()))
	if err != nil {
		return data, err
	}
	return decoded, nil
}

func stripBOM(data []byte) []byte {
	bom := []byte{0xEF, 0xBB, 0xBF} // UTF-8 BOM
	if bytes.HasPrefix(data, bom) {
		return data[3:]
	}
	return data
}
