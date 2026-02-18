package claude

import (
	"testing"
)

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
	output := Block("Please fix issues")
	if output.Decision != "block" {
		t.Errorf("Decision = %q, want %q", output.Decision, "block")
	}
	if output.Reason != "Please fix issues" {
		t.Errorf("Reason = %q, want %q", output.Reason, "Please fix issues")
	}
}

func TestStopWith(t *testing.T) {
	output := StopWith("Stopping due to error")
	if output.Continue == nil || *output.Continue != false {
		t.Error("Continue should be false")
	}
	if output.StopReason != "Stopping due to error" {
		t.Errorf("StopReason = %q, want %q", output.StopReason, "Stopping due to error")
	}
}

// =============================================================================
// PreToolUse Helpers Tests
// =============================================================================

func TestAllow(t *testing.T) {
	output := Allow("Tool is trusted")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q", output.HookSpecificOutput.PermissionDecision, "allow")
	}
	if output.HookSpecificOutput.PermissionDecisionReason != "Tool is trusted" {
		t.Errorf("PermissionDecisionReason = %q, want %q", output.HookSpecificOutput.PermissionDecisionReason, "Tool is trusted")
	}
}

func TestAllowSilent(t *testing.T) {
	output := AllowSilent()
	if !output.SuppressOutput {
		t.Error("SuppressOutput should be true")
	}
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q", output.HookSpecificOutput.PermissionDecision, "allow")
	}
}

func TestAllowWithInput(t *testing.T) {
	updatedInput := map[string]any{"file_path": "/new/path.ts"}
	output := AllowWithInput("Modified input", updatedInput)
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q", output.HookSpecificOutput.PermissionDecision, "allow")
	}
	if output.HookSpecificOutput.UpdatedInput["file_path"] != "/new/path.ts" {
		t.Errorf("UpdatedInput[file_path] = %v, want %v", output.HookSpecificOutput.UpdatedInput["file_path"], "/new/path.ts")
	}
}

func TestDeny(t *testing.T) {
	output := Deny("Tool not allowed")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want %q", output.HookSpecificOutput.PermissionDecision, "deny")
	}
	if output.HookSpecificOutput.PermissionDecisionReason != "Tool not allowed" {
		t.Errorf("PermissionDecisionReason = %q, want %q", output.HookSpecificOutput.PermissionDecisionReason, "Tool not allowed")
	}
}

func TestAsk(t *testing.T) {
	output := Ask("Confirm execution?")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != "ask" {
		t.Errorf("PermissionDecision = %q, want %q", output.HookSpecificOutput.PermissionDecision, "ask")
	}
	if output.HookSpecificOutput.PermissionDecisionReason != "Confirm execution?" {
		t.Errorf("PermissionDecisionReason = %q, want %q", output.HookSpecificOutput.PermissionDecisionReason, "Confirm execution?")
	}
}

func TestAllowWithContext(t *testing.T) {
	output := AllowWithContext("Tool is safe", "Running in production")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want %q", output.HookSpecificOutput.PermissionDecision, "allow")
	}
	if output.HookSpecificOutput.PermissionDecisionReason != "Tool is safe" {
		t.Errorf("PermissionDecisionReason = %q, want %q", output.HookSpecificOutput.PermissionDecisionReason, "Tool is safe")
	}
	if output.HookSpecificOutput.AdditionalContext != "Running in production" {
		t.Errorf("AdditionalContext = %q, want %q", output.HookSpecificOutput.AdditionalContext, "Running in production")
	}
}

func TestPassThrough(t *testing.T) {
	output := PassThrough()
	if output.HookSpecificOutput != nil {
		t.Error("HookSpecificOutput should be nil for PassThrough")
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
	if output.HookSpecificOutput.Decision.Behavior != "allow" {
		t.Errorf("Behavior = %q, want %q", output.HookSpecificOutput.Decision.Behavior, "allow")
	}
}

func TestAllowPermissionWithInput(t *testing.T) {
	updatedInput := map[string]any{"key": "value"}
	output := AllowPermissionWithInput(updatedInput)
	if output.HookSpecificOutput == nil || output.HookSpecificOutput.Decision == nil {
		t.Fatal("HookSpecificOutput and Decision should not be nil")
	}
	if output.HookSpecificOutput.Decision.Behavior != "allow" {
		t.Errorf("Behavior = %q, want %q", output.HookSpecificOutput.Decision.Behavior, "allow")
	}
	if output.HookSpecificOutput.Decision.UpdatedInput["key"] != "value" {
		t.Errorf("UpdatedInput[key] = %v, want %v", output.HookSpecificOutput.Decision.UpdatedInput["key"], "value")
	}
}

func TestDenyPermission(t *testing.T) {
	output := DenyPermission("Not allowed")
	if output.HookSpecificOutput == nil || output.HookSpecificOutput.Decision == nil {
		t.Fatal("HookSpecificOutput and Decision should not be nil")
	}
	if output.HookSpecificOutput.Decision.Behavior != "deny" {
		t.Errorf("Behavior = %q, want %q", output.HookSpecificOutput.Decision.Behavior, "deny")
	}
	if output.HookSpecificOutput.Decision.Message != "Not allowed" {
		t.Errorf("Message = %q, want %q", output.HookSpecificOutput.Decision.Message, "Not allowed")
	}
}

func TestAllowPermissionWithPermissions(t *testing.T) {
	perms := map[string]any{"type": "toolAlwaysAllow", "tool": "Bash"}
	output := AllowPermissionWithPermissions(perms)
	if output.HookSpecificOutput == nil || output.HookSpecificOutput.Decision == nil {
		t.Fatal("HookSpecificOutput and Decision should not be nil")
	}
	if output.HookSpecificOutput.Decision.Behavior != "allow" {
		t.Errorf("Behavior = %q, want %q", output.HookSpecificOutput.Decision.Behavior, "allow")
	}
	if output.HookSpecificOutput.Decision.UpdatedPermissions == nil {
		t.Error("UpdatedPermissions should not be nil")
	}
}

func TestDenyPermissionAndStop(t *testing.T) {
	output := DenyPermissionAndStop("Critical error")
	if output.HookSpecificOutput == nil || output.HookSpecificOutput.Decision == nil {
		t.Fatal("HookSpecificOutput and Decision should not be nil")
	}
	if output.HookSpecificOutput.Decision.Behavior != "deny" {
		t.Errorf("Behavior = %q, want %q", output.HookSpecificOutput.Decision.Behavior, "deny")
	}
	if !output.HookSpecificOutput.Decision.Interrupt {
		t.Error("Interrupt should be true")
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
	output := PostToolBlock("Issue found")
	if output.Decision != "block" {
		t.Errorf("Decision = %q, want %q", output.Decision, "block")
	}
	if output.Reason != "Issue found" {
		t.Errorf("Reason = %q, want %q", output.Reason, "Issue found")
	}
}

func TestPostToolContext(t *testing.T) {
	output := PostToolContext("Additional info")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.AdditionalContext != "Additional info" {
		t.Errorf("AdditionalContext = %q, want %q", output.HookSpecificOutput.AdditionalContext, "Additional info")
	}
}

func TestPostToolReplaceMCPOutput(t *testing.T) {
	replacement := map[string]string{"result": "sanitized"}
	output := PostToolReplaceMCPOutput(replacement)
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.HookEventName != "PostToolUse" {
		t.Errorf("HookEventName = %q, want %q", output.HookSpecificOutput.HookEventName, "PostToolUse")
	}
	if output.HookSpecificOutput.UpdatedMCPToolOutput == nil {
		t.Error("UpdatedMCPToolOutput should not be nil")
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
	output := BlockPrompt("Invalid prompt")
	if output.Decision != "block" {
		t.Errorf("Decision = %q, want %q", output.Decision, "block")
	}
	if output.Reason != "Invalid prompt" {
		t.Errorf("Reason = %q, want %q", output.Reason, "Invalid prompt")
	}
}

func TestAddContext(t *testing.T) {
	output := AddContext("Project uses Go 1.21")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.AdditionalContext != "Project uses Go 1.21" {
		t.Errorf("AdditionalContext = %q, want %q", output.HookSpecificOutput.AdditionalContext, "Project uses Go 1.21")
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
	output := SessionStartContext("Welcome context")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.AdditionalContext != "Welcome context" {
		t.Errorf("AdditionalContext = %q, want %q", output.HookSpecificOutput.AdditionalContext, "Welcome context")
	}
}

// =============================================================================
// Other Helpers Tests
// =============================================================================

func TestSessionEndOK(t *testing.T) {
	output := SessionEndOK()
	_ = output // Just verify it returns without error
}

func TestNotificationOK(t *testing.T) {
	output := NotificationOK()
	if output.HookSpecificOutput != nil {
		t.Error("HookSpecificOutput should be nil for NotificationOK")
	}
}

func TestNotificationContext(t *testing.T) {
	output := NotificationContext("User needs help")
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.HookEventName != "Notification" {
		t.Errorf("HookEventName = %q, want %q", output.HookSpecificOutput.HookEventName, "Notification")
	}
	if output.HookSpecificOutput.AdditionalContext != "User needs help" {
		t.Errorf("AdditionalContext = %q, want %q", output.HookSpecificOutput.AdditionalContext, "User needs help")
	}
}

func TestPreCompactOK(t *testing.T) {
	output := PreCompactOK()
	_ = output
}
