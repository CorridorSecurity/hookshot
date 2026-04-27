# Hooks Documentation Reference

## Source Links

- Claude Code: https://code.claude.com/docs/en/hooks.md (fetched successfully)
- Cursor: https://cursor.com/docs/hooks.md (403 - access denied)
- Windsurf: https://docs.windsurf.com/windsurf/cascade/hooks.md (403 - access denied)
- Factory: https://docs.factory.ai/cli/configuration/hooks-guide.md (403 - access denied)

---

# Claude Code Hooks Reference - Complete Documentation

## Overview

Hooks are user-defined shell commands, HTTP endpoints, or LLM prompts that execute automatically at specific points in Claude Code's lifecycle. They enable automation, validation, and control over tool execution and session management.

## Hook Lifecycle

Hooks fire at three cadences:
- **Once per session**: `SessionStart`, `SessionEnd`
- **Once per turn**: `UserPromptSubmit`, `Stop`, `StopFailure`
- **Per tool call**: `PreToolUse`, `PostToolUse`, `PostToolUseFailure`, `PermissionRequest`, `PermissionDenied`

### Key Events Summary

| Event | When it fires |
|-------|---------------|
| `SessionStart` | When a session begins or resumes |
| `UserPromptSubmit` | Before Claude processes a prompt |
| `UserPromptExpansion` | When a slash command expands |
| `PreToolUse` | Before a tool call executes (can block) |
| `PermissionRequest` | When a permission dialog appears |
| `PostToolUse` | After a tool call succeeds |
| `PostToolUseFailure` | After a tool call fails |
| `PostToolBatch` | After parallel tool calls resolve |
| `Stop` | When Claude finishes responding |
| `SessionEnd` | When a session terminates |

## Configuration Structure

Hooks are defined in JSON with three nesting levels:

```json
{
  "hooks": {
    "EVENT_NAME": [
      {
        "matcher": "ToolName",
        "hooks": [
          {
            "type": "command|http|mcp_tool|prompt|agent",
            "command": "script.sh"
          }
        ]
      }
    ]
  }
}
```

### Hook Locations

| Location | Scope | Shareable |
|----------|-------|-----------|
| `~/.claude/settings.json` | All projects | No |
| `.claude/settings.json` | Single project | Yes |
| `.claude/settings.local.json` | Single project | No |
| Plugin `hooks/hooks.json` | When plugin enabled | Yes |
| Skill/Agent frontmatter | Component lifetime | Yes |

## Hook Handler Types

### 1. Command Hooks (`type: "command"`)

Run shell commands with JSON input on stdin:

```json
{
  "type": "command",
  "command": "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/validate.sh",
  "async": false,
  "shell": "bash"
}
```

**Fields:**
- `command` (required): Shell command to execute
- `async`: Run in background without blocking
- `asyncRewake`: Run in background, wake Claude on exit code 2
- `shell`: `"bash"` (default) or `"powershell"`

**Environment Variables:**
- `$CLAUDE_PROJECT_DIR`: Project root
- `${CLAUDE_PLUGIN_ROOT}`: Plugin installation directory
- `${CLAUDE_PLUGIN_DATA}`: Plugin persistent data directory

### 2. HTTP Hooks (`type: "http"`)

Send JSON to HTTP endpoints:

```json
{
  "type": "http",
  "url": "http://localhost:8080/hooks/validate",
  "headers": {
    "Authorization": "Bearer $MY_TOKEN"
  },
  "allowedEnvVars": ["MY_TOKEN"],
  "timeout": 30
}
```

**Fields:**
- `url` (required): POST endpoint URL
- `headers`: Custom HTTP headers with environment variable interpolation
- `allowedEnvVars`: Environment variables permitted in headers
- `timeout`: Seconds before canceling (default: 30)

### 3. MCP Tool Hooks (`type: "mcp_tool"`)

Call tools on connected MCP servers:

```json
{
  "type": "mcp_tool",
  "server": "my_server",
  "tool": "security_scan",
  "input": { "file_path": "${tool_input.file_path}" }
}
```

**Fields:**
- `server` (required): Configured MCP server name
- `tool` (required): Tool name on that server
- `input`: Tool arguments with `${path}` substitution support

### 4. Prompt Hooks (`type: "prompt"`)

Send prompts to Claude for evaluation:

```json
{
  "type": "prompt",
  "prompt": "Is this command safe? Command: $ARGUMENTS",
  "model": "fast-model",
  "timeout": 30
}
```

**Fields:**
- `prompt` (required): Prompt text with `$ARGUMENTS` placeholder
- `model`: Model for evaluation
- `timeout`: Seconds before canceling (default: 30)

### 5. Agent Hooks (`type: "agent"`)

Spawn subagents for verification (experimental):

```json
{
  "type": "agent",
  "prompt": "Verify this deployment is safe",
  "timeout": 60
}
```

## Common Hook Fields

Applied to all hook types:

```json
{
  "type": "command",
  "if": "Bash(git *)",
  "timeout": 600,
  "statusMessage": "Validating...",
  "once": false
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | yes | Hook type |
| `if` | no | Permission rule filter (tool events only) |
| `timeout` | no | Seconds before canceling |
| `statusMessage` | no | Custom spinner message |
| `once` | no | Run once per session then remove |

## Matcher Patterns

Matchers filter when hooks fire based on tool name, event source, or other criteria:

```json
{
  "matcher": "Bash",           // Exact match
  "matcher": "Edit|Write",     // Multiple tools (pipe-separated)
  "matcher": "^Notebook",      // Regex pattern
  "matcher": "mcp__.*__write"  // MCP tool pattern
}
```

**Matcher Evaluation:**
- `"*"`, `""`, or omitted: Match all occurrences
- Letters, digits, `_`, `|` only: Exact string or list
- Any other character: JavaScript regex

**Event-Specific Matchers:**

| Event | Matches | Examples |
|-------|---------|----------|
| Tool events | Tool name | `Bash`, `Edit\|Write`, `mcp__.*` |
| `SessionStart` | How session started | `startup`, `resume`, `clear`, `compact` |
| `Notification` | Notification type | `permission_prompt`, `auth_success` |
| `SubagentStart/Stop` | Agent type | `Explore`, `Plan`, custom names |
| `FileChanged` | Literal filenames | `.envrc\|.env` |
| `StopFailure` | Error type | `rate_limit`, `authentication_failed` |

## Hook Input and Output

### Common Input Fields

All hooks receive these JSON fields on stdin (command hooks) or POST body (HTTP hooks):

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/working/directory",
  "permission_mode": "default",
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": { "command": "npm test" }
}
```

**Subagent-specific fields:**
- `agent_id`: Unique subagent identifier
- `agent_type`: Agent name (e.g., `"Explore"`)

### Exit Code Output

```bash
exit 0    # Success - parse stdout for JSON decisions
exit 2    # Blocking error - read stderr, block action
exit 1+   # Non-blocking error - continue execution
```

### JSON Output Format

Exit 0 allows returning structured decisions:

```json
{
  "continue": true,
  "stopReason": "Optional message",
  "suppressOutput": false,
  "systemMessage": "Optional warning",
  "decision": "block",
  "reason": "Why blocked",
  "hookSpecificOutput": {
    "hookEventName": "EventName",
    "additionalContext": "Context for Claude"
  }
}
```

## Core Hook Events

### SessionStart

Runs when session begins or resumes. Fast execution recommended.

**Input:**
```json
{
  "source": "startup|resume|clear|compact",
  "model": "claude-sonnet-4-6"
}
```

**Output:**
```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "Development context"
  }
}
```

**Persist Environment Variables:**
```bash
#!/bin/bash
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo 'export NODE_ENV=production' >> "$CLAUDE_ENV_FILE"
  echo 'export DEBUG_LOG=true' >> "$CLAUDE_ENV_FILE"
fi
```

### UserPromptSubmit

Runs before Claude processes a user prompt. Can add context or block.

**Input:**
```json
{
  "prompt": "User's prompt text"
}
```

**Output (Block):**
```json
{
  "decision": "block",
  "reason": "Explanation to user",
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit",
    "additionalContext": "Context for Claude",
    "sessionTitle": "Auto-generated title"
  }
}
```

**Output (Add Context):**
Plain text stdout or JSON `additionalContext` field.

### PreToolUse

Runs before tool execution. Can allow, deny, ask, or defer.

**Matchers:** `Bash`, `Edit`, `Write`, `Read`, `Glob`, `Grep`, `Agent`, `WebFetch`, `WebSearch`, `AskUserQuestion`, and MCP tools.

**Example - Block dangerous commands:**
```bash
#!/bin/bash
COMMAND=$(jq -r '.tool_input.command')

if echo "$COMMAND" | grep -q 'rm -rf'; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "Destructive command blocked"
    }
  }'
else
  exit 0
fi
```

**Output:**
```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow|deny|ask|defer",
    "permissionDecisionReason": "Reason string",
    "updatedInput": { "command": "modified command" },
    "additionalContext": "Context for Claude"
  }
}
```

**Decision Precedence:** `deny` > `defer` > `ask` > `allow`

**Defer Tool Calls (Non-interactive Mode):**

For integrations using Claude Code SDK with `-p` flag:

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "defer"
  }
}
```

Process exits with `stop_reason: "tool_deferred"` and `deferred_tool_use` field containing the tool call. Resume later with `--resume` and return `"allow"` with answer in `updatedInput`.

### PermissionRequest

Runs when permission dialog appears. Allows auto-approval or auto-denial.

**Input:** Same `tool_name` and `tool_input` as PreToolUse, plus optional `permission_suggestions`.

**Output:**
```json
{
  "hookSpecificOutput": {
    "hookEventName": "PermissionRequest",
    "decision": {
      "behavior": "allow|deny",
      "updatedInput": { "command": "modified" },
      "updatedPermissions": [
        {
          "type": "addRules|replaceRules|removeRules|setMode|addDirectories",
          "rules": [{ "toolName": "Bash", "ruleContent": "npm *" }],
          "behavior": "allow",
          "destination": "localSettings|projectSettings|userSettings|session"
        }
      ],
      "message": "Why denied"
    }
  }
}
```

### PostToolUse

Runs after successful tool execution. Provides feedback to Claude.

**Input:** Includes `tool_input`, `tool_response`, `tool_use_id`, `duration_ms`.

**Output:**
```json
{
  "decision": "block",
  "reason": "Test suite failed",
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "additionalContext": "Action required",
    "updatedMCPToolOutput": "Modified output"
  }
}
```

### PostToolUseFailure

Runs when tool execution fails. Log failures or provide recovery guidance.

**Input:** `tool_input`, `error`, `is_interrupt`, `duration_ms`.

**Output:**
```json
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUseFailure",
    "additionalContext": "Recovery suggestions"
  }
}
```

### Stop

Runs when Claude finishes responding. Can prevent stopping.

**Output:**
```json
{
  "decision": "block",
  "reason": "Validation failed, continue"
}
```

### FileChanged

Watches specific files and fires on changes. Useful with direnv or environment setup.

**Matcher:** Literal filenames to watch (`.envrc|.env`)

**Access `CLAUDE_ENV_FILE`** to persist environment variable changes.

## Advanced Features

### Run Hooks in Background

Use `async` for non-blocking background execution:

```json
{
  "type": "command",
  "command": "long-running-task.sh",
  "async": true
}
```

Use `asyncRewake` to run in background and wake Claude on exit code 2:

```json
{
  "type": "command",
  "command": "background-validation.sh",
  "asyncRewake": true
}
```

### Async Hooks

Events that support async execution: `SessionStart`, `InstructionsLoaded`, `ConfigChange`, `CwdChanged`, `FileChanged`, `Notification`, `SubagentStart`, `SubagentStop`, `PostToolUse`, `PostToolUseFailure`, `PostToolBatch`, `PostCompact`, `SessionEnd`, `WorktreeRemove`.

### Conditional Execution with `if` Field

Filter hook execution using permission rule syntax:

```json
{
  "type": "command",
  "if": "Bash(git push *)",
  "command": "verify-git-safety.sh"
}
```

Runs only for Bash commands matching `git push *` pattern.

### Match MCP Tools

MCP tools use naming pattern: `mcp__<server>__<tool>`

Examples:
- `mcp__memory__create_entities`
- `mcp__filesystem__read_file`
- `mcp__memory__.*` (all memory server tools)
- `mcp__.*__write.*` (write tools from any server)

```json
{
  "matcher": "mcp__memory__.*",
  "hooks": [
    {
      "type": "command",
      "command": "echo 'Memory operation' >> ~/mcp.log"
    }
  ]
}
```

### Hooks in Skills and Agents

Define hooks in skill/agent frontmatter (YAML):

```yaml
---
name: secure-operations
description: Operations with security checks
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "./scripts/security-check.sh"
---
```

## Management

### View Configured Hooks

Type `/hooks` in Claude Code to browse all configured hooks with source location and full configuration.

### Disable Hooks

Set in settings file:
```json
{
  "disableAllHooks": true
}
```

Respects managed settings hierarchy - only affects user/project/local hooks, not managed policy hooks.

### Remove Hooks

Delete from settings JSON file. Direct edits are automatically detected by file watcher.

## Complete Example: Bash Command Validator

```bash
#!/bin/bash
# .claude/hooks/validate-commands.sh

COMMAND=$(jq -r '.tool_input.command')

# Block dangerous patterns
if echo "$COMMAND" | grep -qE '(rm -rf|:(){:|fork();|dd if=)'; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "Dangerous command pattern detected"
    }
  }'
  exit 0
fi

# Allow safe commands
if echo "$COMMAND" | grep -qE '^(ls|cat|npm test|npm run lint)'; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "allow"
    }
  }'
  exit 0
fi

# Ask for others
jq -n '{
  hookSpecificOutput: {
    hookEventName: "PreToolUse",
    permissionDecision: "ask",
    permissionDecisionReason: "Command requires approval"
  }
}'
exit 0
```

**Configuration:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/validate-commands.sh"
          }
        ]
      }
    ]
  }
}
```
