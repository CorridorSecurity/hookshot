package claude

// =============================================================================
// Stop / SubagentStop Helpers
// =============================================================================

// Continue allows Claude to stop normally.
func Continue() StopOutput {
	return StopOutput{}
}

// Block prevents Claude from stopping and sends a message.
func Block(reason string) StopOutput {
	return StopOutput{
		Decision: "block",
		Reason:   reason,
	}
}

// StopWith creates a StopOutput that halts Claude entirely.
func StopWith(reason string) StopOutput {
	cont := false
	return StopOutput{
		BaseOutput: BaseOutput{
			Continue:   &cont,
			StopReason: reason,
		},
	}
}

// =============================================================================
// PreToolUse Helpers
// =============================================================================

// Allow permits the tool to execute without asking the user.
func Allow(reason string) PreToolUseOutput {
	return PreToolUseOutput{
		HookSpecificOutput: &PreToolUseHookOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "allow",
			PermissionDecisionReason: reason,
		},
	}
}

// AllowSilent permits the tool to execute without showing output.
func AllowSilent() PreToolUseOutput {
	return PreToolUseOutput{
		BaseOutput: BaseOutput{
			SuppressOutput: true,
		},
		HookSpecificOutput: &PreToolUseHookOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
		},
	}
}

// AllowWithInput permits the tool with modified input parameters.
func AllowWithInput(reason string, updatedInput map[string]any) PreToolUseOutput {
	return PreToolUseOutput{
		HookSpecificOutput: &PreToolUseHookOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "allow",
			PermissionDecisionReason: reason,
			UpdatedInput:             updatedInput,
		},
	}
}

// Deny blocks the tool from executing.
func Deny(reason string) PreToolUseOutput {
	return PreToolUseOutput{
		HookSpecificOutput: &PreToolUseHookOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: reason,
		},
	}
}

// Ask prompts the user to confirm the tool execution.
func Ask(reason string) PreToolUseOutput {
	return PreToolUseOutput{
		HookSpecificOutput: &PreToolUseHookOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "ask",
			PermissionDecisionReason: reason,
		},
	}
}

// PassThrough returns an empty output, letting the normal permission flow proceed.
func PassThrough() PreToolUseOutput {
	return PreToolUseOutput{}
}

// =============================================================================
// PermissionRequest Helpers
// =============================================================================

// AllowPermission grants the permission request.
func AllowPermission() PermissionRequestOutput {
	return PermissionRequestOutput{
		HookSpecificOutput: &PermissionRequestHookOutput{
			HookEventName: "PermissionRequest",
			Decision: &PermissionRequestDecision{
				Behavior: "allow",
			},
		},
	}
}

// AllowPermissionWithInput grants the permission with modified tool input.
func AllowPermissionWithInput(updatedInput map[string]any) PermissionRequestOutput {
	return PermissionRequestOutput{
		HookSpecificOutput: &PermissionRequestHookOutput{
			HookEventName: "PermissionRequest",
			Decision: &PermissionRequestDecision{
				Behavior:     "allow",
				UpdatedInput: updatedInput,
			},
		},
	}
}

// DenyPermission rejects the permission request.
func DenyPermission(message string) PermissionRequestOutput {
	return PermissionRequestOutput{
		HookSpecificOutput: &PermissionRequestHookOutput{
			HookEventName: "PermissionRequest",
			Decision: &PermissionRequestDecision{
				Behavior: "deny",
				Message:  message,
			},
		},
	}
}

// DenyPermissionAndStop rejects the permission and stops Claude.
func DenyPermissionAndStop(message string) PermissionRequestOutput {
	return PermissionRequestOutput{
		HookSpecificOutput: &PermissionRequestHookOutput{
			HookEventName: "PermissionRequest",
			Decision: &PermissionRequestDecision{
				Behavior:  "deny",
				Message:   message,
				Interrupt: true,
			},
		},
	}
}

// =============================================================================
// PostToolUse Helpers
// =============================================================================

// PostToolOK returns an empty output, allowing normal flow to continue.
func PostToolOK() PostToolUseOutput {
	return PostToolUseOutput{}
}

// PostToolBlock provides feedback to Claude that something needs attention.
func PostToolBlock(reason string) PostToolUseOutput {
	return PostToolUseOutput{
		Decision: "block",
		Reason:   reason,
	}
}

// PostToolContext adds context for Claude to consider.
func PostToolContext(context string) PostToolUseOutput {
	return PostToolUseOutput{
		HookSpecificOutput: &PostToolUseHookOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: context,
		},
	}
}

// =============================================================================
// UserPromptSubmit Helpers
// =============================================================================

// AllowPrompt allows the prompt to be processed normally.
func AllowPrompt() UserPromptSubmitOutput {
	return UserPromptSubmitOutput{}
}

// BlockPrompt prevents the prompt from being processed.
func BlockPrompt(reason string) UserPromptSubmitOutput {
	return UserPromptSubmitOutput{
		Decision: "block",
		Reason:   reason,
	}
}

// AddContext allows the prompt and adds context for Claude.
func AddContext(context string) UserPromptSubmitOutput {
	return UserPromptSubmitOutput{
		HookSpecificOutput: &UserPromptSubmitHookOutput{
			HookEventName:     "UserPromptSubmit",
			AdditionalContext: context,
		},
	}
}

// =============================================================================
// SessionStart Helpers
// =============================================================================

// SessionStartOK returns an empty output for session start.
func SessionStartOK() SessionStartOutput {
	return SessionStartOutput{}
}

// SessionStartContext adds context at the start of a session.
func SessionStartContext(context string) SessionStartOutput {
	return SessionStartOutput{
		HookSpecificOutput: &SessionStartHookOutput{
			HookEventName:     "SessionStart",
			AdditionalContext: context,
		},
	}
}

// =============================================================================
// SessionEnd Helpers
// =============================================================================

// SessionEndOK returns an empty output for session end.
func SessionEndOK() SessionEndOutput {
	return SessionEndOutput{}
}

// =============================================================================
// Notification Helpers
// =============================================================================

// NotificationOK returns an empty output for notifications.
func NotificationOK() NotificationOutput {
	return NotificationOutput{}
}

// =============================================================================
// PreCompact Helpers
// =============================================================================

// PreCompactOK returns an empty output for pre-compact.
func PreCompactOK() PreCompactOutput {
	return PreCompactOutput{}
}
