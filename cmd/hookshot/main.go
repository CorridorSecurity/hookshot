// hookshot CLI for building hooks.
//
// Usage:
//
//	hookshot build [flags]      Build hooks binary
//
// Build:
//
//	hookshot build                    # Build for current platform
//	hookshot build -all               # Build for all platforms
//	hookshot build -output ./dist     # Custom output directory
package main

import (
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
