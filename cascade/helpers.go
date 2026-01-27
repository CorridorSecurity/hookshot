package cascade

// =============================================================================
// Pre-Hook Decision Helpers (shared by PreRunCommand, PreWriteCode, etc.)
// =============================================================================

// Allow permits the action to proceed.
func Allow() BaseOutput {
	return BaseOutput{
		Decision: "allow",
	}
}

// Deny blocks the action with a message.
func Deny(message string) BaseOutput {
	return BaseOutput{
		Decision: "deny",
		Message:  message,
	}
}

// Ask prompts the user to confirm the action.
func Ask(message string) BaseOutput {
	return BaseOutput{
		Decision: "ask",
		Message:  message,
	}
}

// =============================================================================
// PreRunCommand Helpers
// =============================================================================

// AllowCommand permits the command to execute.
func AllowCommand() PreRunCommandOutput {
	return PreRunCommandOutput{
		BaseOutput: Allow(),
	}
}

// DenyCommand blocks the command from executing.
func DenyCommand(message string) PreRunCommandOutput {
	return PreRunCommandOutput{
		BaseOutput: Deny(message),
	}
}

// AskCommand prompts the user to confirm the command.
func AskCommand(message string) PreRunCommandOutput {
	return PreRunCommandOutput{
		BaseOutput: Ask(message),
	}
}

// =============================================================================
// PreWriteCode Helpers
// =============================================================================

// AllowWrite permits the file write.
func AllowWrite() PreWriteCodeOutput {
	return PreWriteCodeOutput{
		BaseOutput: Allow(),
	}
}

// DenyWrite blocks the file write.
func DenyWrite(message string) PreWriteCodeOutput {
	return PreWriteCodeOutput{
		BaseOutput: Deny(message),
	}
}

// AskWrite prompts the user to confirm the write.
func AskWrite(message string) PreWriteCodeOutput {
	return PreWriteCodeOutput{
		BaseOutput: Ask(message),
	}
}

// =============================================================================
// PreReadCode Helpers
// =============================================================================

// AllowRead permits the file read.
func AllowRead() PreReadCodeOutput {
	return PreReadCodeOutput{
		BaseOutput: Allow(),
	}
}

// DenyRead blocks the file read.
func DenyRead(message string) PreReadCodeOutput {
	return PreReadCodeOutput{
		BaseOutput: Deny(message),
	}
}

// AskRead prompts the user to confirm the read.
func AskRead(message string) PreReadCodeOutput {
	return PreReadCodeOutput{
		BaseOutput: Ask(message),
	}
}

// =============================================================================
// PreMCPToolUse Helpers
// =============================================================================

// AllowMCP permits the MCP tool to execute.
func AllowMCP() PreMCPToolUseOutput {
	return PreMCPToolUseOutput{
		BaseOutput: Allow(),
	}
}

// DenyMCP blocks the MCP tool from executing.
func DenyMCP(message string) PreMCPToolUseOutput {
	return PreMCPToolUseOutput{
		BaseOutput: Deny(message),
	}
}

// AskMCP prompts the user to confirm the MCP tool.
func AskMCP(message string) PreMCPToolUseOutput {
	return PreMCPToolUseOutput{
		BaseOutput: Ask(message),
	}
}

// =============================================================================
// PreUserPrompt Helpers
// =============================================================================

// AllowPrompt allows the prompt to be processed.
func AllowPrompt() PreUserPromptOutput {
	return PreUserPromptOutput{
		BaseOutput: Allow(),
	}
}

// BlockPrompt prevents the prompt from being processed.
func BlockPrompt(message string) PreUserPromptOutput {
	return PreUserPromptOutput{
		BaseOutput: Deny(message),
	}
}

// =============================================================================
// Post-Hook Helpers (all are fire-and-forget)
// =============================================================================

// PostRunCommandOK returns an empty output for post-run-command.
func PostRunCommandOK() PostRunCommandOutput {
	return PostRunCommandOutput{}
}

// PostWriteCodeOK returns an empty output for post-write-code.
func PostWriteCodeOK() PostWriteCodeOutput {
	return PostWriteCodeOutput{}
}

// PostReadCodeOK returns an empty output for post-read-code.
func PostReadCodeOK() PostReadCodeOutput {
	return PostReadCodeOutput{}
}

// PostMCPToolUseOK returns an empty output for post-mcp-tool-use.
func PostMCPToolUseOK() PostMCPToolUseOutput {
	return PostMCPToolUseOutput{}
}

// PostCascadeResponseOK returns an empty output for post-cascade-response.
func PostCascadeResponseOK() PostCascadeResponseOutput {
	return PostCascadeResponseOutput{}
}

// PostSetupWorktreeOK returns an empty output for post-setup-worktree.
func PostSetupWorktreeOK() PostSetupWorktreeOutput {
	return PostSetupWorktreeOutput{}
}
