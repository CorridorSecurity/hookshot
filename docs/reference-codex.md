# OpenAI Codex API Reference

Platform-specific types and helpers for OpenAI Codex hooks. Use these when you need features not available in the unified API.

Codex hooks use the same JSON wire format as Claude Code hooks, configured in `~/.codex/hooks.json` or inline `[hooks]` tables in `~/.codex/config.toml`. The `codex` Go package re-exports the relevant `claude` types so your code can stay platform-explicit while still benefiting from the shared types.

Codex hooks are behind a feature flag — make sure `~/.codex/config.toml` contains:

```toml
[features]
codex_hooks = true
```

See the upstream spec at https://developers.openai.com/codex/hooks.

## Events

| Event | Input | Output | Description |
|-------|-------|--------|-------------|
| SessionStart | `SessionStartInput` | `SessionStartOutput` | Session started or resumed |
| PreToolUse | `PreToolUseInput` | `PreToolUseOutput` | Before tool execution (Bash, apply_patch, MCP) |
| PermissionRequest | `PermissionRequestInput` | `PermissionRequestOutput` | Approval prompt about to surface |
| PostToolUse | `PostToolUseInput` | `PostToolUseOutput` | After tool execution |
| UserPromptSubmit | `UserPromptSubmitInput` | `UserPromptSubmitOutput` | User submitted a prompt |
| Stop | `StopInput` | `StopOutput` | Turn finished responding |

---

## Common Types

### BaseInput

All Codex hook inputs include these fields:

```go
type BaseInput struct {
    SessionID      string `json:"session_id"`
    TranscriptPath string `json:"transcript_path"`
    Cwd            string `json:"cwd"`
    PermissionMode string `json:"permission_mode"`
    HookEventName  string `json:"hook_event_name"`
}
```

Codex also sends a `model` field (active model slug) on every hook event, and a `turn_id` field on turn-scoped events (`PreToolUse`, `PermissionRequest`, `PostToolUse`, `UserPromptSubmit`, `Stop`). These fields aren't represented on the shared `BaseInput` struct, but you can read them by binding the raw JSON yourself with `hookshot.ReadRawInput`.

### BaseOutput

Common fields for any hook output:

```go
type BaseOutput struct {
    Continue       *bool  `json:"continue,omitempty"`
    StopReason     string `json:"stopReason,omitempty"`
    SuppressOutput bool   `json:"suppressOutput,omitempty"` // parsed but not enforced today
    SystemMessage  string `json:"systemMessage,omitempty"`
}
```

Codex enforces `continue: false` on `SessionStart`, `UserPromptSubmit`, `PostToolUse`, and `Stop`. For `PreToolUse` and `PermissionRequest`, `continue`, `stopReason`, and `suppressOutput` are parsed but currently fail open.

---

## SessionStart

Called when a session starts or resumes. The `matcher` regex is applied to the `source` field. Current runtime values are `startup`, `resume`, and `clear`.

### SessionStartInput

```go
type SessionStartInput struct {
    BaseInput
    Source string `json:"source"` // "startup", "resume", "clear"
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
hookshot.Register("codex-session-start", func() {
    hookshot.Run(func(input codex.SessionStartInput) codex.SessionStartOutput {
        return codex.SessionStartContext("Project uses Go 1.21+")
    })
})
```

---

## PreToolUse

Called before Codex executes a tool. Currently intercepts simple Bash commands, file edits performed through `apply_patch`, and MCP tool calls. The `matcher` regex is applied to `tool_name` and matcher aliases — `apply_patch` also matches `Edit` and `Write`.

### PreToolUseInput

```go
type PreToolUseInput struct {
    BaseInput
    ToolName  string          `json:"tool_name"` // "Bash", "apply_patch", or "mcp__server__tool"
    ToolInput json.RawMessage `json:"tool_input"`
    ToolUseID string          `json:"tool_use_id"`
}
```

For `Bash` and `apply_patch`, the `tool_input` includes a `command` field. For MCP tools it carries all the arguments passed to the MCP call.

### PreToolUseOutput

```go
type PreToolUseOutput struct {
    BaseOutput
    HookSpecificOutput *PreToolUseHookOutput `json:"hookSpecificOutput,omitempty"`
}

type PreToolUseHookOutput struct {
    HookEventName            string         `json:"hookEventName,omitempty"`
    PermissionDecision       string         `json:"permissionDecision,omitempty"`
    PermissionDecisionReason string         `json:"permissionDecisionReason,omitempty"`
    UpdatedInput             map[string]any `json:"updatedInput,omitempty"`
}
```

Codex honors `permissionDecision: "deny"` (or the older `decision: "block"` shape) on Bash and `apply_patch`. `"allow"`, `"ask"`, `updatedInput`, `additionalContext`, `continue: false`, `stopReason`, and `suppressOutput` are parsed but fail open today.

> **Fail-closed `Ask` on the unified API.** Because Codex doesn't enforce `"ask"` yet, `hookshot.OnBeforeExecution` returning `AskExecution(...)` is translated to a `Deny` on Codex so policies that require user confirmation aren't silently bypassed. If you call the platform-level `codex.Ask` helper directly, the output JSON still encodes `"ask"` (so it round-trips with the upstream protocol) — that's only useful for forward-compat testing today.

### Helper Functions

```go
func Deny(reason string) PreToolUseOutput     // Enforced: blocks Bash and apply_patch
func Allow(reason string) PreToolUseOutput    // Parsed but currently falls through
func AllowSilent() PreToolUseOutput           // Parsed but currently falls through
func Ask(reason string) PreToolUseOutput      // Parsed but currently falls through (unified API rewrites to Deny)
func PassThrough() PreToolUseOutput           // Empty output, normal flow
```

### Example

```go
hookshot.Register("codex-pre-tool-use", func() {
    hookshot.Run(func(input codex.PreToolUseInput) codex.PreToolUseOutput {
        if input.ToolName == "Bash" {
            var ti struct{ Command string `json:"command"` }
            json.Unmarshal(input.ToolInput, &ti)
            if strings.Contains(ti.Command, "rm -rf /") {
                return codex.Deny("Destructive command blocked by hook.")
            }
        }
        return codex.PassThrough()
    })
})
```

You can also use exit code `2` with the reason written to stderr instead of returning the JSON output, which is what `RunE` does for you when the handler returns an error.

---

## PermissionRequest

Called when Codex is about to ask for approval (shell escalation, managed-network approval). It doesn't run for commands that don't need approval.

### PermissionRequestInput

```go
type PermissionRequestInput struct {
    BaseInput
    ToolName  string          `json:"tool_name"`
    ToolInput json.RawMessage `json:"tool_input"`
    ToolUseID string          `json:"tool_use_id"`
}
```

The `tool_input` may include a `description` field with a human-readable approval reason.

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
    Behavior string `json:"behavior"`           // "allow" or "deny"
    Message  string `json:"message,omitempty"`  // For "deny"
}
```

If multiple matching hooks return decisions, any `deny` wins. Otherwise, an `allow` lets the request proceed without surfacing the approval prompt. If no matching hook decides, Codex uses the normal approval flow. `updatedInput`, `updatedPermissions`, and `interrupt` are reserved for future behavior and fail closed today.

### Helper Functions

```go
func AllowPermission() PermissionRequestOutput
func DenyPermission(message string) PermissionRequestOutput
```

---

## PostToolUse

Called after Bash, `apply_patch`, or MCP tool calls produce output. For Bash, also runs after non-zero exits. Can't undo side effects.

### PostToolUseInput

```go
type PostToolUseInput struct {
    BaseInput
    ToolName     string          `json:"tool_name"`
    ToolInput    json.RawMessage `json:"tool_input"`
    ToolResponse json.RawMessage `json:"tool_response"`
    ToolUseID    string          `json:"tool_use_id"`
}
```

### PostToolUseOutput

```go
type PostToolUseOutput struct {
    BaseOutput
    Decision           string                 `json:"decision,omitempty"`
    Reason             string                 `json:"reason,omitempty"`
    HookSpecificOutput *PostToolUseHookOutput `json:"hookSpecificOutput,omitempty"`
}

type PostToolUseHookOutput struct {
    HookEventName     string `json:"hookEventName,omitempty"`
    AdditionalContext string `json:"additionalContext,omitempty"`
}
```

`decision: "block"` doesn't undo the completed tool call. Codex records the feedback, replaces the tool result with it, and continues the model from the hook-provided message. To stop normal processing of the original tool result, also return `continue: false`. `updatedMCPToolOutput` and `suppressOutput` are parsed but not supported today.

### Helper Functions

```go
func PostToolOK() PostToolUseOutput
func PostToolBlock(reason string) PostToolUseOutput
func PostToolContext(context string) PostToolUseOutput
```

### apply_patch parsing on the unified API

`hookshot.OnAfterFileEdit` parses Codex `apply_patch` events by unpacking the unified-diff envelope in `tool_input.command` and invoking your handler **once per file** mentioned in the patch. Each invocation receives a fully populated `FileEditContext`:

- `FilePath` is the path declared in the `*** Add File:`, `*** Update File:`, or `*** Delete File:` section.
- `Edits` is `[{OldString: "", NewString: <added content>}]` for Add, one `FileEdit` per hunk for Update (with removed lines as `OldString` and added lines as `NewString`), and empty for Delete.

If any of those per-file invocations returns `FileEditBlock`, the unified bridge concatenates the reasons and emits a single `PostToolBlock` so Codex replaces the tool result with the combined feedback. The platform-level `codex.PostToolUseInput` retains the raw `tool_input.command` if you'd rather parse the patch yourself.

### Example

```go
hookshot.Register("codex-post-tool-use", func() {
    hookshot.Run(func(input codex.PostToolUseInput) codex.PostToolUseOutput {
        if input.ToolName == "apply_patch" {
            return codex.PostToolContext("Generated files were updated.")
        }
        return codex.PostToolOK()
    })
})
```

---

## UserPromptSubmit

Called when the user submits a prompt. `matcher` is ignored for this event.

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
    Decision           string                      `json:"decision,omitempty"`
    Reason             string                      `json:"reason,omitempty"`
    HookSpecificOutput *UserPromptSubmitHookOutput `json:"hookSpecificOutput,omitempty"`
}

type UserPromptSubmitHookOutput struct {
    HookEventName     string `json:"hookEventName,omitempty"`
    AdditionalContext string `json:"additionalContext,omitempty"`
}
```

Return `decision: "block"` to reject the prompt. Otherwise, plain text on stdout (or `additionalContext` in JSON) is added as extra developer context.

### Helper Functions

```go
func AllowPrompt() UserPromptSubmitOutput
func BlockPrompt(reason string) UserPromptSubmitOutput
func AddContext(context string) UserPromptSubmitOutput
```

### Example

```go
hookshot.Register("codex-user-prompt-submit", func() {
    hookshot.Run(func(input codex.UserPromptSubmitInput) codex.UserPromptSubmitOutput {
        if strings.Contains(input.Prompt, "api_key=") {
            return codex.BlockPrompt("Don't include API keys in prompts")
        }
        return codex.AllowPrompt()
    })
})
```

---

## Stop

Called when the turn finishes responding. `matcher` is ignored for this event. Codex expects JSON on stdout when the hook exits 0 — plain text is invalid here.

### StopInput

```go
type StopInput struct {
    BaseInput
    StopHookActive bool `json:"stop_hook_active"`
}
```

Codex also sends `last_assistant_message` (latest assistant message text) on Stop input. Read it via `hookshot.ReadRawInput` if needed.

### StopOutput

```go
type StopOutput struct {
    BaseOutput
    Decision string `json:"decision,omitempty"` // "block" to continue the turn
    Reason   string `json:"reason,omitempty"`
}
```

For this event, `decision: "block"` doesn't reject the turn. Instead, it tells Codex to continue and creates a new continuation prompt that acts as a new user prompt, using `reason` as that prompt text. If any matching Stop hook returns `continue: false`, that takes precedence over continuation decisions from other matching Stop hooks.

### Helper Functions

```go
func Continue() StopOutput               // Allow stopping
func Block(reason string) StopOutput     // Continue the turn with reason as the next prompt
func StopWith(reason string) StopOutput  // Halt Codex entirely (continue=false)
```

### Example

```go
hookshot.Register("codex-stop", func() {
    hookshot.Run(func(input codex.StopInput) codex.StopOutput {
        // IMPORTANT: Check StopHookActive to prevent infinite loops
        if input.StopHookActive {
            return codex.Continue()
        }
        return codex.Continue()
    })
})
```

---

## Configuration Example

`~/.codex/config.toml`:

```toml
[features]
codex_hooks = true
```

`~/.codex/hooks.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash|apply_patch|mcp__.*",
        "hooks": [
          { "type": "command", "command": "/path/to/my-hooks codex-pre-tool-use" }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "apply_patch|Edit|Write|mcp__.*",
        "hooks": [
          { "type": "command", "command": "/path/to/my-hooks codex-post-tool-use" }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          { "type": "command", "command": "/path/to/my-hooks codex-user-prompt-submit" }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          { "type": "command", "command": "/path/to/my-hooks codex-stop", "timeout": 30 }
        ]
      }
    ]
  }
}
```

`hookshot install --codex --binary /path/to/my-hooks` will generate this layout for you, but it will not toggle the `codex_hooks` feature flag — set that yourself.

> **Why `mcp__.*` is in the matcher.** Codex passes MCP tool names to PreToolUse / PostToolUse using the `mcp__server__tool` convention. Omitting the `mcp__.*` alternative would mean Codex never invokes the hook binary for MCP calls, which would silently bypass any `OnBeforeExecution` policy meant to enforce MCP allowlists.
