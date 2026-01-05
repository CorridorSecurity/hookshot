package cursor

import (
	"testing"
)

// =============================================================================
// Stop Helpers Tests
// =============================================================================

func TestContinue(t *testing.T) {
	output := Continue()
	if output.FollowupMessage != "" {
		t.Errorf("Continue should have empty FollowupMessage, got %q", output.FollowupMessage)
	}
}

func TestFollowup(t *testing.T) {
	output := Followup("Please verify changes")
	if output.FollowupMessage != "Please verify changes" {
		t.Errorf("Followup message = %q, want %q", output.FollowupMessage, "Please verify changes")
	}
}

// =============================================================================
// Before Execution Helpers Tests
// =============================================================================

func TestAllow(t *testing.T) {
	output := Allow()
	if output.Permission != "allow" {
		t.Errorf("Permission = %q, want %q", output.Permission, "allow")
	}
	if output.UserMessage != "" || output.AgentMessage != "" {
		t.Error("Allow should have empty UserMessage and AgentMessage")
	}
}

func TestAllowWithMessage(t *testing.T) {
	output := AllowWithMessage("Action permitted")
	if output.Permission != "allow" {
		t.Errorf("Permission = %q, want %q", output.Permission, "allow")
	}
	if output.UserMessage != "Action permitted" {
		t.Errorf("UserMessage = %q, want %q", output.UserMessage, "Action permitted")
	}
}

func TestDeny(t *testing.T) {
	output := Deny("User msg", "Agent msg")
	if output.Permission != "deny" {
		t.Errorf("Permission = %q, want %q", output.Permission, "deny")
	}
	if output.UserMessage != "User msg" {
		t.Errorf("UserMessage = %q, want %q", output.UserMessage, "User msg")
	}
	if output.AgentMessage != "Agent msg" {
		t.Errorf("AgentMessage = %q, want %q", output.AgentMessage, "Agent msg")
	}
}

func TestDenyWithUserMessage(t *testing.T) {
	output := DenyWithUserMessage("Blocked")
	if output.Permission != "deny" {
		t.Errorf("Permission = %q, want %q", output.Permission, "deny")
	}
	if output.UserMessage != "Blocked" {
		t.Errorf("UserMessage = %q, want %q", output.UserMessage, "Blocked")
	}
	if output.AgentMessage != "" {
		t.Errorf("AgentMessage should be empty, got %q", output.AgentMessage)
	}
}

func TestDenyWithAgentMessage(t *testing.T) {
	output := DenyWithAgentMessage("Not allowed")
	if output.Permission != "deny" {
		t.Errorf("Permission = %q, want %q", output.Permission, "deny")
	}
	if output.AgentMessage != "Not allowed" {
		t.Errorf("AgentMessage = %q, want %q", output.AgentMessage, "Not allowed")
	}
	if output.UserMessage != "" {
		t.Errorf("UserMessage should be empty, got %q", output.UserMessage)
	}
}

func TestAsk(t *testing.T) {
	output := Ask("Confirm action?")
	if output.Permission != "ask" {
		t.Errorf("Permission = %q, want %q", output.Permission, "ask")
	}
	if output.UserMessage != "Confirm action?" {
		t.Errorf("UserMessage = %q, want %q", output.UserMessage, "Confirm action?")
	}
}

// =============================================================================
// Before Read File Helpers Tests
// =============================================================================

func TestAllowRead(t *testing.T) {
	output := AllowRead()
	if output.Permission != "allow" {
		t.Errorf("Permission = %q, want %q", output.Permission, "allow")
	}
}

func TestDenyRead(t *testing.T) {
	output := DenyRead("Cannot read", "File is private")
	if output.Permission != "deny" {
		t.Errorf("Permission = %q, want %q", output.Permission, "deny")
	}
	if output.UserMessage != "Cannot read" {
		t.Errorf("UserMessage = %q, want %q", output.UserMessage, "Cannot read")
	}
	if output.AgentMessage != "File is private" {
		t.Errorf("AgentMessage = %q, want %q", output.AgentMessage, "File is private")
	}
}

func TestAskRead(t *testing.T) {
	output := AskRead("Allow file access?")
	if output.Permission != "ask" {
		t.Errorf("Permission = %q, want %q", output.Permission, "ask")
	}
	if output.UserMessage != "Allow file access?" {
		t.Errorf("UserMessage = %q, want %q", output.UserMessage, "Allow file access?")
	}
}

// =============================================================================
// Tab File Helpers Tests
// =============================================================================

func TestAllowTabRead(t *testing.T) {
	output := AllowTabRead()
	if output.Permission != "allow" {
		t.Errorf("Permission = %q, want %q", output.Permission, "allow")
	}
}

func TestDenyTabRead(t *testing.T) {
	output := DenyTabRead()
	if output.Permission != "deny" {
		t.Errorf("Permission = %q, want %q", output.Permission, "deny")
	}
}

// =============================================================================
// Before Submit Prompt Helpers Tests
// =============================================================================

func TestAllowPrompt(t *testing.T) {
	output := AllowPrompt()
	if !output.Continue {
		t.Error("Continue should be true")
	}
}

func TestBlockPrompt(t *testing.T) {
	output := BlockPrompt("Invalid prompt")
	if output.Continue {
		t.Error("Continue should be false")
	}
	if output.UserMessage != "Invalid prompt" {
		t.Errorf("UserMessage = %q, want %q", output.UserMessage, "Invalid prompt")
	}
}

// =============================================================================
// After Execution Helpers Tests
// =============================================================================

func TestAfterShellOK(t *testing.T) {
	// Just verify it returns without error
	output := AfterShellOK()
	_ = output
}

func TestAfterMCPOK(t *testing.T) {
	output := AfterMCPOK()
	_ = output
}

func TestAfterFileEditOK(t *testing.T) {
	output := AfterFileEditOK()
	_ = output
}

func TestAfterTabFileEditOK(t *testing.T) {
	output := AfterTabFileEditOK()
	_ = output
}

func TestAfterAgentResponseOK(t *testing.T) {
	output := AfterAgentResponseOK()
	_ = output
}

func TestAfterAgentThoughtOK(t *testing.T) {
	output := AfterAgentThoughtOK()
	_ = output
}
