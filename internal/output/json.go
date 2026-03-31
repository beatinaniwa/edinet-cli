package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// PrintJSON writes v as pretty-printed JSON to w, followed by a newline.
func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// PrintError writes a structured error to w in the CLI's standard error format.
func PrintError(w io.Writer, code string, message string) {
	payload := struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	payload.Error.Code = code
	payload.Error.Message = message
	data, err := json.Marshal(payload)
	if err != nil {
		_, _ = fmt.Fprintf(w, `{"error":{"code":"INTERNAL_ERROR","message":"failed to marshal error: %s"}}`, err)
		return
	}
	_, _ = fmt.Fprintln(w, string(data))
}
