package droid

import "github.com/CorridorSecurity/hookshot/claude"

// =============================================================================
// Stop / SubagentStop Helpers (re-exported from claude)
// =============================================================================

// Continue allows Droid to stop normally.
var Continue = claude.Continue

// Block prevents Droid from stopping and sends a message.
var Block = claude.Block

// StopWith creates a StopOutput that halts Droid entirely.
var StopWith = claude.StopWith

// =============================================================================
// PreToolUse Helpers (re-exported from claude)
// =============================================================================

// Allow permits the tool to execute without asking the user.
var Allow = claude.Allow

// AllowSilent permits the tool to execute without showing output.
var AllowSilent = claude.AllowSilent

// AllowWithInput permits the tool with modified input parameters.
var AllowWithInput = claude.AllowWithInput

// Deny blocks the tool from executing.
var Deny = claude.Deny

// Ask prompts the user to confirm the tool execution.
var Ask = claude.Ask

// PassThrough returns an empty output, letting the normal permission flow proceed.
var PassThrough = claude.PassThrough

// =============================================================================
// PermissionRequest Helpers (re-exported from claude)
// =============================================================================

// AllowPermission grants the permission request.
var AllowPermission = claude.AllowPermission

// AllowPermissionWithInput grants the permission with modified tool input.
var AllowPermissionWithInput = claude.AllowPermissionWithInput

// DenyPermission rejects the permission request.
var DenyPermission = claude.DenyPermission

// DenyPermissionAndStop rejects the permission and stops Droid.
var DenyPermissionAndStop = claude.DenyPermissionAndStop

// =============================================================================
// PostToolUse Helpers (re-exported from claude)
// =============================================================================

// PostToolOK returns an empty output, allowing normal flow to continue.
var PostToolOK = claude.PostToolOK

// PostToolBlock provides feedback to Droid that something needs attention.
var PostToolBlock = claude.PostToolBlock

// PostToolContext adds context for Droid to consider.
var PostToolContext = claude.PostToolContext

// =============================================================================
// UserPromptSubmit Helpers (re-exported from claude)
// =============================================================================

// AllowPrompt allows the prompt to be processed normally.
var AllowPrompt = claude.AllowPrompt

// BlockPrompt prevents the prompt from being processed.
var BlockPrompt = claude.BlockPrompt

// AddContext allows the prompt and adds context for Droid.
var AddContext = claude.AddContext

// =============================================================================
// SessionStart Helpers (re-exported from claude)
// =============================================================================

// SessionStartOK returns an empty output for session start.
var SessionStartOK = claude.SessionStartOK

// SessionStartContext adds context at the start of a session.
var SessionStartContext = claude.SessionStartContext

// =============================================================================
// SessionEnd Helpers (re-exported from claude)
// =============================================================================

// SessionEndOK returns an empty output for session end.
var SessionEndOK = claude.SessionEndOK

// =============================================================================
// Notification Helpers (re-exported from claude)
// =============================================================================

// NotificationOK returns an empty output for notifications.
var NotificationOK = claude.NotificationOK

// =============================================================================
// PreCompact Helpers (re-exported from claude)
// =============================================================================

// PreCompactOK returns an empty output for pre-compact.
var PreCompactOK = claude.PreCompactOK
