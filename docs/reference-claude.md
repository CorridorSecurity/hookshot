# Claude Code API Reference

Platform-specific types and helpers for Claude Code hooks. Use these when you need features not available in the unified API.

See [Claude Code Hooks Documentation](https://docs.claude.com/en/docs/claude-code/hooks) for official docs.

## Events

| Event | Input | Output | Description |
|-------|-------|--------|-------------|
| Stop | `StopInput` | `StopOutput` | Agent finished responding |
| SubagentStop | `SubagentStopInput` | `SubagentStopOutput` | Subagent (Task tool) finished |
| SessionStart | `SessionStartInput` | `SessionStartOutput` | Session started or resumed |
| SessionEnd | `SessionEndInput` | `SessionEndOutput` | Session ending |
| PreToolUse | `PreToolUseInput` | `PreToolUseOutput` | Before tool execution |
| PostToolUse | `PostToolUseInput` | `PostToolUseOutput` | After tool execution |
| PermissionRequest | `PermissionRequestInput` | `PermissionRequestOutput` | Permission dialog shown |
| UserPromptSubmit | `UserPromptSubmitInput` | `UserPromptSubmitOutput` | User submitted a prompt |
| Notification | `NotificationInput` | `NotificationOutput` | Notification sent |
| PreCompact | `PreCompactInput` | `PreCompactOutput` | Before context compaction |

---

## Stop / SubagentStop

Called when the agent finishes responding.

### StopInput

```go
type StopInput struct {
    SessionID      string `json:"session_id"`
    Cwd            string `json:"cwd"`
    StopHookActive bool   `json:"stop_hook_active"`
}
```

### StopOutput

```go
type StopOutput struct {
    Decision string `json:"decision,omitempty"` // "block" or "stop"
    Reason   string `json:"reason,omitempty"`
}
```

### Helper Functions

```go
func Continue() StopOutput              // Allow stopping
func Block(reason string) StopOutput    // Prevent stopping, continue working
func StopWith(reason string) StopOutput // Halt Claude entirely
```

### Example

```go
hookshot.Register("claude-stop", func() {
    hookshot.Run(func(input claude.StopInput) claude.StopOutput {
        // IMPORTANT: Check StopHookActive to prevent infinite loops
        if input.StopHookActive {
            return claude.Continue()
        }
        return claude.Continue()
    })
})
```

---

## SessionStart

Called when a session starts or resumes.

### SessionStartInput

```go
type SessionStartInput struct {
    SessionID string `json:"session_id"`
    Source    string `json:"source"` // "startup", "resume", "clear", "compact"
    Cwd       string `json:"cwd"`
}
```

### SessionStartOutput

```go
type SessionStartOutput struct {
    Context string `json:"context,omitempty"`
}
```

### Helper Functions

```go
func SessionStartOK() SessionStartOutput
func SessionStartContext(context string) SessionStartOutput
```

### Example

```go
hookshot.Register("claude-session-start", func() {
    hookshot.Run(func(input claude.SessionStartInput) claude.SessionStartOutput {
        return claude.SessionStartContext("Project uses Go 1.21+")
    })
})
```

---

## SessionEnd

Called when a session is ending.

### SessionEndInput

```go
type SessionEndInput struct {
    SessionID string `json:"session_id"`
}
```

### SessionEndOutput

```go
type SessionEndOutput struct{}
```

### Helper Functions

```go
func SessionEndOK() SessionEndOutput
```

---

## PreToolUse

Called before a tool is executed.

### PreToolUseInput

```go
type PreToolUseInput struct {
    SessionID string          `json:"session_id"`
    ToolName  string          `json:"tool_name"`
    ToolInput json.RawMessage `json:"tool_input"`
    Cwd       string          `json:"cwd"`
}
```

### PreToolUseOutput

```go
type PreToolUseOutput struct {
    Decision       string         `json:"decision,omitempty"`
    Reason         string         `json:"reason,omitempty"`
    ModifiedInput  map[string]any `json:"modified_input,omitempty"`
}
```

### Helper Functions

```go
func Allow(reason string) PreToolUseOutput
func AllowSilent() PreToolUseOutput
func AllowWithInput(reason string, modifiedInput map[string]any) PreToolUseOutput
func Deny(reason string) PreToolUseOutput
func Ask(reason string) PreToolUseOutput
func PassThrough() PreToolUseOutput
```

### Example

```go
hookshot.Register("claude-pre-tool-use", func() {
    hookshot.Run(func(input claude.PreToolUseInput) claude.PreToolUseOutput {
        // Block specific MCP servers
        if strings.HasPrefix(input.ToolName, "mcp__blocked__") {
            return claude.Deny("MCP server not allowed")
        }

        // Auto-approve Read tool
        if input.ToolName == "Read" {
            return claude.AllowSilent()
        }

        return claude.PassThrough()
    })
})
```

---

## PostToolUse

Called after a tool is executed.

### PostToolUseInput

```go
type PostToolUseInput struct {
    SessionID  string          `json:"session_id"`
    ToolName   string          `json:"tool_name"`
    ToolInput  json.RawMessage `json:"tool_input"`
    ToolResult json.RawMessage `json:"tool_result"`
    Cwd        string          `json:"cwd"`
}
```

### PostToolUseOutput

```go
type PostToolUseOutput struct {
    Decision string `json:"decision,omitempty"` // "block"
    Reason   string `json:"reason,omitempty"`
    Context  string `json:"context,omitempty"`
}
```

### Helper Functions

```go
func PostToolOK() PostToolUseOutput
func PostToolBlock(reason string) PostToolUseOutput
func PostToolContext(context string) PostToolUseOutput
```

---

## PermissionRequest

Called when a permission dialog is shown.

### PermissionRequestInput

```go
type PermissionRequestInput struct {
    SessionID string          `json:"session_id"`
    ToolName  string          `json:"tool_name"`
    ToolInput json.RawMessage `json:"tool_input"`
}
```

### PermissionRequestOutput

```go
type PermissionRequestOutput struct {
    Decision      string         `json:"decision,omitempty"`
    Message       string         `json:"message,omitempty"`
    ModifiedInput map[string]any `json:"modified_input,omitempty"`
}
```

### Helper Functions

```go
func AllowPermission() PermissionRequestOutput
func AllowPermissionWithInput(modifiedInput map[string]any) PermissionRequestOutput
func DenyPermission(message string) PermissionRequestOutput
func DenyPermissionAndStop(message string) PermissionRequestOutput
```

---

## UserPromptSubmit

Called when a user submits a prompt.

### UserPromptSubmitInput

```go
type UserPromptSubmitInput struct {
    SessionID string `json:"session_id"`
    Prompt    string `json:"prompt"`
}
```

### UserPromptSubmitOutput

```go
type UserPromptSubmitOutput struct {
    Decision string `json:"decision,omitempty"` // "block"
    Reason   string `json:"reason,omitempty"`
    Context  string `json:"context,omitempty"`
}
```

### Helper Functions

```go
func AllowPrompt() UserPromptSubmitOutput
func BlockPrompt(reason string) UserPromptSubmitOutput
func AddContext(context string) UserPromptSubmitOutput
```

---

## Notification

Called when a notification is sent.

### NotificationInput

```go
type NotificationInput struct {
    SessionID        string `json:"session_id"`
    NotificationType string `json:"notification_type"`
    // Types: permission_prompt, idle_prompt, auth_success, elicitation_dialog
}
```

### NotificationOutput

```go
type NotificationOutput struct{}
```

### Helper Functions

```go
func NotificationOK() NotificationOutput
```

---

## PreCompact

Called before context compaction.

### PreCompactInput

```go
type PreCompactInput struct {
    SessionID string `json:"session_id"`
}
```

### PreCompactOutput

```go
type PreCompactOutput struct {
    Summary string `json:"summary,omitempty"`
}
```

### Helper Functions

```go
func PreCompactOK() PreCompactOutput
func PreCompactSummary(summary string) PreCompactOutput
```
