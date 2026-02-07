package agenttrace

import (
	"encoding/json"
	"os"
)

// WriteTrace writes a single trace record to the given file path as
// indented JSON. It creates or truncates the file.
func WriteTrace(path string, record TraceRecord) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// ReadTrace reads a single trace record from the given file path.
func ReadTrace(path string) (TraceRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TraceRecord{}, err
	}
	var record TraceRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return TraceRecord{}, err
	}
	return record, nil
}
