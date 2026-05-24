// Package cursor provides types and helpers for Cursor hooks.
//
// Cursor hooks are configured in hooks.json files and execute at various points
// in the agent loop. See https://cursor.com/docs/agent/hooks
package cursor

import "encoding/json"

// =============================================================================
// Common Types
// =============================================================================

// BaseInput contains fields present in all Cursor hook inputs.
type BaseInput struct {
	ConversationID string   `json:"conversation_id"`
	GenerationID   string   `json:"generation_id"`
	Model          string   `json:"model"`
	HookEventName  string   `json:"hook_event_name"`
	CursorVersion  string   `json:"cursor_version"`
	WorkspaceRoots []string `json:"workspace_roots"`
	UserEmail      string   `json:"user_email"`
}

// =============================================================================
// Before Shell/MCP Execution
// =============================================================================

// BeforeShellExecutionInput is received before a shell command executes.
type BeforeShellExecutionInput struct {
	BaseInput
	Command string `json:"command"`
	Cwd     string `json:"cwd"`
}

// BeforeMCPExecutionInput is received before an MCP tool executes.
type BeforeMCPExecutionInput struct {
	BaseInput
	ToolName  string `json:"tool_name"`
	ToolInput string `json:"tool_input"` // JSON string of params

	// Server identification (one of these will be set)
	URL     string `json:"url,omitempty"`     // For URL-based MCP servers
	Command string `json:"command,omitempty"` // For command-based MCP servers
}

// BeforeExecutionOutput controls whether shell/MCP execution should proceed.
type BeforeExecutionOutput struct {
	Permission   string `json:"permission"`              // "allow", "deny", "ask"
	UserMessage  string `json:"user_message,omitempty"`  // Shown to user in client
	AgentMessage string `json:"agent_message,omitempty"` // Sent to agent
}

// =============================================================================
// After Shell/MCP Execution
// =============================================================================

// AfterShellExecutionInput is received after a shell command executes.
type AfterShellExecutionInput struct {
	BaseInput
	Command  string `json:"command"`
	Output   string `json:"output"`
	Duration int64  `json:"duration"` // Milliseconds
}

// AfterShellExecutionOutput has no decision control.
type AfterShellExecutionOutput struct{}

// AfterMCPExecutionInput is received after an MCP tool executes.
type AfterMCPExecutionInput struct {
	BaseInput
	ToolName   string `json:"tool_name"`
	ToolInput  string `json:"tool_input"`
	ResultJSON string `json:"result_json"`
	Duration   int64  `json:"duration"` // Milliseconds
}

// AfterMCPExecutionOutput has no decision control.
type AfterMCPExecutionOutput struct{}

// =============================================================================
// File Operations (Agent)
// =============================================================================

// BeforeReadFileInput is received before the agent reads a file.
type BeforeReadFileInput struct {
	BaseInput
	FilePath    string       `json:"file_path"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a file or rule attachment in prompts.
type Attachment struct {
	Type     string `json:"type"` // "file" or "rule"
	FilePath string `json:"filePath"`
}

// BeforeReadFileOutput controls whether file reading should proceed.
type BeforeReadFileOutput struct {
	Permission   string `json:"permission"`              // "allow", "deny", "ask"
	UserMessage  string `json:"user_message,omitempty"`  // Shown to user
	AgentMessage string `json:"agent_message,omitempty"` // Sent to agent
}

// AfterFileEditInput is received after the agent edits a file.
type AfterFileEditInput struct {
	BaseInput
	FilePath string `json:"file_path"`
	Edits    []Edit `json:"edits"`
}

// Edit represents a single edit operation.
type Edit struct {
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

// AfterFileEditOutput has no decision control.
type AfterFileEditOutput struct{}

// =============================================================================
// Tab Operations (Inline Completions)
// =============================================================================

// BeforeTabFileReadInput is received before Tab reads a file for completions.
type BeforeTabFileReadInput struct {
	BaseInput
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// BeforeTabFileReadOutput controls Tab's file access.
type BeforeTabFileReadOutput struct {
	Permission string `json:"permission"` // "allow" or "deny"
}

// AfterTabFileEditInput is received after Tab edits a file.
type AfterTabFileEditInput struct {
	BaseInput
	FilePath string    `json:"file_path"`
	Edits    []TabEdit `json:"edits"`
}

// TabEdit represents a Tab edit with detailed position information.
type TabEdit struct {
	OldString string    `json:"old_string"`
	NewString string    `json:"new_string"`
	Range     EditRange `json:"range"`
	OldLine   string    `json:"old_line"`
	NewLine   string    `json:"new_line"`
}

// EditRange specifies the location of an edit.
type EditRange struct {
	StartLineNumber int `json:"start_line_number"`
	StartColumn     int `json:"start_column"`
	EndLineNumber   int `json:"end_line_number"`
	EndColumn       int `json:"end_column"`
}

// AfterTabFileEditOutput has no decision control.
type AfterTabFileEditOutput struct{}

// =============================================================================
// Prompt Submission
// =============================================================================

// BeforeSubmitPromptInput is received when the user submits a prompt.
type BeforeSubmitPromptInput struct {
	BaseInput
	Prompt      string       `json:"prompt"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// BeforeSubmitPromptOutput controls whether the prompt should be submitted.
type BeforeSubmitPromptOutput struct {
	Continue    bool   `json:"continue"`
	UserMessage string `json:"user_message,omitempty"` // Shown when blocked
}

// =============================================================================
// SessionStart
// =============================================================================

// SessionStartInput is received when a Cursor agent session starts.
// See https://cursor.com/docs/agent/hooks (sessionStart).
type SessionStartInput struct {
	BaseInput
}

// SessionStartOutput can inject context at session start.
type SessionStartOutput struct {
	AdditionalContext string `json:"additional_context,omitempty"`
}

// =============================================================================
// Stop
// =============================================================================

// StopInput is received when the agent loop ends.
type StopInput struct {
	BaseInput
	Status    string `json:"status"`     // "completed", "aborted", "error"
	LoopCount int    `json:"loop_count"` // Number of previous auto follow-ups
}

// StopOutput can trigger an automatic follow-up message.
type StopOutput struct {
	FollowupMessage string `json:"followup_message,omitempty"`
}

// =============================================================================
// Agent Response/Thought
// =============================================================================

// AfterAgentResponseInput is received after the agent completes a message.
type AfterAgentResponseInput struct {
	BaseInput
	Text string `json:"text"`
}

// AfterAgentResponseOutput has no decision control.
type AfterAgentResponseOutput struct{}

// AfterAgentThoughtInput is received after the agent completes a thinking block.
type AfterAgentThoughtInput struct {
	BaseInput
	Text       string `json:"text"`
	DurationMs int64  `json:"duration_ms,omitempty"`
}

// AfterAgentThoughtOutput has no decision control.
type AfterAgentThoughtOutput struct{}

// =============================================================================
// Generic JSON Input (for custom parsing)
// =============================================================================

// RawInput can be used when you need to access the raw JSON input.
type RawInput struct {
	BaseInput
	Raw json.RawMessage `json:"-"` // Set by custom unmarshaling if needed
}
