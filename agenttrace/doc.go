// Package agenttrace provides types and utilities for generating Agent Trace
// records, an open specification for tracing AI-generated code.
//
// Agent Trace (https://github.com/cursor/agent-trace) defines a vendor-neutral
// JSON format for recording which parts of code were written by AI vs humans,
// linking to agent conversations, and tracking line-level attribution.
//
// # Types
//
// The package mirrors the Agent Trace v0.1.0 JSON schema:
//
//   - [TraceRecord]: Top-level record with version, id, timestamp, files
//   - [File]: A file path with associated conversations
//   - [Conversation]: Links to a conversation with ranges and contributor info
//   - [Range]: Line range within a file attributed to a conversation
//   - [Contributor]: Attribution type (human, ai, mixed, unknown) with optional model
//   - [Tool]: The AI tool that produced the code (name and version)
//   - [Vcs]: Version control info (type and revision)
//
// # Recording Traces
//
// Use [Recorder] to accumulate file edits and produce a [TraceRecord]:
//
//	rec := agenttrace.NewRecorder(&agenttrace.Tool{Name: "claude-code", Version: "1.0"})
//
//	rec.Record("src/main.go", "https://example.com/conversation/123",
//	    agenttrace.Contributor{Type: agenttrace.ContributorAI, ModelID: "anthropic/claude-sonnet-4-5-20250929"},
//	    []agenttrace.Range{{StartLine: 10, EndLine: 25}},
//	)
//
//	trace := rec.Build(&agenttrace.Vcs{Type: agenttrace.VcsGit, Revision: "abc123"})
//
// # Storing Traces
//
// Use [WriteTrace] and [ReadTrace] for JSON file I/O:
//
//	err := agenttrace.WriteTrace("traces/trace.json", trace)
//	loaded, err := agenttrace.ReadTrace("traces/trace.json")
//
// # Integration with Hookshot
//
// Use the Recorder from hookshot's [OnAfterFileEdit] handler:
//
//	rec := agenttrace.NewRecorder(&agenttrace.Tool{Name: "claude-code"})
//
//	hookshot.OnAfterFileEdit(func(ctx hookshot.FileEditContext) hookshot.FileEditDecision {
//	    rec.Record(ctx.FilePath, "", agenttrace.Contributor{Type: agenttrace.ContributorAI}, nil)
//	    return hookshot.FileEditOK()
//	})
//
// See https://github.com/cursor/agent-trace for the full specification.
package agenttrace
