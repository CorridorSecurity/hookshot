package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// install_test.go captures two clobbering bugs in the install* functions
// in main.go. Each test asserts the *correct* behavior — preserving the
// user's pre-existing config — and therefore FAILS against the current
// code. Fixing the bugs (R/M/W everywhere + per-event-key merge with a
// stable marker to filter prior hookshot entries) makes them pass.
//
// Bug 1: installCursor and installCascade build a fresh config map and
//        write it without reading existing file content, so any pre-
//        existing user entries in ~/.cursor/hooks.json or
//        ~/.codeium/windsurf/hooks.json are destroyed.
//
// Bug 2: installClaude, installDroid, and installCodex read the existing
//        config but then do `config["hooks"] = hooks`, replacing the
//        entire hooks subtree. Sibling top-level settings (theme,
//        permissions) survive, but user-managed hook entries — including
//        events hookshot does not manage at all — do not.

// setHome redirects ~/.claude, ~/.cursor, etc. to a temp directory for
// the duration of the test.
func setHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows
	return home
}

func writeJSON(t *testing.T, path string, data any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, bytes, 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return out
}

// hasClaudeStyleEntry reports whether the array of Claude-style hook
// entries at hooks[event] contains a hook command with the given command
// string. Claude-style entries are {matcher?, hooks: [{type, command}]}.
func hasClaudeStyleEntry(hooks map[string]any, event, command string) bool {
	arr, ok := hooks[event].([]any)
	if !ok {
		return false
	}
	for _, entry := range arr {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		inner, ok := entryMap["hooks"].([]any)
		if !ok {
			continue
		}
		for _, h := range inner {
			hMap, ok := h.(map[string]any)
			if !ok {
				continue
			}
			if cmd, _ := hMap["command"].(string); cmd == command {
				return true
			}
		}
	}
	return false
}

// hasCursorStyleEntry reports whether the array of Cursor/Cascade-style
// hook entries at hooks[event] contains an entry with the given command
// string. Cursor-style entries are flat {command}.
func hasCursorStyleEntry(hooks map[string]any, event, command string) bool {
	arr, ok := hooks[event].([]any)
	if !ok {
		return false
	}
	for _, entry := range arr {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if cmd, _ := entryMap["command"].(string); cmd == command {
			return true
		}
	}
	return false
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestInstallCodex_PreservesUserConfig demonstrates that installCodex
// clobbers user-managed entries in ~/.codex/hooks.json. The outer R/M/W
// preserves unrelated top-level keys, but the entire `hooks` subtree is
// replaced with hookshot's four events.
func TestInstallCodex_PreservesUserConfig(t *testing.T) {
	home := setHome(t)
	configPath := filepath.Join(home, ".codex", "hooks.json")

	// Pre-existing user config:
	//   - unrelated top-level key (should survive)
	//   - Notification hook (event hookshot does not manage)
	//   - additional PreToolUse hook (same event hookshot manages)
	existing := map[string]any{
		"ui": map[string]any{"theme": "dark"},
		"hooks": map[string]any{
			"Notification": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"type": "command", "command": "/usr/local/bin/user-notify"},
					},
				},
			},
			"PreToolUse": []any{
				map[string]any{
					"matcher": "Bash",
					"hooks": []any{
						map[string]any{"type": "command", "command": "/usr/local/bin/user-pre-bash"},
					},
				},
			},
		},
	}
	writeJSON(t, configPath, existing)

	if err := installCodex("/usr/local/bin/hookshot-test"); err != nil {
		t.Fatalf("installCodex: %v", err)
	}

	got := readJSON(t, configPath)

	// Sanity: the outer R/M/W preserves unrelated siblings.
	if _, ok := got["ui"]; !ok {
		t.Errorf("unrelated top-level field 'ui' was lost")
	}

	hooks, ok := got["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("hooks subtree is missing or wrong type: %T", got["hooks"])
	}

	// Bug: the user's Notification event is gone because installCodex
	// replaces the entire hooks subtree with hookshot's four keys.
	if _, ok := hooks["Notification"]; !ok {
		t.Errorf("user's Notification hook was clobbered; remaining events: %v", keysOf(hooks))
	}

	// Bug: even for events hookshot manages, the user's own entry is
	// destroyed because installCodex assigns the whole PreToolUse array
	// rather than appending to it.
	if !hasClaudeStyleEntry(hooks, "PreToolUse", "/usr/local/bin/user-pre-bash") {
		t.Errorf("user's PreToolUse hook (/usr/local/bin/user-pre-bash) was clobbered")
	}

	// Sanity: hookshot's own entry was installed.
	if !hasClaudeStyleEntry(hooks, "PreToolUse", "/usr/local/bin/hookshot-test codex-pre-tool-use") {
		t.Errorf("hookshot's PreToolUse hook was not installed")
	}
}

// TestInstallCursor_PreservesUserConfig demonstrates that installCursor
// destroys any pre-existing ~/.cursor/hooks.json because it never reads
// the file before writing — it builds a fresh config literal and
// overwrites.
func TestInstallCursor_PreservesUserConfig(t *testing.T) {
	home := setHome(t)
	configPath := filepath.Join(home, ".cursor", "hooks.json")

	// Pre-existing user config:
	//   - beforeShellExecution hook (event hookshot also manages)
	//   - beforeReadFile hook (event hookshot does not manage)
	existing := map[string]any{
		"version": 1,
		"hooks": map[string]any{
			"beforeShellExecution": []any{
				map[string]any{"command": "/usr/local/bin/user-shell-guard"},
			},
			"beforeReadFile": []any{
				map[string]any{"command": "/usr/local/bin/user-read-guard"},
			},
		},
	}
	writeJSON(t, configPath, existing)

	if err := installCursor("/usr/local/bin/hookshot-test"); err != nil {
		t.Fatalf("installCursor: %v", err)
	}

	got := readJSON(t, configPath)

	hooks, ok := got["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("hooks subtree is missing or wrong type: %T", got["hooks"])
	}

	// Bug: installCursor builds a fresh config and writes it without
	// reading existing content, so unrelated events vanish entirely.
	if _, ok := hooks["beforeReadFile"]; !ok {
		t.Errorf("user's beforeReadFile hook was clobbered; remaining events: %v", keysOf(hooks))
	}

	// Bug: the user's beforeShellExecution entry is destroyed even though
	// hookshot also writes an entry for that event.
	if !hasCursorStyleEntry(hooks, "beforeShellExecution", "/usr/local/bin/user-shell-guard") {
		t.Errorf("user's beforeShellExecution hook (/usr/local/bin/user-shell-guard) was clobbered")
	}

	// Sanity: hookshot's own entry was installed.
	if !hasCursorStyleEntry(hooks, "beforeShellExecution", "/usr/local/bin/hookshot-test cursor-before-shell") {
		t.Errorf("hookshot's beforeShellExecution hook was not installed")
	}
}

// TestInstallClaude_PreservesUserConfig demonstrates that installClaude
// has the same clobbering shape as installCodex. The outer R/M/W
// preserves sibling settings, but the entire `hooks` subtree is replaced.
func TestInstallClaude_PreservesUserConfig(t *testing.T) {
	home := setHome(t)
	configPath := filepath.Join(home, ".claude", "settings.json")

	existing := map[string]any{
		"theme": "dark",
		"permissions": map[string]any{
			"allow": []any{"Bash(*)"},
		},
		"hooks": map[string]any{
			"Notification": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"type": "command", "command": "/usr/local/bin/user-notify"},
					},
				},
			},
		},
	}
	writeJSON(t, configPath, existing)

	if err := installClaude("/usr/local/bin/hookshot-test"); err != nil {
		t.Fatalf("installClaude: %v", err)
	}

	got := readJSON(t, configPath)

	// Sanity: sibling settings survive the R/M/W.
	if got["theme"] != "dark" {
		t.Errorf("user's theme setting was lost: got %v", got["theme"])
	}
	if _, ok := got["permissions"]; !ok {
		t.Errorf("user's permissions setting was lost")
	}

	hooks, ok := got["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("hooks subtree is missing or wrong type: %T", got["hooks"])
	}

	// Bug: user's Notification hook is gone because installClaude does
	// `config["hooks"] = hooks`, replacing the whole subtree.
	if _, ok := hooks["Notification"]; !ok {
		t.Errorf("user's Notification hook was clobbered; remaining events: %v", keysOf(hooks))
	}
}
