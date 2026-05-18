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

// AllowSilent permits the tool to execute. Codex does NOT support
// suppressOutput — the Codex CLI rejects the hook output with
// "PreToolUse hook returned unsupported suppressOutput" — so on Codex the
// "silent" variant has no equivalent. This helper emits an empty {} which
// Codex treats as success. Use this instead of claude.AllowSilent in Codex
// code paths.
func AllowSilent() PreToolUseOutput {
	return PreToolUseOutput{}
}

// AllowWithInput is a fail-closed shim. Codex does NOT support updatedInput
// — the CLI rejects the hook output with "PreToolUse hook returned
// unsupported updatedInput" — so there is no Codex-supported way to mutate
// tool_input from a PreToolUse hook today. If a caller is using
// AllowWithInput it is almost certainly to sanitize a dangerous payload
// (strip --no-verify, rewrite a Bash command, etc.); silently dropping
// the sanitization and falling through to Allow would execute the
// unsanitized payload. To prevent that, this helper denies the call and
// surfaces the loss of input rewriting in the reason. The signature
// matches claude.AllowWithInput for source compatibility — callers that
// need to inject model-visible context on Codex should use
// AllowWithContext / additionalContext instead.
func AllowWithInput(reason string, _ map[string]any) PreToolUseOutput {
	msg := reason
	if msg == "" {
		msg = "blocked"
	}
	return claude.Deny(msg + " (input rewriting not supported on Codex; fail-closed)")
}

// Deny blocks the tool from executing. This is enforced by Codex for Bash
// and apply_patch tools.
var Deny = claude.Deny

// Ask is a fail-closed shim. Codex parses but does not enforce
// permissionDecision: "ask" for PreToolUse today, so emitting "ask" would
// silently execute the tool without surfacing an approval prompt — turning
// a "require confirmation" policy into a free pass. To match the security
// posture of the hookshot unified bridge (which already rewrites
// AskExecution to Deny on Codex; see unified.go), this helper returns a
// Deny with the caller's reason. If you actually want Codex's approval
// flow, register a PermissionRequest hook instead.
func Ask(reason string) PreToolUseOutput {
	return claude.Deny(reason)
}

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
