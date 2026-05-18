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
//
// Configure in OpenAI Codex (~/.codex/hooks.json; hooks are enabled by
// default in current Codex releases):
//
//	{
//	  "hooks": {
//	    "Stop": [{ "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-stop" }] }],
//	    "PreToolUse": [{ "matcher": "Bash|apply_patch|mcp__.*", "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-pre-tool-use" }] }],
//	    "PostToolUse": [{ "matcher": "apply_patch|mcp__.*", "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-post-tool-use" }] }],
//	    "UserPromptSubmit": [{ "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-user-prompt-submit" }] }]
//	  }
//	}
package main

import (
	"fmt"
	"strings"

	"github.com/CorridorSecurity/hookshot"
	"github.com/CorridorSecurity/hookshot/cascade"
	"github.com/CorridorSecurity/hookshot/claude"
	"github.com/CorridorSecurity/hookshot/cursor"
)

func main() {
	// ==========================================================================
	// UNIFIED HANDLERS
	// Write once, works on Claude Code, Cursor, Windsurf Cascade,
	// Factory Droid, and OpenAI Codex automatically.
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
	// WINDSURF CASCADE: pre-write-code (no unified equivalent)
	// Other Cascade hooks (pre-run-command, pre-mcp-tool-use, pre-user-prompt,
	// post-cascade-response, post-write-code) are covered by unified handlers.
	// ==========================================================================

	hookshot.Register("cascade-pre-write-code", handleCascadePreWriteCode)

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
// PLATFORM-SPECIFIC: Windsurf Cascade pre-write-code
// No unified equivalent exists for this hook.
// Cascade uses exit code 2 to block actions, so we use RunE with errors.
// =============================================================================

func handleCascadePreWriteCode() {
	hookshot.RunE(func(input cascade.PreWriteCodeInput) (cascade.PreWriteCodeOutput, error) {
		// Block writes to sensitive files
		if strings.HasSuffix(input.ToolInfo.FilePath, ".env") {
			return cascade.PreWriteCodeOutput{}, fmt.Errorf("Cannot write to .env files")
		}
		return cascade.AllowWrite(), nil
	})
}
