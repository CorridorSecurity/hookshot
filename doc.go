// Package hookshot provides a framework for building hooks for AI coding agents
// like Cursor and Claude Code.
//
// # Quick Start
//
// Use the unified handlers to write cross-platform hooks:
//
//	package main
//
//	import "github.com/CorridorSecurity/hookshot"
//
//	func main() {
//	    hookshot.OnStop(func(ctx hookshot.StopContext) hookshot.StopDecision {
//	        if ctx.ShouldSkip() {
//	            return hookshot.AllowStop()
//	        }
//	        return hookshot.PreventStop("Please verify changes")
//	    })
//	    hookshot.RunCommand()
//	}
//
// # Platform-Specific Handlers
//
// For platform-specific features, use [Register] with [Run]:
//
//	func main() {
//	    // Unified handlers
//	    hookshot.OnStop(handleStop)
//
//	    // Platform-specific: Cursor Tab completion (no Claude equivalent)
//	    hookshot.Register("cursor-before-tab-read", func() {
//	        hookshot.Run(func(input cursor.BeforeTabFileReadInput) cursor.BeforeTabFileReadOutput {
//	            return cursor.AllowTabRead()
//	        })
//	    })
//
//	    hookshot.RunCommand()
//	}
//
// Run with: ./my-hooks claude-stop or ./my-hooks cursor-before-tab-read
//
// # Unified API
//
// For cross-platform handlers, use the unified API which works across both Claude Code
// and Cursor with a single handler:
//
//	hookshot.OnStop(func(ctx hookshot.StopContext) hookshot.StopDecision {
//	    if ctx.ShouldSkip() {
//	        return hookshot.AllowStop()
//	    }
//	    return hookshot.PreventStop("Please verify changes")
//	})
//
//	hookshot.OnBeforeExecution(func(ctx hookshot.ExecutionContext) hookshot.ExecutionDecision {
//	    if ctx.Type == hookshot.ExecutionShell && strings.Contains(ctx.Command, "rm -rf") {
//	        return hookshot.DenyExecution("Dangerous command")
//	    }
//	    return hookshot.AllowExecution()
//	})
//
// Unified handlers:
//   - [OnStop]: Stop hooks for both platforms
//   - [OnBeforeExecution]: Shell/MCP execution for both platforms
//   - [OnAfterFileEdit]: File edit hooks for both platforms
//   - [OnPromptSubmit]: Prompt submission for both platforms
//   - [OnSessionStart]: Session start (Claude Code only)
//
// # Error Handling
//
// Use [RunE] when your handler can fail:
//
//	hookshot.RunE(func(input claude.PreToolUseInput) (claude.PreToolUseOutput, error) {
//	    if err := validate(input); err != nil {
//	        return claude.PreToolUseOutput{}, err // Exits with code 2
//	    }
//	    return claude.Allow("Validated"), nil
//	})
//
// # Building and Installing
//
// Use the hookshot CLI for building and installing:
//
//	go install github.com/CorridorSecurity/hookshot/cmd/hookshot@latest
//	hookshot build -all -output ./dist
//	hookshot install --binary ./dist/darwin-arm64/my-hooks
//
// # Configuration
//
// Configure hooks in Claude Code (~/.claude/settings.json):
//
//	{
//	  "hooks": {
//	    "Stop": [{
//	      "hooks": [{ "type": "command", "command": "/path/to/my-hooks claude-stop" }]
//	    }]
//	  }
//	}
//
// Configure hooks in Cursor (~/.cursor/hooks.json):
//
//	{
//	  "version": 1,
//	  "hooks": {
//	    "stop": [{ "command": "/path/to/my-hooks cursor-stop" }]
//	  }
//	}
//
// # Packages
//
// The hookshot module consists of:
//
//   - hookshot (this package): Core Run/Register/RunCommand functions
//   - hookshot/claude: Types and helpers for Claude Code hooks
//   - hookshot/cursor: Types and helpers for Cursor hooks
//   - hookshot/build: Cross-platform build tool
//   - hookshot/internal: Internal JSON I/O (not for external use)
//
// See the sub-packages for platform-specific documentation.
package hookshot
