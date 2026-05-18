// Package codex provides types and helpers for OpenAI Codex hooks.
//
// Codex hooks use the same JSON protocol as Claude Code hooks.
// This package re-exports the claude package types for convenience.
//
// See https://developers.openai.com/codex/hooks for the upstream
// specification.
package codex

import "github.com/CorridorSecurity/hookshot/claude"

// =============================================================================
// Common Types (re-exported from claude)
// =============================================================================

// BaseInput contains fields present in all Codex hook inputs.
//
// In Codex, every command hook receives one JSON object on stdin with these
// shared fields (see https://developers.openai.com/codex/hooks). Codex also
// passes a "model" field and turn-scoped hooks include a "turn_id"; those
// fields are accessible through the raw JSON if needed.
type BaseInput = claude.BaseInput

// BaseOutput contains common fields that can be included in any hook output.
type BaseOutput = claude.BaseOutput

// =============================================================================
// Stop (re-exported from claude)
// =============================================================================

// StopInput is received when the main Codex turn finishes responding.
type StopInput = claude.StopInput

// StopOutput controls whether Codex should stop or continue the turn.
//
// For this event, decision: "block" doesn't reject the turn. Instead, it tells
// Codex to continue and automatically creates a new continuation prompt that
// acts as a new user prompt, using the reason as that prompt text.
type StopOutput = claude.StopOutput

// =============================================================================
// SessionStart (re-exported from claude)
// =============================================================================

// SessionStartInput is received when Codex starts or resumes a session.
//
// The "source" field is one of "startup", "resume", or "clear".
type SessionStartInput = claude.SessionStartInput

// SessionStartOutput can add developer context to the session.
type SessionStartOutput = claude.SessionStartOutput

// SessionStartHookOutput contains session-start-specific output fields.
type SessionStartHookOutput = claude.SessionStartHookOutput

// =============================================================================
// PreToolUse (re-exported from claude)
// =============================================================================

// PreToolUseInput is received before Codex executes a tool.
//
// Codex PreToolUse can intercept Bash, file edits performed through
// apply_patch, and MCP tool calls. The canonical ToolName values include
// "Bash", "apply_patch", and MCP tool names like "mcp__server__tool".
type PreToolUseInput = claude.PreToolUseInput

// PreToolUseOutput controls whether the tool should execute.
//
// Codex currently honors permissionDecision: "deny" (or the older
// decision: "block" shape) for Bash and apply_patch. "allow" and "ask"
// values are parsed but fall open today.
type PreToolUseOutput = claude.PreToolUseOutput

// PreToolUseHookOutput contains pre-tool-use-specific output fields.
type PreToolUseHookOutput = claude.PreToolUseHookOutput

// =============================================================================
// PermissionRequest (re-exported from claude)
// =============================================================================

// PermissionRequestInput is received when Codex is about to ask for approval,
// such as a shell escalation or managed-network approval. It doesn't run for
// commands that don't need approval.
type PermissionRequestInput = claude.PermissionRequestInput

// PermissionRequestOutput controls the permission dialog response.
//
// Codex respects "allow" and "deny" behaviors. Returning updatedInput,
// updatedPermissions, or interrupt is reserved for future behavior and
// currently fails closed.
type PermissionRequestOutput = claude.PermissionRequestOutput

// PermissionRequestHookOutput contains permission-request-specific output fields.
type PermissionRequestHookOutput = claude.PermissionRequestHookOutput

// PermissionRequestDecision controls how the permission request is handled.
type PermissionRequestDecision = claude.PermissionRequestDecision

// =============================================================================
// PostToolUse (re-exported from claude)
// =============================================================================

// PostToolUseInput is received after a tool executes.
//
// For Bash, PostToolUse also runs after commands that exit with a non-zero
// status. The canonical ToolName values include "Bash", "apply_patch", and
// MCP tool names like "mcp__server__tool".
type PostToolUseInput = claude.PostToolUseInput

// PostToolUseOutput can provide feedback to Codex after tool execution.
//
// For this event, decision: "block" doesn't undo the completed tool call.
// Codex records the feedback, replaces the tool result with it, and continues
// the model from the hook-provided message.
type PostToolUseOutput = claude.PostToolUseOutput

// PostToolUseHookOutput contains post-tool-use-specific output fields.
type PostToolUseHookOutput = claude.PostToolUseHookOutput

// =============================================================================
// UserPromptSubmit (re-exported from claude)
// =============================================================================

// UserPromptSubmitInput is received when the user submits a prompt.
type UserPromptSubmitInput = claude.UserPromptSubmitInput

// UserPromptSubmitOutput controls whether the prompt should be processed
// and can add developer context.
type UserPromptSubmitOutput = claude.UserPromptSubmitOutput

// UserPromptSubmitHookOutput contains user-prompt-submit-specific output fields.
type UserPromptSubmitHookOutput = claude.UserPromptSubmitHookOutput
