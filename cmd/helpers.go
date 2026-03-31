package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/beatinaniwa/edinet-cli/internal/api"
	"github.com/beatinaniwa/edinet-cli/internal/output"
)

// outputResult writes v as JSON or table to w based on the current format flag.
func outputResult(w io.Writer, v any) error {
	if app != nil && app.Config != nil && app.Config.Format == "table" {
		return outputTable(w, v)
	}
	return output.PrintJSON(w, v)
}

// outputTable renders v as a human-readable table.
// It converts the value to a []map[string]any and extracts headers from keys.
func outputTable(w io.Writer, v any) error {
	// Marshal then unmarshal to get a generic representation
	data, err := json.Marshal(v)
	if err != nil {
		return output.PrintJSON(w, v) // fallback to JSON
	}

	// Try as array of objects
	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err == nil && len(rows) > 0 {
		return renderMapTable(w, rows)
	}

	// Try as single object with "results" key
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err == nil {
		if results, ok := obj["results"]; ok {
			if resultData, err := json.Marshal(results); err == nil {
				if err := json.Unmarshal(resultData, &rows); err == nil && len(rows) > 0 {
					return renderMapTable(w, rows)
				}
			}
		}
	}

	// Fallback to JSON for non-tabular data
	return output.PrintJSON(w, v)
}

func renderMapTable(w io.Writer, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}
	// Collect headers from first row
	var headers []string
	for k := range rows[0] {
		headers = append(headers, k)
	}

	var tableRows [][]string
	for _, row := range rows {
		var cells []string
		for _, h := range headers {
			cells = append(cells, fmt.Sprintf("%v", row[h]))
		}
		tableRows = append(tableRows, cells)
	}

	output.PrintTable(w, headers, tableRows)
	return nil
}

// exitError writes a structured error to w and returns the appropriate exit code.
func exitError(w io.Writer, err error) int {
	if edinetErr, ok := err.(*api.EDINETError); ok {
		output.PrintError(w, string(edinetErr.Code), edinetErr.Message)
		return edinetErr.ExitCode()
	}
	output.PrintError(w, string(api.ErrInternal), err.Error())
	return api.ExitGeneral
}
