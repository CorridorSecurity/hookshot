package cascade

import (
	"testing"
)

// =============================================================================
// Base Helpers Tests
// =============================================================================

func TestAllow(t *testing.T) {
	output := Allow()
	if output.Decision != "allow" {
		t.Errorf("Decision = %q, want %q", output.Decision, "allow")
	}
	if output.Message != "" {
		t.Errorf("Message should be empty, got %q", output.Message)
	}
}

func TestDeny(t *testing.T) {
	output := Deny("Action blocked")
	if output.Decision != "deny" {
		t.Errorf("Decision = %q, want %q", output.Decision, "deny")
	}
	if output.Message != "Action blocked" {
		t.Errorf("Message = %q, want %q", output.Message, "Action blocked")
	}
}

func TestBlock(t *testing.T) {
	output, err := Block("Action blocked")
	if output.Decision != "deny" {
		t.Errorf("Decision = %q, want %q", output.Decision, "deny")
	}
	if output.Message != "Action blocked" {
		t.Errorf("Message = %q, want %q", output.Message, "Action blocked")
	}
	if err == nil || err.Error() != "Action blocked" {
		t.Fatalf("error = %v, want Action blocked", err)
	}
}

func TestAsk(t *testing.T) {
	output := Ask("Confirm action?")
	if output.Decision != "ask" {
		t.Errorf("Decision = %q, want %q", output.Decision, "ask")
	}
	if output.Message != "Confirm action?" {
		t.Errorf("Message = %q, want %q", output.Message, "Confirm action?")
	}
}

// =============================================================================
// PreRunCommand Helpers Tests
// =============================================================================

func TestAllowCommand(t *testing.T) {
	output := AllowCommand()
	if output.Decision != "allow" {
		t.Errorf("Decision = %q, want %q", output.Decision, "allow")
	}
	if output.Message != "" {
		t.Errorf("Message should be empty, got %q", output.Message)
	}
}

func TestDenyCommand(t *testing.T) {
	output := DenyCommand("Command not allowed")
	if output.Decision != "deny" {
		t.Errorf("Decision = %q, want %q", output.Decision, "deny")
	}
	if output.Message != "Command not allowed" {
		t.Errorf("Message = %q, want %q", output.Message, "Command not allowed")
	}
}

func TestBlockCommand(t *testing.T) {
	output, err := BlockCommand("Command not allowed")
	if output.Decision != "deny" || output.Message != "Command not allowed" {
		t.Fatalf("output = %+v, want deny with message", output)
	}
	if err == nil || err.Error() != "Command not allowed" {
		t.Fatalf("error = %v, want Command not allowed", err)
	}
}

func TestAskCommand(t *testing.T) {
	output := AskCommand("Run this command?")
	if output.Decision != "ask" {
		t.Errorf("Decision = %q, want %q", output.Decision, "ask")
	}
	if output.Message != "Run this command?" {
		t.Errorf("Message = %q, want %q", output.Message, "Run this command?")
	}
}

// =============================================================================
// PreWriteCode Helpers Tests
// =============================================================================

func TestAllowWrite(t *testing.T) {
	output := AllowWrite()
	if output.Decision != "allow" {
		t.Errorf("Decision = %q, want %q", output.Decision, "allow")
	}
	if output.Message != "" {
		t.Errorf("Message should be empty, got %q", output.Message)
	}
}

func TestDenyWrite(t *testing.T) {
	output := DenyWrite("Write not permitted")
	if output.Decision != "deny" {
		t.Errorf("Decision = %q, want %q", output.Decision, "deny")
	}
	if output.Message != "Write not permitted" {
		t.Errorf("Message = %q, want %q", output.Message, "Write not permitted")
	}
}

func TestBlockWrite(t *testing.T) {
	output, err := BlockWrite("Write not permitted")
	if output.Decision != "deny" || output.Message != "Write not permitted" {
		t.Fatalf("output = %+v, want deny with message", output)
	}
	if err == nil || err.Error() != "Write not permitted" {
		t.Fatalf("error = %v, want Write not permitted", err)
	}
}

func TestAskWrite(t *testing.T) {
	output := AskWrite("Allow file write?")
	if output.Decision != "ask" {
		t.Errorf("Decision = %q, want %q", output.Decision, "ask")
	}
	if output.Message != "Allow file write?" {
		t.Errorf("Message = %q, want %q", output.Message, "Allow file write?")
	}
}

// =============================================================================
// PreReadCode Helpers Tests
// =============================================================================

func TestAllowRead(t *testing.T) {
	output := AllowRead()
	if output.Decision != "allow" {
		t.Errorf("Decision = %q, want %q", output.Decision, "allow")
	}
	if output.Message != "" {
		t.Errorf("Message should be empty, got %q", output.Message)
	}
}

func TestDenyRead(t *testing.T) {
	output := DenyRead("Cannot read file")
	if output.Decision != "deny" {
		t.Errorf("Decision = %q, want %q", output.Decision, "deny")
	}
	if output.Message != "Cannot read file" {
		t.Errorf("Message = %q, want %q", output.Message, "Cannot read file")
	}
}

func TestBlockRead(t *testing.T) {
	output, err := BlockRead("Cannot read file")
	if output.Decision != "deny" || output.Message != "Cannot read file" {
		t.Fatalf("output = %+v, want deny with message", output)
	}
	if err == nil || err.Error() != "Cannot read file" {
		t.Fatalf("error = %v, want Cannot read file", err)
	}
}

func TestAskRead(t *testing.T) {
	output := AskRead("Allow file read?")
	if output.Decision != "ask" {
		t.Errorf("Decision = %q, want %q", output.Decision, "ask")
	}
	if output.Message != "Allow file read?" {
		t.Errorf("Message = %q, want %q", output.Message, "Allow file read?")
	}
}

// =============================================================================
// PreMCPToolUse Helpers Tests
// =============================================================================

func TestAllowMCP(t *testing.T) {
	output := AllowMCP()
	if output.Decision != "allow" {
		t.Errorf("Decision = %q, want %q", output.Decision, "allow")
	}
	if output.Message != "" {
		t.Errorf("Message should be empty, got %q", output.Message)
	}
}

func TestDenyMCP(t *testing.T) {
	output := DenyMCP("MCP tool blocked")
	if output.Decision != "deny" {
		t.Errorf("Decision = %q, want %q", output.Decision, "deny")
	}
	if output.Message != "MCP tool blocked" {
		t.Errorf("Message = %q, want %q", output.Message, "MCP tool blocked")
	}
}

func TestBlockMCP(t *testing.T) {
	output, err := BlockMCP("MCP tool blocked")
	if output.Decision != "deny" || output.Message != "MCP tool blocked" {
		t.Fatalf("output = %+v, want deny with message", output)
	}
	if err == nil || err.Error() != "MCP tool blocked" {
		t.Fatalf("error = %v, want MCP tool blocked", err)
	}
}

func TestAskMCP(t *testing.T) {
	output := AskMCP("Run MCP tool?")
	if output.Decision != "ask" {
		t.Errorf("Decision = %q, want %q", output.Decision, "ask")
	}
	if output.Message != "Run MCP tool?" {
		t.Errorf("Message = %q, want %q", output.Message, "Run MCP tool?")
	}
}

// =============================================================================
// PreUserPrompt Helpers Tests
// =============================================================================

func TestAllowPrompt(t *testing.T) {
	output := AllowPrompt()
	if output.Decision != "allow" {
		t.Errorf("Decision = %q, want %q", output.Decision, "allow")
	}
	if output.Message != "" {
		t.Errorf("Message should be empty, got %q", output.Message)
	}
}

func TestBlockPrompt(t *testing.T) {
	output := BlockPrompt("Prompt blocked")
	if output.Decision != "deny" {
		t.Errorf("Decision = %q, want %q", output.Decision, "deny")
	}
	if output.Message != "Prompt blocked" {
		t.Errorf("Message = %q, want %q", output.Message, "Prompt blocked")
	}
}

func TestBlockPromptWithError(t *testing.T) {
	output, err := BlockPromptWithError("Prompt blocked")
	if output.Decision != "deny" || output.Message != "Prompt blocked" {
		t.Fatalf("output = %+v, want deny with message", output)
	}
	if err == nil || err.Error() != "Prompt blocked" {
		t.Fatalf("error = %v, want Prompt blocked", err)
	}
}

// =============================================================================
// Post-Hook Helpers Tests
// =============================================================================

func TestPostRunCommandOK(t *testing.T) {
	output := PostRunCommandOK()
	// Just verify it returns without error (empty struct)
	_ = output
}

func TestPostWriteCodeOK(t *testing.T) {
	output := PostWriteCodeOK()
	_ = output
}

func TestPostReadCodeOK(t *testing.T) {
	output := PostReadCodeOK()
	_ = output
}

func TestPostMCPToolUseOK(t *testing.T) {
	output := PostMCPToolUseOK()
	_ = output
}

func TestPostCascadeResponseOK(t *testing.T) {
	output := PostCascadeResponseOK()
	_ = output
}

func TestPostSetupWorktreeOK(t *testing.T) {
	output := PostSetupWorktreeOK()
	_ = output
}
