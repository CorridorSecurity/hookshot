package internal

import (
	"encoding/json"
	"io"
	"os"
)

// ReadJSON reads JSON from stdin and unmarshals it into v.
// Returns an error if stdin cannot be read or JSON is invalid.
func ReadJSON(v any) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// WriteJSON marshals v to JSON and writes it to stdout.
func WriteJSON(v any) error {
	return json.NewEncoder(os.Stdout).Encode(v)
}
