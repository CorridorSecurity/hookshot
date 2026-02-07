// Example multi-hook binary demonstrating hookshot's patterns.
//
// Build:
//
//	go build -o my-hooks .
//
// Cross-platform build:
//
//	go run github.com/CorridorSecurity/hookshot/build -all -output ./dist -name my-hooks
//
// Configure in Claude Code (~/.claude/settings.json):
//
//	{
//	  "hooks": {
//	    "Stop": [{ "hooks": [{ "type": "command", "command": "/path/to/my-hooks claude-stop" }] }],
//	    "PreToolUse": [{ "matcher": "*", "hooks": [{ "type": "command", "command": "/path/to/my-hooks claude-pre-tool-use" }] }],
//	    "PostToolUse": [{ "matcher": "Write|Edit", "hooks": [{ "type": "command", "command": "/path/to/my-hooks claude-after-file-edit" }] }],
//	    "UserPromptSubmit": [{ "hooks": [{ "type": "command", "command": "/path/to/my-hooks claude-user-prompt-submit" }] }],
//	    "SessionStart": [{ "hooks": [{ "type": "command", "command": "/path/to/my-hooks claude-session-start" }] }]
//	  }
//	}
//
// Configure in Cursor (~/.cursor/hooks.json):
//
//	{
//	  "version": 1,
//	  "hooks": {
//	    "stop": [{ "command": "/path/to/my-hooks cursor-stop" }],
//	    "beforeShellExecution": [{ "command": "/path/to/my-hooks cursor-before-shell" }],
//	    "beforeMCPExecution": [{ "command": "/path/to/my-hooks cursor-before-mcp" }],
//	    "afterFileEdit": [{ "command": "/path/to/my-hooks cursor-after-file-edit" }],
//	    "beforeSubmitPrompt": [{ "command": "/path/to/my-hooks cursor-before-submit-prompt" }],
//	    "beforeTabFileRead": [{ "command": "/path/to/my-hooks cursor-before-tab-read" }]
//	  }
//	}
//
// Configure in Windsurf Cascade (hooks.json in workspace):
//
//	{
//	  "hooks": {
//	    "pre-run-command": { "command": "/path/to/my-hooks cascade-pre-run-command" },
//	    "pre-write-code": { "command": "/path/to/my-hooks cascade-pre-write-code" },
//	    "pre-user-prompt": { "command": "/path/to/my-hooks cascade-pre-user-prompt" }
//	  }
//	}
//
// Configure in Factory Droid (similar to Claude Code):
//
//	{
//	  "hooks": {
//	    "Stop": [{ "command": "/path/to/my-hooks droid-stop" }],
//	    "PreToolUse": [{ "matcher": "*", "command": "/path/to/my-hooks droid-pre-tool-use" }],
//	    "UserPromptSubmit": [{ "command": "/path/to/my-hooks droid-user-prompt-submit" }]
//	  }
//	}
package main

import (
	"fmt"
	"strings"

	"github.com/CorridorSecurity/hookshot"
	"github.com/CorridorSecurity/hookshot/agenttrace"
	"github.com/CorridorSecurity/hookshot/cascade"
	"github.com/CorridorSecurity/hookshot/claude"
	"github.com/CorridorSecurity/hookshot/cursor"
	"github.com/CorridorSecurity/hookshot/droid"
)

// traceRecorder accumulates AI file edit events for agent-trace attribution.
var traceRecorder = agenttrace.NewRecorder(&agenttrace.Tool{Name: "hookshot-example"})

func main() {
	// ==========================================================================
	// UNIFIED HANDLERS
	// Write once, works on both Claude Code and Cursor automatically.
	// ==========================================================================

	hookshot.OnStop(handleStop)
	hookshot.OnBeforeExecution(handleBeforeExecution)
	hookshot.OnAfterFileEdit(handleAfterFileEdit)
	hookshot.OnPromptSubmit(handlePromptSubmit)

	// ==========================================================================
	// PLATFORM-SPECIFIC HANDLERS
	// Use these for features that only exist on one platform.
	// ==========================================================================

	// Claude Code only: SessionStart (Cursor has no equivalent)
	hookshot.Register("claude-session-start", handleClaudeSessionStart)

	// Cursor only: Tab completion file read (Claude Code has no equivalent)
	hookshot.Register("cursor-before-tab-read", handleCursorBeforeTabRead)

	// ==========================================================================
	// WINDSURF CASCADE HANDLERS
	// ==========================================================================

	hookshot.Register("cascade-pre-run-command", handleCascadePreRunCommand)
	hookshot.Register("cascade-pre-write-code", handleCascadePreWriteCode)
	hookshot.Register("cascade-pre-user-prompt", handleCascadePreUserPrompt)

	// ==========================================================================
	// FACTORY DROID HANDLERS
	// ==========================================================================

	hookshot.Register("droid-stop", handleDroidStop)
	hookshot.Register("droid-pre-tool-use", handleDroidPreToolUse)
	hookshot.Register("droid-user-prompt-submit", handleDroidUserPromptSubmit)

	hookshot.RunCommand()
}

// =============================================================================
// UNIFIED HANDLERS
// =============================================================================

func handleStop(ctx hookshot.StopContext) hookshot.StopDecision {
	// Prevent infinite loops
	if ctx.ShouldSkip() {
		return hookshot.AllowStop()
	}

	// Example: Request verification on Cursor's first completion
	if ctx.Platform == hookshot.PlatformCursor {
		if ctx.Status == "completed" && ctx.LoopCount == 0 {
			return hookshot.PreventStop("Please verify the changes are correct")
		}
	}

	return hookshot.AllowStop()
}

func handleBeforeExecution(ctx hookshot.ExecutionContext) hookshot.ExecutionDecision {
	// Block dangerous shell commands
	if ctx.Type == hookshot.ExecutionShell {
		if strings.Contains(ctx.Command, "rm -rf /") {
			return hookshot.DenyExecution("Dangerous command blocked")
		}
	}

	// Block specific MCP servers
	if ctx.IsMCP() {
		if strings.HasPrefix(ctx.ToolName, "mcp__blocked__") {
			return hookshot.DenyExecution("This MCP server is not allowed")
		}
	}

	return hookshot.AllowExecution()
}

func handleAfterFileEdit(ctx hookshot.FileEditContext) hookshot.FileEditDecision {
	// Log file edits
	fmt.Printf("File edited: %s\n", ctx.FilePath)

	// Record the edit for agent-trace attribution
	traceRecorder.Record(ctx.FilePath, "",
		agenttrace.Contributor{Type: agenttrace.ContributorAI},
		nil, // no line-level ranges in this example
	)

	// Claude Code: Add context if TODO found
	if ctx.Platform == hookshot.PlatformClaude {
		for _, edit := range ctx.Edits {
			if strings.Contains(edit.NewString, "TODO") {
				return hookshot.FileEditAddContext("File contains TODO comments")
			}
		}
	}

	return hookshot.FileEditOK()
}

func handlePromptSubmit(ctx hookshot.PromptContext) hookshot.PromptDecision {
	// Block prompts with API keys
	if strings.Contains(strings.ToLower(ctx.Prompt), "api_key=") {
		return hookshot.BlockPromptDecision("Don't include API keys in prompts")
	}

	// Claude Code: Add project context
	if ctx.Platform == hookshot.PlatformClaude {
		return hookshot.AddPromptContext("Project uses Go 1.21+")
	}

	return hookshot.AllowPromptDecision()
}

// =============================================================================
// PLATFORM-SPECIFIC: Claude Code SessionStart
// =============================================================================

func handleClaudeSessionStart() {
	hookshot.Run(func(input claude.SessionStartInput) claude.SessionStartOutput {
		// Inject context at session start
		context := fmt.Sprintf("Session started via %s in %s", input.Source, input.Cwd)
		return claude.SessionStartContext(context)
	})
}

// =============================================================================
// PLATFORM-SPECIFIC: Cursor Tab Completion
// =============================================================================

func handleCursorBeforeTabRead() {
	hookshot.Run(func(input cursor.BeforeTabFileReadInput) cursor.BeforeTabFileReadOutput {
		// Block Tab from reading sensitive files
		if strings.Contains(strings.ToLower(input.FilePath), ".env") {
			return cursor.DenyTabRead()
		}
		return cursor.AllowTabRead()
	})
}

// =============================================================================
// PLATFORM-SPECIFIC: Windsurf Cascade
// Cascade uses exit code 2 to block actions, so we use RunE with errors
// =============================================================================

func handleCascadePreRunCommand() {
	hookshot.RunE(func(input cascade.PreRunCommandInput) (cascade.PreRunCommandOutput, error) {
		// Block dangerous shell commands
		if strings.Contains(input.ToolInfo.CommandLine, "rm -rf /") {
			return cascade.PreRunCommandOutput{}, fmt.Errorf("Dangerous command blocked")
		}
		return cascade.AllowCommand(), nil
	})
}

func handleCascadePreWriteCode() {
	hookshot.RunE(func(input cascade.PreWriteCodeInput) (cascade.PreWriteCodeOutput, error) {
		// Block writes to sensitive files
		if strings.HasSuffix(input.ToolInfo.FilePath, ".env") {
			return cascade.PreWriteCodeOutput{}, fmt.Errorf("Cannot write to .env files")
		}
		return cascade.AllowWrite(), nil
	})
}

func handleCascadePreUserPrompt() {
	hookshot.RunE(func(input cascade.PreUserPromptInput) (cascade.PreUserPromptOutput, error) {
		// Block prompts with API keys
		if strings.Contains(strings.ToLower(input.ToolInfo.Prompt), "api_key=") {
			return cascade.PreUserPromptOutput{}, fmt.Errorf("Don't include API keys in prompts")
		}
		return cascade.AllowPrompt(), nil
	})
}

// =============================================================================
// PLATFORM-SPECIFIC: Factory Droid
// =============================================================================

func handleDroidStop() {
	hookshot.Run(func(input droid.StopInput) droid.StopOutput {
		// IMPORTANT: Check StopHookActive to prevent infinite loops
		if input.StopHookActive {
			return droid.Continue()
		}
		return droid.Continue()
	})
}

func handleDroidPreToolUse() {
	hookshot.Run(func(input droid.PreToolUseInput) droid.PreToolUseOutput {
		// Block specific MCP servers
		if strings.HasPrefix(input.ToolName, "mcp__blocked__") {
			return droid.Deny("MCP server not allowed")
		}

		// Auto-approve Read tool
		if input.ToolName == "Read" {
			return droid.AllowSilent()
		}

		return droid.PassThrough()
	})
}

func handleDroidUserPromptSubmit() {
	hookshot.Run(func(input droid.UserPromptSubmitInput) droid.UserPromptSubmitOutput {
		// Block prompts with API keys
		if strings.Contains(strings.ToLower(input.Prompt), "api_key=") {
			return droid.BlockPrompt("Don't include API keys in prompts")
		}
		return droid.AllowPrompt()
	})
}
