# OpenAI Codex API Reference

Codex hooks use the same JSON wire format as Claude Code hooks. The `codex`
Go package re-exports the relevant types and helpers from `claude` so your
code can stay platform-explicit. **For type definitions and helper-function
signatures see [reference-claude.md](reference-claude.md)** — this page only
documents what differs on Codex.

See the upstream spec at <https://developers.openai.com/codex/hooks>.

## Configuration

Codex hooks are enabled by default (the `hooks` feature flag is stable). If
your organization disabled hooks, set `[features].hooks = true` in
`~/.codex/config.toml` to re-enable. The older `codex_hooks` key still works
as a deprecated alias.

Hook commands live in `~/.codex/hooks.json` (or an inline `[hooks]` table in
`~/.codex/config.toml`). The minimum useful layout:

```json
{
  "hooks": {
    "PreToolUse":       [{ "matcher": "Bash|apply_patch|mcp__.*", "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-pre-tool-use" }] }],
    "PostToolUse":      [{ "matcher": "Bash|apply_patch|mcp__.*", "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-post-tool-use" }] }],
    "UserPromptSubmit": [{                                          "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-user-prompt-submit" }] }],
    "Stop":             [{                                          "hooks": [{ "type": "command", "command": "/path/to/my-hooks codex-stop", "timeout": 30 }] }]
  }
}
```

`hookshot install --codex --binary /path/to/my-hooks` will generate this for
you.

> **Why `mcp__.*` is in the matcher.** Codex passes MCP tool names to
> PreToolUse / PostToolUse using the `mcp__server__tool` convention.
> Omitting `mcp__.*` would silently bypass any `OnBeforeExecution` policy
> meant to enforce MCP allowlists.
>
> **Why `Edit|Write` is not in the matcher.** Codex emits `Edit` and `Write`
> as *matcher aliases* for `apply_patch`. The canonical `tool_name` Codex
> sends to the hook is always `apply_patch`, so a matcher of `apply_patch`
> alone covers every file-edit call.

## Events

| Event | Types | Codex notes |
|---|---|---|
| SessionStart | `codex.SessionStartInput` / `Output` | `source` is `startup`, `resume`, or `clear`. |
| PreToolUse | `codex.PreToolUseInput` / `Output` | `tool_name` is `Bash`, `apply_patch`, or `mcp__server__tool`. See **Ask is not enforced** below. |
| PermissionRequest | `codex.PermissionRequestInput` / `Output` | Codex-specific. Fires only when Codex is about to surface an approval prompt; see below. |
| PostToolUse | `codex.PostToolUseInput` / `Output` | See **apply_patch on the unified API** below. |
| UserPromptSubmit | `codex.UserPromptSubmitInput` / `Output` | Same shape as Claude. |
| Stop | `codex.StopInput` / `Output` | Same shape as Claude; Codex expects JSON on stdout (not plain text). |

Codex also sends a `model` field (active model slug) on every hook event and
a `turn_id` field on turn-scoped events (`PreToolUse`, `PermissionRequest`,
`PostToolUse`, `UserPromptSubmit`, `Stop`). These aren't on the shared
`BaseInput` struct — read them with `hookshot.ReadRawInput` if you need
them. Stop also carries `last_assistant_message`.

Codex enforces `continue: false` on `SessionStart`, `UserPromptSubmit`,
`PostToolUse`, and `Stop`. For `PreToolUse` and `PermissionRequest`, Codex
rejects `continue`, `stopReason`, and `suppressOutput` with an
`unsupported <field>` error and discards the whole hook output — these
fields fail **closed** and must be omitted. The upstream
[Codex hooks doc](https://developers.openai.com/codex/hooks) currently
describes these as fail-open; the runtime behavior is fail-closed.

## PreToolUse: which output fields actually work

Codex honors `permissionDecision: "deny"` (or the older `decision: "block"`
shape) on Bash and `apply_patch`. Codex also honors
`hookSpecificOutput.additionalContext`, which is injected as developer
context without blocking the call (the upstream
[Codex hooks doc](https://developers.openai.com/codex/hooks#pretooluse)
shows this case as a first-class example).

`permissionDecision: "ask"` is parsed but currently fails open.
`hookshot.OnBeforeExecution` returning `AskExecution(...)` is rewritten to
`Deny` on Codex so policies that require user confirmation aren't silently
bypassed. The platform-level `codex.Ask` helper still emits `"ask"` in the
JSON for forward-compat testing. If you want to react when Codex is
actually about to prompt the user, register a separate handler for the
**PermissionRequest** event below — that event's enforcement is supported.

**`updatedInput`, `continue: false`, `stopReason`, and `suppressOutput`
fail closed** — Codex rejects the whole hook output with
`PreToolUse hook returned unsupported <field>` rather than ignoring the
field, so these must be omitted. Concretely, this means:

- `codex.Allow(reason)`, `codex.Deny(reason)`, and `codex.PassThrough()`
  are all safe on Codex.
- `codex.AllowSilent()` is **not safe**: it sets `suppressOutput: true`,
  which Codex rejects. Use `codex.PassThrough()` for an empty
  no-side-effects allow.
- `codex.AllowWithInput(reason, input)` is **not safe**: it sets
  `updatedInput`, which Codex rejects. There's no Codex-supported way to
  mutate `tool_input` from a PreToolUse hook today — fall back to
  injecting state through `additionalContext` instead.
- To attach model-visible context to an allow, return
  `codex.AllowWithContext(reason, context)` (or build the output by hand
  with `hookSpecificOutput.additionalContext`).

The hookshot unified bridge already strips `suppressOutput` from Codex
`OnBeforeExecution(AllowExecution())` outputs, but if you call the
platform-level helpers directly you need to pick safe ones yourself.

## PermissionRequest (Codex-only)

Fires when Codex is about to ask for approval (shell escalation, managed-
network approval). Doesn't run for commands that don't need approval. The
`tool_input` may include a `description` field with a human-readable reason.

If multiple matching hooks return decisions, any `deny` wins. Otherwise an
`allow` lets the request proceed without surfacing the approval prompt. If
no matching hook decides, Codex uses the normal approval flow. `updatedInput`,
`updatedPermissions`, and `interrupt` are reserved for future behavior and
fail closed today.

Helpers: `codex.AllowPermission()` and `codex.DenyPermission(message string)`.

## apply_patch on the unified API

`hookshot.OnAfterFileEdit` parses Codex `apply_patch` events by unpacking
the unified-diff envelope in `tool_input.command` and invoking your handler
**once per file** mentioned in the patch. Each invocation receives a fully
populated `FileEditContext`:

- `FilePath` is the path declared in the `*** Add File:`,
  `*** Update File:`, or `*** Delete File:` section.
- `NewFilePath` is the destination path for rename operations
  (`*** Move to:`); empty otherwise.
- `Edits` is `[{OldString: "", NewString: <added content>}]` for Add, one
  `FileEdit` per hunk for Update, and empty for Delete.

For renames (`*** Update File: <src>` followed by `*** Move to: <dst>`) the
handler is invoked **twice** — once with `FilePath` set to the source and
once with `FilePath` set to the destination — and `NewFilePath` is populated
on both. This means a FilePath-only allowlist that permits the benign source
still receives a separate call for the destination so it can deny moves to
sensitive locations like `../../.ssh/authorized_keys`. Policies that want
to react specifically to renames should check
`ctx.NewFilePath != "" && ctx.NewFilePath != ctx.FilePath`.

If any per-file invocation returns `FileEditBlock`, the unified bridge
concatenates the reasons and emits a single `PostToolBlock`.

The same parser is also exported as
`codex.ParseApplyPatch(rawCommand string) []codex.PatchFile` for callers
that want to parse a patch envelope themselves — for example from the raw
`codex.PostToolUseInput.ToolInput`.

## PostToolUse semantics

`decision: "block"` doesn't undo the completed tool call. Codex records the
feedback, replaces the tool result with it, and continues the model from
the hook-provided message. To stop normal processing of the original tool
result, also return `continue: false`. `updatedMCPToolOutput` and
`suppressOutput` are parsed but not supported today.

## Stop semantics

`decision: "block"` doesn't reject the turn. Instead it tells Codex to
continue and creates a new continuation prompt that acts as a new user
prompt, using `reason` as that prompt text. If any matching Stop hook
returns `continue: false`, that takes precedence over continuation
decisions from other matching Stop hooks.

## Example

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

You can also use exit code `2` with the reason written to stderr instead of
returning the JSON output — that's what `RunE` does for you when the
handler returns an error.
