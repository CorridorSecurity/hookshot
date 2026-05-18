// Package codex provides types and helpers for OpenAI Codex hooks.
//
// OpenAI Codex hooks use the same JSON wire format as Claude Code hooks, so
// this package re-exports types and helpers from the claude package for
// convenience. Codex adds the apply_patch tool name to PreToolUse and
// PostToolUse events, and exposes a few Codex-specific fields (turn_id,
// model, last_assistant_message) that are not represented in the shared
// types but are still available as raw JSON when needed.
//
// See https://developers.openai.com/codex/hooks for the upstream
// specification.
//
// # Configuration
//
// Codex hooks are enabled by default in current releases (the `hooks`
// feature flag is stable). If your organization disabled them, set
// `[features].hooks = true` in ~/.codex/config.toml to re-enable.
//
// Configure hook commands in ~/.codex/hooks.json (or inline under [hooks] in
// ~/.codex/config.toml):
//
//	{
//	  "hooks": {
//	    "PreToolUse": [
//	      {
//	        "matcher": "Bash|apply_patch|mcp__.*",
//	        "hooks": [{ "type": "command", "command": "/path/to/hook codex-pre-tool-use" }]
//	      }
//	    ],
//	    "PostToolUse": [
//	      {
//	        "matcher": "Bash|apply_patch|mcp__.*",
//	        "hooks": [{ "type": "command", "command": "/path/to/hook codex-post-tool-use" }]
//	      }
//	    ],
//	    "UserPromptSubmit": [
//	      { "hooks": [{ "type": "command", "command": "/path/to/hook codex-user-prompt-submit" }] }
//	    ],
//	    "Stop": [
//	      { "hooks": [{ "type": "command", "command": "/path/to/hook codex-stop" }] }
//	    ]
//	  }
//	}
//
// # Example Usage
//
//	package main
//
//	import (
//	    "github.com/CorridorSecurity/hookshot"
//	    "github.com/CorridorSecurity/hookshot/codex"
//	)
//
//	func main() {
//	    hookshot.Register("codex-pre-tool-use", func() {
//	        hookshot.Run(func(input codex.PreToolUseInput) codex.PreToolUseOutput {
//	            if input.ToolName == "Bash" {
//	                return codex.Allow("Bash command reviewed by hook")
//	            }
//	            return codex.PassThrough()
//	        })
//	    })
//	    hookshot.RunCommand()
//	}
package codex
