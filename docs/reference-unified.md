# Unified API Reference

The unified API provides cross-platform handlers that work on both Claude Code and Cursor. Write once, run on both platforms.

## Platform Constants

```go
type Platform string

const (
    PlatformClaude Platform = "claude"
    PlatformCursor Platform = "cursor"
)
```

## OnStop

Handles stop events when the agent is about to finish.

**Registers:** `claude-stop`, `cursor-stop`

### StopContext

```go
type StopContext struct {
    Platform  Platform
    SessionID string // Claude: session_id, Cursor: conversation_id
    Cwd       string // Working directory (Claude only, empty for Cursor)

    // Claude-specific
    StopHookActive bool // True if already continuing from a previous stop hook

    // Cursor-specific
    Status    string // "completed", "aborted", or "error"
    LoopCount int    // Number of previous auto follow-ups (max 5)
}
```

### Methods

```go
// ShouldSkip returns true if the stop hook should be skipped to prevent loops.
// Claude: checks StopHookActive
// Cursor: checks LoopCount >= 3
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

**Registers:** `claude-pre-tool-use`, `cursor-before-shell`, `cursor-before-mcp`

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

    // For shell execution (Cursor beforeShellExecution, Claude Code Bash tool)
    // Also used for local MCP servers on Cursor (command-based MCP servers)
    // NOTE: Only populated for Cursor, not Claude Code
    Command string
    Cwd     string // Working directory

    // For MCP execution
    ToolName  string          // MCP tool name (e.g., "mcp__server__tool")
    ToolInput json.RawMessage // Tool input as JSON
    ServerURL string          // MCP server URL (Cursor only, for URL-based servers)

    // Raw access
    RawClaudeCode *claude.PreToolUseInput
    RawCursor     any // *cursor.BeforeShellExecutionInput or *cursor.BeforeMCPExecutionInput
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

**Registers:** `claude-after-file-edit`, `cursor-after-file-edit`

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
    Platform  Platform
    SessionID string // Claude: session_id, Cursor: conversation_id
    FilePath  string
    Edits     []FileEdit
    Cwd       string

    // Raw access
    RawClaudeCode *claude.PostToolUseInput
    RawCursor     *cursor.AfterFileEditInput
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

**Registers:** `claude-user-prompt-submit`, `cursor-before-submit-prompt`

### PromptContext

```go
type PromptContext struct {
    Platform  Platform
    SessionID string // Claude: session_id, Cursor: conversation_id
    Prompt    string

    // Raw access
    RawClaudeCode *claude.UserPromptSubmitInput
    RawCursor     *cursor.BeforeSubmitPromptInput
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
