// Package droid provides types and helpers for Factory Droid hooks.
//
// Factory Droid hooks use the same API as Claude Code hooks.
// This package re-exports the claude package types for convenience.
package droid

import "github.com/CorridorSecurity/hookshot/claude"

// =============================================================================
// Common Types (re-exported from claude)
// =============================================================================

// BaseInput contains fields present in all Droid hook inputs.
type BaseInput = claude.BaseInput

// BaseOutput contains common fields that can be included in any hook output.
type BaseOutput = claude.BaseOutput

// =============================================================================
// Stop / SubagentStop (re-exported from claude)
// =============================================================================

// StopInput is received when the main Droid agent finishes responding.
type StopInput = claude.StopInput

// StopOutput controls whether Droid should stop or continue.
type StopOutput = claude.StopOutput

// SubagentStopInput is received when a Droid subagent (Task tool) finishes.
type SubagentStopInput = claude.SubagentStopInput

// SubagentStopOutput controls whether a subagent should stop or continue.
type SubagentStopOutput = claude.SubagentStopOutput

// =============================================================================
// SessionStart (re-exported from claude)
// =============================================================================

// SessionStartInput is received when Droid starts or resumes a session.
type SessionStartInput = claude.SessionStartInput

// SessionStartOutput can add context to the session.
type SessionStartOutput = claude.SessionStartOutput

// SessionStartHookOutput contains session-start-specific output fields.
type SessionStartHookOutput = claude.SessionStartHookOutput

// =============================================================================
// SessionEnd (re-exported from claude)
// =============================================================================

// SessionEndInput is received when a Droid session ends.
type SessionEndInput = claude.SessionEndInput

// SessionEndOutput has no decision control - hooks just run for cleanup.
type SessionEndOutput = claude.SessionEndOutput

// =============================================================================
// PreToolUse (re-exported from claude)
// =============================================================================

// PreToolUseInput is received before Droid executes a tool.
type PreToolUseInput = claude.PreToolUseInput

// PreToolUseOutput controls whether the tool should execute.
type PreToolUseOutput = claude.PreToolUseOutput

// PreToolUseHookOutput contains pre-tool-use-specific output fields.
type PreToolUseHookOutput = claude.PreToolUseHookOutput

// =============================================================================
// PermissionRequest (re-exported from claude)
// =============================================================================

// PermissionRequestInput is received when the user is shown a permission dialog.
type PermissionRequestInput = claude.PermissionRequestInput

// PermissionRequestOutput controls the permission dialog response.
type PermissionRequestOutput = claude.PermissionRequestOutput

// PermissionRequestHookOutput contains permission-request-specific output fields.
type PermissionRequestHookOutput = claude.PermissionRequestHookOutput

// PermissionRequestDecision controls how the permission request is handled.
type PermissionRequestDecision = claude.PermissionRequestDecision

// =============================================================================
// PostToolUse (re-exported from claude)
// =============================================================================

// PostToolUseInput is received after a tool executes successfully.
type PostToolUseInput = claude.PostToolUseInput

// PostToolUseOutput can provide feedback to Droid after tool execution.
type PostToolUseOutput = claude.PostToolUseOutput

// PostToolUseHookOutput contains post-tool-use-specific output fields.
type PostToolUseHookOutput = claude.PostToolUseHookOutput

// =============================================================================
// UserPromptSubmit (re-exported from claude)
// =============================================================================

// UserPromptSubmitInput is received when the user submits a prompt.
type UserPromptSubmitInput = claude.UserPromptSubmitInput

// UserPromptSubmitOutput controls whether the prompt should be processed.
type UserPromptSubmitOutput = claude.UserPromptSubmitOutput

// UserPromptSubmitHookOutput contains user-prompt-submit-specific output fields.
type UserPromptSubmitHookOutput = claude.UserPromptSubmitHookOutput

// =============================================================================
// Notification (re-exported from claude)
// =============================================================================

// NotificationInput is received when Droid sends a notification.
type NotificationInput = claude.NotificationInput

// NotificationOutput has no decision control - hooks just run for side effects.
type NotificationOutput = claude.NotificationOutput

// =============================================================================
// PreCompact (re-exported from claude)
// =============================================================================

// PreCompactInput is received before Droid runs a compact operation.
type PreCompactInput = claude.PreCompactInput

// PreCompactOutput has no decision control - hooks just run for side effects.
type PreCompactOutput = claude.PreCompactOutput
