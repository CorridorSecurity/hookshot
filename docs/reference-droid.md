# Factory Droid API Reference

Platform-specific types and helpers for Factory Droid hooks. Use these when you need features not available in the unified API.

Factory Droid hooks use the same event model as Claude Code, configured in settings files that execute shell commands at various points in the agent loop.

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

## Common Types

### BaseInput

All Droid hook inputs include these fields:

```go
type BaseInput struct {
    SessionID      string `json:"session_id"`
    TranscriptPath string `json:"transcript_path"`
    Cwd            string `json:"cwd"`
    PermissionMode string `json:"permission_mode"` // "default", "plan", "acceptEdits", "bypassPermissions"
    HookEventName  string `json:"hook_event_name"`
}
```

### BaseOutput

Common fields for any hook output:

```go
type BaseOutput struct {
    Continue       *bool  `json:"continue,omitempty"`       // false to stop Droid
    StopReason     string `json:"stopReason,omitempty"`     // Shown when Continue is false
    SuppressOutput bool   `json:"suppressOutput,omitempty"` // Hide stdout from transcript
    SystemMessage  string `json:"systemMessage,omitempty"`  // Warning message shown to user
}
```

---

## Stop / SubagentStop

Called when the agent finishes responding.

### StopInput

```go
type StopInput struct {
    BaseInput
    StopHookActive bool `json:"stop_hook_active"` // Check to prevent infinite loops
}

type SubagentStopInput = StopInput
```

### StopOutput

```go
type StopOutput struct {
    BaseOutput
    Decision string `json:"decision,omitempty"` // "block" to prevent stopping
    Reason   string `json:"reason,omitempty"`   // Shown to Droid when blocked
}

type SubagentStopOutput = StopOutput
```

### Helper Functions

```go
func Continue() StopOutput              // Allow stopping
func Block(reason string) StopOutput    // Prevent stopping, continue working
func StopWith(reason string) StopOutput // Halt Droid entirely
```

### Example

```go
hookshot.Register("droid-stop", func() {
    hookshot.Run(func(input droid.StopInput) droid.StopOutput {
        // IMPORTANT: Check StopHookActive to prevent infinite loops
        if input.StopHookActive {
            return droid.Continue()
        }
        return droid.Continue()
    })
})
```

---

## SessionStart

Called when a session starts or resumes.

### SessionStartInput

```go
type SessionStartInput struct {
    BaseInput
    Source string `json:"source"` // "startup", "resume", "clear", "compact"
}
```

### SessionStartOutput

```go
type SessionStartOutput struct {
    BaseOutput
    HookSpecificOutput *SessionStartHookOutput `json:"hookSpecificOutput,omitempty"`
}

type SessionStartHookOutput struct {
    HookEventName     string `json:"hookEventName,omitempty"`
    AdditionalContext string `json:"additionalContext,omitempty"`
}
```

### Helper Functions

```go
func SessionStartOK() SessionStartOutput
func SessionStartContext(context string) SessionStartOutput
```

### Example

```go
hookshot.Register("droid-session-start", func() {
    hookshot.Run(func(input droid.SessionStartInput) droid.SessionStartOutput {
        return droid.SessionStartContext("Project uses Go 1.21+")
    })
})
```

---

## SessionEnd

Called when a session is ending.

### SessionEndInput

```go
type SessionEndInput struct {
    BaseInput
    Reason string `json:"reason"` // "clear", "logout", "prompt_input_exit", "other"
}
```

### SessionEndOutput

```go
type SessionEndOutput struct {
    BaseOutput
}
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
    BaseInput
    ToolName  string          `json:"tool_name"`
    ToolInput json.RawMessage `json:"tool_input"` // Tool-specific JSON
    ToolUseID string          `json:"tool_use_id"`
}
```

### PreToolUseOutput

```go
type PreToolUseOutput struct {
    BaseOutput
    HookSpecificOutput *PreToolUseHookOutput `json:"hookSpecificOutput,omitempty"`
}

type PreToolUseHookOutput struct {
    HookEventName            string         `json:"hookEventName,omitempty"`
    PermissionDecision       string         `json:"permissionDecision,omitempty"`       // "allow", "deny", "ask"
    PermissionDecisionReason string         `json:"permissionDecisionReason,omitempty"` // Shown to user or Droid
    UpdatedInput             map[string]any `json:"updatedInput,omitempty"`             // Modified tool input
}
```

### Helper Functions

```go
func Allow(reason string) PreToolUseOutput
func AllowSilent() PreToolUseOutput
func AllowWithInput(reason string, updatedInput map[string]any) PreToolUseOutput
func Deny(reason string) PreToolUseOutput
func Ask(reason string) PreToolUseOutput
func PassThrough() PreToolUseOutput
```

### Example

```go
hookshot.Register("droid-pre-tool-use", func() {
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
})
```

---

## PermissionRequest

Called when a permission dialog is shown.

### PermissionRequestInput

```go
type PermissionRequestInput struct {
    BaseInput
    ToolName  string          `json:"tool_name"`
    ToolInput json.RawMessage `json:"tool_input"`
    ToolUseID string          `json:"tool_use_id"`
}
```

### PermissionRequestOutput

```go
type PermissionRequestOutput struct {
    BaseOutput
    HookSpecificOutput *PermissionRequestHookOutput `json:"hookSpecificOutput,omitempty"`
}

type PermissionRequestHookOutput struct {
    HookEventName string                     `json:"hookEventName,omitempty"`
    Decision      *PermissionRequestDecision `json:"decision,omitempty"`
}

type PermissionRequestDecision struct {
    Behavior     string         `json:"behavior"`               // "allow" or "deny"
    UpdatedInput map[string]any `json:"updatedInput,omitempty"` // For "allow": modified input
    Message      string         `json:"message,omitempty"`      // For "deny": shown to Droid
    Interrupt    bool           `json:"interrupt,omitempty"`    // For "deny": stop Droid if true
}
```

### Helper Functions

```go
func AllowPermission() PermissionRequestOutput
func AllowPermissionWithInput(updatedInput map[string]any) PermissionRequestOutput
func DenyPermission(message string) PermissionRequestOutput
func DenyPermissionAndStop(message string) PermissionRequestOutput
```

---

## PostToolUse

Called after a tool is executed.

### PostToolUseInput

```go
type PostToolUseInput struct {
    BaseInput
    ToolName     string          `json:"tool_name"`
    ToolInput    json.RawMessage `json:"tool_input"`
    ToolResponse json.RawMessage `json:"tool_response"` // Tool-specific response JSON
    ToolUseID    string          `json:"tool_use_id"`
}
```

### PostToolUseOutput

```go
type PostToolUseOutput struct {
    BaseOutput
    Decision           string                 `json:"decision,omitempty"` // "block" to provide feedback
    Reason             string                 `json:"reason,omitempty"`   // Shown to Droid when blocked
    HookSpecificOutput *PostToolUseHookOutput `json:"hookSpecificOutput,omitempty"`
}

type PostToolUseHookOutput struct {
    HookEventName     string `json:"hookEventName,omitempty"`
    AdditionalContext string `json:"additionalContext,omitempty"` // Added to context for Droid
}
```

### Helper Functions

```go
func PostToolOK() PostToolUseOutput
func PostToolBlock(reason string) PostToolUseOutput
func PostToolContext(context string) PostToolUseOutput
```

---

## UserPromptSubmit

Called when the user submits a prompt.

### UserPromptSubmitInput

```go
type UserPromptSubmitInput struct {
    BaseInput
    Prompt string `json:"prompt"`
}
```

### UserPromptSubmitOutput

```go
type UserPromptSubmitOutput struct {
    BaseOutput
    Decision           string                      `json:"decision,omitempty"` // "block" to prevent
    Reason             string                      `json:"reason,omitempty"`   // Shown to user when blocked
    HookSpecificOutput *UserPromptSubmitHookOutput `json:"hookSpecificOutput,omitempty"`
}

type UserPromptSubmitHookOutput struct {
    HookEventName     string `json:"hookEventName,omitempty"`
    AdditionalContext string `json:"additionalContext,omitempty"` // Added to context
}
```

### Helper Functions

```go
func AllowPrompt() UserPromptSubmitOutput
func BlockPrompt(reason string) UserPromptSubmitOutput
func AddContext(context string) UserPromptSubmitOutput
```

### Example

```go
hookshot.Register("droid-user-prompt-submit", func() {
    hookshot.Run(func(input droid.UserPromptSubmitInput) droid.UserPromptSubmitOutput {
        if strings.Contains(input.Prompt, "api_key=") {
            return droid.BlockPrompt("Don't include API keys in prompts")
        }
        return droid.AllowPrompt()
    })
})
```

---

## Notification

Called when a notification is sent.

### NotificationInput

```go
type NotificationInput struct {
    BaseInput
    Message          string `json:"message"`
    NotificationType string `json:"notification_type"` // "permission_prompt", "idle_prompt", "auth_success", "elicitation_dialog"
}
```

### NotificationOutput

```go
type NotificationOutput struct {
    BaseOutput
}
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
    BaseInput
    Trigger            string `json:"trigger"`             // "manual" or "auto"
    CustomInstructions string `json:"custom_instructions"` // For manual trigger only
}
```

### PreCompactOutput

```go
type PreCompactOutput struct {
    BaseOutput
}
```

### Helper Functions

```go
func PreCompactOK() PreCompactOutput
```
