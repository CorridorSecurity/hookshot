// Package cascade provides types and helpers for Windsurf Cascade hooks.
//
// Windsurf Cascade hooks are configured in hooks.json files and execute at various
// points in the agent loop. See Windsurf documentation for details.
package cascade

import "encoding/json"

// =============================================================================
// Common Types
// =============================================================================

// BaseInput contains fields present in all Cascade hook inputs.
type BaseInput struct {
	AgentActionName string `json:"agent_action_name"` // The hook event name
	TrajectoryID    string `json:"trajectory_id"`     // Session identifier
	ExecutionID     string `json:"execution_id"`      // Unique execution identifier
	Timestamp       string `json:"timestamp"`         // ISO 8601 timestamp
}

// BaseOutput contains common fields for hook outputs that support decisions.
type BaseOutput struct {
	// Decision controls whether the action should proceed.
	// Values: "allow", "deny", "ask" (where supported)
	Decision string `json:"decision,omitempty"`

	// Message is shown to the user or agent when Decision is "deny".
	Message string `json:"message,omitempty"`
}

// =============================================================================
// PreRunCommand
// =============================================================================

// PreRunCommandToolInfo contains details about the command to be executed.
type PreRunCommandToolInfo struct {
	CommandLine string `json:"command_line"`
	Cwd         string `json:"cwd"`
}

// PreRunCommandInput is received before a shell command executes.
type PreRunCommandInput struct {
	BaseInput
	ToolInfo PreRunCommandToolInfo `json:"tool_info"`
}

// PreRunCommandOutput controls whether the command should execute.
type PreRunCommandOutput struct {
	BaseOutput
}

// =============================================================================
// PostRunCommand
// =============================================================================

// PostRunCommandToolInfo contains details about the executed command.
type PostRunCommandToolInfo struct {
	CommandLine string `json:"command_line"`
	Cwd         string `json:"cwd"`
	ExitCode    int    `json:"exit_code"`
	Output      string `json:"output,omitempty"`
}

// PostRunCommandInput is received after a shell command executes.
type PostRunCommandInput struct {
	BaseInput
	ToolInfo PostRunCommandToolInfo `json:"tool_info"`
}

// PostRunCommandOutput has no decision control - hooks just run for side effects.
type PostRunCommandOutput struct{}

// =============================================================================
// PreWriteCode
// =============================================================================

// CascadeEdit represents a single edit operation in Cascade write hooks.
type CascadeEdit struct {
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

// PreWriteCodeToolInfo contains details about the file write operation.
type PreWriteCodeToolInfo struct {
	FilePath string        `json:"file_path"`
	Edits    []CascadeEdit `json:"edits"`
}

// PreWriteCodeInput is received before writing a code file.
type PreWriteCodeInput struct {
	BaseInput
	ToolInfo PreWriteCodeToolInfo `json:"tool_info"`
}

// PreWriteCodeOutput controls whether the write should proceed.
type PreWriteCodeOutput struct {
	BaseOutput
}

// =============================================================================
// PostWriteCode
// =============================================================================

// PostWriteCodeToolInfo contains details about the written file.
type PostWriteCodeToolInfo struct {
	FilePath string        `json:"file_path"`
	Edits    []CascadeEdit `json:"edits"`
}

// PostWriteCodeInput is received after writing a code file.
type PostWriteCodeInput struct {
	BaseInput
	ToolInfo PostWriteCodeToolInfo `json:"tool_info"`
}

// PostWriteCodeOutput has no decision control - hooks just run for side effects.
type PostWriteCodeOutput struct{}

// =============================================================================
// PreReadCode
// =============================================================================

// PreReadCodeToolInfo contains details about the file read operation.
type PreReadCodeToolInfo struct {
	FilePath string `json:"file_path"`
}

// PreReadCodeInput is received before reading a code file.
type PreReadCodeInput struct {
	BaseInput
	ToolInfo PreReadCodeToolInfo `json:"tool_info"`
}

// PreReadCodeOutput controls whether the read should proceed.
type PreReadCodeOutput struct {
	BaseOutput
}

// =============================================================================
// PostReadCode
// =============================================================================

// PostReadCodeToolInfo contains details about the read file.
type PostReadCodeToolInfo struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content,omitempty"`
}

// PostReadCodeInput is received after reading a code file.
type PostReadCodeInput struct {
	BaseInput
	ToolInfo PostReadCodeToolInfo `json:"tool_info"`
}

// PostReadCodeOutput has no decision control - hooks just run for side effects.
type PostReadCodeOutput struct{}

// =============================================================================
// PreMCPToolUse
// =============================================================================

// PreMCPToolUseToolInfo contains details about the MCP tool call.
type PreMCPToolUseToolInfo struct {
	MCPServerName    string          `json:"mcp_server_name"`
	MCPToolName      string          `json:"mcp_tool_name"`
	MCPToolArguments json.RawMessage `json:"mcp_tool_arguments"`
}

// PreMCPToolUseInput is received before an MCP tool executes.
type PreMCPToolUseInput struct {
	BaseInput
	ToolInfo PreMCPToolUseToolInfo `json:"tool_info"`
}

// PreMCPToolUseOutput controls whether the MCP tool should execute.
type PreMCPToolUseOutput struct {
	BaseOutput
}

// =============================================================================
// PostMCPToolUse
// =============================================================================

// PostMCPToolUseToolInfo contains details about the executed MCP tool.
type PostMCPToolUseToolInfo struct {
	MCPServerName    string          `json:"mcp_server_name"`
	MCPToolName      string          `json:"mcp_tool_name"`
	MCPToolArguments json.RawMessage `json:"mcp_tool_arguments"`
	MCPResult        string          `json:"mcp_result,omitempty"`
}

// PostMCPToolUseInput is received after an MCP tool executes.
type PostMCPToolUseInput struct {
	BaseInput
	ToolInfo PostMCPToolUseToolInfo `json:"tool_info"`
}

// PostMCPToolUseOutput has no decision control - hooks just run for side effects.
type PostMCPToolUseOutput struct{}

// =============================================================================
// PreUserPrompt
// =============================================================================

// PreUserPromptToolInfo contains details about the user's prompt.
type PreUserPromptToolInfo struct {
	UserPrompt string `json:"user_prompt"`
}

// PreUserPromptInput is received when the user submits a prompt.
type PreUserPromptInput struct {
	BaseInput
	ToolInfo PreUserPromptToolInfo `json:"tool_info"`
}

// PreUserPromptOutput controls whether the prompt should be processed.
type PreUserPromptOutput struct {
	BaseOutput
}

// =============================================================================
// PostCascadeResponse
// =============================================================================

// PostCascadeResponseToolInfo contains details about Cascade's response.
type PostCascadeResponseToolInfo struct {
	Response string `json:"response"`
}

// PostCascadeResponseInput is received after Cascade completes a response.
type PostCascadeResponseInput struct {
	BaseInput
	ToolInfo PostCascadeResponseToolInfo `json:"tool_info"`
}

// PostCascadeResponseOutput has no decision control - hooks just run for side effects.
type PostCascadeResponseOutput struct{}

// =============================================================================
// PostSetupWorktree
// =============================================================================

// PostSetupWorktreeToolInfo contains details about the worktree setup.
type PostSetupWorktreeToolInfo struct {
	WorktreePath string `json:"worktree_path"`
}

// PostSetupWorktreeInput is received after worktree setup completes.
type PostSetupWorktreeInput struct {
	BaseInput
	ToolInfo PostSetupWorktreeToolInfo `json:"tool_info"`
}

// PostSetupWorktreeOutput has no decision control - hooks just run for side effects.
type PostSetupWorktreeOutput struct{}
