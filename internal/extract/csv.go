package extract

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
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
	r := csv.NewReader(bytes.NewReader(entry.Data))
	r.FieldsPerRecord = -1 // Allow uneven rows
	r.LazyQuotes = true

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
