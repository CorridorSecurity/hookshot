# Cursor API Reference

Platform-specific types and helpers for Cursor hooks. Use these when you need features not available in the unified API.

See [Cursor Hooks Documentation](https://cursor.com/docs/agent/hooks) for official docs.

## Events

### Agent Hooks (Cmd+K/Agent Chat)

| Event | Input | Output | Description |
|-------|-------|--------|-------------|
| stop | `StopInput` | `StopOutput` | Agent loop ended |
| beforeShellExecution | `BeforeShellExecutionInput` | `BeforeExecutionOutput` | Before shell command |
| afterShellExecution | `AfterShellExecutionInput` | `AfterShellExecutionOutput` | After shell command |
| beforeMCPExecution | `BeforeMCPExecutionInput` | `BeforeExecutionOutput` | Before MCP tool |
| afterMCPExecution | `AfterMCPExecutionInput` | `AfterMCPExecutionOutput` | After MCP tool |
| beforeReadFile | `BeforeReadFileInput` | `BeforeReadFileOutput` | Before agent reads file |
| afterFileEdit | `AfterFileEditInput` | `AfterFileEditOutput` | After agent edits file |
| beforeSubmitPrompt | `BeforeSubmitPromptInput` | `BeforeSubmitPromptOutput` | Before prompt submission |
| afterAgentResponse | `AfterAgentResponseInput` | `AfterAgentResponseOutput` | After agent message |
| afterAgentThought | `AfterAgentThoughtInput` | `AfterAgentThoughtOutput` | After thinking block |

### Tab Hooks (Inline Completions)

| Event | Input | Output | Description |
|-------|-------|--------|-------------|
| beforeTabFileRead | `BeforeTabFileReadInput` | `BeforeTabFileReadOutput` | Before Tab reads file |
| afterTabFileEdit | `AfterTabFileEditInput` | `AfterTabFileEditOutput` | After Tab edits file |

---

## Stop

Called when the agent loop ends.

### StopInput

```go
type StopInput struct {
    ConversationID string `json:"conversation_id"`
    Status         string `json:"status"` // "completed", "aborted", "error"
    LoopCount      int    `json:"loop_count"`
}
```

### StopOutput

```go
type StopOutput struct {
    FollowupMessage string `json:"followup_message,omitempty"`
}
```

### Helper Functions

```go
func Continue() StopOutput              // Allow stopping
func Followup(message string) StopOutput // Auto follow-up (max 5)
```

### Example

```go
hookshot.Register("cursor-stop", func() {
    hookshot.Run(func(input cursor.StopInput) cursor.StopOutput {
        // Limit follow-ups to prevent loops (max 5 enforced by Cursor)
        if input.LoopCount >= 3 {
            return cursor.Continue()
        }

        if input.Status == "completed" && input.LoopCount == 0 {
            return cursor.Followup("Please verify the changes")
        }

        return cursor.Continue()
    })
})
```

---

## Before Shell Execution

Called before a shell command is executed.

### BeforeShellExecutionInput

```go
type BeforeShellExecutionInput struct {
    ConversationID string `json:"conversation_id"`
    Command        string `json:"command"`
    Cwd            string `json:"cwd"`
}
```

### BeforeExecutionOutput

```go
type BeforeExecutionOutput struct {
    Decision     string `json:"decision,omitempty"` // "deny", "ask"
    UserMessage  string `json:"user_message,omitempty"`
    AgentMessage string `json:"agent_message,omitempty"`
}
```

### Helper Functions

```go
func Allow() BeforeExecutionOutput
func AllowWithMessage(msg string) BeforeExecutionOutput
func Deny(userMsg, agentMsg string) BeforeExecutionOutput
func DenyWithUserMessage(msg string) BeforeExecutionOutput
func DenyWithAgentMessage(msg string) BeforeExecutionOutput
func Ask(msg string) BeforeExecutionOutput
```

### Example

```go
hookshot.Register("cursor-before-shell", func() {
    hookshot.Run(func(input cursor.BeforeShellExecutionInput) cursor.BeforeExecutionOutput {
        if strings.Contains(input.Command, "rm -rf /") {
            return cursor.Deny(
                "Command blocked for safety",
                "Dangerous command not allowed",
            )
        }
        return cursor.Allow()
    })
})
```

---

## After Shell Execution

Called after a shell command completes.

### AfterShellExecutionInput

```go
type AfterShellExecutionInput struct {
    ConversationID string `json:"conversation_id"`
    Command        string `json:"command"`
    Cwd            string `json:"cwd"`
    ExitCode       int    `json:"exit_code"`
    Stdout         string `json:"stdout"`
    Stderr         string `json:"stderr"`
    Duration       int    `json:"duration"` // milliseconds
}
```

### AfterShellExecutionOutput

```go
type AfterShellExecutionOutput struct{}
```

### Helper Functions

```go
func AfterShellOK() AfterShellExecutionOutput
```

---

## Before MCP Execution

Called before an MCP tool is executed.

### BeforeMCPExecutionInput

```go
type BeforeMCPExecutionInput struct {
    ConversationID string `json:"conversation_id"`
    URL            string `json:"url"`
    ToolName       string `json:"tool_name"`
    ToolInput      string `json:"tool_input"` // JSON string
}
```

Uses `BeforeExecutionOutput` (same as shell).

### Example

```go
hookshot.Register("cursor-before-mcp", func() {
    hookshot.Run(func(input cursor.BeforeMCPExecutionInput) cursor.BeforeExecutionOutput {
        if strings.Contains(input.URL, "blocked.com") {
            return cursor.Deny(
                "MCP server blocked",
                "This MCP server is not allowed",
            )
        }
        return cursor.Allow()
    })
})
```

---

## After MCP Execution

Called after an MCP tool completes.

### AfterMCPExecutionInput

```go
type AfterMCPExecutionInput struct {
    ConversationID string `json:"conversation_id"`
    URL            string `json:"url"`
    ToolName       string `json:"tool_name"`
    ToolInput      string `json:"tool_input"`
    ToolResult     string `json:"tool_result"`
    Duration       int    `json:"duration"` // milliseconds
}
```

### AfterMCPExecutionOutput

```go
type AfterMCPExecutionOutput struct{}
```

### Helper Functions

```go
func AfterMCPOK() AfterMCPExecutionOutput
```

---

## Before Read File

Called before the agent reads a file.

### BeforeReadFileInput

```go
type BeforeReadFileInput struct {
    ConversationID string `json:"conversation_id"`
    FilePath       string `json:"file_path"`
}
```

### BeforeReadFileOutput

```go
type BeforeReadFileOutput struct {
    Decision     string `json:"decision,omitempty"` // "deny", "ask"
    UserMessage  string `json:"user_message,omitempty"`
    AgentMessage string `json:"agent_message,omitempty"`
}
```

### Helper Functions

```go
func AllowRead() BeforeReadFileOutput
func DenyRead(userMsg, agentMsg string) BeforeReadFileOutput
func AskRead(msg string) BeforeReadFileOutput
```

### Example

```go
hookshot.Register("cursor-before-read-file", func() {
    hookshot.Run(func(input cursor.BeforeReadFileInput) cursor.BeforeReadFileOutput {
        if strings.Contains(input.FilePath, ".env") {
            return cursor.DenyRead(
                "Cannot read sensitive files",
                "File access blocked",
            )
        }
        return cursor.AllowRead()
    })
})
```

---

## After File Edit

Called after the agent edits a file.

### AfterFileEditInput

```go
type AfterFileEditInput struct {
    ConversationID string `json:"conversation_id"`
    FilePath       string `json:"file_path"`
    Edits          []Edit `json:"edits"`
}

type Edit struct {
    OldString string `json:"old_string"`
    NewString string `json:"new_string"`
}
```

### AfterFileEditOutput

```go
type AfterFileEditOutput struct{}
```

### Helper Functions

```go
func AfterFileEditOK() AfterFileEditOutput
```

---

## Before Submit Prompt

Called before a prompt is submitted.

### BeforeSubmitPromptInput

```go
type BeforeSubmitPromptInput struct {
    ConversationID string `json:"conversation_id"`
    Prompt         string `json:"prompt"`
}
```

### BeforeSubmitPromptOutput

```go
type BeforeSubmitPromptOutput struct {
    Decision    string `json:"decision,omitempty"` // "block"
    UserMessage string `json:"user_message,omitempty"`
}
```

### Helper Functions

```go
func AllowPrompt() BeforeSubmitPromptOutput
func BlockPrompt(msg string) BeforeSubmitPromptOutput
```

---

## Tab Hooks

### Before Tab File Read

Called before Tab reads a file for context.

```go
type BeforeTabFileReadInput struct {
    FilePath string `json:"file_path"`
}

type BeforeTabFileReadOutput struct {
    Decision string `json:"decision,omitempty"` // "deny"
}
```

### Helper Functions

```go
func AllowTabRead() BeforeTabFileReadOutput
func DenyTabRead() BeforeTabFileReadOutput
```

### Example

```go
hookshot.Register("cursor-before-tab-read", func() {
    hookshot.Run(func(input cursor.BeforeTabFileReadInput) cursor.BeforeTabFileReadOutput {
        if strings.Contains(input.FilePath, ".env") {
            return cursor.DenyTabRead()
        }
        return cursor.AllowTabRead()
    })
})
```

### After Tab File Edit

Called after Tab edits a file.

```go
type AfterTabFileEditInput struct {
    FilePath string `json:"file_path"`
    Edits    []Edit `json:"edits"`
}

type AfterTabFileEditOutput struct{}
```

### Helper Functions

```go
func AfterTabEditOK() AfterTabFileEditOutput
```
