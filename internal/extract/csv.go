package extract

import (
	"bytes"
	"encoding/csv"
	"encoding/binary"
	"fmt"
	"io"
	"path/filepath"
	"unicode/utf16"
)

// CSVDataResult is the output structure for extracted CSV financial data.
type CSVDataResult struct {
	Files []CSVFile `json:"files"`
}

// CSVFile represents one CSV file extracted from a ZIP archive.
type CSVFile struct {
	Filename string     `json:"filename"`
	Headers  []string   `json:"headers"`
	Rows     [][]string `json:"rows"`
}

// ExtractCSVData extracts CSV financial data from a type=5 ZIP archive.
// Reads all CSV files under XBRL_TO_CSV/ and returns them as structured data.
func ExtractCSVData(zipData []byte) (*CSVDataResult, error) {
	entries, err := ReadFromZip(zipData, "XBRL_TO_CSV/*.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV from ZIP: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("directory not found in archive: XBRL_TO_CSV/ (no CSV files found)")
	}

	result := &CSVDataResult{}
	for _, entry := range entries {
		csvFile, err := parseCSV(entry)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name, err)
		}
		result.Files = append(result.Files, *csvFile)
	}

	return result, nil
}

func parseCSV(entry ZipEntry) (*CSVFile, error) {
	// EDINET CSVs may be UTF-16LE with BOM. Detect and convert to UTF-8.
	data := decodeToUTF8(entry.Data)

	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1 // Allow uneven rows
	r.LazyQuotes = true

	// EDINET CSVs may use tab delimiter
	if bytes.Contains(data[:min(len(data), 1024)], []byte("\t")) {
		r.Comma = '\t'
	}

	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	var rows [][]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}
		rows = append(rows, record)
	}

	return &CSVFile{
		Filename: filepath.Base(entry.Name),
		Headers:  headers,
		Rows:     rows,
	}, nil
}

// decodeToUTF8 detects UTF-16LE/BE BOM and converts to UTF-8. Returns as-is if already UTF-8.
func decodeToUTF8(data []byte) []byte {
	if len(data) < 2 {
		return data
	}
	// UTF-16LE BOM: FF FE
	if data[0] == 0xFF && data[1] == 0xFE {
		return utf16LEToUTF8(data[2:])
	}
	// UTF-16BE BOM: FE FF
	if data[0] == 0xFE && data[1] == 0xFF {
		return utf16BEToUTF8(data[2:])
	}
	// UTF-8 BOM: EF BB BF
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

func utf16LEToUTF8(data []byte) []byte {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}
	u16 := make([]uint16, len(data)/2)
	for i := range u16 {
		u16[i] = binary.LittleEndian.Uint16(data[2*i:])
	}
	runes := utf16.Decode(u16)
	var buf bytes.Buffer
	for _, r := range runes {
		buf.WriteRune(r)
	}
	return buf.Bytes()
}

func utf16BEToUTF8(data []byte) []byte {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}
	u16 := make([]uint16, len(data)/2)
	for i := range u16 {
		u16[i] = binary.BigEndian.Uint16(data[2*i:])
	}
	runes := utf16.Decode(u16)
	var buf bytes.Buffer
	for _, r := range runes {
		buf.WriteRune(r)
	}
	return buf.Bytes()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
