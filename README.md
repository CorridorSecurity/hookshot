# hookshot

A Go library for building hooks for AI coding agents like [Cursor](https://cursor.com/docs/agent/hooks), [Claude Code](https://docs.claude.com/en/docs/claude-code/hooks), [Windsurf Cascade](https://docs.codeium.com/windsurf/memories#hooks), [Factory Droid](https://docs.factory.ai/reference/hooks-reference), and [OpenAI Codex](https://developers.openai.com/codex/hooks).

Hooks are a key component of [Agentic Coding Security Management (ACSM)](https://corridor.dev/blog/introducing-acsm/) — they let you observe, control, and secure AI agent behavior in your development environment.

## Installation

```bash
go get github.com/CorridorSecurity/hookshot
```

## Quick Start

```go
package main

import (
    "strings"
    "github.com/CorridorSecurity/hookshot"
)

func main() {
    // Stop hooks - prevent agent from stopping prematurely
    hookshot.OnStop(func(ctx hookshot.StopContext) hookshot.StopDecision {
        if ctx.ShouldSkip() {
            return hookshot.AllowStop()
        }
        return hookshot.PreventStop("Please verify the changes")
    })

    // Execution hooks - control shell commands and MCP tools
    hookshot.OnBeforeExecution(func(ctx hookshot.ExecutionContext) hookshot.ExecutionDecision {
        if ctx.Type == hookshot.ExecutionShell {
            if strings.Contains(ctx.Command, "rm -rf /") {
                return hookshot.DenyExecution("Dangerous command blocked")
            }
        }
        return hookshot.AllowExecution()
    })

    // File edit hooks - react to file changes
    hookshot.OnAfterFileEdit(func(ctx hookshot.FileEditContext) hookshot.FileEditDecision {
        return hookshot.FileEditOK()
    })

    // Prompt hooks - validate user prompts
    hookshot.OnPromptSubmit(func(ctx hookshot.PromptContext) hookshot.PromptDecision {
        return hookshot.AllowPromptDecision()
    })

    hookshot.RunCommand()
}
```

Run with: `./my-hooks claude-stop` or `./my-hooks cursor-stop`

## Building

```bash
# Current platform
go build -o my-hooks .

# All platforms
hookshot build -all -output ./dist
```

## Installing Hooks

```bash
# Install to Claude Code and Cursor config files
hookshot install --binary /path/to/my-hooks
```

## Configuration

### Claude Code (`~/.claude/settings.json`)

```json
{
  "hooks": {
    "Stop": [{ "hooks": [{ "type": "command", "command": "/path/to/my-hooks claude-stop" }] }],
    "PreToolUse": [{ "matcher": "*", "hooks": [{ "type": "command", "command": "/path/to/my-hooks claude-pre-tool-use" }] }]
  }
}
```

### Cursor (`~/.cursor/hooks.json`)

```json
{
  "version": 1,
  "hooks": {
    "stop": [{ "command": "/path/to/my-hooks cursor-stop" }],
    "beforeShellExecution": [{ "command": "/path/to/my-hooks cursor-before-shell" }]
  }
}
```

### OpenAI Codex (`~/.codex/hooks.json`)

Codex hooks are enabled by default (the `hooks` feature flag in Codex is
stable and on). No `~/.codex/config.toml` change is required. If your
organization disabled hooks, set `[features].hooks = true` to turn them
back on — `codex_hooks` is a deprecated alias for the same flag.

```json
{
  "hooks": {
    "Stop": [{ "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-stop" }] }],
    "PreToolUse": [{ "matcher": "Bash|apply_patch|mcp__.*", "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-pre-tool-use" }] }],
    "PostToolUse": [{ "matcher": "apply_patch|mcp__.*", "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-post-tool-use" }] }]
  }
}
```

## Unified Handlers

Write once, run on all five platforms:

| Handler | Claude Code | Cursor | Windsurf Cascade | Factory Droid | OpenAI Codex |
|---------|-------------|--------|------------------|---------------|--------------|
| `OnStop` | Stop | stop | post-cascade-response | Stop | Stop |
| `OnBeforeExecution` | PreToolUse | beforeShellExecution, beforeMCPExecution | pre-run-command, pre-mcp-tool-use | PreToolUse | PreToolUse |
| `OnAfterFileEdit` | PostToolUse | afterFileEdit | post-write-code | PostToolUse | PostToolUse |
| `OnPromptSubmit` | UserPromptSubmit | beforeSubmitPrompt | pre-user-prompt | UserPromptSubmit | UserPromptSubmit |

## Platform-Specific Handlers

For platform-specific features, use `Register`:

```go
// Claude Code only: SessionStart
hookshot.Register("claude-session-start", func() {
    hookshot.Run(func(input claude.SessionStartInput) claude.SessionStartOutput {
        return claude.SessionStartContext("Project uses Go 1.21+")
    })
})

// Cursor only: Tab completion
hookshot.Register("cursor-before-tab-read", func() {
    hookshot.Run(func(input cursor.BeforeTabFileReadInput) cursor.BeforeTabFileReadOutput {
        return cursor.AllowTabRead()
    })
})

// Windsurf Cascade: Pre-write-code (not covered by unified API)
hookshot.Register("cascade-pre-write-code", func() {
    hookshot.RunE(func(input cascade.PreWriteCodeInput) (cascade.PreWriteCodeOutput, error) {
        return cascade.AllowWrite(), nil
    })
})

// Factory Droid: Pre-tool use
hookshot.Register("droid-pre-tool-use", func() {
    hookshot.Run(func(input droid.PreToolUseInput) droid.PreToolUseOutput {
        return droid.PassThrough()
    })
})

// OpenAI Codex: Pre-tool use (matches Bash, apply_patch, and MCP tools)
hookshot.Register("codex-pre-tool-use", func() {
    hookshot.Run(func(input codex.PreToolUseInput) codex.PreToolUseOutput {
        return codex.PassThrough()
    })
})
```

## Documentation

- [Unified API Reference](docs/reference-unified.md)
- [Claude Code Reference](docs/reference-claude.md)
- [Cursor Reference](docs/reference-cursor.md)
- [Windsurf Cascade Reference](docs/reference-cascade.md)
- [Factory Droid Reference](docs/reference-droid.md)
- [OpenAI Codex Reference](docs/reference-codex.md)

Full API documentation is available via godoc:

```bash
go doc github.com/CorridorSecurity/hookshot
go doc github.com/CorridorSecurity/hookshot/claude
go doc github.com/CorridorSecurity/hookshot/cursor
go doc github.com/CorridorSecurity/hookshot/cascade
go doc github.com/CorridorSecurity/hookshot/droid
go doc github.com/CorridorSecurity/hookshot/codex
```

Or view online at [pkg.go.dev/github.com/CorridorSecurity/hookshot](https://pkg.go.dev/github.com/CorridorSecurity/hookshot).

## License

MIT
