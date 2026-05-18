package hookshot

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/CorridorSecurity/hookshot/cascade"
	"github.com/CorridorSecurity/hookshot/claude"
	"github.com/CorridorSecurity/hookshot/codex"
	"github.com/CorridorSecurity/hookshot/cursor"
	"github.com/CorridorSecurity/hookshot/droid"
	"github.com/CorridorSecurity/hookshot/internal"
)

// Platform identifies which AI coding tool triggered the hook.
type Platform string

const (
	PlatformClaude  Platform = "claude"
	PlatformCursor  Platform = "cursor"
	PlatformDroid   Platform = "droid"
	PlatformCascade Platform = "cascade"
	PlatformCodex   Platform = "codex"
)

// =============================================================================
// Unified Stop Handler
// =============================================================================

// StopContext provides a unified view of stop events from both platforms.
type StopContext struct {
	Platform  Platform
	SessionID string // Claude Code: session_id, Cursor: conversation_id
	Cwd       string // Working directory (all platforms)

	// Claude Code-specific fields
	StopHookActive bool // True if already continuing from a previous stop hook

	// Cursor-specific fields
	Status    string // "completed", "aborted", or "error"
	LoopCount int    // Number of previous auto follow-ups (max 5)
}

// ShouldSkip returns true if the stop hook should be skipped to prevent loops.
// For Claude Code, Droid, and Codex, this checks StopHookActive. For Cursor,
// this checks LoopCount >= 3. For Cascade, there is no loop prevention
// mechanism (returns false).
func (c StopContext) ShouldSkip() bool {
	if c.Platform == PlatformClaude || c.Platform == PlatformDroid || c.Platform == PlatformCodex {
		return c.StopHookActive
	}
	if c.Platform == PlatformCursor {
		return c.LoopCount >= 3
	}
	// Cascade has no loop prevention mechanism
	return false
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

// OnStop registers a unified handler for stop events on all platforms.
// It automatically registers handlers for "claude-stop", "cursor-stop",
// "droid-stop", "cascade-post-cascade-response", and "codex-stop".
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
				Cwd:       cursorWorkspaceRoot(input.WorkspaceRoots),
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

	Register("droid-stop", func() {
		Run(func(input droid.StopInput) droid.StopOutput {
			ctx := StopContext{
				Platform:       PlatformDroid,
				SessionID:      input.SessionID,
				Cwd:            input.Cwd,
				StopHookActive: input.StopHookActive,
			}
			decision := handler(ctx)
			if decision.Continue {
				return droid.Continue()
			}
			return droid.Block(decision.Message)
		})
	})

	// Cascade uses post_cascade_response as the closest equivalent to stop hooks
	Register("cascade-post-cascade-response", func() {
		Run(func(input cascade.PostCascadeResponseInput) cascade.PostCascadeResponseOutput {
			ctx := StopContext{
				Platform:  PlatformCascade,
				SessionID: input.TrajectoryID,
			}
			// Cascade post hooks are fire-and-forget, but we still call the handler
			// for side effects (logging, telemetry, etc.)
			handler(ctx)
			return cascade.PostCascadeResponseOK()
		})
	})

	// Codex uses the same JSON wire protocol as Claude Code but with stricter
	// validation. Use codex.* helpers (not claude.*) so any Codex-specific
	// quirks (e.g. rejected suppressOutput / updatedInput) are handled in
	// one place — the codex package.
	Register("codex-stop", func() {
		Run(func(input codex.StopInput) codex.StopOutput {
			ctx := StopContext{
				Platform:       PlatformCodex,
				SessionID:      input.SessionID,
				Cwd:            input.Cwd,
				StopHookActive: input.StopHookActive,
			}
			decision := handler(ctx)
			if decision.Continue {
				return codex.Continue()
			}
			return codex.Block(decision.Message)
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
	// NOTE: Only populated for Cursor and Cascade, not Claude Code or Droid
	Command string
	Cwd     string // Working directory

	// For MCP execution
	ToolName  string          // MCP tool name (e.g., "mcp__server__tool")
	ToolInput json.RawMessage // Tool input parameters as JSON
	ServerURL string          // MCP server URL (Cursor/Cascade only, for URL-based servers)

	// Raw input for advanced use cases
	RawClaudeCode *claude.PreToolUseInput
	RawCursor     any // *cursor.BeforeShellExecutionInput or *cursor.BeforeMCPExecutionInput
	RawDroid      *droid.PreToolUseInput
	RawCascade    any // *cascade.PreRunCommandInput or *cascade.PreMCPToolUseInput
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
//   - "droid-pre-tool-use" (filters to Bash and mcp__* tools)
//   - "cascade-pre-run-command"
//   - "cascade-pre-mcp-tool-use"
//   - "codex-pre-tool-use" (filters to Bash, apply_patch, and mcp__* tools)
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
			cwd := input.Cwd
			if cwd == "" {
				cwd = cursorWorkspaceRoot(input.WorkspaceRoots)
			}
			ctx := ExecutionContext{
				Platform:  PlatformCursor,
				Type:      ExecutionShell,
				Command:   input.Command,
				Cwd:       cwd,
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
				Cwd:       cursorWorkspaceRoot(input.WorkspaceRoots),
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

	// Droid PreToolUse (for Bash and MCP tools)
	// Uses RunE so that blocking decisions exit with code 2, which is how
	// Factory Droid detects that a hook has denied an action.
	Register("droid-pre-tool-use", func() {
		RunE(func(input droid.PreToolUseInput) (droid.PreToolUseOutput, error) {
			// Determine execution type
			// Droid MCP tools use serverName___toolName format (e.g. corridor___listProjects)
			// unlike Claude's mcp__server__tool format
			var execType ExecutionType
			if input.ToolName == "Bash" {
				execType = ExecutionShell
			} else if len(input.ToolName) > 5 && input.ToolName[:5] == "mcp__" {
				execType = ExecutionMCP
			} else if strings.Contains(input.ToolName, "___") {
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
				Platform:  PlatformDroid,
				Type:      execType,
				Command:   command,
				Cwd:       input.Cwd,
				ToolName:  input.ToolName,
				ToolInput: input.ToolInput,
				RawDroid:  &input,
			}

			decision := handler(ctx)
			if decision.Allow {
				if decision.Reason != "" {
					return droid.Allow(decision.Reason), nil
				}
				return droid.AllowSilent(), nil
			}
			// Exit code 2 + stderr message for blocking (per Factory docs)
			return droid.PreToolUseOutput{}, errors.New(decision.Reason)
		})
	})

	// Cascade preRunCommand
	// Uses RunE so that blocking decisions exit with code 2, which is how
	// Windsurf Cascade detects that a hook has denied an action.
	Register("cascade-pre-run-command", func() {
		RunE(func(input cascade.PreRunCommandInput) (cascade.PreRunCommandOutput, error) {
			ctx := ExecutionContext{
				Platform:   PlatformCascade,
				Type:       ExecutionShell,
				Command:    input.ToolInfo.CommandLine,
				Cwd:        input.ToolInfo.Cwd,
				RawCascade: &input,
			}

			decision := handler(ctx)
			if decision.Allow {
				return cascade.AllowCommand(), nil
			}
			return cascade.PreRunCommandOutput{}, errors.New(decision.Reason)
		})
	})

	// Cascade preMCPToolUse
	// Uses RunE so that blocking decisions exit with code 2, which is how
	// Windsurf Cascade detects that a hook has denied an action.
	Register("cascade-pre-mcp-tool-use", func() {
		RunE(func(input cascade.PreMCPToolUseInput) (cascade.PreMCPToolUseOutput, error) {
			ctx := ExecutionContext{
				Platform:   PlatformCascade,
				Type:       ExecutionMCP,
				ToolName:   input.ToolInfo.MCPToolName,
				ToolInput:  input.ToolInfo.MCPToolArguments,
				ServerURL:  input.ToolInfo.MCPServerName,
				RawCascade: &input,
			}

			decision := handler(ctx)
			if decision.Allow {
				return cascade.AllowMCP(), nil
			}
			return cascade.PreMCPToolUseOutput{}, errors.New(decision.Reason)
		})
	})

	// Codex PreToolUse (same JSON wire protocol as Claude Code, stricter
	// validation). Codex tool names include "Bash", "apply_patch", and MCP
	// names like "mcp__server__tool". apply_patch is classified as
	// ExecutionTool because it represents a file edit rather than a shell
	// command or MCP call. For apply_patch the underlying tool_input.command
	// is parsed and exposed via ExecutionContext.Command so policies can
	// inspect the patch text. Uses codex.* helpers so Codex quirks (no
	// suppressOutput, no updatedInput) live in the codex package.
	Register("codex-pre-tool-use", func() {
		Run(func(input codex.PreToolUseInput) codex.PreToolUseOutput {
			var execType ExecutionType
			if input.ToolName == "Bash" {
				execType = ExecutionShell
			} else if len(input.ToolName) > 5 && input.ToolName[:5] == "mcp__" {
				execType = ExecutionMCP
			} else {
				execType = ExecutionTool
			}

			var command string
			if execType == ExecutionShell || input.ToolName == "apply_patch" {
				var cmdInput struct {
					Command string `json:"command"`
				}
				json.Unmarshal(input.ToolInput, &cmdInput)
				command = cmdInput.Command
			}

			ctx := ExecutionContext{
				Platform:      PlatformCodex,
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
					return codex.Allow(decision.Reason)
				}
				// codex.AllowSilent is a Codex-safe no-op (emits {}) — it
				// does NOT set suppressOutput like claude.AllowSilent,
				// because Codex rejects that field with "PreToolUse hook
				// returned unsupported suppressOutput".
				return codex.AllowSilent()
			}
			// Codex currently parses but does not enforce permissionDecision
			// "ask" for PreToolUse, so an Ask decision would silently fail
			// open. Until Codex enforces it, fail closed by denying — this
			// matches the security posture of the other platforms where
			// Ask actually surfaces an approval prompt.
			if decision.Ask {
				return codex.Deny(decision.Reason)
			}
			return codex.Deny(decision.Reason)
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
	SessionID string // Claude Code: session_id, Cursor: conversation_id, Cascade: trajectory_id
	FilePath  string
	// NewFilePath is the destination path when the edit also renames the
	// file (Codex apply_patch "*** Move to:" today; future platforms may
	// surface their own rename semantics). It is empty for in-place edits.
	//
	// For Codex moves, the unified bridge invokes OnAfterFileEdit twice —
	// once with FilePath set to the source and once with FilePath set to
	// the destination — and populates NewFilePath on both invocations so
	// path-based policies can never be bypassed by inspecting only
	// FilePath. Handlers that want to detect a rename should check
	// `ctx.NewFilePath != "" && ctx.NewFilePath != ctx.FilePath`.
	NewFilePath string
	Edits       []FileEdit
	Cwd         string

	// Raw input for advanced use cases
	RawClaudeCode *claude.PostToolUseInput
	RawCursor     *cursor.AfterFileEditInput
	RawDroid      *droid.PostToolUseInput
	RawCascade    *cascade.PostWriteCodeInput
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
//   - "droid-after-file-edit" (PostToolUse for Write/Edit)
//   - "cascade-post-write-code"
//   - "codex-post-tool-use" (PostToolUse for Write/Edit/apply_patch)
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
				Cwd:       cursorWorkspaceRoot(input.WorkspaceRoots),
				RawCursor: &input,
			}

			// Cursor afterFileEdit has no decision control, but we still call the handler
			// for side effects (logging, telemetry, etc.)
			handler(ctx)
			return cursor.AfterFileEditOK()
		})
	})

	// Droid PostToolUse (for Write/Edit)
	Register("droid-after-file-edit", func() {
		Run(func(input droid.PostToolUseInput) droid.PostToolUseOutput {
			// Only handle Write and Edit tools
			if input.ToolName != "Write" && input.ToolName != "Edit" {
				return droid.PostToolOK()
			}

			// Extract file path and edits from tool input
			// Factory Droid uses "old_str"/"new_str" (not "old_string"/"new_string" like Claude Code)
			var toolInput struct {
				FilePath string `json:"file_path"`
				Content  string `json:"content"`
				OldStr   string `json:"old_str"`
				NewStr   string `json:"new_str"`
			}
			json.Unmarshal(input.ToolInput, &toolInput)

			var edits []FileEdit
			if input.ToolName == "Edit" {
				edits = []FileEdit{{OldString: toolInput.OldStr, NewString: toolInput.NewStr}}
			} else {
				edits = []FileEdit{{OldString: "", NewString: toolInput.Content}}
			}

			ctx := FileEditContext{
				Platform:  PlatformDroid,
				SessionID: input.SessionID,
				FilePath:  toolInput.FilePath,
				Edits:     edits,
				Cwd:       input.Cwd,
				RawDroid:  &input,
			}

			decision := handler(ctx)
			if decision.Block {
				return droid.PostToolBlock(decision.Reason)
			}
			if decision.Context != "" {
				return droid.PostToolContext(decision.Context)
			}
			return droid.PostToolOK()
		})
	})

	// Cascade postWriteCode
	Register("cascade-post-write-code", func() {
		Run(func(input cascade.PostWriteCodeInput) cascade.PostWriteCodeOutput {
			var edits []FileEdit
			for _, e := range input.ToolInfo.Edits {
				edits = append(edits, FileEdit{OldString: e.OldString, NewString: e.NewString})
			}

			ctx := FileEditContext{
				Platform:   PlatformCascade,
				SessionID:  input.TrajectoryID,
				FilePath:   input.ToolInfo.FilePath,
				Edits:      edits,
				RawCascade: &input,
			}

			// Cascade post hooks are fire-and-forget, but we still call the handler
			// for side effects (logging, telemetry, etc.)
			handler(ctx)
			return cascade.PostWriteCodeOK()
		})
	})

	// Codex PostToolUse (same JSON protocol as Claude Code).
	// Codex uses apply_patch in addition to Write/Edit; apply_patch carries
	// a unified-diff envelope under tool_input.command that may touch
	// multiple files in a single call. For each file in the patch we invoke
	// the user's handler exactly once with a populated FileEditContext, and
	// we combine the decisions across files: a Block from any file wins,
	// otherwise context strings are concatenated. The parser used here is
	// also exported as codex.ParseApplyPatch for callers that want raw
	// access. Configure the hook with matcher "apply_patch|mcp__.*" in
	// hooks.json — "Edit" and "Write" matcher aliases exist but are
	// redundant with "apply_patch".
	Register("codex-post-tool-use", func() {
		Run(func(input codex.PostToolUseInput) codex.PostToolUseOutput {
			if input.ToolName != "Write" && input.ToolName != "Edit" && input.ToolName != "apply_patch" {
				return codex.PostToolOK()
			}

			// Write/Edit use the Claude-style schema.
			if input.ToolName == "Write" || input.ToolName == "Edit" {
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
					Platform:      PlatformCodex,
					SessionID:     input.SessionID,
					FilePath:      toolInput.FilePath,
					Edits:         edits,
					Cwd:           input.Cwd,
					RawClaudeCode: &input,
				}

				decision := handler(ctx)
				if decision.Block {
					return codex.PostToolBlock(decision.Reason)
				}
				if decision.Context != "" {
					return codex.PostToolContext(decision.Context)
				}
				return codex.PostToolOK()
			}

			// apply_patch: tool_input is {"command": "*** Begin Patch ..."}.
			var applyInput struct {
				Command string `json:"command"`
			}
			json.Unmarshal(input.ToolInput, &applyInput)

			files := codex.ParseApplyPatch(applyInput.Command)
			if len(files) == 0 {
				// We could not parse anything actionable out of the patch.
				// Fall back to invoking the handler once with whatever raw
				// information we have so policies still see a Codex
				// PostToolUse event rather than nothing.
				ctx := FileEditContext{
					Platform:      PlatformCodex,
					SessionID:     input.SessionID,
					Cwd:           input.Cwd,
					Edits:         []FileEdit{{OldString: "", NewString: applyInput.Command}},
					RawClaudeCode: &input,
				}
				decision := handler(ctx)
				if decision.Block {
					return codex.PostToolBlock(decision.Reason)
				}
				if decision.Context != "" {
					return codex.PostToolContext(decision.Context)
				}
				return codex.PostToolOK()
			}

			var (
				blockReasons []string
				contexts     []string
			)
			invoke := func(filePath string, f codex.PatchFile) {
				edits := make([]FileEdit, 0, len(f.Edits))
				for _, e := range f.Edits {
					edits = append(edits, FileEdit{OldString: e.OldString, NewString: e.NewString})
				}
				ctx := FileEditContext{
					Platform:      PlatformCodex,
					SessionID:     input.SessionID,
					FilePath:      filePath,
					NewFilePath:   f.NewFilePath,
					Edits:         edits,
					Cwd:           input.Cwd,
					RawClaudeCode: &input,
				}
				decision := handler(ctx)
				if decision.Block {
					blockReasons = append(blockReasons, decision.Reason)
				} else if decision.Context != "" {
					contexts = append(contexts, decision.Context)
				}
			}
			for _, f := range files {
				// Always invoke for the declared source path.
				invoke(f.FilePath, f)
				// For renames, invoke again with the destination so policies
				// that only inspect ctx.FilePath cannot be bypassed by a
				// "*** Move to:" pointing at a sensitive path (e.g.
				// "../../.ssh/authorized_keys"). NewFilePath is populated on
				// both invocations so policies that want to detect the
				// rename relationship can.
				if f.NewFilePath != "" && f.NewFilePath != f.FilePath {
					invoke(f.NewFilePath, f)
				}
			}
			if len(blockReasons) > 0 {
				return codex.PostToolBlock(strings.Join(blockReasons, "\n"))
			}
			if len(contexts) > 0 {
				return codex.PostToolContext(strings.Join(contexts, "\n"))
			}
			return codex.PostToolOK()
		})
	})
}

// =============================================================================
// Unified Prompt Submit Handler
// =============================================================================

// PromptContext provides a unified view of prompt submission events.
type PromptContext struct {
	Platform  Platform
	SessionID string // Claude Code: session_id, Cursor: conversation_id, Cascade: trajectory_id
	Prompt    string
	Cwd       string // Working directory (Cursor: from workspace_roots[0])

	// Raw input for advanced use cases
	RawClaudeCode *claude.UserPromptSubmitInput
	RawCursor     *cursor.BeforeSubmitPromptInput
	RawDroid      *droid.UserPromptSubmitInput
	RawCascade    *cascade.PreUserPromptInput
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
//   - "droid-user-prompt-submit"
//   - "cascade-pre-user-prompt"
//   - "codex-user-prompt-submit"
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
				Cwd:       cursorWorkspaceRoot(input.WorkspaceRoots),
				RawCursor: &input,
			}

			decision := handler(ctx)
			if !decision.Allow {
				return cursor.BlockPrompt(decision.Reason)
			}
			return cursor.AllowPrompt()
		})
	})

	// Droid UserPromptSubmit
	Register("droid-user-prompt-submit", func() {
		Run(func(input droid.UserPromptSubmitInput) droid.UserPromptSubmitOutput {
			ctx := PromptContext{
				Platform:  PlatformDroid,
				SessionID: input.SessionID,
				Prompt:    input.Prompt,
				RawDroid:  &input,
			}

			decision := handler(ctx)
			if !decision.Allow {
				return droid.BlockPrompt(decision.Reason)
			}
			if decision.Context != "" {
				return droid.AddContext(decision.Context)
			}
			return droid.AllowPrompt()
		})
	})

	// Cascade preUserPrompt
	// Uses RunE so that blocking decisions exit with code 2, which is how
	// Windsurf Cascade detects that a hook has denied an action.
	Register("cascade-pre-user-prompt", func() {
		RunE(func(input cascade.PreUserPromptInput) (cascade.PreUserPromptOutput, error) {
			ctx := PromptContext{
				Platform:   PlatformCascade,
				SessionID:  input.TrajectoryID,
				Prompt:     input.ToolInfo.UserPrompt,
				RawCascade: &input,
			}

			decision := handler(ctx)
			if !decision.Allow {
				return cascade.PreUserPromptOutput{}, errors.New(decision.Reason)
			}
			return cascade.AllowPrompt(), nil
		})
	})

	// Codex UserPromptSubmit (same JSON wire protocol as Claude Code,
	// stricter validation — use codex.* helpers).
	Register("codex-user-prompt-submit", func() {
		Run(func(input codex.UserPromptSubmitInput) codex.UserPromptSubmitOutput {
			ctx := PromptContext{
				Platform:      PlatformCodex,
				SessionID:     input.SessionID,
				Prompt:        input.Prompt,
				Cwd:           input.Cwd,
				RawClaudeCode: &input,
			}

			decision := handler(ctx)
			if !decision.Allow {
				return codex.BlockPrompt(decision.Reason)
			}
			if decision.Context != "" {
				return codex.AddContext(decision.Context)
			}
			return codex.AllowPrompt()
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

// cursorWorkspaceRoot returns the first workspace root from a Cursor hook
// input's workspace_roots array. Cursor launches hooks from its config
// directory (~/.cursor), so os.Getwd() returns a misleading path. The
// workspace_roots field always contains the actual project directories.
func cursorWorkspaceRoot(roots []string) string {
	if len(roots) > 0 {
		return roots[0]
	}
	return ""
}
