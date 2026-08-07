# Hooks Configuration & Environment Variable Issues

Based on analysis of hookshot against the official Claude Code hooks documentation.

---

## Issue 1: Missing `Cwd` field in PromptContext for Claude Code

**File:** `unified.go:693-698`

**Description:** The `PromptContext` for Claude Code doesn't populate the `Cwd` field from `input.Cwd`, even though `claude.UserPromptSubmitInput` inherits from `claude.BaseInput` which includes `Cwd`.

```go
ctx := PromptContext{
    Platform:      PlatformClaude,
    SessionID:     input.SessionID,
    Prompt:        input.Prompt,
    RawClaudeCode: &input,
    // Missing: Cwd: input.Cwd,
}
```

---

## Issue 2: Missing `duration_ms` field in PostToolUseInput

**File:** `claude/types.go:190-198`

**Description:** The official docs mention `duration_ms` is included in PostToolUse input, but it's not defined in the struct.

```go
type PostToolUseInput struct {
    BaseInput
    ToolName     string          `json:"tool_name"`
    ToolInput    json.RawMessage `json:"tool_input"`
    ToolResponse json.RawMessage `json:"tool_response"`
    ToolUseID    string          `json:"tool_use_id"`
    // Missing: DurationMs int64 `json:"duration_ms"`
}
```

---

## Issue 3: Missing `defer` permission decision helper

**File:** `claude/helpers.go`

**Description:** The official docs mention `permissionDecision: "defer"` for PreToolUse hooks in non-interactive mode (SDK with `-p` flag), but there's no `Defer()` helper function.

**Reference:** Per docs, "defer" causes process to exit with `stop_reason: "tool_deferred"`.

---

## Issue 4: Missing environment variable documentation

**Files:** `claude/doc.go`, `doc.go`, `examples/multi-hook/main.go`

**Description:** The official Claude Code documentation lists these environment variables available to command hooks, but they're not documented in hookshot:

- `$CLAUDE_PROJECT_DIR` - Project root directory
- `$CLAUDE_ENV_FILE` - Path to write persistent environment variables (SessionStart only)
- `${CLAUDE_PLUGIN_ROOT}` - Plugin installation directory
- `${CLAUDE_PLUGIN_DATA}` - Plugin persistent data directory

---

## Issue 5: Missing hook events not implemented

**Files:** `claude/types.go`, `unified.go`

**Description:** Several hook events from the official documentation are not implemented:

| Missing Event | Description |
|---------------|-------------|
| `UserPromptExpansion` | When a slash command expands |
| `PermissionDenied` | When permission is denied |
| `PostToolUseFailure` | After a tool call fails |
| `PostToolBatch` | After parallel tool calls resolve |
| `PostCompact` | After context compaction |
| `StopFailure` | When stop fails (rate_limit, auth_failed) |
| `SubagentStart` | When a subagent starts |
| `InstructionsLoaded` | When CLAUDE.md instructions load |
| `ConfigChange` | When settings change |
| `CwdChanged` | When working directory changes |
| `FileChanged` | When watched files change |
| `WorktreeRemove` | When a worktree is removed |

---

## Issue 6: Incorrect documentation URL

**File:** `claude/doc.go:133`

**Description:** The URL points to a non-existent pattern:
```go
// See https://docs.claude.com/en/docs/claude-code/hooks for full documentation.
```

Should match actual URL pattern (e.g., `https://code.claude.com/docs/en/hooks.md`).

---

## Issue 7: Missing `is_interrupt` field in PostToolUseFailure context

**File:** Not implemented

**Description:** Per official docs, `PostToolUseFailure` input should include:
- `tool_input`
- `error`
- `is_interrupt` (boolean)
- `duration_ms`

---

## Issue 8: Factory Droid configuration example may be incorrect

**File:** `examples/multi-hook/main.go:50-55`

**Description:** The Factory Droid configuration shows hooks without the nested `"hooks"` array wrapper:
```json
{
  "hooks": {
    "Stop": [{ "command": "/path/to/my-hooks droid-stop" }]
  }
}
```

But Claude Code (which Droid is based on) uses:
```json
{
  "hooks": {
    "Stop": [{ "hooks": [{ "type": "command", "command": "..." }] }]
  }
}
```

Needs verification against Factory Droid documentation.

---

## Issue 9: Missing `asyncRewake` behavior documentation

**Files:** All documentation

**Description:** The official docs mention `asyncRewake` for command hooks that run in background and wake Claude on exit code 2. This behavior pattern is not documented in hookshot.

---

## Issue 10: PostToolUseFailure not exposed in unified handlers

**File:** `unified.go`

**Description:** `OnAfterFileEdit` only handles successful tool execution (PostToolUse). There's no unified handler for failed tool executions that would wrap `PostToolUseFailure` across platforms.

---

## Issue 11: Missing `updatedInput` support in Droid file edit handler

**File:** `unified.go:576-583`

**Description:** Droid uses `old_str`/`new_str` field names (correctly handled), but the comment could be clearer about this difference from Claude Code's `old_string`/`new_string`.

---

## Issue 12: Cascade `Cwd` field missing in FileEditContext

**File:** `unified.go:612-630`

**Description:** When handling Cascade's `PostWriteCode`, the `Cwd` field is not populated in `FileEditContext`, even though the context struct supports it.

```go
ctx := FileEditContext{
    Platform:   PlatformCascade,
    SessionID:  input.TrajectoryID,
    FilePath:   input.ToolInfo.FilePath,
    Edits:      edits,
    RawCascade: &input,
    // Missing: Cwd field population
}
```

---

## Summary

| Severity | Count | Description |
|----------|-------|-------------|
| High | 3 | Missing required fields (Cwd, duration_ms, is_interrupt) |
| Medium | 5 | Missing hook events and helper functions |
| Low | 4 | Documentation and example inconsistencies |
