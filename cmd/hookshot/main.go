// hookshot CLI for building and installing hooks.
//
// Usage:
//
//	hookshot build [flags]      Build hooks binary
//	hookshot install [flags]    Install hooks to config files
//
// Build:
//
//	hookshot build                    # Build for current platform
//	hookshot build -all               # Build for all platforms
//	hookshot build -output ./dist     # Custom output directory
//
// Install:
//
//	hookshot install --binary ./my-hooks
//	hookshot install --binary ./my-hooks --claude --cursor
//	hookshot install --binary ./my-hooks --codex
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		runBuild(os.Args[2:])
	case "install":
		runInstall(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`hookshot - Build hooks for AI coding agents

Usage:
  hookshot <command> [flags]

Commands:
  build     Build hooks binary for one or more platforms
  install   Install hooks to AI coding agent config files (Claude Code, Cursor, Droid, Cascade, Codex)

Run 'hookshot <command> -h' for command-specific help.`)
}

// =============================================================================
// Build Command
// =============================================================================

type Platform struct {
	OS   string
	Arch string
}

var AllPlatforms = []Platform{
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"windows", "amd64"},
	{"windows", "arm64"},
}

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)

	var (
		sourceDir  string
		outputDir  string
		binaryName string
		all        bool
		platforms  string
		ldflags    string
		verbose    bool
		clean      bool
	)

	fs.StringVar(&sourceDir, "source", ".", "Source directory to build")
	fs.StringVar(&outputDir, "output", "./dist", "Output directory for binaries")
	fs.StringVar(&binaryName, "name", "", "Binary name (default: source dir name)")
	fs.BoolVar(&all, "all", false, "Build for all platforms")
	fs.StringVar(&platforms, "platforms", "", "Comma-separated os/arch pairs")
	fs.StringVar(&ldflags, "ldflags", "-s -w", "Linker flags")
	fs.BoolVar(&verbose, "v", false, "Verbose output")
	fs.BoolVar(&clean, "clean", false, "Clean output directory before building")

	fs.Usage = func() {
		fmt.Println(`Build hooks binary for one or more platforms.

Usage:
  hookshot build [flags]

Examples:
  hookshot build                              # Current platform
  hookshot build -all                         # All platforms
  hookshot build -all -output ./dist          # Custom output
  hookshot build -platforms linux/amd64,darwin/arm64

Flags:`)
		fs.PrintDefaults()
		fmt.Println(`
Supported platforms:
  darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64, windows/arm64`)
	}

	fs.Parse(args)

	// Defaults
	if binaryName == "" {
		absPath, _ := filepath.Abs(sourceDir)
		binaryName = filepath.Base(absPath)
	}

	// Determine platforms
	var targetPlatforms []Platform
	if platforms != "" {
		for _, p := range strings.Split(platforms, ",") {
			parts := strings.Split(strings.TrimSpace(p), "/")
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Invalid platform: %s\n", p)
				os.Exit(1)
			}
			targetPlatforms = append(targetPlatforms, Platform{parts[0], parts[1]})
		}
	} else if all {
		targetPlatforms = AllPlatforms
	} else {
		targetPlatforms = []Platform{{runtime.GOOS, runtime.GOARCH}}
	}

	// Clean
	if clean {
		if verbose {
			fmt.Printf("Cleaning %s\n", outputDir)
		}
		os.RemoveAll(outputDir)
	}

	// Build
	for _, p := range targetPlatforms {
		outDir := filepath.Join(outputDir, fmt.Sprintf("%s-%s", p.OS, p.Arch))
		os.MkdirAll(outDir, 0755)

		name := binaryName
		if p.OS == "windows" {
			name += ".exe"
		}
		outPath := filepath.Join(outDir, name)

		cmdArgs := []string{"build", "-ldflags", ldflags, "-o", outPath, sourceDir}
		cmd := exec.Command("go", cmdArgs...)
		cmd.Env = append(os.Environ(), "GOOS="+p.OS, "GOARCH="+p.Arch, "CGO_ENABLED=0")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if verbose {
			fmt.Printf("GOOS=%s GOARCH=%s go %s\n", p.OS, p.Arch, strings.Join(cmdArgs, " "))
		} else {
			fmt.Printf("Building %s/%s... ", p.OS, p.Arch)
		}

		if err := cmd.Run(); err != nil {
			if !verbose {
				fmt.Println("FAILED")
			}
			os.Exit(1)
		}
		if !verbose {
			fmt.Println("OK")
		}
	}

	fmt.Printf("\nBuild complete! Output: %s\n", outputDir)
}

// =============================================================================
// Install Command
// =============================================================================

func runInstall(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)

	var (
		binaryPath  string
		claudeFlag  bool
		cursorFlag  bool
		droidFlag   bool
		cascadeFlag bool
		codexFlag   bool
	)

	fs.StringVar(&binaryPath, "binary", "", "Path to hooks binary (required)")
	fs.BoolVar(&claudeFlag, "claude", false, "Install to Claude Code only")
	fs.BoolVar(&cursorFlag, "cursor", false, "Install to Cursor only")
	fs.BoolVar(&droidFlag, "droid", false, "Install to Factory Droid only")
	fs.BoolVar(&cascadeFlag, "cascade", false, "Install to Windsurf Cascade only")
	fs.BoolVar(&codexFlag, "codex", false, "Install to OpenAI Codex only")

	fs.Usage = func() {
		fmt.Println(`Install hooks to AI coding agent config files.

Usage:
  hookshot install --binary <path> [flags]

Examples:
  hookshot install --binary ./my-hooks            # Install to all platforms
  hookshot install --binary ./my-hooks --claude   # Claude Code only
  hookshot install --binary ./my-hooks --cursor   # Cursor only
  hookshot install --binary ./my-hooks --droid    # Factory Droid only
  hookshot install --binary ./my-hooks --cascade  # Windsurf Cascade only
  hookshot install --binary ./my-hooks --codex    # OpenAI Codex only

Flags:`)
		fs.PrintDefaults()
	}

	fs.Parse(args)

	if binaryPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --binary is required")
		fs.Usage()
		os.Exit(1)
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Check binary exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: binary not found: %s\n", absPath)
		os.Exit(1)
	}

	// If none specified, install to all
	if !claudeFlag && !cursorFlag && !droidFlag && !cascadeFlag && !codexFlag {
		claudeFlag = true
		cursorFlag = true
		droidFlag = true
		cascadeFlag = true
		codexFlag = true
	}

	if claudeFlag {
		if err := installClaude(absPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing to Claude Code: %v\n", err)
			os.Exit(1)
		}
	}

	if cursorFlag {
		if err := installCursor(absPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing to Cursor: %v\n", err)
			os.Exit(1)
		}
	}

	if droidFlag {
		if err := installDroid(absPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing to Factory Droid: %v\n", err)
			os.Exit(1)
		}
	}

	if cascadeFlag {
		if err := installCascade(absPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing to Windsurf Cascade: %v\n", err)
			os.Exit(1)
		}
	}

	if codexFlag {
		if err := installCodex(absPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing to OpenAI Codex: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("\nInstallation complete!")
}

func installClaude(binaryPath string) error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".claude", "settings.json")

	fmt.Printf("Installing to Claude Code (%s)...\n", configPath)

	// Read existing config or create new
	var config map[string]any
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}
	if config == nil {
		config = make(map[string]any)
	}

	// Build hooks config
	hooks := map[string]any{
		"Stop": []map[string]any{{
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " claude-stop",
			}},
		}},
		"PreToolUse": []map[string]any{{
			"matcher": "*",
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " claude-pre-tool-use",
			}},
		}},
		"PostToolUse": []map[string]any{{
			"matcher": "Write|Edit",
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " claude-after-file-edit",
			}},
		}},
		"UserPromptSubmit": []map[string]any{{
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " claude-user-prompt-submit",
			}},
		}},
	}

	config["hooks"] = hooks

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(configPath), 0755)

	// Write config
	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, output, 0644); err != nil {
		return err
	}

	fmt.Println("  Installed hooks: Stop, PreToolUse, PostToolUse, UserPromptSubmit")
	return nil
}

func installCursor(binaryPath string) error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".cursor", "hooks.json")

	fmt.Printf("Installing to Cursor (%s)...\n", configPath)

	// Build hooks config
	config := map[string]any{
		"version": 1,
		"hooks": map[string]any{
			"stop":                 []map[string]any{{"command": binaryPath + " cursor-stop"}},
			"beforeShellExecution": []map[string]any{{"command": binaryPath + " cursor-before-shell"}},
			"beforeMCPExecution":   []map[string]any{{"command": binaryPath + " cursor-before-mcp"}},
			"afterFileEdit":        []map[string]any{{"command": binaryPath + " cursor-after-file-edit"}},
			"beforeSubmitPrompt":   []map[string]any{{"command": binaryPath + " cursor-before-submit-prompt"}},
		},
	}

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(configPath), 0755)

	// Write config
	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, output, 0644); err != nil {
		return err
	}

	fmt.Println("  Installed hooks: stop, beforeShellExecution, beforeMCPExecution, afterFileEdit, beforeSubmitPrompt")
	return nil
}

func installDroid(binaryPath string) error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".factory", "settings.json")

	fmt.Printf("Installing to Factory Droid (%s)...\n", configPath)

	// Read existing config or create new
	var config map[string]any
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}
	if config == nil {
		config = make(map[string]any)
	}

	// Build hooks config (same structure as Claude Code)
	hooks := map[string]any{
		"Stop": []map[string]any{{
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " droid-stop",
			}},
		}},
		"PreToolUse": []map[string]any{{
			"matcher": "*",
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " droid-pre-tool-use",
			}},
		}},
		"PostToolUse": []map[string]any{{
			"matcher": "Write|Edit",
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " droid-after-file-edit",
			}},
		}},
		"UserPromptSubmit": []map[string]any{{
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " droid-user-prompt-submit",
			}},
		}},
	}

	config["hooks"] = hooks

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(configPath), 0755)

	// Write config
	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, output, 0644); err != nil {
		return err
	}

	fmt.Println("  Installed hooks: Stop, PreToolUse, PostToolUse, UserPromptSubmit")
	return nil
}

func installCodex(binaryPath string) error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".codex", "hooks.json")

	fmt.Printf("Installing to OpenAI Codex (%s)...\n", configPath)

	// Read existing config or create new
	var config map[string]any
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}
	if config == nil {
		config = make(map[string]any)
	}

	// Codex hook config follows the same shape as Claude Code's settings, but
	// lives in ~/.codex/hooks.json. PostToolUse matches "apply_patch|Edit|Write"
	// because Codex uses apply_patch for file edits in addition to Write/Edit
	// aliases.
	hooks := map[string]any{
		"Stop": []map[string]any{{
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " codex-stop",
			}},
		}},
		"PreToolUse": []map[string]any{{
			"matcher": "Bash|apply_patch",
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " codex-pre-tool-use",
			}},
		}},
		"PostToolUse": []map[string]any{{
			"matcher": "apply_patch|Edit|Write",
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " codex-post-tool-use",
			}},
		}},
		"UserPromptSubmit": []map[string]any{{
			"hooks": []map[string]any{{
				"type":    "command",
				"command": binaryPath + " codex-user-prompt-submit",
			}},
		}},
	}

	config["hooks"] = hooks

	os.MkdirAll(filepath.Dir(configPath), 0755)

	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, output, 0644); err != nil {
		return err
	}

	fmt.Println("  Installed hooks: Stop, PreToolUse, PostToolUse, UserPromptSubmit")
	fmt.Println("  Note: Codex hooks require codex_hooks = true under [features] in ~/.codex/config.toml")
	return nil
}

func installCascade(binaryPath string) error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".codeium", "windsurf", "hooks.json")

	fmt.Printf("Installing to Windsurf Cascade (%s)...\n", configPath)

	// Build hooks config
	config := map[string]any{
		"hooks": map[string]any{
			"pre_run_command":         []map[string]any{{"command": binaryPath + " cascade-pre-run-command"}},
			"pre_mcp_tool_use":        []map[string]any{{"command": binaryPath + " cascade-pre-mcp-tool-use"}},
			"pre_user_prompt":         []map[string]any{{"command": binaryPath + " cascade-pre-user-prompt"}},
			"post_write_code":         []map[string]any{{"command": binaryPath + " cascade-post-write-code"}},
			"post_cascade_response":   []map[string]any{{"command": binaryPath + " cascade-post-cascade-response"}},
		},
	}

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(configPath), 0755)

	// Write config
	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, output, 0644); err != nil {
		return err
	}

	fmt.Println("  Installed hooks: pre_run_command, pre_mcp_tool_use, pre_user_prompt, post_write_code, post_cascade_response")
	return nil
}
