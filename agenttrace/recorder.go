package agenttrace

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// Version is the Agent Trace specification version implemented by this package.
const Version = "0.1.0"

// Recorder accumulates file edit information and builds TraceRecord values.
// It is safe for concurrent use.
type Recorder struct {
	mu    sync.Mutex
	tool  *Tool
	files map[string][]Conversation // keyed by file path
}

// NewRecorder creates a Recorder that tags traces with the given tool info.
// The tool parameter may be nil if tool information is not available.
func NewRecorder(tool *Tool) *Recorder {
	return &Recorder{
		tool:  tool,
		files: make(map[string][]Conversation),
	}
}

// Record adds a conversation entry for a file. If ranges is nil or empty,
// the conversation is recorded without specific line attribution.
func (r *Recorder) Record(file, conversationURL string, contributor Contributor, ranges []Range) {
	conv := Conversation{
		URL:    conversationURL,
		Ranges: ranges,
	}
	conv.Contributor = &contributor

	r.mu.Lock()
	r.files[file] = append(r.files[file], conv)
	r.mu.Unlock()
}

// Build produces a complete TraceRecord with a UUID, timestamp, and all
// accumulated file data. The vcs parameter may be nil if VCS info is not
// available.
func (r *Recorder) Build(vcs *Vcs) TraceRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	files := make([]File, 0, len(r.files))
	for path, convs := range r.files {
		files = append(files, File{
			Path:          path,
			Conversations: convs,
		})
	}

	return TraceRecord{
		Version:   Version,
		ID:        newUUID(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Vcs:       vcs,
		Tool:      r.tool,
		Files:     files,
	}
}

// Reset clears all accumulated data so the Recorder can be reused.
func (r *Recorder) Reset() {
	r.mu.Lock()
	r.files = make(map[string][]Conversation)
	r.mu.Unlock()
}

// newUUID generates a v4 UUID string.
func newUUID() string {
	var uuid [16]byte
	_, _ = rand.Read(uuid[:])
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant 1
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}
