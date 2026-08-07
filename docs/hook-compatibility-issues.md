# Hook Compatibility Issues

Concise issues found while comparing the implementation against the official hook documentation links in [hook-documentation-links.md](hook-documentation-links.md).

1. **Cursor generic tool hooks are not modeled.** Current Cursor docs define `preToolUse`, `postToolUse`, and `postToolUseFailure` for all tool types, but `cursor/types.go` only models shell/MCP-specific execution inputs and `OnBeforeExecution` only registers `cursor-before-shell` / `cursor-before-mcp`. This misses Read/Write/Task/tool-wide policies and output features such as `updated_input`.
   - References: `cursor/types.go` (`BeforeShellExecutionInput`, `BeforeMCPExecutionInput`, `BeforeExecutionOutput`), `unified.go` (`OnBeforeExecution`), `cmd/hookshot/main.go` (`installCursor`)

2. **Factory Droid shell commands are classified with Claude tool names.** Droid docs use `Execute` for shell commands, but `OnBeforeExecution` checks `input.ToolName == "Bash"` for Droid. Execute hooks will be surfaced as `ExecutionTool` with no command extracted, so unified command-blocking handlers will not work for Droid shell commands.
   - References: `unified.go` (`OnBeforeExecution`)

3. **Factory Droid file creation is not handled.** Droid docs use `Create|Edit` for file writes, but the unified Droid file-edit adapter and installer use `Write|Edit`. `Create` events are ignored, so `OnAfterFileEdit` misses newly created files on Droid.
   - References: `unified.go` (`OnAfterFileEdit`), `cmd/hookshot/main.go` (`installDroid`)

4. **Cascade helpers imply JSON decisions that the docs do not describe.** Windsurf Cascade docs describe blocking via exit code 2/stderr for pre-hooks; `cascade` helper outputs include JSON `decision` fields (`allow`, `deny`, `ask`) that may be ignored by Cascade when users call `hookshot.Run` instead of `RunE`. The unified adapter uses `RunE`, but platform-specific helper examples can mislead users.
   - References: `cascade/types.go` (`BaseOutput`), `cascade/helpers.go` (`Allow`, `Deny`, `Ask`, `AllowCommand`, `DenyCommand`, `AskCommand`)

5. **Claude documentation links and event coverage are stale.** The repo still points to older Claude hook URLs and only models an older subset of events, while current Claude docs include additional events such as `UserPromptExpansion`, `PermissionDenied`, `PostToolBatch`, `StopFailure`, `SubagentStart`, `TaskCreated`, `TaskCompleted`, `InstructionsLoaded`, `ConfigChange`, `CwdChanged`, `FileChanged`, `WorktreeCreate`, `WorktreeRemove`, `PostCompact`, `Elicitation`, and `ElicitationResult`.
   - References: `README.md` (Documentation links), `claude/types.go` (event type coverage)
