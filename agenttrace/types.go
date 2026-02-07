package agenttrace

// ContributorType identifies whether code was written by a human, AI, or a mix.
type ContributorType string

const (
	ContributorHuman   ContributorType = "human"
	ContributorAI      ContributorType = "ai"
	ContributorMixed   ContributorType = "mixed"
	ContributorUnknown ContributorType = "unknown"
)

// VcsType identifies the version control system.
type VcsType string

const (
	VcsGit VcsType = "git"
	VcsJJ  VcsType = "jj"
	VcsHg  VcsType = "hg"
	VcsSvn VcsType = "svn"
)

// TraceRecord is the top-level Agent Trace record.
type TraceRecord struct {
	Version   string         `json:"version"`
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Vcs       *Vcs           `json:"vcs,omitempty"`
	Tool      *Tool          `json:"tool,omitempty"`
	Files     []File         `json:"files"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// File represents a traced file with its associated conversations.
type File struct {
	Path          string         `json:"path"`
	Conversations []Conversation `json:"conversations"`
}

// Conversation links a conversation URL to the code ranges it produced.
type Conversation struct {
	URL         string            `json:"url,omitempty"`
	Contributor *Contributor      `json:"contributor,omitempty"`
	Ranges      []Range           `json:"ranges"`
	Related     []RelatedResource `json:"related,omitempty"`
}

// Range identifies a line range within a file attributed to a conversation.
type Range struct {
	StartLine   int          `json:"start_line"`
	EndLine     int          `json:"end_line"`
	ContentHash string       `json:"content_hash,omitempty"`
	Contributor *Contributor `json:"contributor,omitempty"`
}

// Contributor describes who wrote the code.
type Contributor struct {
	Type    ContributorType `json:"type"`
	ModelID string          `json:"model_id,omitempty"`
}

// Tool identifies the AI coding tool that produced the code.
type Tool struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// Vcs holds version control information for the trace.
type Vcs struct {
	Type     VcsType `json:"type"`
	Revision string  `json:"revision"`
}

// RelatedResource links to an external resource associated with a conversation.
type RelatedResource struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}
