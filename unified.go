package hookshot

import (
	"encoding/json"

	"github.com/CorridorSecurity/hookshot/claude"
	"github.com/CorridorSecurity/hookshot/cursor"
	"github.com/CorridorSecurity/hookshot/internal"
)

// Platform identifies which AI coding tool triggered the hook.
type Platform string

const (
	PlatformClaude Platform = "claude"
	PlatformCursor Platform = "cursor"
)

// =============================================================================
// Unified Stop Handler
// =============================================================================

// StopContext provides a unified view of stop events from both platforms.
type StopContext struct {
	Platform  Platform
	SessionID string // Claude Code: session_id, Cursor: conversation_id
	Cwd       string // Working directory (Claude Code only, empty for Cursor)

	// Claude Code-specific fields
	StopHookActive bool // True if already continuing from a previous stop hook

	// Cursor-specific fields
	Status    string // "completed", "aborted", or "error"
	LoopCount int    // Number of previous auto follow-ups (max 5)
}

// ShouldSkip returns true if the stop hook should be skipped to prevent loops.
// For Claude Code, this checks StopHookActive. For Cursor, this checks LoopCount >= 3.
func (c StopContext) ShouldSkip() bool {
	if c.Platform == PlatformClaude {
		return c.StopHookActive
	}
	return c.LoopCount >= 3
}

// StopDecision represents the unified decision for stop hooks.
type StopDecision struct {
	// Continue determines whether the agent should stop.
	// true = allow stopping, false = prevent stopping (continue working)
	Continue bool

	// Message is shown to the agent when Continue is false.
	// For Claude Code, this becomes the "reason" field.
	// For Cursor, this becomes the "followup_message" field.
	Message string
}

// AllowStop returns a decision that allows the agent to stop.
func AllowStop() StopDecision {
	return StopDecision{Continue: true}
}

// PreventStop returns a decision that prevents stopping with a message.
func PreventStop(message string) StopDecision {
	return StopDecision{Continue: false, Message: message}
}

// StopHandler is the function signature for unified stop handlers.
type StopHandler func(StopContext) StopDecision

// OnStop registers a unified handler for stop events on both platforms.
// It automatically registers handlers for "claude-stop" and "cursor-stop".
func OnStop(handler StopHandler) {
	Register("claude-stop", func() {
		Run(func(input claude.StopInput) claude.StopOutput {
			ctx := StopContext{
				Platform:       PlatformClaude,
				SessionID:      input.SessionID,
				Cwd:            input.Cwd,
				StopHookActive: input.StopHookActive,
			}
			decision := handler(ctx)
			if decision.Continue {
				return claude.Continue()
			}
			return claude.Block(decision.Message)
		})
	})

	Register("cursor-stop", func() {
		Run(func(input cursor.StopInput) cursor.StopOutput {
			ctx := StopContext{
				Platform:  PlatformCursor,
				SessionID: input.ConversationID,
				Status:    input.Status,
				LoopCount: input.LoopCount,
			}
			decision := handler(ctx)
			if decision.Continue {
				return cursor.Continue()
			}
			return cursor.Followup(decision.Message)
		})
	})
}

// =============================================================================
// Unified Before Execution Handler
// =============================================================================

// ExecutionType identifies what kind of execution is being attempted.
type ExecutionType string

const (
	ExecutionShell ExecutionType = "shell"
	ExecutionMCP   ExecutionType = "mcp"
	ExecutionTool  ExecutionType = "tool" // Claude Code non-MCP tools (Read, Write, etc.)
)

// ExecutionContext provides a unified view of pre-execution events.
type ExecutionContext struct {
	Platform Platform
	Type     ExecutionType

	// For shell execution (Cursor beforeShellExecution, Claude Code Bash tool)
	// Also used for local MCP servers on Cursor (command-based MCP servers)
	// NOTE: Only populated for Cursor, not Claude Code
	Command string
	Cwd     string // Working directory

	// For MCP execution
	ToolName  string          // MCP tool name (e.g., "mcp__server__tool")
	ToolInput json.RawMessage // Tool input parameters as JSON
	ServerURL string          // MCP server URL (Cursor only, for URL-based servers)

	// Raw input for advanced use cases
	RawClaudeCode *claude.PreToolUseInput
	RawCursor     any // *cursor.BeforeShellExecutionInput or *cursor.BeforeMCPExecutionInput
}

// IsMCP returns true if this is an MCP tool execution.
func (c ExecutionContext) IsMCP() bool {
	return c.Type == ExecutionMCP
}

// ExecutionDecision represents the unified decision for execution hooks.
type ExecutionDecision struct {
	// Allow determines whether execution should proceed.
	Allow bool

	// Reason explains the decision.
	// For Allow=true, shown to user (if not empty).
	// For Allow=false, shown to agent.
	Reason string

	// Ask prompts the user to confirm (only if Allow is false and Ask is true).
	Ask bool
}

// AllowExecution returns a decision that permits execution.
func AllowExecution() ExecutionDecision {
	return ExecutionDecision{Allow: true}
}

// AllowExecutionWithReason permits execution with a reason shown to the user.
func AllowExecutionWithReason(reason string) ExecutionDecision {
	return ExecutionDecision{Allow: true, Reason: reason}
}

// DenyExecution blocks execution with a reason shown to the agent.
func DenyExecution(reason string) ExecutionDecision {
	return ExecutionDecision{Allow: false, Reason: reason}
}

// AskExecution prompts the user to confirm execution.
func AskExecution(reason string) ExecutionDecision {
	return ExecutionDecision{Allow: false, Ask: true, Reason: reason}
}

// ExecutionHandler is the function signature for unified execution handlers.
type ExecutionHandler func(ExecutionContext) ExecutionDecision

// OnBeforeExecution registers a unified handler for pre-execution events.
// It automatically registers handlers for:
//   - "claude-pre-tool-use" (filters to Bash and mcp__* tools)
//   - "cursor-before-shell"
//   - "cursor-before-mcp"
func OnBeforeExecution(handler ExecutionHandler) {
	// Claude Code PreToolUse (for Bash and MCP tools)
	Register("claude-pre-tool-use", func() {
		Run(func(input claude.PreToolUseInput) claude.PreToolUseOutput {
			// Determine execution type
			var execType ExecutionType
			if input.ToolName == "Bash" {
				execType = ExecutionShell
			} else if len(input.ToolName) > 5 && input.ToolName[:5] == "mcp__" {
				execType = ExecutionMCP
			} else {
				execType = ExecutionTool
			}

			// Extract command for Bash tool
			var command string
			if execType == ExecutionShell {
				var bashInput struct {
					Command string `json:"command"`
				}
				json.Unmarshal(input.ToolInput, &bashInput)
				command = bashInput.Command
			}

			ctx := ExecutionContext{
				Platform:      PlatformClaude,
				Type:          execType,
				Command:       command,
				Cwd:           input.Cwd,
				ToolName:      input.ToolName,
				ToolInput:     input.ToolInput,
				RawClaudeCode: &input,
			}

			decision := handler(ctx)
			if decision.Allow {
				if decision.Reason != "" {
					return claude.Allow(decision.Reason)
				}
				return claude.AllowSilent()
			}
			if decision.Ask {
				return claude.Ask(decision.Reason)
			}
			return claude.Deny(decision.Reason)
		})
	})

	// Cursor beforeShellExecution
	Register("cursor-before-shell", func() {
		Run(func(input cursor.BeforeShellExecutionInput) cursor.BeforeExecutionOutput {
			ctx := ExecutionContext{
				Platform:  PlatformCursor,
				Type:      ExecutionShell,
				Command:   input.Command,
				Cwd:       input.Cwd,
				RawCursor: &input,
			}

			decision := handler(ctx)
			if decision.Allow {
				if decision.Reason != "" {
					return cursor.AllowWithMessage(decision.Reason)
				}
				return cursor.Allow()
			}
			if decision.Ask {
				return cursor.Ask(decision.Reason)
			}
			return cursor.Deny(decision.Reason, decision.Reason)
		})
	})

	// Cursor beforeMCPExecution
	Register("cursor-before-mcp", func() {
		Run(func(input cursor.BeforeMCPExecutionInput) cursor.BeforeExecutionOutput {
			ctx := ExecutionContext{
				Platform:  PlatformCursor,
				Type:      ExecutionMCP,
				ToolName:  input.ToolName,
				ToolInput: json.RawMessage(input.ToolInput),
				ServerURL: input.URL,
				Command:   input.Command, // For local MCP servers (command-based)
				RawCursor: &input,
			}

			decision := handler(ctx)
			if decision.Allow {
				if decision.Reason != "" {
					return cursor.AllowWithMessage(decision.Reason)
				}
				return cursor.Allow()
			}
			if decision.Ask {
				return cursor.Ask(decision.Reason)
			}
			return cursor.Deny(decision.Reason, decision.Reason)
		})
	})
}

// =============================================================================
// Unified After File Edit Handler
// =============================================================================

// FileEdit represents a single edit operation.
type FileEdit struct {
	OldString string
	NewString string
}

// FileEditContext provides a unified view of file edit events.
type FileEditContext struct {
	Platform  Platform
	SessionID string // Claude Code: session_id, Cursor: conversation_id
	FilePath  string
	Edits     []FileEdit
	Cwd       string

	// Raw input for advanced use cases
	RawClaudeCode *claude.PostToolUseInput
	RawCursor     *cursor.AfterFileEditInput
}

// FileEditDecision represents the unified decision for file edit hooks.
type FileEditDecision struct {
	// Block sends feedback to the agent (Claude Code only).
	Block bool

	// Reason is shown to the agent when Block is true.
	Reason string

	// Context is additional context added for the agent (Claude Code only).
	Context string
}

// FileEditOK returns a decision that allows normal flow.
func FileEditOK() FileEditDecision {
	return FileEditDecision{}
}

// FileEditBlock sends feedback to the agent about the edit.
func FileEditBlock(reason string) FileEditDecision {
	return FileEditDecision{Block: true, Reason: reason}
}

// FileEditAddContext adds context for the agent to consider.
func FileEditAddContext(context string) FileEditDecision {
	return FileEditDecision{Context: context}
}

// FileEditHandler is the function signature for unified file edit handlers.
type FileEditHandler func(FileEditContext) FileEditDecision

// OnAfterFileEdit registers a unified handler for post-file-edit events.
// It automatically registers handlers for:
//   - "claude-after-file-edit" (PostToolUse for Write/Edit)
//   - "cursor-after-file-edit"
func OnAfterFileEdit(handler FileEditHandler) {
	// Claude Code PostToolUse (for Write/Edit)
	Register("claude-after-file-edit", func() {
		Run(func(input claude.PostToolUseInput) claude.PostToolUseOutput {
			// Only handle Write and Edit tools
			if input.ToolName != "Write" && input.ToolName != "Edit" {
				return claude.PostToolOK()
			}

			// Extract file path and edits from tool input
			var toolInput struct {
				FilePath  string `json:"file_path"`
				Content   string `json:"content"`
				OldString string `json:"old_string"`
				NewString string `json:"new_string"`
			}
			json.Unmarshal(input.ToolInput, &toolInput)

			var edits []FileEdit
			if input.ToolName == "Edit" {
				edits = []FileEdit{{OldString: toolInput.OldString, NewString: toolInput.NewString}}
			} else {
				edits = []FileEdit{{OldString: "", NewString: toolInput.Content}}
			}

			ctx := FileEditContext{
				Platform:      PlatformClaude,
				SessionID:     input.SessionID,
				FilePath:      toolInput.FilePath,
				Edits:         edits,
				Cwd:           input.Cwd,
				RawClaudeCode: &input,
			}

			decision := handler(ctx)
			if decision.Block {
				return claude.PostToolBlock(decision.Reason)
			}
			if decision.Context != "" {
				return claude.PostToolContext(decision.Context)
			}
			return claude.PostToolOK()
		})
	})

	// Cursor afterFileEdit
	Register("cursor-after-file-edit", func() {
		Run(func(input cursor.AfterFileEditInput) cursor.AfterFileEditOutput {
			var edits []FileEdit
			for _, e := range input.Edits {
				edits = append(edits, FileEdit{OldString: e.OldString, NewString: e.NewString})
			}

			ctx := FileEditContext{
				Platform:  PlatformCursor,
				SessionID: input.ConversationID,
				FilePath:  input.FilePath,
				Edits:     edits,
				RawCursor: &input,
			}

			// Cursor afterFileEdit has no decision control, but we still call the handler
			// for side effects (logging, telemetry, etc.)
			handler(ctx)
			return cursor.AfterFileEditOK()
		})
	})
}

// =============================================================================
// Unified Prompt Submit Handler
// =============================================================================

// PromptContext provides a unified view of prompt submission events.
type PromptContext struct {
	Platform  Platform
	SessionID string // Claude Code: session_id, Cursor: conversation_id
	Prompt    string

	// Raw input for advanced use cases
	RawClaudeCode *claude.UserPromptSubmitInput
	RawCursor     *cursor.BeforeSubmitPromptInput
}

// PromptDecision represents the unified decision for prompt hooks.
type PromptDecision struct {
	// Allow determines whether the prompt should be processed.
	Allow bool

	// Reason is shown to the user when Allow is false.
	Reason string

	// Context is additional context added for the agent (Claude Code only).
	Context string
}

// AllowPromptDecision returns a decision that allows the prompt.
func AllowPromptDecision() PromptDecision {
	return PromptDecision{Allow: true}
}

// BlockPromptDecision blocks the prompt with a reason.
func BlockPromptDecision(reason string) PromptDecision {
	return PromptDecision{Allow: false, Reason: reason}
}

// AddPromptContext allows the prompt and adds context.
func AddPromptContext(context string) PromptDecision {
	return PromptDecision{Allow: true, Context: context}
}

// PromptHandler is the function signature for unified prompt handlers.
type PromptHandler func(PromptContext) PromptDecision

// OnPromptSubmit registers a unified handler for prompt submission events.
// It automatically registers handlers for:
//   - "claude-user-prompt-submit"
//   - "cursor-before-submit-prompt"
func OnPromptSubmit(handler PromptHandler) {
	// Claude Code UserPromptSubmit
	Register("claude-user-prompt-submit", func() {
		Run(func(input claude.UserPromptSubmitInput) claude.UserPromptSubmitOutput {
			ctx := PromptContext{
				Platform:      PlatformClaude,
				SessionID:     input.SessionID,
				Prompt:        input.Prompt,
				RawClaudeCode: &input,
			}

			decision := handler(ctx)
			if !decision.Allow {
				return claude.BlockPrompt(decision.Reason)
			}
			if decision.Context != "" {
				return claude.AddContext(decision.Context)
			}
			return claude.AllowPrompt()
		})
	})

	// Cursor beforeSubmitPrompt
	Register("cursor-before-submit-prompt", func() {
		Run(func(input cursor.BeforeSubmitPromptInput) cursor.BeforeSubmitPromptOutput {
			ctx := PromptContext{
				Platform:  PlatformCursor,
				SessionID: input.ConversationID,
				Prompt:    input.Prompt,
				RawCursor: &input,
			}

			decision := handler(ctx)
			if !decision.Allow {
				return cursor.BlockPrompt(decision.Reason)
			}
			return cursor.AllowPrompt()
		})
	})
}

// =============================================================================
// Raw Input Access Helper
// =============================================================================

// ReadRawInput reads the raw JSON input from stdin into the provided struct.
// Use this when you need access to platform-specific fields not exposed
// in the unified contexts.
func ReadRawInput(v any) error {
	return internal.ReadJSON(v)
}
