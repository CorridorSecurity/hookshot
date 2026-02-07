package agenttrace

import (
	"encoding/json"
	"testing"
)

func TestTraceRecordRoundTrip(t *testing.T) {
	record := TraceRecord{
		Version:   "0.1.0",
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		Timestamp: "2026-01-25T10:00:00Z",
		Vcs:       &Vcs{Type: VcsGit, Revision: "abc123"},
		Tool:      &Tool{Name: "claude-code", Version: "1.0.0"},
		Files: []File{
			{
				Path: "src/app.ts",
				Conversations: []Conversation{
					{
						URL:         "https://example.com/conversation/1",
						Contributor: &Contributor{Type: ContributorAI, ModelID: "anthropic/claude-sonnet-4-5-20250929"},
						Ranges: []Range{
							{StartLine: 1, EndLine: 50},
							{StartLine: 75, EndLine: 100, ContentHash: "sha256:abc"},
						},
						Related: []RelatedResource{
							{Type: "issue", URL: "https://github.com/org/repo/issues/1"},
						},
					},
				},
			},
		},
		Metadata: map[string]any{"source": "hookshot"},
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got TraceRecord
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
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
	if got.Vcs.Revision != "abc123" {
		t.Errorf("Vcs.Revision = %q, want %q", got.Vcs.Revision, "abc123")
	}
	if got.Tool.Name != "claude-code" {
		t.Errorf("Tool.Name = %q, want %q", got.Tool.Name, "claude-code")
	}
	if len(got.Files) != 1 {
		t.Fatalf("len(Files) = %d, want 1", len(got.Files))
	}
	if got.Files[0].Path != "src/app.ts" {
		t.Errorf("Files[0].Path = %q, want %q", got.Files[0].Path, "src/app.ts")
	}
	if len(got.Files[0].Conversations) != 1 {
		t.Fatalf("len(Conversations) = %d, want 1", len(got.Files[0].Conversations))
	}

	conv := got.Files[0].Conversations[0]
	if conv.URL != "https://example.com/conversation/1" {
		t.Errorf("Conversation.URL = %q, want %q", conv.URL, "https://example.com/conversation/1")
	}
	if conv.Contributor.Type != ContributorAI {
		t.Errorf("Contributor.Type = %q, want %q", conv.Contributor.Type, ContributorAI)
	}
	if conv.Contributor.ModelID != "anthropic/claude-sonnet-4-5-20250929" {
		t.Errorf("Contributor.ModelID = %q, want %q", conv.Contributor.ModelID, "anthropic/claude-sonnet-4-5-20250929")
	}
	if len(conv.Ranges) != 2 {
		t.Fatalf("len(Ranges) = %d, want 2", len(conv.Ranges))
	}
	if conv.Ranges[0].StartLine != 1 || conv.Ranges[0].EndLine != 50 {
		t.Errorf("Range[0] = {%d, %d}, want {1, 50}", conv.Ranges[0].StartLine, conv.Ranges[0].EndLine)
	}
	if conv.Ranges[1].ContentHash != "sha256:abc" {
		t.Errorf("Range[1].ContentHash = %q, want %q", conv.Ranges[1].ContentHash, "sha256:abc")
	}
	if len(conv.Related) != 1 {
		t.Fatalf("len(Related) = %d, want 1", len(conv.Related))
	}
	if conv.Related[0].Type != "issue" {
		t.Errorf("Related[0].Type = %q, want %q", conv.Related[0].Type, "issue")
	}
}

func TestMinimalTraceRecord(t *testing.T) {
	// Matches the minimal valid example from the spec.
	record := TraceRecord{
		Version:   "0.1.0",
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		Timestamp: "2026-01-25T10:00:00Z",
		Files: []File{
			{
				Path: "src/app.ts",
				Conversations: []Conversation{
					{
						Contributor: &Contributor{Type: ContributorAI},
						Ranges:      []Range{{StartLine: 1, EndLine: 50}},
					},
				},
			},
		},
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify omitempty: optional fields should not appear.
	var raw map[string]any
	json.Unmarshal(data, &raw)

	if _, ok := raw["vcs"]; ok {
		t.Error("vcs should be omitted when nil")
	}
	if _, ok := raw["tool"]; ok {
		t.Error("tool should be omitted when nil")
	}
	if _, ok := raw["metadata"]; ok {
		t.Error("metadata should be omitted when nil")
	}

	// Verify round-trip.
	var got TraceRecord
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Version != "0.1.0" {
		t.Errorf("Version = %q, want %q", got.Version, "0.1.0")
	}
	if len(got.Files) != 1 {
		t.Fatalf("len(Files) = %d, want 1", len(got.Files))
	}
}

func TestContributorTypeValues(t *testing.T) {
	tests := []struct {
		ct   ContributorType
		want string
	}{
		{ContributorHuman, "human"},
		{ContributorAI, "ai"},
		{ContributorMixed, "mixed"},
		{ContributorUnknown, "unknown"},
	}

	for _, tt := range tests {
		if string(tt.ct) != tt.want {
			t.Errorf("ContributorType = %q, want %q", tt.ct, tt.want)
		}
	}
}

func TestVcsTypeValues(t *testing.T) {
	tests := []struct {
		vt   VcsType
		want string
	}{
		{VcsGit, "git"},
		{VcsJJ, "jj"},
		{VcsHg, "hg"},
		{VcsSvn, "svn"},
	}

	for _, tt := range tests {
		if string(tt.vt) != tt.want {
			t.Errorf("VcsType = %q, want %q", tt.vt, tt.want)
		}
	}
}

func TestRangeWithContributorOverride(t *testing.T) {
	conv := Conversation{
		Contributor: &Contributor{Type: ContributorAI},
		Ranges: []Range{
			{StartLine: 1, EndLine: 10},
			{StartLine: 11, EndLine: 20, Contributor: &Contributor{Type: ContributorHuman}},
		},
	}

	data, err := json.Marshal(conv)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got Conversation
	json.Unmarshal(data, &got)

	if got.Ranges[0].Contributor != nil {
		t.Error("Range[0].Contributor should be nil (omitted)")
	}
	if got.Ranges[1].Contributor == nil {
		t.Fatal("Range[1].Contributor should not be nil")
	}
	if got.Ranges[1].Contributor.Type != ContributorHuman {
		t.Errorf("Range[1].Contributor.Type = %q, want %q", got.Ranges[1].Contributor.Type, ContributorHuman)
	}
}
