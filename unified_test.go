package hookshot

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// =============================================================================
// StopContext Tests
// =============================================================================

func TestStopContext_ShouldSkip_Claude(t *testing.T) {
	tests := []struct {
		name           string
		stopHookActive bool
		wantSkip       bool
	}{
		{
			name:           "StopHookActive true should skip",
			stopHookActive: true,
			wantSkip:       true,
		},
		{
			name:           "StopHookActive false should not skip",
			stopHookActive: false,
			wantSkip:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := StopContext{
				Platform:       PlatformClaude,
				StopHookActive: tt.stopHookActive,
			}
			if got := ctx.ShouldSkip(); got != tt.wantSkip {
				t.Errorf("ShouldSkip() = %v, want %v", got, tt.wantSkip)
			}
		})
	}
}

func TestStopContext_ShouldSkip_Cursor(t *testing.T) {
	tests := []struct {
		name      string
		loopCount int
		wantSkip  bool
	}{
		{
			name:      "LoopCount 0 should not skip",
			loopCount: 0,
			wantSkip:  false,
		},
		{
			name:      "LoopCount 1 should not skip",
			loopCount: 1,
			wantSkip:  false,
		},
		{
			name:      "LoopCount 2 should not skip",
			loopCount: 2,
			wantSkip:  false,
		},
		{
			name:      "LoopCount 3 should skip",
			loopCount: 3,
			wantSkip:  true,
		},
		{
			name:      "LoopCount 5 should skip",
			loopCount: 5,
			wantSkip:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := StopContext{
				Platform:  PlatformCursor,
				LoopCount: tt.loopCount,
			}
			if got := ctx.ShouldSkip(); got != tt.wantSkip {
				t.Errorf("ShouldSkip() = %v, want %v", got, tt.wantSkip)
			}
		})
	}
}

func TestStopDecision_Helpers(t *testing.T) {
	// Test AllowStop
	allow := AllowStop()
	if !allow.Continue {
		t.Error("AllowStop should have Continue=true")
	}
	if allow.Message != "" {
		t.Error("AllowStop should have empty Message")
	}

	// Test PreventStop
	prevent := PreventStop("Please verify changes")
	if prevent.Continue {
		t.Error("PreventStop should have Continue=false")
	}
	if prevent.Message != "Please verify changes" {
		t.Errorf("PreventStop Message = %q, want %q", prevent.Message, "Please verify changes")
	}
}

// =============================================================================
// ExecutionContext Tests
// =============================================================================

func TestExecutionContext_IsMCP(t *testing.T) {
	tests := []struct {
		name   string
		execType ExecutionType
		want   bool
	}{
		{
			name:   "ExecutionMCP returns true",
			execType: ExecutionMCP,
			want:   true,
		},
		{
			name:   "ExecutionShell returns false",
			execType: ExecutionShell,
			want:   false,
		},
		{
			name:   "ExecutionTool returns false",
			execType: ExecutionTool,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ExecutionContext{Type: tt.execType}
			if got := ctx.IsMCP(); got != tt.want {
				t.Errorf("IsMCP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutionDecision_Helpers(t *testing.T) {
	// Test AllowExecution
	allow := AllowExecution()
	if !allow.Allow {
		t.Error("AllowExecution should have Allow=true")
	}
	if allow.Reason != "" {
		t.Error("AllowExecution should have empty Reason")
	}

	// Test AllowExecutionWithReason
	allowReason := AllowExecutionWithReason("Trusted source")
	if !allowReason.Allow {
		t.Error("AllowExecutionWithReason should have Allow=true")
	}
	if allowReason.Reason != "Trusted source" {
		t.Errorf("AllowExecutionWithReason Reason = %q, want %q", allowReason.Reason, "Trusted source")
	}

	// Test DenyExecution
	deny := DenyExecution("Not allowed")
	if deny.Allow {
		t.Error("DenyExecution should have Allow=false")
	}
	if deny.Reason != "Not allowed" {
		t.Errorf("DenyExecution Reason = %q, want %q", deny.Reason, "Not allowed")
	}

	// Test AskExecution
	ask := AskExecution("Confirm action?")
	if ask.Allow {
		t.Error("AskExecution should have Allow=false")
	}
	if !ask.Ask {
		t.Error("AskExecution should have Ask=true")
	}
	if ask.Reason != "Confirm action?" {
		t.Errorf("AskExecution Reason = %q, want %q", ask.Reason, "Confirm action?")
	}
}

// =============================================================================
// FileEditContext Tests
// =============================================================================

func TestFileEditDecision_Helpers(t *testing.T) {
	// Test FileEditOK
	ok := FileEditOK()
	if ok.Block {
		t.Error("FileEditOK should have Block=false")
	}
	if ok.Reason != "" || ok.Context != "" {
		t.Error("FileEditOK should have empty Reason and Context")
	}

	// Test FileEditBlock
	block := FileEditBlock("Issue found")
	if !block.Block {
		t.Error("FileEditBlock should have Block=true")
	}
	if block.Reason != "Issue found" {
		t.Errorf("FileEditBlock Reason = %q, want %q", block.Reason, "Issue found")
	}

	// Test FileEditAddContext
	ctx := FileEditAddContext("Additional info")
	if ctx.Block {
		t.Error("FileEditAddContext should have Block=false")
	}
	if ctx.Context != "Additional info" {
		t.Errorf("FileEditAddContext Context = %q, want %q", ctx.Context, "Additional info")
	}
}

// =============================================================================
// PromptContext Tests
// =============================================================================

func TestPromptDecision_Helpers(t *testing.T) {
	// Test AllowPromptDecision
	allow := AllowPromptDecision()
	if !allow.Allow {
		t.Error("AllowPromptDecision should have Allow=true")
	}

	// Test BlockPromptDecision
	block := BlockPromptDecision("Invalid prompt")
	if block.Allow {
		t.Error("BlockPromptDecision should have Allow=false")
	}
	if block.Reason != "Invalid prompt" {
		t.Errorf("BlockPromptDecision Reason = %q, want %q", block.Reason, "Invalid prompt")
	}

	// Test AddPromptContext
	ctx := AddPromptContext("Project uses Go 1.21")
	if !ctx.Allow {
		t.Error("AddPromptContext should have Allow=true")
	}
	if ctx.Context != "Project uses Go 1.21" {
		t.Errorf("AddPromptContext Context = %q, want %q", ctx.Context, "Project uses Go 1.21")
	}
}

// =============================================================================
// Platform Constants Tests
// =============================================================================

func TestPlatformConstants(t *testing.T) {
	if PlatformClaude != "claude" {
		t.Errorf("PlatformClaude = %q, want %q", PlatformClaude, "claude")
	}
	if PlatformCursor != "cursor" {
		t.Errorf("PlatformCursor = %q, want %q", PlatformCursor, "cursor")
	}
	if PlatformDroid != "droid" {
		t.Errorf("PlatformDroid = %q, want %q", PlatformDroid, "droid")
	}
	if PlatformCascade != "cascade" {
		t.Errorf("PlatformCascade = %q, want %q", PlatformCascade, "cascade")
	}
	if PlatformCodex != "codex" {
		t.Errorf("PlatformCodex = %q, want %q", PlatformCodex, "codex")
	}
}

func TestStopContext_ShouldSkip_Codex(t *testing.T) {
	// Codex uses StopHookActive like Claude Code and Droid.
	if !(StopContext{Platform: PlatformCodex, StopHookActive: true}).ShouldSkip() {
		t.Error("Codex StopHookActive=true should skip")
	}
	if (StopContext{Platform: PlatformCodex, StopHookActive: false}).ShouldSkip() {
		t.Error("Codex StopHookActive=false should not skip")
	}
}

// =============================================================================
// ExecutionContext Command Field Tests (for MCP local servers)
// =============================================================================

func TestExecutionContext_CommandField(t *testing.T) {
	// Test that Command field can be used for local MCP servers
	ctx := ExecutionContext{
		Platform:  PlatformCursor,
		Type:      ExecutionMCP,
		ToolName:  "analyze",
		ServerURL: "",
		Command:   "node /path/to/server.js",
	}

	if ctx.Command != "node /path/to/server.js" {
		t.Errorf("Command = %q, want %q", ctx.Command, "node /path/to/server.js")
	}

	// Verify IsMCP still works
	if !ctx.IsMCP() {
		t.Error("IsMCP should return true")
	}
}

// =============================================================================
// Unified Handler Registration Tests
// =============================================================================

func TestOnBeforeExecution_RegistersAllHandlers(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	OnBeforeExecution(func(ctx ExecutionContext) ExecutionDecision {
		return AllowExecution()
	})

	expectedHandlers := []string{
		"claude-pre-tool-use",
		"cursor-before-shell",
		"cursor-before-mcp",
		"droid-pre-tool-use",
		"cascade-pre-run-command",
		"cascade-pre-mcp-tool-use",
		"codex-pre-tool-use",
	}

	for _, name := range expectedHandlers {
		if _, ok := handlers[name]; !ok {
			t.Errorf("OnBeforeExecution did not register handler %q", name)
		}
	}
}

func TestOnPromptSubmit_RegistersAllHandlers(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	OnPromptSubmit(func(ctx PromptContext) PromptDecision {
		return AllowPromptDecision()
	})

	expectedHandlers := []string{
		"claude-user-prompt-submit",
		"cursor-before-submit-prompt",
		"droid-user-prompt-submit",
		"cascade-pre-user-prompt",
		"codex-user-prompt-submit",
	}

	for _, name := range expectedHandlers {
		if _, ok := handlers[name]; !ok {
			t.Errorf("OnPromptSubmit did not register handler %q", name)
		}
	}
}

func TestOnStop_RegistersAllHandlers(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	OnStop(func(ctx StopContext) StopDecision {
		return AllowStop()
	})

	expectedHandlers := []string{
		"claude-stop",
		"cursor-stop",
		"droid-stop",
		"cascade-post-cascade-response",
		"codex-stop",
	}

	for _, name := range expectedHandlers {
		if _, ok := handlers[name]; !ok {
			t.Errorf("OnStop did not register handler %q", name)
		}
	}
}

func TestOnAfterFileEdit_RegistersAllHandlers(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	OnAfterFileEdit(func(ctx FileEditContext) FileEditDecision {
		return FileEditOK()
	})

	expectedHandlers := []string{
		"claude-after-file-edit",
		"cursor-after-file-edit",
		"droid-after-file-edit",
		"cascade-post-write-code",
		"codex-post-tool-use",
	}

	for _, name := range expectedHandlers {
		if _, ok := handlers[name]; !ok {
			t.Errorf("OnAfterFileEdit did not register handler %q", name)
		}
	}
}

// =============================================================================
// Cascade RunE Behavior Tests (exit code 2 on deny, 0 on allow)
// =============================================================================

// TestPreHooks_RunE verifies that cascade and droid pre-hooks use RunE so that
// blocking decisions exit with code 2 (how Windsurf Cascade and Factory Droid detect denials).
func TestCascadePreHooks_RunE(t *testing.T) {
	// Subprocess entry point: when invoked as a subprocess, register handlers
	// and execute the specified cascade handler with controlled stdin.
	if handler := os.Getenv("HOOKSHOT_TEST_CASCADE_HANDLER"); handler != "" {
		decision := os.Getenv("HOOKSHOT_TEST_CASCADE_DECISION")

		ClearHandlers()
		OnBeforeExecution(func(ctx ExecutionContext) ExecutionDecision {
			if decision == "deny" {
				return DenyExecution("blocked by test")
			}
			return AllowExecution()
		})
		OnPromptSubmit(func(ctx PromptContext) PromptDecision {
			if decision == "deny" {
				return BlockPromptDecision("blocked by test")
			}
			return AllowPromptDecision()
		})

		handlers[handler]()
		return
	}

	tests := []struct {
		name       string
		handler    string
		stdinJSON  string
		decision   string
		wantExit   int
		wantStderr string // substring expected in stderr (empty = don't check)
	}{
		{
			name:       "cascade-pre-run-command deny exits 2",
			handler:    "cascade-pre-run-command",
			stdinJSON:  `{"trajectory_id":"test","tool_info":{"command_line":"rm -rf /","cwd":"/tmp"}}`,
			decision:   "deny",
			wantExit:   2,
			wantStderr: "blocked by test",
		},
		{
			name:      "cascade-pre-run-command allow exits 0",
			handler:   "cascade-pre-run-command",
			stdinJSON: `{"trajectory_id":"test","tool_info":{"command_line":"ls","cwd":"/tmp"}}`,
			decision:  "allow",
			wantExit:  0,
		},
		{
			name:       "cascade-pre-mcp-tool-use deny exits 2",
			handler:    "cascade-pre-mcp-tool-use",
			stdinJSON:  `{"trajectory_id":"test","tool_info":{"mcp_server_name":"example","mcp_tool_name":"analyze","mcp_tool_arguments":{}}}`,
			decision:   "deny",
			wantExit:   2,
			wantStderr: "blocked by test",
		},
		{
			name:      "cascade-pre-mcp-tool-use allow exits 0",
			handler:   "cascade-pre-mcp-tool-use",
			stdinJSON: `{"trajectory_id":"test","tool_info":{"mcp_server_name":"example","mcp_tool_name":"analyze","mcp_tool_arguments":{}}}`,
			decision:  "allow",
			wantExit:  0,
		},
		{
			name:       "cascade-pre-user-prompt deny exits 2",
			handler:    "cascade-pre-user-prompt",
			stdinJSON:  `{"trajectory_id":"test","tool_info":{"user_prompt":"hello"}}`,
			decision:   "deny",
			wantExit:   2,
			wantStderr: "blocked by test",
		},
		{
			name:      "cascade-pre-user-prompt allow exits 0",
			handler:   "cascade-pre-user-prompt",
			stdinJSON: `{"trajectory_id":"test","tool_info":{"user_prompt":"hello"}}`,
			decision:  "allow",
			wantExit:  0,
		},
		{
			name:       "droid-pre-tool-use deny exits 2",
			handler:    "droid-pre-tool-use",
			stdinJSON:  `{"session_id":"test","tool_name":"mcp__blocked__tool","tool_input":{},"cwd":"/tmp"}`,
			decision:   "deny",
			wantExit:   2,
			wantStderr: "blocked by test",
		},
		{
			name:      "droid-pre-tool-use allow exits 0",
			handler:   "droid-pre-tool-use",
			stdinJSON: `{"session_id":"test","tool_name":"mcp__allowed__tool","tool_input":{},"cwd":"/tmp"}`,
			decision:  "allow",
			wantExit:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestCascadePreHooks_RunE$")
			cmd.Env = append(os.Environ(),
				"HOOKSHOT_TEST_CASCADE_HANDLER="+tt.handler,
				"HOOKSHOT_TEST_CASCADE_DECISION="+tt.decision,
			)
			cmd.Stdin = strings.NewReader(tt.stdinJSON)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			exitCode := 0
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else if err != nil {
				t.Fatalf("unexpected error running subprocess: %v", err)
			}

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d (stderr: %s)", exitCode, tt.wantExit, stderr.String())
			}
			if tt.wantStderr != "" && !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want substring %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

// TestDroidPreToolUse_MCPTypeDetection verifies that droid's tool naming convention
// (serverName___toolName) is correctly classified as ExecutionMCP.
func TestDroidPreToolUse_MCPTypeDetection(t *testing.T) {
	if handler := os.Getenv("HOOKSHOT_TEST_DROID_MCP_HANDLER"); handler != "" {
		var capturedType ExecutionType
		ClearHandlers()
		OnBeforeExecution(func(ctx ExecutionContext) ExecutionDecision {
			capturedType = ctx.Type
			// Write the detected type to stdout so the parent can verify
			os.Stdout.WriteString("type=" + string(capturedType))
			return AllowExecution()
		})
		handlers[handler]()
		return
	}

	tests := []struct {
		name      string
		stdinJSON string
		wantType  string
	}{
		{
			name:      "Droid MCP tool with ___ delimiter",
			stdinJSON: `{"session_id":"test","tool_name":"corridor___listProjects","tool_input":{},"cwd":"/tmp"}`,
			wantType:  "type=mcp",
		},
		{
			name:      "Droid MCP tool with mcp__ prefix",
			stdinJSON: `{"session_id":"test","tool_name":"mcp__context7__query","tool_input":{},"cwd":"/tmp"}`,
			wantType:  "type=mcp",
		},
		{
			name:      "Droid built-in tool",
			stdinJSON: `{"session_id":"test","tool_name":"Read","tool_input":{},"cwd":"/tmp"}`,
			wantType:  "type=tool",
		},
		{
			name:      "Droid Bash tool",
			stdinJSON: `{"session_id":"test","tool_name":"Bash","tool_input":{"command":"ls"},"cwd":"/tmp"}`,
			wantType:  "type=shell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestDroidPreToolUse_MCPTypeDetection$")
			cmd.Env = append(os.Environ(),
				"HOOKSHOT_TEST_DROID_MCP_HANDLER=droid-pre-tool-use",
			)
			cmd.Stdin = strings.NewReader(tt.stdinJSON)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			exitCode := 0
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else if err != nil {
				t.Fatalf("unexpected error: %v (stderr: %s)", err, stderr.String())
			}

			if exitCode != 0 {
				t.Fatalf("exit code = %d, want 0 (stderr: %s)", exitCode, stderr.String())
			}

			// The handler writes the type to stdout, but the droid output JSON
			// may also be written. Check that our type marker is present.
			if !strings.Contains(stdout.String(), tt.wantType) {
				t.Errorf("stdout = %q, want substring %q", stdout.String(), tt.wantType)
			}
		})
	}
}

func TestExecutionContext_ToolInput_JSON(t *testing.T) {
	toolInput := json.RawMessage(`{"file_path":"/test.ts"}`)
	ctx := ExecutionContext{
		Platform:  PlatformClaude,
		Type:      ExecutionMCP,
		ToolName:  "mcp__corridor__analyze",
		ToolInput: toolInput,
	}

	// Parse ToolInput
	var parsed map[string]string
	if err := json.Unmarshal(ctx.ToolInput, &parsed); err != nil {
		t.Fatalf("Failed to parse ToolInput: %v", err)
	}

	if parsed["file_path"] != "/test.ts" {
		t.Errorf("file_path = %q, want %q", parsed["file_path"], "/test.ts")
	}
}

// =============================================================================
// Cursor workspace_roots -> ctx.Cwd Tests
// =============================================================================

func TestCursorWorkspaceRoot(t *testing.T) {
	tests := []struct {
		name  string
		roots []string
		want  string
	}{
		{name: "nil slice returns empty", roots: nil, want: ""},
		{name: "empty slice returns empty", roots: []string{}, want: ""},
		{name: "single root returned", roots: []string{"/work/proj"}, want: "/work/proj"},
		{name: "first of multiple returned", roots: []string{"/a", "/b"}, want: "/a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cursorWorkspaceRoot(tt.roots); got != tt.want {
				t.Errorf("cursorWorkspaceRoot(%v) = %q, want %q", tt.roots, got, tt.want)
			}
		})
	}
}

// runHandlerWithStdin invokes the named registered handler with the given JSON
// piped to os.Stdin, discarding anything the handler writes to os.Stdout.
func runHandlerWithStdin(t *testing.T, name, stdinJSON string) {
	t.Helper()
	h, ok := handlers[name]
	if !ok {
		t.Fatalf("handler %q not registered", name)
	}

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if _, err := stdinW.Write([]byte(stdinJSON)); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	stdinW.Close()
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer stdoutR.Close()

	origStdin, origStdout := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = stdinR, stdoutW
	defer func() {
		stdoutW.Close()
		os.Stdin, os.Stdout = origStdin, origStdout
	}()

	h()
}

func TestOnStop_CursorPopulatesCwd(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	var captured StopContext
	OnStop(func(ctx StopContext) StopDecision {
		captured = ctx
		return AllowStop()
	})

	stdin := `{"conversation_id":"c1","workspace_roots":["/Users/me/proj","/other"],"status":"completed","loop_count":0}`
	runHandlerWithStdin(t, "cursor-stop", stdin)

	if captured.Platform != PlatformCursor {
		t.Errorf("Platform = %q, want %q", captured.Platform, PlatformCursor)
	}
	if captured.Cwd != "/Users/me/proj" {
		t.Errorf("Cwd = %q, want %q", captured.Cwd, "/Users/me/proj")
	}
}

func TestOnBeforeExecution_CursorShellPrefersInputCwd(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	var captured ExecutionContext
	OnBeforeExecution(func(ctx ExecutionContext) ExecutionDecision {
		captured = ctx
		return AllowExecution()
	})

	stdin := `{"conversation_id":"c1","workspace_roots":["/root/from/ws"],"command":"ls","cwd":"/explicit/cwd"}`
	runHandlerWithStdin(t, "cursor-before-shell", stdin)

	if captured.Cwd != "/explicit/cwd" {
		t.Errorf("Cwd = %q, want %q (input.Cwd should win when set)", captured.Cwd, "/explicit/cwd")
	}
}

func TestOnBeforeExecution_CursorShellFallsBackToWorkspaceRoot(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	var captured ExecutionContext
	OnBeforeExecution(func(ctx ExecutionContext) ExecutionDecision {
		captured = ctx
		return AllowExecution()
	})

	stdin := `{"conversation_id":"c1","workspace_roots":["/root/from/ws"],"command":"ls","cwd":""}`
	runHandlerWithStdin(t, "cursor-before-shell", stdin)

	if captured.Cwd != "/root/from/ws" {
		t.Errorf("Cwd = %q, want %q (should fall back to workspace_roots[0])", captured.Cwd, "/root/from/ws")
	}
}

func TestOnBeforeExecution_CursorMCPPopulatesCwd(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	var captured ExecutionContext
	OnBeforeExecution(func(ctx ExecutionContext) ExecutionDecision {
		captured = ctx
		return AllowExecution()
	})

	stdin := `{"conversation_id":"c1","workspace_roots":["/root/from/ws"],"tool_name":"analyze","tool_input":"{}","url":"https://example.com"}`
	runHandlerWithStdin(t, "cursor-before-mcp", stdin)

	if captured.Cwd != "/root/from/ws" {
		t.Errorf("Cwd = %q, want %q", captured.Cwd, "/root/from/ws")
	}
}

func TestOnAfterFileEdit_CursorPopulatesCwd(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	var captured FileEditContext
	OnAfterFileEdit(func(ctx FileEditContext) FileEditDecision {
		captured = ctx
		return FileEditOK()
	})

	stdin := `{"conversation_id":"c1","workspace_roots":["/root/from/ws"],"file_path":"/x.ts","edits":[{"old_string":"a","new_string":"b"}]}`
	runHandlerWithStdin(t, "cursor-after-file-edit", stdin)

	if captured.Cwd != "/root/from/ws" {
		t.Errorf("Cwd = %q, want %q", captured.Cwd, "/root/from/ws")
	}
}

func TestOnPromptSubmit_CursorPopulatesCwd(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	var captured PromptContext
	OnPromptSubmit(func(ctx PromptContext) PromptDecision {
		captured = ctx
		return AllowPromptDecision()
	})

	stdin := `{"conversation_id":"c1","workspace_roots":["/root/from/ws"],"prompt":"hi"}`
	runHandlerWithStdin(t, "cursor-before-submit-prompt", stdin)

	if captured.Cwd != "/root/from/ws" {
		t.Errorf("Cwd = %q, want %q", captured.Cwd, "/root/from/ws")
	}
}

// =============================================================================
// Codex-specific behavior tests
// =============================================================================

func TestCodexPreToolUse_AskFailsClosed(t *testing.T) {
	// Codex currently parses but does not enforce permissionDecision "ask",
	// so the unified Codex pre-tool-use bridge must convert an Ask decision
	// into a Deny so policies still fail closed.
	ClearHandlers()
	defer ClearHandlers()

	OnBeforeExecution(func(ctx ExecutionContext) ExecutionDecision {
		return AskExecution("please confirm")
	})

	// Capture stdout to assert on the emitted JSON.
	stdinR, stdinW, _ := os.Pipe()
	stdinW.Write([]byte(`{"session_id":"s","tool_name":"Bash","tool_input":{"command":"rm -rf /"},"cwd":"/tmp"}`))
	stdinW.Close()
	stdoutR, stdoutW, _ := os.Pipe()
	origStdin, origStdout := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = stdinR, stdoutW
	defer func() {
		stdoutW.Close()
		os.Stdin, os.Stdout = origStdin, origStdout
	}()

	handlers["codex-pre-tool-use"]()
	stdoutW.Close()

	var buf bytes.Buffer
	buf.ReadFrom(stdoutR)
	got := buf.String()
	if !strings.Contains(got, `"permissionDecision":"deny"`) {
		t.Errorf("Codex Ask decision should emit deny, got %q", got)
	}
	if strings.Contains(got, `"permissionDecision":"ask"`) {
		t.Errorf("Codex Ask decision should NOT emit ask, got %q", got)
	}
}

func TestCodexPreToolUse_ApplyPatchPopulatesCommand(t *testing.T) {
	// For apply_patch, the unified handler should pull tool_input.command
	// out as ExecutionContext.Command so policies can inspect the patch.
	ClearHandlers()
	defer ClearHandlers()

	var captured ExecutionContext
	OnBeforeExecution(func(ctx ExecutionContext) ExecutionDecision {
		captured = ctx
		return AllowExecution()
	})

	stdin := `{"session_id":"s","tool_name":"apply_patch","tool_input":{"command":"*** Begin Patch\n*** Add File: foo.txt\n+hi\n*** End Patch"},"cwd":"/tmp"}`
	runHandlerWithStdin(t, "codex-pre-tool-use", stdin)

	if captured.Platform != PlatformCodex {
		t.Errorf("Platform = %q, want %q", captured.Platform, PlatformCodex)
	}
	if captured.Type != ExecutionTool {
		t.Errorf("Type = %q, want %q (apply_patch is a tool, not shell/mcp)", captured.Type, ExecutionTool)
	}
	if captured.ToolName != "apply_patch" {
		t.Errorf("ToolName = %q, want %q", captured.ToolName, "apply_patch")
	}
	if !strings.Contains(captured.Command, "*** Add File: foo.txt") {
		t.Errorf("Command should contain the patch text, got %q", captured.Command)
	}
}

func TestCodexPreToolUse_MCPClassifiedAsMCP(t *testing.T) {
	// Verifies that MCP tools reaching the codex-pre-tool-use handler are
	// classified as ExecutionMCP. (Reaching the handler at all also
	// requires the installer to include mcp__.* in the matcher; see
	// TestCodexInstaller_MatcherIncludesMCP for that side.)
	ClearHandlers()
	defer ClearHandlers()

	var captured ExecutionContext
	OnBeforeExecution(func(ctx ExecutionContext) ExecutionDecision {
		captured = ctx
		return AllowExecution()
	})

	stdin := `{"session_id":"s","tool_name":"mcp__net__fetch","tool_input":{"url":"http://example.com"},"cwd":"/tmp"}`
	runHandlerWithStdin(t, "codex-pre-tool-use", stdin)

	if !captured.IsMCP() {
		t.Errorf("Type = %q, want %q for MCP tool", captured.Type, ExecutionMCP)
	}
	if captured.ToolName != "mcp__net__fetch" {
		t.Errorf("ToolName = %q, want %q", captured.ToolName, "mcp__net__fetch")
	}
}

func TestCodexPostToolUse_ApplyPatchParsesFilesAndEdits(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	var got []FileEditContext
	OnAfterFileEdit(func(ctx FileEditContext) FileEditDecision {
		got = append(got, ctx)
		return FileEditOK()
	})

	// One add, one update, one delete in a single apply_patch call.
	patch := "*** Begin Patch\\n" +
		"*** Add File: secrets/api_key.txt\\n" +
		"+sk-deadbeef\\n" +
		"*** Update File: src/auth.go\\n" +
		"@@\\n" +
		"-old\\n" +
		"+new\\n" +
		"*** Delete File: stale.env\\n" +
		"*** End Patch"
	stdin := `{"session_id":"s","tool_name":"apply_patch","tool_input":{"command":"` + patch + `"},"cwd":"/repo"}`
	runHandlerWithStdin(t, "codex-post-tool-use", stdin)

	if len(got) != 3 {
		t.Fatalf("handler invocations = %d, want 3 (one per file in patch); got=%+v", len(got), got)
	}
	if got[0].FilePath != "secrets/api_key.txt" {
		t.Errorf("got[0].FilePath = %q, want %q", got[0].FilePath, "secrets/api_key.txt")
	}
	if len(got[0].Edits) != 1 || got[0].Edits[0].NewString != "sk-deadbeef" {
		t.Errorf("got[0].Edits = %+v, want [{OldString:\"\", NewString:\"sk-deadbeef\"}]", got[0].Edits)
	}
	if got[1].FilePath != "src/auth.go" {
		t.Errorf("got[1].FilePath = %q, want %q", got[1].FilePath, "src/auth.go")
	}
	wantEdit := FileEdit{OldString: "old", NewString: "new"}
	if len(got[1].Edits) != 1 || got[1].Edits[0] != wantEdit {
		t.Errorf("got[1].Edits = %+v, want [%+v]", got[1].Edits, wantEdit)
	}
	// Delete sections carry no per-edit content; we just verify the file
	// path made it through so path-based policies can still react.
	if got[2].FilePath != "stale.env" {
		t.Errorf("got[2].FilePath = %q, want %q", got[2].FilePath, "stale.env")
	}
	if len(got[2].Edits) != 0 {
		t.Errorf("got[2].Edits = %+v, want empty for delete", got[2].Edits)
	}
	for _, c := range got {
		if c.Platform != PlatformCodex {
			t.Errorf("Platform = %q, want %q", c.Platform, PlatformCodex)
		}
		if c.Cwd != "/repo" {
			t.Errorf("Cwd = %q, want %q", c.Cwd, "/repo")
		}
	}
}

func TestCodexPostToolUse_ApplyPatchBlockPropagates(t *testing.T) {
	// If any file in a multi-file apply_patch is blocked, the unified
	// handler must emit a PostToolBlock so Codex sees feedback rather than
	// silently passing through.
	ClearHandlers()
	defer ClearHandlers()

	OnAfterFileEdit(func(ctx FileEditContext) FileEditDecision {
		if strings.Contains(ctx.FilePath, "secrets") {
			return FileEditBlock("contains secrets")
		}
		return FileEditOK()
	})

	patch := "*** Begin Patch\\n" +
		"*** Add File: secrets/key.txt\\n" +
		"+sk-deadbeef\\n" +
		"*** Update File: src/auth.go\\n" +
		"@@\\n" +
		"-x\\n" +
		"+y\\n" +
		"*** End Patch"
	stdin := `{"session_id":"s","tool_name":"apply_patch","tool_input":{"command":"` + patch + `"},"cwd":"/repo"}`

	stdinR, stdinW, _ := os.Pipe()
	stdinW.Write([]byte(stdin))
	stdinW.Close()
	stdoutR, stdoutW, _ := os.Pipe()
	origStdin, origStdout := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = stdinR, stdoutW
	defer func() {
		stdoutW.Close()
		os.Stdin, os.Stdout = origStdin, origStdout
	}()

	handlers["codex-post-tool-use"]()
	stdoutW.Close()

	var buf bytes.Buffer
	buf.ReadFrom(stdoutR)
	out := buf.String()
	if !strings.Contains(out, `"decision":"block"`) {
		t.Errorf("expected PostToolBlock JSON output, got %q", out)
	}
	if !strings.Contains(out, "contains secrets") {
		t.Errorf("expected block reason to be present, got %q", out)
	}
}

func TestOnStop_CursorEmptyWorkspaceRootsLeavesCwdEmpty(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	var captured StopContext
	OnStop(func(ctx StopContext) StopDecision {
		captured = ctx
		return AllowStop()
	})

	stdin := `{"conversation_id":"c1","workspace_roots":[],"status":"completed","loop_count":0}`
	runHandlerWithStdin(t, "cursor-stop", stdin)

	if captured.Cwd != "" {
		t.Errorf("Cwd = %q, want empty", captured.Cwd)
	}
}
