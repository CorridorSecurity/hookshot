// Package claude provides types and helpers for Claude Code hooks.
//
// Claude Code hooks allow you to observe, control, and extend the AI agent loop
// using custom scripts. Hooks are spawned processes that communicate over stdio
// using JSON.
//
// # Hook Events
//
// Claude Code supports the following hook events:
//
//   - [StopInput]/[StopOutput]: Agent finished responding
//   - [SubagentStopInput]/[SubagentStopOutput]: Subagent (Task tool) finished
//   - [SessionStartInput]/[SessionStartOutput]: Session started or resumed
//   - [SessionEndInput]/[SessionEndOutput]: Session ending
//   - [PreToolUseInput]/[PreToolUseOutput]: Before tool execution
//   - [PostToolUseInput]/[PostToolUseOutput]: After tool execution
//   - [PermissionRequestInput]/[PermissionRequestOutput]: Permission dialog shown
//   - [UserPromptSubmitInput]/[UserPromptSubmitOutput]: User submitted a prompt
//   - [NotificationInput]/[NotificationOutput]: Notification sent
//   - [PreCompactInput]/[PreCompactOutput]: Before context compaction
//
// # Stop Hooks
//
// Stop hooks control whether Claude should stop or continue working:
//
//	hookshot.Run(func(input claude.StopInput) claude.StopOutput {
//	    // IMPORTANT: Always check StopHookActive to prevent infinite loops
//	    if input.StopHookActive {
//	        return claude.Continue()
//	    }
//
//	    if hasUnresolvedIssues() {
//	        return claude.Block("Please fix the security vulnerabilities first")
//	    }
//	    return claude.Continue()
//	})
//
// # PreToolUse Hooks
//
// PreToolUse hooks control tool execution:
//
//	hookshot.Run(func(input claude.PreToolUseInput) claude.PreToolUseOutput {
//	    // Block specific MCP servers
//	    if strings.HasPrefix(input.ToolName, "mcp__blocked__") {
//	        return claude.Deny("This MCP server is not allowed")
//	    }
//
//	    // Auto-approve safe operations
//	    if input.ToolName == "Read" {
//	        return claude.AllowSilent()
//	    }
//
//	    // Let normal flow proceed
//	    return claude.PassThrough()
//	})
//
// # UserPromptSubmit Hooks
//
// UserPromptSubmit hooks validate or augment user prompts:
//
//	hookshot.Run(func(input claude.UserPromptSubmitInput) claude.UserPromptSubmitOutput {
//	    // Block sensitive content
//	    if containsSecrets(input.Prompt) {
//	        return claude.BlockPrompt("Please don't include secrets in prompts")
//	    }
//
//	    // Add context to all prompts
//	    return claude.AddContext("Project uses Go 1.21+")
//	})
//
// # SessionStart Hooks
//
// SessionStart hooks can inject context at the beginning of a session:
//
//	hookshot.Run(func(input claude.SessionStartInput) claude.SessionStartOutput {
//	    context := loadProjectContext()
//	    return claude.SessionStartContext(context)
//	})
//
// # Helper Functions
//
// This package provides helper functions for common responses:
//
// Stop/SubagentStop:
//   - [Continue]: Allow stopping
//   - [Block]: Prevent stopping with a reason
//   - [StopWith]: Halt Claude entirely
//
// PreToolUse:
//   - [Allow]: Permit execution with a reason
//   - [AllowSilent]: Permit without showing output
//   - [AllowWithInput]: Permit with modified tool input
//   - [Deny]: Block execution with a reason
//   - [Ask]: Prompt user to confirm
//   - [PassThrough]: Let normal permission flow proceed
//
// PermissionRequest:
//   - [AllowPermission]: Grant the permission
//   - [AllowPermissionWithInput]: Grant with modified input
//   - [DenyPermission]: Reject with message
//   - [DenyPermissionAndStop]: Reject and stop Claude
//
// PostToolUse:
//   - [PostToolOK]: Normal flow
//   - [PostToolBlock]: Send feedback to Claude
//   - [PostToolContext]: Add context for Claude
//
// UserPromptSubmit:
//   - [AllowPrompt]: Process normally
//   - [BlockPrompt]: Reject with reason
//   - [AddContext]: Add context to the prompt
//
// SessionStart:
//   - [SessionStartOK]: Normal flow
//   - [SessionStartContext]: Add session context
//
// # Configuration
//
// Configure Claude Code hooks in ~/.claude/settings.json:
//
//	{
//	  "hooks": {
//	    "Stop": [{
//	      "hooks": [{ "type": "command", "command": "/path/to/hooks claude-stop" }]
//	    }],
//	    "PreToolUse": [{
//	      "matcher": "mcp__.*",
//	      "hooks": [{ "type": "command", "command": "/path/to/hooks claude-pre-tool" }]
//	    }]
//	  }
//	}
//
// See https://docs.claude.com/en/docs/claude-code/hooks for full documentation.
package claude
