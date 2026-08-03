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
//	hookshot.RunE(func(input cascade.PreRunCommandInput) (cascade.PreRunCommandOutput, error) {
//	    // Block dangerous commands
//	    if strings.Contains(input.ToolInfo.CommandLine, "rm -rf") {
//	        return cascade.BlockCommand("Dangerous command blocked")
//	    }
//	    return cascade.AllowCommand(), nil
//	})
//
// # PreWriteCode Hooks
//
// PreWriteCode hooks control file modifications:
//
//	hookshot.RunE(func(input cascade.PreWriteCodeInput) (cascade.PreWriteCodeOutput, error) {
//	    // Block writes to sensitive files
//	    if strings.Contains(input.ToolInfo.FilePath, ".env") {
//	        return cascade.BlockWrite("Cannot modify environment files")
//	    }
//	    return cascade.AllowWrite(), nil
//	})
//
// # PreUserPrompt Hooks
//
// PreUserPrompt hooks validate or modify user prompts:
//
//	hookshot.RunE(func(input cascade.PreUserPromptInput) (cascade.PreUserPromptOutput, error) {
//	    if containsSecrets(input.ToolInfo.UserPrompt) {
//	        return cascade.BlockPromptWithError("Please don't include secrets in prompts")
//	    }
//	    return cascade.AllowPrompt(), nil
//	})
//
// # Helper Functions
//
// This package provides helper functions for common responses:
//
// Pre-hooks (PreRunCommand, PreWriteCode, PreReadCode, PreMCPToolUse):
//   - [Allow]: Permit the action
//   - [Deny]: Build deny JSON for the action
//   - [Block]: Build deny JSON plus the error required for [hookshot.RunE]
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
