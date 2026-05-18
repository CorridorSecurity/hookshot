# Unified API Reference

The unified API provides cross-platform handlers that work on Claude Code, Cursor, Windsurf Cascade, Factory Droid, and OpenAI Codex. Write once, run on all platforms.

## Platform Constants

```go
type Platform string

const (
    PlatformClaude  Platform = "claude"
    PlatformCursor  Platform = "cursor"
    PlatformDroid   Platform = "droid"
    PlatformCascade Platform = "cascade"
    PlatformCodex   Platform = "codex"
)
```

## OnStop

Handles stop events when the agent is about to finish.

**Registers:** `claude-stop`, `cursor-stop`, `droid-stop`, `cascade-post-cascade-response`, `codex-stop`

### StopContext

```go
type StopContext struct {
    Platform  Platform
    SessionID string // Claude/Droid/Codex: session_id, Cursor: conversation_id, Cascade: trajectory_id
    Cwd       string // Working directory (Claude/Droid/Codex only, empty for Cursor/Cascade)

    // Claude/Droid/Codex-specific
    StopHookActive bool // True if already continuing from a previous stop hook

    // Cursor-specific
    Status    string // "completed", "aborted", or "error"
    LoopCount int    // Number of previous auto follow-ups (max 5)
}
```

### Methods

```go
// ShouldSkip returns true if the stop hook should be skipped to prevent loops.
// Claude/Droid/Codex: checks StopHookActive
// Cursor: checks LoopCount >= 3
// Cascade: always returns false (no loop prevention mechanism)
func (c StopContext) ShouldSkip() bool
```

### StopDecision

```go
type StopDecision struct {
    Continue bool   // true = allow stopping, false = prevent stopping
    Message  string // Shown to agent when Continue is false
}
```

### Helper Functions

```go
func AllowStop() StopDecision
func PreventStop(message string) StopDecision
```

### Example

```go
hookshot.OnStop(func(ctx hookshot.StopContext) hookshot.StopDecision {
    if ctx.ShouldSkip() {
        return hookshot.AllowStop()
    }

    if ctx.Platform == hookshot.PlatformCursor && ctx.Status == "completed" {
        return hookshot.PreventStop("Please verify the changes")
    }

    return hookshot.AllowStop()
})
```

---

## OnBeforeExecution

Handles pre-execution events for shell commands and MCP tools.

**Registers:** `claude-pre-tool-use`, `cursor-before-shell`, `cursor-before-mcp`, `droid-pre-tool-use`, `cascade-pre-run-command`, `cascade-pre-mcp-tool-use`, `codex-pre-tool-use`

For Codex, `apply_patch` is classified as `ExecutionTool` (not `ExecutionShell` or `ExecutionMCP`); use `ctx.ToolName == "apply_patch"` to detect it. The patch text is exposed via `ctx.Command` so policies can inspect it without re-parsing `ToolInput`.

Codex does not currently enforce `permissionDecision: "ask"`. To avoid a silent fail-open, the unified bridge rewrites `AskExecution(...)` decisions to a `Deny` on Codex; on every other platform `Ask` still surfaces an approval prompt as before.

### ExecutionType

```go
type ExecutionType string

const (
    ExecutionShell ExecutionType = "shell" // Shell commands
    ExecutionMCP   ExecutionType = "mcp"   // MCP tool calls
    ExecutionTool  ExecutionType = "tool"  // Claude non-MCP tools (Read, Write, etc.)
)
```

### ExecutionContext

```go
type ExecutionContext struct {
    Platform Platform
    Type     ExecutionType

    // For shell execution (Cursor beforeShellExecution, Claude Code/Codex Bash tool)
    // Also used for local MCP servers on Cursor (command-based MCP servers)
    // NOTE: For Claude Code, Droid, and Codex, Command is parsed from tool_input.command for Bash.
    Command string
    Cwd     string // Working directory

    // For MCP execution
    ToolName  string          // MCP tool name (e.g., "mcp__server__tool")
    ToolInput json.RawMessage // Tool input as JSON
    ServerURL string          // MCP server URL (Cursor/Cascade only, for URL-based servers)

    // Raw access
    // For Codex, the raw input is shared with Claude Code (RawClaudeCode) because
    // Codex uses the same JSON wire format.
    RawClaudeCode *claude.PreToolUseInput
    RawCursor     any // *cursor.BeforeShellExecutionInput or *cursor.BeforeMCPExecutionInput
    RawDroid      *droid.PreToolUseInput
    RawCascade    any // *cascade.PreRunCommandInput or *cascade.PreMCPToolUseInput
}
```

### Methods

```go
// IsMCP returns true if this is an MCP tool execution
func (c ExecutionContext) IsMCP() bool
```

### ExecutionDecision

```go
type ExecutionDecision struct {
    Allow  bool   // true = permit, false = block
    Reason string // Explanation shown to user (allow) or agent (deny)
    Ask    bool   // Prompt user to confirm (only when Allow is false)
}
```

### Helper Functions

```go
func AllowExecution() ExecutionDecision
func AllowExecutionWithReason(reason string) ExecutionDecision
func DenyExecution(reason string) ExecutionDecision
func AskExecution(reason string) ExecutionDecision
```

### Example

```go
hookshot.OnBeforeExecution(func(ctx hookshot.ExecutionContext) hookshot.ExecutionDecision {
    // Block dangerous shell commands
    if ctx.Type == hookshot.ExecutionShell {
        if strings.Contains(ctx.Command, "rm -rf /") {
            return hookshot.DenyExecution("Dangerous command blocked")
        }
    }

    // Block specific MCP servers
    if ctx.IsMCP() && strings.HasPrefix(ctx.ToolName, "mcp__blocked__") {
        return hookshot.DenyExecution("MCP server not allowed")
    }

    return hookshot.AllowExecution()
})
```

---

## OnAfterFileEdit

Handles post-file-edit events.

**Registers:** `claude-after-file-edit`, `cursor-after-file-edit`, `droid-after-file-edit`, `cascade-post-write-code`, `codex-post-tool-use`

For Codex, the underlying PostToolUse handler matches `apply_patch` (which is how Codex performs all file edits). Configure the hook with `matcher: "apply_patch|mcp__.*"` in `~/.codex/hooks.json`.

For Codex `apply_patch`, the unified bridge parses the unified-diff envelope in `tool_input.command` and invokes your handler **once per file** in the patch, with `FilePath` set to the file declared in the `*** Add/Update/Delete File:` header and `Edits` populated from each hunk. If any of those invocations returns `FileEditBlock`, the reasons are concatenated and emitted as a single `PostToolBlock`.

For renames (`*** Update File: <src>` + `*** Move to: <dst>`) the handler is invoked **twice** — once with `FilePath` set to the source and once with `FilePath` set to the destination — and `NewFilePath` is populated on both invocations. This means a FilePath-only allowlist that permits the benign source still receives a separate call for the destination so it can deny something like `../../.ssh/authorized_keys`. Handlers that want to detect the rename relationship should check `ctx.NewFilePath != "" && ctx.NewFilePath != ctx.FilePath`.

The same parser is exported as `codex.ParseApplyPatch` for callers that want to parse a patch envelope themselves.

### FileEdit

```go
type FileEdit struct {
    OldString string
    NewString string
}
```

### FileEditContext

```go
type FileEditContext struct {
    Platform    Platform
    SessionID   string // Claude/Droid: session_id, Cursor: conversation_id, Cascade: trajectory_id
    FilePath    string
    NewFilePath string // Destination path for rename operations (Codex apply_patch "*** Move to:"); empty otherwise.
    Edits       []FileEdit
    Cwd         string

    // Raw access
    RawClaudeCode *claude.PostToolUseInput
    RawCursor     *cursor.AfterFileEditInput
    RawDroid      *droid.PostToolUseInput
    RawCascade    *cascade.PostWriteCodeInput
}
```

### FileEditDecision

```go
type FileEditDecision struct {
    Block   bool   // Send feedback to agent (Claude only)
    Reason  string // Shown to agent when Block is true
    Context string // Additional context for agent (Claude only)
}
```

### Helper Functions

```go
func FileEditOK() FileEditDecision
func FileEditBlock(reason string) FileEditDecision
func FileEditAddContext(context string) FileEditDecision
```

### Example

```go
hookshot.OnAfterFileEdit(func(ctx hookshot.FileEditContext) hookshot.FileEditDecision {
    // Log all file edits
    fmt.Printf("File edited: %s\n", ctx.FilePath)

    // Add context if TODO found (Claude only)
    if ctx.Platform == hookshot.PlatformClaude {
        for _, edit := range ctx.Edits {
            if strings.Contains(edit.NewString, "TODO") {
                return hookshot.FileEditAddContext("File contains TODO comments")
            }
        }
    }

    return hookshot.FileEditOK()
})
```

---

## OnPromptSubmit

Handles prompt submission events.

**Registers:** `claude-user-prompt-submit`, `cursor-before-submit-prompt`, `droid-user-prompt-submit`, `cascade-pre-user-prompt`, `codex-user-prompt-submit`

### PromptContext

```go
type PromptContext struct {
    Platform  Platform
    SessionID string // Claude/Droid: session_id, Cursor: conversation_id, Cascade: trajectory_id
    Prompt    string

    // Raw access
    RawClaudeCode *claude.UserPromptSubmitInput
    RawCursor     *cursor.BeforeSubmitPromptInput
    RawDroid      *droid.UserPromptSubmitInput
    RawCascade    *cascade.PreUserPromptInput
}
```

### PromptDecision

```go
type PromptDecision struct {
    Allow   bool   // true = process, false = block
    Reason  string // Shown to user when Allow is false
    Context string // Additional context for agent (Claude only)
}
```

### Helper Functions

```go
func AllowPromptDecision() PromptDecision
func BlockPromptDecision(reason string) PromptDecision
func AddPromptContext(context string) PromptDecision
```

### Example

```go
hookshot.OnPromptSubmit(func(ctx hookshot.PromptContext) hookshot.PromptDecision {
    // Block prompts with API keys
    if strings.Contains(ctx.Prompt, "api_key=") {
        return hookshot.BlockPromptDecision("Don't include API keys")
    }

    // Add project context (Claude only)
    if ctx.Platform == hookshot.PlatformClaude {
        return hookshot.AddPromptContext("Project uses Go 1.21+")
    }

    return hookshot.AllowPromptDecision()
})
```

---

## Raw Input Access

For advanced use cases, access raw platform-specific data:

```go
func ReadRawInput(v any) error
```

Example:

```go
hookshot.OnBeforeExecution(func(ctx hookshot.ExecutionContext) hookshot.ExecutionDecision {
    if ctx.Platform == hookshot.PlatformClaude && ctx.RawClaudeCode != nil {
        // Access Claude-specific fields
        sessionID := ctx.RawClaudeCode.SessionID
    }
    return hookshot.AllowExecution()
})
```
