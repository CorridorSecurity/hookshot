package hookshot

import (
	"encoding/json"
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
