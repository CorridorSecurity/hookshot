package agenttrace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReadTrace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trace.json")

	record := TraceRecord{
		Version:   "0.1.0",
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		Timestamp: "2026-01-25T10:00:00Z",
		Vcs:       &Vcs{Type: VcsGit, Revision: "abc123"},
		Tool:      &Tool{Name: "claude-code", Version: "1.0.0"},
		Files: []File{
			{
				Path: "src/main.go",
				Conversations: []Conversation{
					{
						URL:         "https://example.com/conv/1",
						Contributor: &Contributor{Type: ContributorAI},
						Ranges:      []Range{{StartLine: 1, EndLine: 50}},
					},
				},
			},
		},
	}

	if err := WriteTrace(path, record); err != nil {
		t.Fatalf("WriteTrace failed: %v", err)
	}

	got, err := ReadTrace(path)
	if err != nil {
		t.Fatalf("ReadTrace failed: %v", err)
	}

	if got.Version != record.Version {
		t.Errorf("Version = %q, want %q", got.Version, record.Version)
	}
	if got.ID != record.ID {
		t.Errorf("ID = %q, want %q", got.ID, record.ID)
	}
	if got.Timestamp != record.Timestamp {
		t.Errorf("Timestamp = %q, want %q", got.Timestamp, record.Timestamp)
	}
	if got.Vcs.Type != VcsGit {
		t.Errorf("Vcs.Type = %q, want %q", got.Vcs.Type, VcsGit)
	}
	if got.Tool.Name != "claude-code" {
		t.Errorf("Tool.Name = %q, want %q", got.Tool.Name, "claude-code")
	}
	if len(got.Files) != 1 {
		t.Fatalf("len(Files) = %d, want 1", len(got.Files))
	}
	if got.Files[0].Path != "src/main.go" {
		t.Errorf("Path = %q, want %q", got.Files[0].Path, "src/main.go")
	}
}

func TestReadTraceFileNotFound(t *testing.T) {
	_, err := ReadTrace("/nonexistent/path/trace.json")
	if err == nil {
		t.Error("ReadTrace should fail for nonexistent file")
	}
}

func TestReadTraceInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("not json"), 0644)

	_, err := ReadTrace(path)
	if err == nil {
		t.Error("ReadTrace should fail for invalid JSON")
	}
}

func TestWriteTraceCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "trace.json")

	// WriteTrace should fail if parent directory doesn't exist (expected behavior).
	record := TraceRecord{
		Version:   "0.1.0",
		ID:        "test-id",
		Timestamp: "2026-01-25T10:00:00Z",
		Files:     []File{},
	}

	err := WriteTrace(path, record)
	if err == nil {
		t.Error("WriteTrace should fail when parent directory doesn't exist")
	}
}

func TestWriteTraceIndentedJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trace.json")

	record := TraceRecord{
		Version:   "0.1.0",
		ID:        "test-id",
		Timestamp: "2026-01-25T10:00:00Z",
		Files:     []File{},
	}

	if err := WriteTrace(path, record); err != nil {
		t.Fatalf("WriteTrace failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	// Verify the output is indented (contains newlines and spaces).
	if len(content) < 10 {
		t.Fatal("Output too short to be indented JSON")
	}
	if content[0] != '{' {
		t.Errorf("Expected JSON to start with '{', got %q", string(content[0]))
	}
	if content[len(content)-1] != '\n' {
		t.Error("Expected trailing newline")
	}
}
