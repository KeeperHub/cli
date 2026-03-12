package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteJSON marshals data as indented JSON and writes it to w with a trailing newline.
func WriteJSON(w io.Writer, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", b)
	return err
}

// WriteJSONError writes {"error": message, "code": code} as indented JSON to w.
func WriteJSONError(w io.Writer, message string, code int) error {
	return WriteJSON(w, map[string]any{
		"error": message,
		"code":  code,
	})
}
