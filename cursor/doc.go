// Package cursor provides types and helpers for Cursor hooks.
//
// Cursor hooks let you observe, control, and extend the agent loop using custom
// scripts. Hooks are spawned processes that communicate over stdio using JSON.
// They work with both Cursor Agent (Cmd+K/Agent Chat) and Cursor Tab (inline
// completions).
//
// # Hook Events
//
// Cursor supports the following hook events:
//
// Agent hooks (Cmd+K/Agent Chat):
//   - [BeforeShellExecutionInput]/[BeforeExecutionOutput]: Before shell command
//   - [AfterShellExecutionInput]/[AfterShellExecutionOutput]: After shell command
//   - [BeforeMCPExecutionInput]/[BeforeExecutionOutput]: Before MCP tool
//   - [AfterMCPExecutionInput]/[AfterMCPExecutionOutput]: After MCP tool
//   - [BeforeReadFileInput]/[BeforeReadFileOutput]: Before agent reads file
//   - [AfterFileEditInput]/[AfterFileEditOutput]: After agent edits file
//   - [BeforeSubmitPromptInput]/[BeforeSubmitPromptOutput]: Before prompt submission
//   - [StopInput]/[StopOutput]: Agent loop ended
//   - [AfterAgentResponseInput]/[AfterAgentResponseOutput]: After agent message
//   - [AfterAgentThoughtInput]/[AfterAgentThoughtOutput]: After thinking block
//
// Tab hooks (inline completions):
//   - [BeforeTabFileReadInput]/[BeforeTabFileReadOutput]: Before Tab reads file
//   - [AfterTabFileEditInput]/[AfterTabFileEditOutput]: After Tab edits file
//
// # Stop Hooks
//
// Stop hooks can trigger automatic follow-up messages:
//
//	hookshot.Run(func(input cursor.StopInput) cursor.StopOutput {
//	    // Limit auto follow-ups (max 5 enforced by Cursor)
//	    if input.LoopCount >= 3 {
//	        return cursor.Continue()
//	    }
//
//	    // Request verification on first completion
//	    if input.Status == "completed" && input.LoopCount == 0 {
//	        return cursor.Followup("Please verify the changes")
//	    }
//
//	    return cursor.Continue()
//	})
//
// # Before Execution Hooks
//
// BeforeShellExecution and BeforeMCPExecution hooks control command execution:
//
//	hookshot.Run(func(input cursor.BeforeShellExecutionInput) cursor.BeforeExecutionOutput {
//	    // Block dangerous commands
//	    if strings.Contains(input.Command, "rm -rf /") {
//	        return cursor.Deny(
//	            "Command blocked for safety",     // shown to user
//	            "Dangerous command not allowed",  // sent to agent
//	        )
//	    }
//
//	    // Auto-approve safe commands
//	    if strings.HasPrefix(input.Command, "ls ") {
//	        return cursor.Allow()
//	    }
//
//	    // Prompt user for confirmation
//	    return cursor.Ask("Confirm execution?")
//	})
//
// # File Access Hooks
//
// Control file reading and react to file edits:
//
//	hookshot.Run(func(input cursor.BeforeReadFileInput) cursor.BeforeReadFileOutput {
//	    // Block reading sensitive files
//	    if strings.Contains(input.FilePath, ".env") {
//	        return cursor.DenyRead(
//	            "Cannot read environment files",
//	            "File access blocked: sensitive data",
//	        )
//	    }
//	    return cursor.AllowRead()
//	})
//
// # Prompt Validation
//
// Validate prompts before submission:
//
//	hookshot.Run(func(input cursor.BeforeSubmitPromptInput) cursor.BeforeSubmitPromptOutput {
//	    if len(input.Prompt) > 10000 {
//	        return cursor.BlockPrompt("Prompt too long")
//	    }
//	    return cursor.AllowPrompt()
//	})
//
// # Helper Functions
//
// This package provides helper functions for common responses:
//
// Stop:
//   - [Continue]: Allow agent to stop
//   - [Followup]: Trigger automatic follow-up message
//
// Before execution (Shell/MCP):
//   - [Allow]: Permit execution
//   - [AllowWithMessage]: Permit with user message
//   - [Deny]: Block execution
//   - [DenyWithUserMessage]: Block with user message
//   - [DenyWithAgentMessage]: Block with agent message
//   - [Ask]: Prompt user to confirm
//
// File read:
//   - [AllowRead]: Permit file read
//   - [DenyRead]: Block file read
//   - [AskRead]: Prompt user to confirm
//
// Tab file:
//   - [AllowTabRead]: Permit Tab to read
//   - [DenyTabRead]: Block Tab from reading
//
// Prompt:
//   - [AllowPrompt]: Submit prompt
//   - [BlockPrompt]: Block submission
//
// After hooks (for completeness):
//   - [AfterShellOK], [AfterMCPOK], [AfterFileEditOK], etc.
//
// # Configuration
//
// Configure Cursor hooks in ~/.cursor/hooks.json:
//
//	{
//	  "version": 1,
//	  "hooks": {
//	    "stop": [{ "command": "/path/to/hooks cursor-stop" }],
//	    "beforeShellExecution": [{ "command": "/path/to/hooks cursor-shell" }],
//	    "beforeMCPExecution": [{ "command": "/path/to/hooks cursor-mcp" }]
//	  }
//	}
//
// See https://cursor.com/docs/agent/hooks for full documentation.
package cursor
