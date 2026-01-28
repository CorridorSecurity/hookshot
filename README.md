# hookshot

A Go library for building hooks for AI coding agents like [Cursor](https://cursor.com/docs/agent/hooks), [Claude Code](https://docs.claude.com/en/docs/claude-code/hooks), [Windsurf Cascade](https://docs.codeium.com/windsurf/memories#hooks), and Factory Droid.

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

## Unified Handlers

Write once, run on both platforms:

| Handler | Claude Code | Cursor |
|---------|-------------|--------|
| `OnStop` | Stop | stop |
| `OnBeforeExecution` | PreToolUse | beforeShellExecution, beforeMCPExecution |
| `OnAfterFileEdit` | PostToolUse | afterFileEdit |
| `OnPromptSubmit` | UserPromptSubmit | beforeSubmitPrompt |

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

// Windsurf Cascade: Pre-run command
hookshot.Register("cascade-pre-run-command", func() {
    hookshot.Run(func(input cascade.PreRunCommandInput) cascade.PreRunCommandOutput {
        return cascade.AllowCommand()
    })
})

// Factory Droid: Pre-tool use
hookshot.Register("droid-pre-tool-use", func() {
    hookshot.Run(func(input droid.PreToolUseInput) droid.PreToolUseOutput {
        return droid.PassThrough()
    })
})
```

## Documentation

- [Unified API Reference](docs/reference-unified.md)
- [Claude Code Reference](docs/reference-claude.md)
- [Cursor Reference](docs/reference-cursor.md)
- [Windsurf Cascade Reference](docs/reference-cascade.md)
- [Factory Droid Reference](docs/reference-droid.md)

Full API documentation is available via godoc:

```bash
go doc github.com/CorridorSecurity/hookshot
go doc github.com/CorridorSecurity/hookshot/claude
go doc github.com/CorridorSecurity/hookshot/cursor
go doc github.com/CorridorSecurity/hookshot/cascade
go doc github.com/CorridorSecurity/hookshot/droid
```

Or view online at [pkg.go.dev/github.com/CorridorSecurity/hookshot](https://pkg.go.dev/github.com/CorridorSecurity/hookshot).

## License

MIT
