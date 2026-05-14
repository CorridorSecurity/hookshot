package codex

import "github.com/CorridorSecurity/hookshot/claude"

// =============================================================================
// Stop Helpers (re-exported from claude)
// =============================================================================

// Continue allows Codex to stop normally.
var Continue = claude.Continue

// Block prevents Codex from stopping and asks it to continue the turn.
// The reason is used as the continuation prompt text.
var Block = claude.Block

// StopWith creates a StopOutput that halts Codex entirely (continue=false).
var StopWith = claude.StopWith

// =============================================================================
// PreToolUse Helpers (re-exported from claude)
// =============================================================================

// Allow permits the tool to execute. Note that Codex currently parses but
// does not enforce permissionDecision: "allow" for PreToolUse, so this
// effectively falls through to the normal flow.
var Allow = claude.Allow

// AllowSilent permits the tool to execute without showing output.
var AllowSilent = claude.AllowSilent

// AllowWithInput permits the tool with modified input parameters. Note that
// Codex currently parses but does not enforce updatedInput for PreToolUse.
var AllowWithInput = claude.AllowWithInput

// Deny blocks the tool from executing. This is enforced by Codex for Bash
// and apply_patch tools.
var Deny = claude.Deny

// Ask prompts the user to confirm the tool execution. Note that Codex
// currently parses but does not enforce permissionDecision: "ask" for
// PreToolUse, so this falls open today.
var Ask = claude.Ask

// PassThrough returns an empty output, letting the normal permission flow proceed.
var PassThrough = claude.PassThrough

// =============================================================================
// PermissionRequest Helpers (re-exported from claude)
// =============================================================================

// AllowPermission grants the permission request without surfacing the
// approval prompt.
var AllowPermission = claude.AllowPermission

// DenyPermission rejects the permission request with a message shown to Codex.
var DenyPermission = claude.DenyPermission

// =============================================================================
// PostToolUse Helpers (re-exported from claude)
// =============================================================================

// PostToolOK returns an empty output, allowing normal flow to continue.
var PostToolOK = claude.PostToolOK

// PostToolBlock provides feedback to Codex that replaces the tool result and
// continues the model from the hook-provided message.
var PostToolBlock = claude.PostToolBlock

// PostToolContext adds developer context after the tool runs.
var PostToolContext = claude.PostToolContext

// =============================================================================
// UserPromptSubmit Helpers (re-exported from claude)
// =============================================================================

// AllowPrompt allows the prompt to be processed normally.
var AllowPrompt = claude.AllowPrompt

// BlockPrompt prevents the prompt from being processed.
var BlockPrompt = claude.BlockPrompt

// AddContext allows the prompt and adds developer context for Codex.
var AddContext = claude.AddContext

// =============================================================================
// SessionStart Helpers (re-exported from claude)
// =============================================================================

// SessionStartOK returns an empty output for session start.
var SessionStartOK = claude.SessionStartOK

// SessionStartContext adds developer context at the start of a session.
var SessionStartContext = claude.SessionStartContext
