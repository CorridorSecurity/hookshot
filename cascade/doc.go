// Package cascade provides types and helpers for Windsurf Cascade hooks.
//
// Windsurf Cascade hooks allow you to observe, control, and extend the AI agent loop
// using custom scripts. Hooks are spawned processes that communicate over stdio
// using JSON.
//
// # Hook Events
//
// Cascade supports the following hook events:
//
// Pre-hooks (can block or modify):
//   - [PreReadCodeInput]/[PreReadCodeOutput]: Before reading code files
//   - [PreWriteCodeInput]/[PreWriteCodeOutput]: Before writing code files
//   - [PreRunCommandInput]/[PreRunCommandOutput]: Before running shell commands
//   - [PreMCPToolUseInput]/[PreMCPToolUseOutput]: Before MCP tool execution
//   - [PreUserPromptInput]/[PreUserPromptOutput]: Before processing user prompt
//
// Post-hooks (observe only):
//   - [PostReadCodeInput]/[PostReadCodeOutput]: After reading code files
//   - [PostWriteCodeInput]/[PostWriteCodeOutput]: After writing code files
//   - [PostRunCommandInput]/[PostRunCommandOutput]: After running shell commands
//   - [PostMCPToolUseInput]/[PostMCPToolUseOutput]: After MCP tool execution
//   - [PostCascadeResponseInput]/[PostCascadeResponseOutput]: After Cascade responds
//   - [PostSetupWorktreeInput]/[PostSetupWorktreeOutput]: After worktree setup
//
// # PreRunCommand Hooks
//
// PreRunCommand hooks control shell command execution:
//
//	hookshot.Run(func(input cascade.PreRunCommandInput) cascade.PreRunCommandOutput {
//	    // Block dangerous commands
//	    if strings.Contains(input.ToolInfo.CommandLine, "rm -rf") {
//	        return cascade.Deny("Dangerous command blocked")
//	    }
//	    return cascade.Allow()
//	})
//
// # PreWriteCode Hooks
//
// PreWriteCode hooks control file modifications:
//
//	hookshot.Run(func(input cascade.PreWriteCodeInput) cascade.PreWriteCodeOutput {
//	    // Block writes to sensitive files
//	    if strings.Contains(input.ToolInfo.FilePath, ".env") {
//	        return cascade.Deny("Cannot modify environment files")
//	    }
//	    return cascade.Allow()
//	})
//
// # PreUserPrompt Hooks
//
// PreUserPrompt hooks validate or modify user prompts:
//
//	hookshot.Run(func(input cascade.PreUserPromptInput) cascade.PreUserPromptOutput {
//	    if containsSecrets(input.ToolInfo.UserPrompt) {
//	        return cascade.BlockPrompt("Please don't include secrets in prompts")
//	    }
//	    return cascade.AllowPrompt()
//	})
//
// # Helper Functions
//
// This package provides helper functions for common responses:
//
// Pre-hooks (PreRunCommand, PreWriteCode, PreReadCode, PreMCPToolUse):
//   - [Allow]: Permit the action
//   - [Deny]: Block the action with a reason
//   - [Ask]: Prompt user to confirm (where supported)
//
// PreUserPrompt:
//   - [AllowPrompt]: Process normally
//   - [BlockPrompt]: Reject with reason
//
// Post-hooks:
//   - [PostOK]: Normal flow (all post-hooks are fire-and-forget)
//
// # Configuration
//
// Configure Cascade hooks in ~/.codeium/windsurf/hooks.json:
//
//	{
//	  "hooks": {
//	    "pre_run_command": [
//	      { "command": "/path/to/hooks cascade-pre-run-command" }
//	    ],
//	    "pre_write_code": [
//	      { "command": "/path/to/hooks cascade-pre-write-code" }
//	    ],
//	    "pre_user_prompt": [
//	      { "command": "/path/to/hooks cascade-pre-user-prompt" }
//	    ]
//	  }
//	}
//
// See Windsurf documentation for full details.
package cascade
