# Windsurf Cascade API Reference

Platform-specific types and helpers for Windsurf Cascade hooks. Use these when you need features not available in the unified API.

See [Windsurf Cascade Hooks Documentation](https://docs.codeium.com/windsurf/memories#hooks) for official docs.

## Events

| Event | Input | Output | Description |
|-------|-------|--------|-------------|
| pre-run-command | `PreRunCommandInput` | `PreRunCommandOutput` | Before shell command |
| post-run-command | `PostRunCommandInput` | `PostRunCommandOutput` | After shell command |
| pre-write-code | `PreWriteCodeInput` | `PreWriteCodeOutput` | Before file write |
| post-write-code | `PostWriteCodeInput` | `PostWriteCodeOutput` | After file write |
| pre-read-code | `PreReadCodeInput` | `PreReadCodeOutput` | Before file read |
| post-read-code | `PostReadCodeInput` | `PostReadCodeOutput` | After file read |
| pre-mcp-tool-use | `PreMCPToolUseInput` | `PreMCPToolUseOutput` | Before MCP tool |
| post-mcp-tool-use | `PostMCPToolUseInput` | `PostMCPToolUseOutput` | After MCP tool |
| pre-user-prompt | `PreUserPromptInput` | `PreUserPromptOutput` | Before prompt processing |
| post-cascade-response | `PostCascadeResponseInput` | `PostCascadeResponseOutput` | After Cascade response |
| post-setup-worktree | `PostSetupWorktreeInput` | `PostSetupWorktreeOutput` | After worktree setup |

---

## Common Types

### BaseInput

All Cascade hook inputs include these fields:

```go
type BaseInput struct {
    AgentActionName string `json:"agent_action_name"` // The hook event name
    TrajectoryID    string `json:"trajectory_id"`     // Session identifier
    ExecutionID     string `json:"execution_id"`      // Unique execution identifier
    Timestamp       string `json:"timestamp"`         // ISO 8601 timestamp
}
```

### BaseOutput

Pre-hooks that support decisions use this structure:

```go
type BaseOutput struct {
    Decision string `json:"decision,omitempty"` // "allow", "deny", "ask"
    Message  string `json:"message,omitempty"`  // Shown when denied
}
```

### Base Helper Functions

```go
func Allow() BaseOutput              // Permit the action
func Deny(message string) BaseOutput // Block the action
func Ask(message string) BaseOutput  // Prompt user to confirm
```

---

## PreRunCommand

Called before a shell command executes.

### PreRunCommandInput

```go
type PreRunCommandToolInfo struct {
    CommandLine string `json:"command_line"`
    Cwd         string `json:"cwd"`
}

type PreRunCommandInput struct {
    BaseInput
    ToolInfo PreRunCommandToolInfo `json:"tool_info"`
}
```

### PreRunCommandOutput

```go
type PreRunCommandOutput struct {
    BaseOutput
}
```

### Helper Functions

```go
func AllowCommand() PreRunCommandOutput
func DenyCommand(message string) PreRunCommandOutput
func AskCommand(message string) PreRunCommandOutput
```

### Example

Cascade uses exit code 2 to block actions, so pre-hooks should use `RunE`. When the handler returns an error, the process exits with code 2 and Cascade blocks the action.

```go
hookshot.Register("cascade-pre-run-command", func() {
    hookshot.RunE(func(input cascade.PreRunCommandInput) (cascade.PreRunCommandOutput, error) {
        if strings.Contains(input.ToolInfo.CommandLine, "rm -rf /") {
            return cascade.PreRunCommandOutput{}, fmt.Errorf("Dangerous command blocked")
        }
        return cascade.AllowCommand(), nil
    })
})
```

> **Note:** The unified handler `OnBeforeExecution` already handles this correctly — it uses `RunE` for Cascade pre-hooks automatically. You only need `Register` for Cascade-specific hooks not covered by the unified API.

---

## PostRunCommand

Called after a shell command executes.

### PostRunCommandInput

```go
type PostRunCommandToolInfo struct {
    CommandLine string `json:"command_line"`
    Cwd         string `json:"cwd"`
    ExitCode    int    `json:"exit_code"`
    Output      string `json:"output,omitempty"`
}

type PostRunCommandInput struct {
    BaseInput
    ToolInfo PostRunCommandToolInfo `json:"tool_info"`
}
```

### PostRunCommandOutput

```go
type PostRunCommandOutput struct{} // Fire-and-forget
```

### Helper Functions

```go
func PostRunCommandOK() PostRunCommandOutput
```

---

## PreWriteCode

Called before writing a file.

### PreWriteCodeInput

```go
type PreWriteCodeToolInfo struct {
    FilePath string `json:"file_path"`
    Content  string `json:"content"`
}

type PreWriteCodeInput struct {
    BaseInput
    ToolInfo PreWriteCodeToolInfo `json:"tool_info"`
}
```

### PreWriteCodeOutput

```go
type PreWriteCodeOutput struct {
    BaseOutput
}
```

### Helper Functions

```go
func AllowWrite() PreWriteCodeOutput
func DenyWrite(message string) PreWriteCodeOutput
func AskWrite(message string) PreWriteCodeOutput
```

### Example

```go
hookshot.Register("cascade-pre-write-code", func() {
    hookshot.RunE(func(input cascade.PreWriteCodeInput) (cascade.PreWriteCodeOutput, error) {
        if strings.HasSuffix(input.ToolInfo.FilePath, ".env") {
            return cascade.PreWriteCodeOutput{}, fmt.Errorf("Cannot write to .env files")
        }
        return cascade.AllowWrite(), nil
    })
})
```

---

## PostWriteCode

Called after writing a file.

### PostWriteCodeInput

```go
type PostWriteCodeToolInfo struct {
    FilePath string `json:"file_path"`
    Content  string `json:"content"`
}

type PostWriteCodeInput struct {
    BaseInput
    ToolInfo PostWriteCodeToolInfo `json:"tool_info"`
}
```

### PostWriteCodeOutput

```go
type PostWriteCodeOutput struct{} // Fire-and-forget
```

### Helper Functions

```go
func PostWriteCodeOK() PostWriteCodeOutput
```

---

## PreReadCode

Called before reading a file.

### PreReadCodeInput

```go
type PreReadCodeToolInfo struct {
    FilePath string `json:"file_path"`
}

type PreReadCodeInput struct {
    BaseInput
    ToolInfo PreReadCodeToolInfo `json:"tool_info"`
}
```

### PreReadCodeOutput

```go
type PreReadCodeOutput struct {
    BaseOutput
}
```

### Helper Functions

```go
func AllowRead() PreReadCodeOutput
func DenyRead(message string) PreReadCodeOutput
func AskRead(message string) PreReadCodeOutput
```

### Example

```go
hookshot.Register("cascade-pre-read-code", func() {
    hookshot.RunE(func(input cascade.PreReadCodeInput) (cascade.PreReadCodeOutput, error) {
        if strings.Contains(input.ToolInfo.FilePath, "secrets") {
            return cascade.PreReadCodeOutput{}, fmt.Errorf("Cannot read secret files")
        }
        return cascade.AllowRead(), nil
    })
})
```

---

## PostReadCode

Called after reading a file.

### PostReadCodeInput

```go
type PostReadCodeToolInfo struct {
    FilePath string `json:"file_path"`
    Content  string `json:"content,omitempty"`
}

type PostReadCodeInput struct {
    BaseInput
    ToolInfo PostReadCodeToolInfo `json:"tool_info"`
}
```

### PostReadCodeOutput

```go
type PostReadCodeOutput struct{} // Fire-and-forget
```

### Helper Functions

```go
func PostReadCodeOK() PostReadCodeOutput
```

---

## PreMCPToolUse

Called before an MCP tool executes.

### PreMCPToolUseInput

```go
type PreMCPToolUseToolInfo struct {
    ToolName  string `json:"tool_name"`
    ToolInput string `json:"tool_input"` // JSON string of parameters
    ServerURL string `json:"server_url,omitempty"`
}

type PreMCPToolUseInput struct {
    BaseInput
    ToolInfo PreMCPToolUseToolInfo `json:"tool_info"`
}
```

### PreMCPToolUseOutput

```go
type PreMCPToolUseOutput struct {
    BaseOutput
}
```

### Helper Functions

```go
func AllowMCP() PreMCPToolUseOutput
func DenyMCP(message string) PreMCPToolUseOutput
func AskMCP(message string) PreMCPToolUseOutput
```

### Example

```go
hookshot.Register("cascade-pre-mcp-tool-use", func() {
    hookshot.RunE(func(input cascade.PreMCPToolUseInput) (cascade.PreMCPToolUseOutput, error) {
        if strings.Contains(input.ToolInfo.ServerURL, "blocked.com") {
            return cascade.PreMCPToolUseOutput{}, fmt.Errorf("MCP server not allowed")
        }
        return cascade.AllowMCP(), nil
    })
})
```

---

## PostMCPToolUse

Called after an MCP tool executes.

### PostMCPToolUseInput

```go
type PostMCPToolUseToolInfo struct {
    ToolName   string `json:"tool_name"`
    ToolInput  string `json:"tool_input"`
    ToolOutput string `json:"tool_output,omitempty"`
}

type PostMCPToolUseInput struct {
    BaseInput
    ToolInfo PostMCPToolUseToolInfo `json:"tool_info"`
}
```

### PostMCPToolUseOutput

```go
type PostMCPToolUseOutput struct{} // Fire-and-forget
```

### Helper Functions

```go
func PostMCPToolUseOK() PostMCPToolUseOutput
```

---

## PreUserPrompt

Called when the user submits a prompt.

### PreUserPromptInput

```go
type PreUserPromptToolInfo struct {
    Prompt string `json:"prompt"`
}

type PreUserPromptInput struct {
    BaseInput
    ToolInfo PreUserPromptToolInfo `json:"tool_info"`
}
```

### PreUserPromptOutput

```go
type PreUserPromptOutput struct {
    BaseOutput
}
```

### Helper Functions

```go
func AllowPrompt() PreUserPromptOutput
func BlockPrompt(message string) PreUserPromptOutput
```

### Example

```go
hookshot.Register("cascade-pre-user-prompt", func() {
    hookshot.RunE(func(input cascade.PreUserPromptInput) (cascade.PreUserPromptOutput, error) {
        if strings.Contains(input.ToolInfo.Prompt, "api_key=") {
            return cascade.PreUserPromptOutput{}, fmt.Errorf("Don't include API keys in prompts")
        }
        return cascade.AllowPrompt(), nil
    })
})
```

---

## PostCascadeResponse

Called after Cascade completes a response.

### PostCascadeResponseInput

```go
type PostCascadeResponseToolInfo struct {
    Response string `json:"response"`
}

type PostCascadeResponseInput struct {
    BaseInput
    ToolInfo PostCascadeResponseToolInfo `json:"tool_info"`
}
```

### PostCascadeResponseOutput

```go
type PostCascadeResponseOutput struct{} // Fire-and-forget
```

### Helper Functions

```go
func PostCascadeResponseOK() PostCascadeResponseOutput
```

---

## PostSetupWorktree

Called after worktree setup completes.

### PostSetupWorktreeInput

```go
type PostSetupWorktreeToolInfo struct {
    WorktreePath string `json:"worktree_path"`
}

type PostSetupWorktreeInput struct {
    BaseInput
    ToolInfo PostSetupWorktreeToolInfo `json:"tool_info"`
}
```

### PostSetupWorktreeOutput

```go
type PostSetupWorktreeOutput struct{} // Fire-and-forget
```

### Helper Functions

```go
func PostSetupWorktreeOK() PostSetupWorktreeOutput
```
