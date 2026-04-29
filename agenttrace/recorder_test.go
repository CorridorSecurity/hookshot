package agenttrace

import (
	"strings"
	"testing"
)

func TestNewRecorder(t *testing.T) {
	rec := NewRecorder(&Tool{Name: "test-tool", Version: "1.0"})
	if rec == nil {
		t.Fatal("NewRecorder returned nil")
	}
}

func TestRecordAndBuild(t *testing.T) {
	rec := NewRecorder(&Tool{Name: "claude-code", Version: "1.0"})

	rec.Record("src/main.go", "https://example.com/conv/1",
		Contributor{Type: ContributorAI, ModelID: "anthropic/claude-sonnet-4-5-20250929"},
		[]Range{{StartLine: 10, EndLine: 25}},
	)

	trace := rec.Build(&Vcs{Type: VcsGit, Revision: "abc123"})

	if trace.Version != Version {
		t.Errorf("Version = %q, want %q", trace.Version, Version)
	}
	if trace.ID == "" {
		t.Error("ID should not be empty")
	}
	if trace.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
	if trace.Tool.Name != "claude-code" {
		t.Errorf("Tool.Name = %q, want %q", trace.Tool.Name, "claude-code")
	}
	if trace.Vcs.Type != VcsGit {
		t.Errorf("Vcs.Type = %q, want %q", trace.Vcs.Type, VcsGit)
	}
	if trace.Vcs.Revision != "abc123" {
		t.Errorf("Vcs.Revision = %q, want %q", trace.Vcs.Revision, "abc123")
	}
	if len(trace.Files) != 1 {
		t.Fatalf("len(Files) = %d, want 1", len(trace.Files))
	}
	if trace.Files[0].Path != "src/main.go" {
		t.Errorf("Files[0].Path = %q, want %q", trace.Files[0].Path, "src/main.go")
	}
	if len(trace.Files[0].Conversations) != 1 {
		t.Fatalf("len(Conversations) = %d, want 1", len(trace.Files[0].Conversations))
	}

	conv := trace.Files[0].Conversations[0]
	if conv.URL != "https://example.com/conv/1" {
		t.Errorf("URL = %q, want %q", conv.URL, "https://example.com/conv/1")
	}
	if conv.Contributor.Type != ContributorAI {
		t.Errorf("Contributor.Type = %q, want %q", conv.Contributor.Type, ContributorAI)
	}
	if len(conv.Ranges) != 1 {
		t.Fatalf("len(Ranges) = %d, want 1", len(conv.Ranges))
	}
	if conv.Ranges[0].StartLine != 10 || conv.Ranges[0].EndLine != 25 {
		t.Errorf("Range = {%d, %d}, want {10, 25}", conv.Ranges[0].StartLine, conv.Ranges[0].EndLine)
	}
}

func TestMultipleRecords(t *testing.T) {
	rec := NewRecorder(nil)

	rec.Record("a.go", "", Contributor{Type: ContributorAI}, nil)
	rec.Record("b.go", "", Contributor{Type: ContributorHuman}, nil)
	rec.Record("a.go", "https://example.com/conv/2", Contributor{Type: ContributorMixed}, nil)

	trace := rec.Build(nil)

	if len(trace.Files) != 2 {
		t.Fatalf("len(Files) = %d, want 2", len(trace.Files))
	}

	// Find a.go — it should have 2 conversations.
	var aFile *File
	for i := range trace.Files {
		if trace.Files[i].Path == "a.go" {
			aFile = &trace.Files[i]
			break
		}
	}
	if aFile == nil {
		t.Fatal("File a.go not found")
	}
	if len(aFile.Conversations) != 2 {
		t.Errorf("a.go conversations = %d, want 2", len(aFile.Conversations))
	}
}

func TestBuildWithNilVcsAndTool(t *testing.T) {
	rec := NewRecorder(nil)
	rec.Record("file.go", "", Contributor{Type: ContributorUnknown}, nil)

	trace := rec.Build(nil)

	if trace.Vcs != nil {
		t.Error("Vcs should be nil")
	}
	if trace.Tool != nil {
		t.Error("Tool should be nil")
	}
	if trace.Version != Version {
		t.Errorf("Version = %q, want %q", trace.Version, Version)
	}
}

func TestReset(t *testing.T) {
	rec := NewRecorder(&Tool{Name: "test"})

	rec.Record("a.go", "", Contributor{Type: ContributorAI}, nil)
	rec.Record("b.go", "", Contributor{Type: ContributorAI}, nil)

	rec.Reset()

	trace := rec.Build(nil)
	if len(trace.Files) != 0 {
		t.Errorf("after Reset, len(Files) = %d, want 0", len(trace.Files))
	}
}

func TestBuildGeneratesUniqueIDs(t *testing.T) {
	rec := NewRecorder(nil)
	rec.Record("file.go", "", Contributor{Type: ContributorAI}, nil)

	trace1 := rec.Build(nil)
	trace2 := rec.Build(nil)

	if trace1.ID == trace2.ID {
		t.Errorf("Build should generate unique IDs, got same: %s", trace1.ID)
	}
}

func TestBuildTimestampFormat(t *testing.T) {
	rec := NewRecorder(nil)
	rec.Record("file.go", "", Contributor{Type: ContributorAI}, nil)

	trace := rec.Build(nil)

	// RFC3339 timestamps contain "T" and end with "Z" (UTC).
	if !strings.Contains(trace.Timestamp, "T") {
		t.Errorf("Timestamp %q does not look like RFC3339", trace.Timestamp)
	}
	if !strings.HasSuffix(trace.Timestamp, "Z") {
		t.Errorf("Timestamp %q should end with Z (UTC)", trace.Timestamp)
	}
}
