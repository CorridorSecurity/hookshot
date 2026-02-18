// Package claude provides types and helpers for Claude Code hooks.
//
// Claude Code hooks are configured in settings files and execute shell commands
// at various points in the agent loop. See https://docs.claude.com/en/docs/claude-code/hooks
package claude

import "encoding/json"

// =============================================================================
// Common Types
// =============================================================================

// BaseInput contains fields present in all Claude Code hook inputs.
type BaseInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"` // "default", "plan", "acceptEdits", "dontAsk", "bypassPermissions"
	HookEventName  string `json:"hook_event_name"`
}

// BaseOutput contains common fields that can be included in any hook output.
type BaseOutput struct {
	// Continue controls whether Claude should continue after hook execution.
	// Default is true. Set to false to stop Claude entirely.
	Continue *bool `json:"continue,omitempty"`

	// StopReason is shown to the user when Continue is false.
	StopReason string `json:"stopReason,omitempty"`

	// SuppressOutput hides stdout from transcript mode when true.
	SuppressOutput bool `json:"suppressOutput,omitempty"`

	// SystemMessage is an optional warning message shown to the user.
	SystemMessage string `json:"systemMessage,omitempty"`
}

// =============================================================================
// Stop / SubagentStop
// =============================================================================

// StopInput is received when the main Claude Code agent finishes responding.
type StopInput struct {
	BaseInput
	// StopHookActive is true when Claude is already continuing as a result of a stop hook.
	// Check this to prevent infinite loops.
	StopHookActive bool `json:"stop_hook_active"`
}

// StopOutput controls whether Claude should stop or continue.
type StopOutput struct {
	BaseOutput
	// Decision should be "block" to prevent stopping, or omitted to allow stopping.
	Decision string `json:"decision,omitempty"`
	// Reason is shown to Claude when Decision is "block".
	Reason string `json:"reason,omitempty"`
}

// SubagentStopInput is received when a Claude Code subagent (Task tool) finishes.
type SubagentStopInput struct {
	BaseInput
	// StopHookActive is true when the subagent is already continuing as a result of a stop hook.
	StopHookActive bool `json:"stop_hook_active"`
	// AgentID is the unique identifier for the subagent.
	AgentID string `json:"agent_id"`
	// AgentType is the agent type name (e.g. "Bash", "Explore", "Plan", or custom agent names).
	// This is the value used for matcher filtering.
	AgentType string `json:"agent_type"`
	// AgentTranscriptPath is the path to the subagent's own transcript file.
	AgentTranscriptPath string `json:"agent_transcript_path,omitempty"`
}

// SubagentStopOutput controls whether a subagent should stop or continue.
// Uses the same decision format as StopOutput.
type SubagentStopOutput = StopOutput

// =============================================================================
// SessionStart
// =============================================================================

// SessionStartInput is received when Claude Code starts or resumes a session.
type SessionStartInput struct {
	BaseInput
	// Source indicates how the session started: "startup", "resume", "clear", "compact"
	Source string `json:"source"`
	// Model is the model identifier (e.g. "claude-sonnet-4-6").
	Model string `json:"model"`
	// AgentType is the agent name when started via `claude --agent <name>`. May be empty.
	AgentType string `json:"agent_type,omitempty"`
}

// SessionStartOutput can add context to the session.
type SessionStartOutput struct {
	BaseOutput
	HookSpecificOutput *SessionStartHookOutput `json:"hookSpecificOutput,omitempty"`
}

// SessionStartHookOutput contains session-start-specific output fields.
type SessionStartHookOutput struct {
	HookEventName     string `json:"hookEventName,omitempty"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

// =============================================================================
// SessionEnd
// =============================================================================

// SessionEndInput is received when a Claude Code session ends.
type SessionEndInput struct {
	BaseInput
	// Reason indicates why the session ended: "clear", "logout", "prompt_input_exit", "bypass_permissions_disabled", "other"
	Reason string `json:"reason"`
}

// SessionEndOutput has no decision control - hooks just run for cleanup.
type SessionEndOutput struct {
	BaseOutput
}

// =============================================================================
// PreToolUse
// =============================================================================

// PreToolUseInput is received before Claude executes a tool.
type PreToolUseInput struct {
	BaseInput
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"` // Tool-specific JSON
	ToolUseID string          `json:"tool_use_id"`
}

// PreToolUseOutput controls whether the tool should execute.
type PreToolUseOutput struct {
	BaseOutput
	HookSpecificOutput *PreToolUseHookOutput `json:"hookSpecificOutput,omitempty"`
}

// PreToolUseHookOutput contains pre-tool-use-specific output fields.
type PreToolUseHookOutput struct {
	HookEventName            string         `json:"hookEventName,omitempty"`
	PermissionDecision       string         `json:"permissionDecision,omitempty"`       // "allow", "deny", "ask"
	PermissionDecisionReason string         `json:"permissionDecisionReason,omitempty"` // Shown to user (allow/ask) or Claude (deny)
	UpdatedInput             map[string]any `json:"updatedInput,omitempty"`             // Modified tool input
	AdditionalContext        string         `json:"additionalContext,omitempty"`        // Added to Claude's context before tool executes
}

// =============================================================================
// PermissionRequest
// =============================================================================

// PermissionSuggestion represents an "always allow" option from the permission dialog.
type PermissionSuggestion struct {
	Type string `json:"type"`
	Tool string `json:"tool,omitempty"`
}

// PermissionRequestInput is received when the user is shown a permission dialog.
type PermissionRequestInput struct {
	BaseInput
	ToolName              string                 `json:"tool_name"`
	ToolInput             json.RawMessage        `json:"tool_input"`
	ToolUseID             string                 `json:"tool_use_id"`
	PermissionSuggestions []PermissionSuggestion `json:"permission_suggestions,omitempty"`
}

// PermissionRequestOutput controls the permission dialog response.
type PermissionRequestOutput struct {
	BaseOutput
	HookSpecificOutput *PermissionRequestHookOutput `json:"hookSpecificOutput,omitempty"`
}

// PermissionRequestHookOutput contains permission-request-specific output fields.
type PermissionRequestHookOutput struct {
	HookEventName string                     `json:"hookEventName,omitempty"`
	Decision      *PermissionRequestDecision `json:"decision,omitempty"`
}

// PermissionRequestDecision controls how the permission request is handled.
type PermissionRequestDecision struct {
	Behavior           string         `json:"behavior"`                     // "allow" or "deny"
	UpdatedInput       map[string]any `json:"updatedInput,omitempty"`       // For "allow": modified tool input
	UpdatedPermissions any            `json:"updatedPermissions,omitempty"` // For "allow": permission rule updates (equivalent to "always allow")
	Message            string         `json:"message,omitempty"`            // For "deny": shown to Claude
	Interrupt          bool           `json:"interrupt,omitempty"`          // For "deny": stop Claude if true
}

// =============================================================================
// PostToolUse
// =============================================================================

// PostToolUseInput is received after a tool executes successfully.
type PostToolUseInput struct {
	BaseInput
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"` // Tool-specific response JSON
	ToolUseID    string          `json:"tool_use_id"`
}

// PostToolUseOutput can provide feedback to Claude after tool execution.
type PostToolUseOutput struct {
	BaseOutput
	// Decision should be "block" to provide feedback to Claude, or omitted.
	Decision string `json:"decision,omitempty"`
	// Reason is shown to Claude when Decision is "block".
	Reason             string                  `json:"reason,omitempty"`
	HookSpecificOutput *PostToolUseHookOutput `json:"hookSpecificOutput,omitempty"`
}

// PostToolUseHookOutput contains post-tool-use-specific output fields.
type PostToolUseHookOutput struct {
	HookEventName       string `json:"hookEventName,omitempty"`
	AdditionalContext   string `json:"additionalContext,omitempty"`   // Added to context for Claude
	UpdatedMCPToolOutput any   `json:"updatedMCPToolOutput,omitempty"` // For MCP tools: replaces the tool's output
}

// =============================================================================
// UserPromptSubmit
// =============================================================================

// UserPromptSubmitInput is received when the user submits a prompt.
type UserPromptSubmitInput struct {
	BaseInput
	Prompt string `json:"prompt"`
}

// UserPromptSubmitOutput controls whether the prompt should be processed.
type UserPromptSubmitOutput struct {
	BaseOutput
	// Decision should be "block" to prevent processing, or omitted.
	Decision string `json:"decision,omitempty"`
	// Reason is shown to the user when Decision is "block".
	Reason             string                       `json:"reason,omitempty"`
	HookSpecificOutput *UserPromptSubmitHookOutput `json:"hookSpecificOutput,omitempty"`
}

// UserPromptSubmitHookOutput contains user-prompt-submit-specific output fields.
type UserPromptSubmitHookOutput struct {
	HookEventName     string `json:"hookEventName,omitempty"`
	AdditionalContext string `json:"additionalContext,omitempty"` // Added to context
}

// =============================================================================
// Notification
// =============================================================================

// NotificationInput is received when Claude Code sends a notification.
type NotificationInput struct {
	BaseInput
	Message          string `json:"message"`
	Title            string `json:"title,omitempty"`
	NotificationType string `json:"notification_type"` // "permission_prompt", "idle_prompt", "auth_success", "elicitation_dialog"
}

// NotificationOutput can add context to the conversation.
type NotificationOutput struct {
	BaseOutput
	HookSpecificOutput *NotificationHookOutput `json:"hookSpecificOutput,omitempty"`
}

// NotificationHookOutput contains notification-specific output fields.
type NotificationHookOutput struct {
	HookEventName     string `json:"hookEventName,omitempty"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

// =============================================================================
// PreCompact
// =============================================================================

// PreCompactInput is received before Claude Code runs a compact operation.
type PreCompactInput struct {
	BaseInput
	Trigger            string `json:"trigger"`             // "manual" or "auto"
	CustomInstructions string `json:"custom_instructions"` // For manual trigger only
}

// PreCompactOutput has no decision control - hooks just run for side effects.
type PreCompactOutput struct {
	BaseOutput
}
