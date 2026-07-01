package cascade

import "errors"

// =============================================================================
// Pre-Hook Decision Helpers (shared by PreRunCommand, PreWriteCode, etc.)
// =============================================================================

// Allow permits the action to proceed.
func Allow() BaseOutput {
	return BaseOutput{
		Decision: "allow",
	}
}

// Deny returns deny JSON with a message. Cascade pre-hooks must return this
// from RunE with a non-nil error to block the action.
func Deny(message string) BaseOutput {
	return BaseOutput{
		Decision: "deny",
		Message:  message,
	}
}

// Block returns a deny output plus the error RunE needs to make Cascade
// block the pre-hook with exit code 2.
func Block(message string) (BaseOutput, error) {
	return Deny(message), errors.New(message)
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

// DenyCommand returns deny JSON for a command.
func DenyCommand(message string) PreRunCommandOutput {
	return PreRunCommandOutput{
		BaseOutput: Deny(message),
	}
}

// BlockCommand blocks the command from executing when returned from RunE.
func BlockCommand(message string) (PreRunCommandOutput, error) {
	return DenyCommand(message), errors.New(message)
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

// DenyWrite returns deny JSON for a file write.
func DenyWrite(message string) PreWriteCodeOutput {
	return PreWriteCodeOutput{
		BaseOutput: Deny(message),
	}
}

// BlockWrite blocks the file write when returned from RunE.
func BlockWrite(message string) (PreWriteCodeOutput, error) {
	return DenyWrite(message), errors.New(message)
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

// DenyRead returns deny JSON for a file read.
func DenyRead(message string) PreReadCodeOutput {
	return PreReadCodeOutput{
		BaseOutput: Deny(message),
	}
}

// BlockRead blocks the file read when returned from RunE.
func BlockRead(message string) (PreReadCodeOutput, error) {
	return DenyRead(message), errors.New(message)
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

// DenyMCP returns deny JSON for an MCP tool.
func DenyMCP(message string) PreMCPToolUseOutput {
	return PreMCPToolUseOutput{
		BaseOutput: Deny(message),
	}
}

// BlockMCP blocks the MCP tool when returned from RunE.
func BlockMCP(message string) (PreMCPToolUseOutput, error) {
	return DenyMCP(message), errors.New(message)
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

// BlockPromptWithError blocks the user prompt when returned from RunE.
func BlockPromptWithError(message string) (PreUserPromptOutput, error) {
	return BlockPrompt(message), errors.New(message)
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
