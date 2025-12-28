// Build tool for cross-compiling hookshot binaries.
//
// Usage:
//
//	go run ./build                           # Build for current platform
//	go run ./build -all                      # Build for all platforms
//	go run ./build -output ./dist            # Custom output directory
//	go run ./build -name my-hooks            # Custom binary name
//	go run ./build -all -output ./dist -name my-hooks
//
// This tool can also be imported as a library:
//
//	import "github.com/CorridorSecurity/hookshot/build"
//
//	func main() {
//	    build.Build(build.Options{
//	        OutputDir:  "./dist",
//	        BinaryName: "my-hooks",
//	        All:        true,
//	    })
//	}
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

// Platform represents a target OS/architecture combination.
type Platform struct {
	OS   string
	Arch string
}

// AllPlatforms lists all supported build targets.
var AllPlatforms = []Platform{
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"windows", "amd64"},
	{"windows", "arm64"},
}

// Options configures the build process.
type Options struct {
	// SourceDir is the directory containing the Go source to build.
	// Defaults to current directory.
	SourceDir string

	// OutputDir is where binaries will be written.
	// Defaults to "./dist".
	OutputDir string

	// BinaryName is the name of the output binary (without extension).
	// Defaults to the source directory name.
	BinaryName string

	// All builds for all platforms. If false, builds for current platform only.
	All bool

	// Platforms to build for. If empty and All is true, uses AllPlatforms.
	Platforms []Platform

	// LDFlags are additional linker flags. Defaults to "-s -w" (strip symbols).
	LDFlags string

	// Verbose prints build commands.
	Verbose bool

	// Clean removes output directory before building.
	Clean bool
}

// Build compiles the Go binary for the specified platforms.
func Build(opts Options) error {
	// Apply defaults
	if opts.SourceDir == "" {
		opts.SourceDir = "."
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "./dist"
	}
	if opts.LDFlags == "" {
		opts.LDFlags = "-s -w"
	}
	if opts.BinaryName == "" {
		absPath, _ := filepath.Abs(opts.SourceDir)
		opts.BinaryName = filepath.Base(absPath)
	}

	// Determine platforms
	platforms := opts.Platforms
	if len(platforms) == 0 {
		if opts.All {
			platforms = AllPlatforms
		} else {
			platforms = []Platform{{runtime.GOOS, runtime.GOARCH}}
		}
	}

	// Clean if requested
	if opts.Clean {
		if opts.Verbose {
			fmt.Printf("Cleaning %s\n", opts.OutputDir)
		}
		os.RemoveAll(opts.OutputDir)
	}

	// Build for each platform
	for _, p := range platforms {
		if err := buildPlatform(opts, p); err != nil {
			return fmt.Errorf("build %s/%s failed: %w", p.OS, p.Arch, err)
		}
	}

	fmt.Printf("\nBuild complete! Output: %s\n", opts.OutputDir)
	return nil
}

func buildPlatform(opts Options, p Platform) error {
	// Create output directory
	outDir := filepath.Join(opts.OutputDir, fmt.Sprintf("%s-%s", p.OS, p.Arch))
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Determine binary name
	binaryName := opts.BinaryName
	if p.OS == "windows" {
		binaryName += ".exe"
	}
	outPath := filepath.Join(outDir, binaryName)

	// Build command
	args := []string{"build"}
	if opts.LDFlags != "" {
		args = append(args, "-ldflags", opts.LDFlags)
	}
	args = append(args, "-o", outPath, opts.SourceDir)

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(),
		"GOOS="+p.OS,
		"GOARCH="+p.Arch,
		"CGO_ENABLED=0",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if opts.Verbose {
		fmt.Printf("GOOS=%s GOARCH=%s go %s\n", p.OS, p.Arch, strings.Join(args, " "))
	} else {
		fmt.Printf("Building %s/%s... ", p.OS, p.Arch)
	}

	if err := cmd.Run(); err != nil {
		if !opts.Verbose {
			fmt.Println("FAILED")
		}
		return err
	}

	if !opts.Verbose {
		fmt.Println("OK")
	}
	return nil
}

// CurrentPlatform returns the current OS/architecture.
func CurrentPlatform() Platform {
	return Platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
}

func main() {
	var opts Options
	var platforms string

	flag.StringVar(&opts.SourceDir, "source", ".", "Source directory to build")
	flag.StringVar(&opts.OutputDir, "output", "./dist", "Output directory for binaries")
	flag.StringVar(&opts.BinaryName, "name", "", "Binary name (default: source dir name)")
	flag.BoolVar(&opts.All, "all", false, "Build for all platforms")
	flag.StringVar(&platforms, "platforms", "", "Comma-separated list of os/arch pairs (e.g., linux/amd64,darwin/arm64)")
	flag.StringVar(&opts.LDFlags, "ldflags", "-s -w", "Linker flags")
	flag.BoolVar(&opts.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&opts.Clean, "clean", false, "Clean output directory before building")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Hookshot Build Tool

Cross-compile Go binaries for Cursor and Claude Code hooks.

Usage:
  go run ./build [flags]

Examples:
  go run ./build                    # Build for current platform
  go run ./build -all               # Build for all platforms
  go run ./build -all -clean        # Clean and build all platforms
  go run ./build -output ./dist     # Custom output directory
  go run ./build -name my-hooks     # Custom binary name
  go run ./build -platforms linux/amd64,darwin/arm64

Flags:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Supported platforms:
  darwin/amd64   - macOS Intel
  darwin/arm64   - macOS Apple Silicon
  linux/amd64    - Linux x86_64
  linux/arm64    - Linux ARM64
  windows/amd64  - Windows x86_64
  windows/arm64  - Windows ARM64
`)
	}

	flag.Parse()

	// Parse custom platforms
	if platforms != "" {
		opts.Platforms = nil
		for _, p := range strings.Split(platforms, ",") {
			parts := strings.Split(strings.TrimSpace(p), "/")
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Invalid platform format: %s (expected os/arch)\n", p)
				os.Exit(1)
			}
			opts.Platforms = append(opts.Platforms, Platform{OS: parts[0], Arch: parts[1]})
		}
	}

	if err := Build(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
