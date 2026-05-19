package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// normalizeJSON re-marshals JSON to normalize key order so comparisons
// are stable regardless of map iteration order.
func normalizeJSON(t *testing.T, data []byte) string {
	t.Helper()
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("normalizeJSON: %v", err)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("normalizeJSON marshal: %v", err)
	}
	return string(out)
}

func TestInstallClaude(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := installClaude("/usr/local/bin/hooks"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks claude-stop"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks claude-pre-tool-use"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks claude-after-file-edit"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks claude-user-prompt-submit"
          }
        ]
      }
    ]
  }
}`

	if normalizeJSON(t, got) != normalizeJSON(t, []byte(want)) {
		t.Errorf("Claude config mismatch.\ngot:\n%s\nwant:\n%s", normalizeJSON(t, got), normalizeJSON(t, []byte(want)))
	}
}

func TestInstallClaude_PreservesExistingKeys(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	claudeDir := filepath.Join(home, ".claude")
	os.MkdirAll(claudeDir, 0755)

	existing := `{
  "permissions": {
    "allow": ["Bash(*)"]
  },
  "mcpServers": {
    "my-server": {"command": "npx my-server"}
  }
}`
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(existing), 0644)

	if err := installClaude("/usr/local/bin/hooks"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatal(err)
	}

	if _, ok := parsed["permissions"]; !ok {
		t.Error("existing 'permissions' key was not preserved")
	}
	if _, ok := parsed["mcpServers"]; !ok {
		t.Error("existing 'mcpServers' key was not preserved")
	}
	if _, ok := parsed["hooks"]; !ok {
		t.Error("'hooks' key was not written")
	}
}

func TestInstallCursor(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := installCursor("/usr/local/bin/hooks"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(home, ".cursor", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "version": 1,
  "hooks": {
    "stop": [
      {
        "command": "/usr/local/bin/hooks cursor-stop"
      }
    ],
    "beforeShellExecution": [
      {
        "command": "/usr/local/bin/hooks cursor-before-shell"
      }
    ],
    "beforeMCPExecution": [
      {
        "command": "/usr/local/bin/hooks cursor-before-mcp"
      }
    ],
    "afterFileEdit": [
      {
        "command": "/usr/local/bin/hooks cursor-after-file-edit"
      }
    ],
    "beforeSubmitPrompt": [
      {
        "command": "/usr/local/bin/hooks cursor-before-submit-prompt"
      }
    ]
  }
}`

	if normalizeJSON(t, got) != normalizeJSON(t, []byte(want)) {
		t.Errorf("Cursor config mismatch.\ngot:\n%s\nwant:\n%s", normalizeJSON(t, got), normalizeJSON(t, []byte(want)))
	}
}

func TestInstallDroid(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := installDroid("/usr/local/bin/hooks"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(home, ".factory", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks droid-stop"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks droid-pre-tool-use"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks droid-after-file-edit"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks droid-user-prompt-submit"
          }
        ]
      }
    ]
  }
}`

	if normalizeJSON(t, got) != normalizeJSON(t, []byte(want)) {
		t.Errorf("Droid config mismatch.\ngot:\n%s\nwant:\n%s", normalizeJSON(t, got), normalizeJSON(t, []byte(want)))
	}
}

func TestInstallCodex(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := installCodex("/usr/local/bin/hooks"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(home, ".codex", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks codex-stop"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash|apply_patch|mcp__.*",
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks codex-pre-tool-use"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Bash|apply_patch|mcp__.*",
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks codex-post-tool-use"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/hooks codex-user-prompt-submit"
          }
        ]
      }
    ]
  }
}`

	if normalizeJSON(t, got) != normalizeJSON(t, []byte(want)) {
		t.Errorf("Codex config mismatch.\ngot:\n%s\nwant:\n%s", normalizeJSON(t, got), normalizeJSON(t, []byte(want)))
	}
}

func TestInstallCascade(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := installCascade("/usr/local/bin/hooks"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(home, ".codeium", "windsurf", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "hooks": {
    "pre_run_command": [
      {
        "command": "/usr/local/bin/hooks cascade-pre-run-command"
      }
    ],
    "pre_mcp_tool_use": [
      {
        "command": "/usr/local/bin/hooks cascade-pre-mcp-tool-use"
      }
    ],
    "pre_user_prompt": [
      {
        "command": "/usr/local/bin/hooks cascade-pre-user-prompt"
      }
    ],
    "post_write_code": [
      {
        "command": "/usr/local/bin/hooks cascade-post-write-code"
      }
    ],
    "post_cascade_response": [
      {
        "command": "/usr/local/bin/hooks cascade-post-cascade-response"
      }
    ]
  }
}`

	if normalizeJSON(t, got) != normalizeJSON(t, []byte(want)) {
		t.Errorf("Cascade config mismatch.\ngot:\n%s\nwant:\n%s", normalizeJSON(t, got), normalizeJSON(t, []byte(want)))
	}
}

func TestInstallCodex_PreservesExistingKeys(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	codexDir := filepath.Join(home, ".codex")
	os.MkdirAll(codexDir, 0755)

	existing := `{
  "features": {
    "hooks": true
  }
}`
	os.WriteFile(filepath.Join(codexDir, "hooks.json"), []byte(existing), 0644)

	if err := installCodex("/usr/local/bin/hooks"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(codexDir, "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatal(err)
	}

	if _, ok := parsed["features"]; !ok {
		t.Error("existing 'features' key was not preserved")
	}
	if _, ok := parsed["hooks"]; !ok {
		t.Error("'hooks' key was not written")
	}
}
