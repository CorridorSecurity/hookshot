package codex

import (
	"strings"
	"testing"
)

// These tests verify that the codex helpers correctly re-export the Claude
// Code helpers and produce the wire shapes documented for Codex hooks.

// =============================================================================
// Stop Helpers Tests
// =============================================================================

func TestContinue(t *testing.T) {
	output := Continue()
	if output.Decision != "" {
		t.Errorf("Decision should be empty, got %q", output.Decision)
	}
	if output.Reason != "" {
		t.Errorf("Reason should be empty, got %q", output.Reason)
	}
}

func TestBlock(t *testing.T) {
	output := Block("Run one more pass over the failing tests")
	if output.Decision != "block" {
		t.Errorf("Decision = %q, want %q", output.Decision, "block")
	}
	if output.Reason != "Run one more pass over the failing tests" {
		t.Errorf("Reason = %q, want %q", output.Reason, "Run one more pass over the failing tests")
	}
}

func TestStopWith(t *testing.T) {
	output := StopWith("Stopping due to policy violation")
	if output.Continue == nil || *output.Continue != false {
		t.Error("Continue should be false")
	}
	if output.StopReason != "Stopping due to policy violation" {
		t.Errorf("StopReason = %q, want %q", output.StopReason, "Stopping due to policy violation")
	}
}

// =============================================================================
// PreToolUse Helpers Tests
// =============================================================================

func TestDeny(t *testing.T) {
	output := Deny("Destructive command blocked by hook.")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q", output.HookSpecificOutput.PermissionDecision, "deny")
	}
	if output.HookSpecificOutput.PermissionDecisionReason != "Destructive command blocked by hook." {
		t.Errorf("PermissionDecisionReason = %q, want %q", output.HookSpecificOutput.PermissionDecisionReason, "Destructive command blocked by hook.")
	}
}

func TestAllow(t *testing.T) {
	output := Allow("Trusted command")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q", output.HookSpecificOutput.PermissionDecision, "allow")
	}
}

func TestPassThrough(t *testing.T) {
	output := PassThrough()
	if output.HookSpecificOutput != nil {
		t.Error("HookSpecificOutput should be nil for PassThrough")
	}
}

func TestAllowSilent_DoesNotEmitSuppressOutput(t *testing.T) {
	// Codex rejects suppressOutput with
	// "PreToolUse hook returned unsupported suppressOutput".
	// codex.AllowSilent must NOT set SuppressOutput (unlike claude.AllowSilent).
	output := AllowSilent()
	if output.SuppressOutput {
		t.Error("codex.AllowSilent must not set SuppressOutput (Codex rejects suppressOutput)")
	}
	if output.HookSpecificOutput != nil {
		t.Error("codex.AllowSilent should emit empty {} on Codex (no hookSpecificOutput)")
	}
}

func TestAllowWithInput_FailsClosed(t *testing.T) {
	// Codex rejects updatedInput with
	// "PreToolUse hook returned unsupported updatedInput", so there is no
	// safe way to apply the sanitized input. The helper must fail closed
	// (Deny) rather than silently fall through to Allow with the original,
	// unsanitized tool_input.
	output := AllowWithInput("stripped --no-verify", map[string]any{"file_path": "/tmp/x"})
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q (fail-closed)", output.HookSpecificOutput.PermissionDecision, "deny")
	}
	if output.HookSpecificOutput.UpdatedInput != nil {
		t.Error("codex.AllowWithInput must not set UpdatedInput (Codex rejects updatedInput)")
	}
	if !strings.Contains(output.HookSpecificOutput.PermissionDecisionReason, "stripped --no-verify") {
		t.Errorf("PermissionDecisionReason should include caller reason, got %q", output.HookSpecificOutput.PermissionDecisionReason)
	}
	if !strings.Contains(output.HookSpecificOutput.PermissionDecisionReason, "input rewriting not supported on Codex") {
		t.Errorf("PermissionDecisionReason should explain why input rewriting was dropped, got %q", output.HookSpecificOutput.PermissionDecisionReason)
	}
}

func TestAllowWithInput_EmptyReasonStillDenies(t *testing.T) {
	output := AllowWithInput("", map[string]any{"k": "v"})
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil (fail-closed)")
	}
	if output.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q (fail-closed)", output.HookSpecificOutput.PermissionDecision, "deny")
	}
}

func TestAsk_FailsClosed(t *testing.T) {
	// Codex parses but does not enforce permissionDecision: "ask" for
	// PreToolUse, so emitting "ask" would silently execute the tool
	// without surfacing an approval prompt. codex.Ask must fail closed
	// (Deny) to match the unified bridge's behaviour on Codex.
	output := Ask("Confirm before deleting the production database.")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q (Codex does not enforce ask)", output.HookSpecificOutput.PermissionDecision, "deny")
	}
	if output.HookSpecificOutput.PermissionDecisionReason != "Confirm before deleting the production database." {
		t.Errorf("PermissionDecisionReason = %q, want caller reason preserved", output.HookSpecificOutput.PermissionDecisionReason)
	}
}

// =============================================================================
// PermissionRequest Helpers Tests
// =============================================================================

func TestAllowPermission(t *testing.T) {
	output := AllowPermission()
	if output.HookSpecificOutput == nil || output.HookSpecificOutput.Decision == nil {
		t.Fatal("HookSpecificOutput and Decision should not be nil")
	}
	if output.HookSpecificOutput.HookEventName != "PermissionRequest" {
		t.Errorf("HookEventName = %q, want %q", output.HookSpecificOutput.HookEventName, "PermissionRequest")
	}
	if output.HookSpecificOutput.Decision.Behavior != "allow" {
		t.Errorf("Behavior = %q, want %q", output.HookSpecificOutput.Decision.Behavior, "allow")
	}
}

func TestDenyPermission(t *testing.T) {
	output := DenyPermission("Blocked by repository policy.")
	if output.HookSpecificOutput == nil || output.HookSpecificOutput.Decision == nil {
		t.Fatal("HookSpecificOutput and Decision should not be nil")
	}
	if output.HookSpecificOutput.Decision.Behavior != "deny" {
		t.Errorf("Behavior = %q, want %q", output.HookSpecificOutput.Decision.Behavior, "deny")
	}
	if output.HookSpecificOutput.Decision.Message != "Blocked by repository policy." {
		t.Errorf("Message = %q, want %q", output.HookSpecificOutput.Decision.Message, "Blocked by repository policy.")
	}
}

// =============================================================================
// PostToolUse Helpers Tests
// =============================================================================

func TestPostToolOK(t *testing.T) {
	output := PostToolOK()
	if output.Decision != "" {
		t.Errorf("Decision should be empty, got %q", output.Decision)
	}
}

func TestPostToolBlock(t *testing.T) {
	output := PostToolBlock("Output needs review before continuing.")
	if output.Decision != "block" {
		t.Errorf("Decision = %q, want %q", output.Decision, "block")
	}
	if output.Reason != "Output needs review before continuing." {
		t.Errorf("Reason = %q, want %q", output.Reason, "Output needs review before continuing.")
	}
}

func TestPostToolContext(t *testing.T) {
	output := PostToolContext("The command updated generated files.")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.HookEventName != "PostToolUse" {
		t.Errorf("HookEventName = %q, want %q", output.HookSpecificOutput.HookEventName, "PostToolUse")
	}
	if output.HookSpecificOutput.AdditionalContext != "The command updated generated files." {
		t.Errorf("AdditionalContext = %q, want %q", output.HookSpecificOutput.AdditionalContext, "The command updated generated files.")
	}
}

// =============================================================================
// UserPromptSubmit Helpers Tests
// =============================================================================

func TestAllowPrompt(t *testing.T) {
	output := AllowPrompt()
	if output.Decision != "" {
		t.Errorf("Decision should be empty, got %q", output.Decision)
	}
}

func TestBlockPrompt(t *testing.T) {
	output := BlockPrompt("Ask for confirmation before doing that.")
	if output.Decision != "block" {
		t.Errorf("Decision = %q, want %q", output.Decision, "block")
	}
	if output.Reason != "Ask for confirmation before doing that." {
		t.Errorf("Reason = %q, want %q", output.Reason, "Ask for confirmation before doing that.")
	}
}

func TestAddContext(t *testing.T) {
	output := AddContext("Ask for a clearer reproduction before editing files.")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.HookEventName != "UserPromptSubmit" {
		t.Errorf("HookEventName = %q, want %q", output.HookSpecificOutput.HookEventName, "UserPromptSubmit")
	}
	if output.HookSpecificOutput.AdditionalContext != "Ask for a clearer reproduction before editing files." {
		t.Errorf("AdditionalContext = %q, want %q", output.HookSpecificOutput.AdditionalContext, "Ask for a clearer reproduction before editing files.")
	}
}

// =============================================================================
// SessionStart Helpers Tests
// =============================================================================

func TestSessionStartOK(t *testing.T) {
	output := SessionStartOK()
	if output.HookSpecificOutput != nil {
		t.Error("HookSpecificOutput should be nil for SessionStartOK")
	}
}

func TestSessionStartContext(t *testing.T) {
	output := SessionStartContext("Load the workspace conventions before editing.")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Errorf("HookEventName = %q, want %q", output.HookSpecificOutput.HookEventName, "SessionStart")
	}
	if output.HookSpecificOutput.AdditionalContext != "Load the workspace conventions before editing." {
		t.Errorf("AdditionalContext = %q, want %q", output.HookSpecificOutput.AdditionalContext, "Load the workspace conventions before editing.")
	}
}
