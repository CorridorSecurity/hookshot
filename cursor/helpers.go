package cursor

// =============================================================================
// Stop Helpers
// =============================================================================

// Continue allows the agent to stop normally.
func Continue() StopOutput {
	return StopOutput{}
}

// Followup triggers an automatic follow-up message.
// Note: Maximum 5 auto follow-ups are enforced by Cursor.
func Followup(message string) StopOutput {
	return StopOutput{
		FollowupMessage: message,
	}
}

// =============================================================================
// Before Execution Helpers (Shell/MCP)
// =============================================================================

// Allow permits the execution without asking the user.
func Allow() BeforeExecutionOutput {
	return BeforeExecutionOutput{
		Permission: "allow",
	}
}

// AllowWithMessage permits execution and shows a message.
func AllowWithMessage(userMsg string) BeforeExecutionOutput {
	return BeforeExecutionOutput{
		Permission:  "allow",
		UserMessage: userMsg,
	}
}

// Deny blocks the execution.
func Deny(userMsg, agentMsg string) BeforeExecutionOutput {
	return BeforeExecutionOutput{
		Permission:   "deny",
		UserMessage:  userMsg,
		AgentMessage: agentMsg,
	}
}

// DenyWithUserMessage blocks execution with a message shown to the user.
func DenyWithUserMessage(msg string) BeforeExecutionOutput {
	return BeforeExecutionOutput{
		Permission:  "deny",
		UserMessage: msg,
	}
}

// DenyWithAgentMessage blocks execution with a message sent to the agent.
func DenyWithAgentMessage(msg string) BeforeExecutionOutput {
	return BeforeExecutionOutput{
		Permission:   "deny",
		AgentMessage: msg,
	}
}

// Ask prompts the user to confirm the execution.
func Ask(userMsg string) BeforeExecutionOutput {
	return BeforeExecutionOutput{
		Permission:  "ask",
		UserMessage: userMsg,
	}
}

// =============================================================================
// Before Read File Helpers
// =============================================================================

// AllowRead permits the file read.
func AllowRead() BeforeReadFileOutput {
	return BeforeReadFileOutput{
		Permission: "allow",
	}
}

// DenyRead blocks the file read.
func DenyRead(userMsg, agentMsg string) BeforeReadFileOutput {
	return BeforeReadFileOutput{
		Permission:   "deny",
		UserMessage:  userMsg,
		AgentMessage: agentMsg,
	}
}

// AskRead prompts the user to confirm the file read.
func AskRead(userMsg string) BeforeReadFileOutput {
	return BeforeReadFileOutput{
		Permission:  "ask",
		UserMessage: userMsg,
	}
}

// =============================================================================
// Tab File Helpers
// =============================================================================

// AllowTabRead permits Tab to read the file.
func AllowTabRead() BeforeTabFileReadOutput {
	return BeforeTabFileReadOutput{
		Permission: "allow",
	}
}

// DenyTabRead blocks Tab from reading the file.
func DenyTabRead() BeforeTabFileReadOutput {
	return BeforeTabFileReadOutput{
		Permission: "deny",
	}
}

// =============================================================================
// Before Submit Prompt Helpers
// =============================================================================

// AllowPrompt allows the prompt to be submitted.
func AllowPrompt() BeforeSubmitPromptOutput {
	return BeforeSubmitPromptOutput{
		Continue: true,
	}
}

// BlockPrompt prevents the prompt from being submitted.
func BlockPrompt(userMsg string) BeforeSubmitPromptOutput {
	return BeforeSubmitPromptOutput{
		Continue:    false,
		UserMessage: userMsg,
	}
}

// =============================================================================
// After Execution Helpers (for completeness)
// =============================================================================

// AfterShellOK returns an empty output for after-shell-execution.
func AfterShellOK() AfterShellExecutionOutput {
	return AfterShellExecutionOutput{}
}

// AfterMCPOK returns an empty output for after-mcp-execution.
func AfterMCPOK() AfterMCPExecutionOutput {
	return AfterMCPExecutionOutput{}
}

// AfterFileEditOK returns an empty output for after-file-edit.
func AfterFileEditOK() AfterFileEditOutput {
	return AfterFileEditOutput{}
}

// AfterTabFileEditOK returns an empty output for after-tab-file-edit.
func AfterTabFileEditOK() AfterTabFileEditOutput {
	return AfterTabFileEditOutput{}
}

// AfterAgentResponseOK returns an empty output for after-agent-response.
func AfterAgentResponseOK() AfterAgentResponseOutput {
	return AfterAgentResponseOutput{}
}

// AfterAgentThoughtOK returns an empty output for after-agent-thought.
func AfterAgentThoughtOK() AfterAgentThoughtOutput {
	return AfterAgentThoughtOutput{}
}
