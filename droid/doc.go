// Package droid provides types and helpers for Factory Droid hooks.
//
// Factory Droid hooks allow you to observe, control, and extend the AI agent loop
// using custom scripts. Hooks are spawned processes that communicate over stdio
// using JSON. Droid uses the same event model as Claude Code.
//
// # Hook Events
//
// Droid supports the following hook events:
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
// Stop hooks control whether Droid should stop or continue working:
//
//	hookshot.Run(func(input droid.StopInput) droid.StopOutput {
//	    // IMPORTANT: Always check StopHookActive to prevent infinite loops
//	    if input.StopHookActive {
//	        return droid.Continue()
//	    }
//
//	    if hasUnresolvedIssues() {
//	        return droid.Block("Please fix the security vulnerabilities first")
//	    }
//	    return droid.Continue()
//	})
//
// # PreToolUse Hooks
//
// PreToolUse hooks control tool execution:
//
//	hookshot.Run(func(input droid.PreToolUseInput) droid.PreToolUseOutput {
//	    // Block specific MCP servers
//	    if strings.HasPrefix(input.ToolName, "mcp__blocked__") {
//	        return droid.Deny("This MCP server is not allowed")
//	    }
//
//	    // Auto-approve safe operations
//	    if input.ToolName == "Read" {
//	        return droid.AllowSilent()
//	    }
//
//	    // Let normal flow proceed
//	    return droid.PassThrough()
//	})
//
// # UserPromptSubmit Hooks
//
// UserPromptSubmit hooks validate or augment user prompts:
//
//	hookshot.Run(func(input droid.UserPromptSubmitInput) droid.UserPromptSubmitOutput {
//	    // Block sensitive content
//	    if containsSecrets(input.Prompt) {
//	        return droid.BlockPrompt("Please don't include secrets in prompts")
//	    }
//
//	    // Add context to all prompts
//	    return droid.AddContext("Project uses Go 1.21+")
//	})
//
// # SessionStart Hooks
//
// SessionStart hooks can inject context at the beginning of a session:
//
//	hookshot.Run(func(input droid.SessionStartInput) droid.SessionStartOutput {
//	    context := loadProjectContext()
//	    return droid.SessionStartContext(context)
//	})
//
// # Helper Functions
//
// This package provides helper functions for common responses:
//
// Stop/SubagentStop:
//   - [Continue]: Allow stopping
//   - [Block]: Prevent stopping with a reason
//   - [StopWith]: Halt Droid entirely
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
//   - [DenyPermissionAndStop]: Reject and stop Droid
//
// PostToolUse:
//   - [PostToolOK]: Normal flow
//   - [PostToolBlock]: Send feedback to Droid
//   - [PostToolContext]: Add context for Droid
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
// Configure Droid hooks in ~/.factory/settings.json:
//
//	{
//	  "hooks": {
//	    "Stop": [{
//	      "hooks": [{ "type": "command", "command": "/path/to/hooks droid-stop" }]
//	    }],
//	    "PreToolUse": [{
//	      "matcher": "mcp__.*",
//	      "hooks": [{ "type": "command", "command": "/path/to/hooks droid-pre-tool" }]
//	    }]
//	  }
//	}
//
// See Factory Droid documentation for full details.
package droid
