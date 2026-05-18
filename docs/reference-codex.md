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
| PostToolUse | `codex.PostToolUseInput` / `Output` | Codex 0.130.0+ routes most file edits through `Bash` rather than `apply_patch`/`Write`/`Edit`. See **Bash bridge** and **apply_patch on the unified API** below. |
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

`permissionDecision: "ask"` is parsed by Codex but currently not enforced
at the platform level — emitting it would silently allow the tool to run.
To prevent "require confirmation" policies from turning into free passes,
both the high-level and platform-level APIs fail closed:

- `hookshot.OnBeforeExecution` returning `AskExecution(...)` is rewritten
  to `Deny` on Codex.
- `codex.Ask(reason)` is a fail-closed shim that returns `Deny(reason)` —
  it does not emit `"ask"`.

If you want to react when Codex is actually about to prompt the user,
register a separate handler for the **PermissionRequest** event below —
that event's enforcement is supported.

**`updatedInput`, `continue: false`, `stopReason`, and `suppressOutput`
fail closed** — Codex rejects the whole hook output with
`PreToolUse hook returned unsupported <field>` rather than ignoring the
field, so these must be omitted. Concretely, this means:

- `codex.Allow(reason)`, `codex.Deny(reason)`, and `codex.PassThrough()`
  are all safe on Codex.
- `codex.AllowSilent()` is **not safe**: it sets `suppressOutput: true`,
  which Codex rejects. Use `codex.PassThrough()` for an empty
  no-side-effects allow.
- `codex.AllowWithInput(reason, input)` is a **fail-closed shim**: there
  is no Codex-supported way to mutate `tool_input` from a PreToolUse
  hook (the CLI rejects `updatedInput`), and silently allowing the call
  with the original, unsanitized input would defeat the purpose of
  rewriting it. The helper therefore returns `Deny(reason + " (input
  rewriting not supported on Codex; fail-closed)")`. If you need to
  inject model-visible context on Codex, use `codex.AllowWithContext`
  (or `additionalContext` directly) instead.
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

## Bash bridge: file edits routed through Bash

Codex `0.130.0`+ routes most file operations through plain `Bash`
invocations rather than the native `apply_patch`, `Write`, or `Edit`
tools. The unified `hookshot.OnAfterFileEdit` bridge recognizes this
and dispatches the same per-file pipeline as the native shapes, so
policy handlers see one `FileEditContext` per file regardless of how
Codex chose to encode the edit.

The four shapes the bridge recognizes today:

| `tool_name`    | `tool_input.command` shape                                            | Parser                          |
|----------------|-----------------------------------------------------------------------|---------------------------------|
| `Write` / `Edit` | (native fields — uncommon on 0.130.0+)                              | inline                          |
| `apply_patch`  | unified-diff envelope, no Bash wrapper                                | `codex.ParseApplyPatch`         |
| `Bash`         | `apply_patch <<'PATCH' … *** End Patch … PATCH`                       | `codex.ParseApplyPatchFromBash` |
| `Bash`         | `cat <<'EOF' > FILE … EOF` or `tee FILE <<'EOF' … EOF`                | `codex.ParseBashRedirectWrite`  |

This is why the `PostToolUse` matcher in `~/.codex/hooks.json` must
include `Bash` (not just `apply_patch`):

```jsonc
"PostToolUse": [{ "matcher": "Bash|apply_patch|mcp__.*", /* … */ }]
```

`hookshot install --codex` writes this matcher by default. If you
configured hooks before adding the `Bash` token, you'll see zero
`afterFileEdit` events for any Codex session that edits or creates
files — that's the symptom that motivated this bridge.

### `apply_patch` via Bash heredoc

Edits to existing files almost always arrive as a Bash heredoc that
pipes a unified-diff envelope into `apply_patch`. Two shape variants
show up in practice:

- **Built-in name**: `apply_patch <<'PATCH' … *** End Patch … PATCH`
- **Per-session shim**: `/Users/me/.codex/tmp/arg0/codex-arg0…/apply_patch <<'PATCH' …`

`codex.ParseApplyPatchFromBash` accepts both. Detection requires the
literal token `apply_patch` followed by a heredoc operator (`<<` or
`<<-`) somewhere before the `*** Begin Patch` marker, which keeps
plain Bash commands that merely *mention* `apply_patch` (filenames,
docs, log lines) from spuriously firing the file-edit handler.

Once detected, the envelope is fed to `codex.ParseApplyPatch` and
the per-file dispatch behaves exactly like the native `apply_patch`
case described above — including the rename double-invocation for
`*** Move to:` paths.

### Greenfield writes via `cat` / `tee` heredoc

New-file creation typically arrives as `cat <<'EOF' > FILE … EOF`
(or, less commonly, `tee FILE <<'EOF' … EOF`). There's no prior
content to diff against, so `codex.ParseBashRedirectWrite`
synthesizes a single `PatchEdit{OldString: "", NewString: <body>}`
and a `PatchFile{Operation: "add", FilePath: <FILE>, Edits: …}`,
then dispatches it through the same per-file pipeline. Variants
currently recognized:

- quoted (`<<'EOF'`, `<<"EOF"`) and unquoted (`<<EOF`) delimiters,
- `<<-` tab-stripping (POSIX: strips leading **tabs**, not spaces,
  from each body line and from the delimiter line),
- `>` (overwrite) and `>>` (append) redirections,
- `cd … && cat <<'EOF' > FILE … EOF` cwd-prefixed forms.

File paths are surfaced verbatim — `../`, `~/`, and absolute paths
pass through unchanged so downstream policies can apply their own
canonicalization rules (the bridge can't safely guess the agent's
cwd, so it doesn't try).

If a Codex Bash command matches neither parser, the
`codex-post-tool-use` handler returns `PostToolOK()` without invoking
`OnAfterFileEdit`. That matches the platform semantics of every
other backend: non-edit Bash commands belong to `OnBeforeExecution`,
not `OnAfterFileEdit`.

### Calling the parsers from your own handler

Every parser used by the bridge is exported from the `codex` package
so handlers that want full control can short-circuit the unified
dispatch:

```go
func ParseApplyPatch(rawCommand string) []PatchFile
func ParseApplyPatchFromBash(bashCommand string) (files []PatchFile, ok bool)
func ParseBashRedirectWrite(bashCommand string) (files []PatchFile, ok bool)

type PatchFile struct {
    Operation   string // "add", "update", or "delete"
    FilePath    string
    NewFilePath string // set for "*** Move to:" renames
    Edits       []PatchEdit
}

type PatchEdit struct {
    OldString string
    NewString string
}
```

Wiring them up manually:

```go
hookshot.Register("codex-post-tool-use", func() {
    hookshot.Run(func(input codex.PostToolUseInput) codex.PostToolUseOutput {
        if input.ToolName != "Bash" {
            return codex.PostToolOK()
        }
        var bash struct{ Command string `json:"command"` }
        json.Unmarshal(input.ToolInput, &bash)

        if files, ok := codex.ParseApplyPatchFromBash(bash.Command); ok {
            return scanPatches(files)
        }
        if files, ok := codex.ParseBashRedirectWrite(bash.Command); ok {
            return scanPatches(files)
        }
        return codex.PostToolOK()
    })
})
```

### Known limitations

The Bash bridge is deliberately conservative — it short-circuits on
ambiguous input rather than guessing. Known gaps today:

- **`echo … > FILE` and `printf … > FILE` writes are not
  recognized.** Codex prefers heredocs for any multi-line content,
  so these are rare in practice. If you observe them in your
  sessions please open an issue with the captured
  `tool_input.command`.
- **Only the first `cat`/`tee` heredoc redirect in a single Bash
  command is currently surfaced.** A command that performs several
  heredoc writes in sequence (e.g.
  `cat <<EOF > a.txt … EOF; cat <<EOF > b.txt … EOF`) will fire
  `OnAfterFileEdit` only for the first file. As defence in depth
  against this gap, also register an `OnBeforeExecution` policy
  that grep/regex-scans `ctx.Command` for sensitive paths in Bash
  commands — that handler runs *before* any heredoc writes execute
  and isn't subject to the first-match restriction.
- **The bridge does not normalize file paths.** `../`, `~/`, and
  symbolic links are passed through verbatim. Handlers that want to
  enforce a containment policy should apply `filepath.Clean` and a
  cwd-escape check themselves (the bridge can't safely assume the
  agent's working directory matches the hook process's).

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
