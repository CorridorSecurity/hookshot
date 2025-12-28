// Package hookshot provides a framework for building hooks for AI coding agents
// like Cursor and Claude Code.
//
// Hookshot handles the boilerplate of reading JSON from stdin and writing JSON to stdout,
// letting you focus on your hook logic. All hooks must use the multi-hook pattern with
// Register/RunCommand.
//
// Basic usage with unified handlers:
//
//	package main
//
//	import "github.com/CorridorSecurity/hookshot"
//
//	func main() {
//	    hookshot.OnStop(func(ctx hookshot.StopContext) hookshot.StopDecision {
//	        if ctx.ShouldSkip() {
//	            return hookshot.AllowStop()
//	        }
//	        return hookshot.PreventStop("Please verify changes")
//	    })
//	    hookshot.RunCommand()
//	}
//
// Platform-specific handlers:
//
//	hookshot.Register("cursor-before-tab-read", func() {
//	    hookshot.Run(func(input cursor.BeforeTabFileReadInput) cursor.BeforeTabFileReadOutput {
//	        return cursor.AllowTabRead()
//	    })
//	})
package hookshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CorridorSecurity/hookshot/internal"
)

// Run executes a hook handler function.
// It reads JSON input from stdin, passes it to the handler, and writes the output to stdout.
// If stdin cannot be read or contains invalid JSON, the program exits with code 0
// (graceful failure - hooks should not block on parse errors).
func Run[I any, O any](handler func(I) O) {
	var input I
	if err := internal.ReadJSON(&input); err != nil {
		os.Exit(0)
	}
	output := handler(input)
	if err := internal.WriteJSON(output); err != nil {
		os.Exit(0)
	}
}

// RunE is like Run but allows the handler to return an error.
// If the handler returns an error, stderr is written and the program exits with code 2
// (blocking error that will be shown to the AI).
func RunE[I any, O any](handler func(I) (O, error)) {
	var input I
	if err := internal.ReadJSON(&input); err != nil {
		os.Exit(0)
	}
	output, err := handler(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	if err := internal.WriteJSON(output); err != nil {
		os.Exit(0)
	}
}

// Handler is a function that can be registered for multi-hook binaries.
type Handler func()

var handlers = map[string]Handler{}

// Register adds a named handler for use with RunCommand.
// Use this to build a single binary that handles multiple hook types.
//
// Example:
//
//	func main() {
//	    hookshot.Register("claude-stop", handleClaudeStop)
//	    hookshot.Register("cursor-stop", handleCursorStop)
//	    hookshot.RunCommand()
//	}
func Register(name string, handler Handler) {
	handlers[name] = handler
}

// RunCommand executes the handler matching the command name.
// The command is determined from os.Args[1] if provided, otherwise from the binary name.
// If no matching handler is found, usage information is printed and the program exits with code 1.
func RunCommand() {
	cmd := getCommand()

	if handler, ok := handlers[cmd]; ok {
		handler()
		return
	}

	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
	fmt.Fprintln(os.Stderr, "Usage: <binary> <command>")
	fmt.Fprintln(os.Stderr, "Commands:")
	for name := range handlers {
		fmt.Fprintf(os.Stderr, "  %s\n", name)
	}
	os.Exit(1)
}

// getCommand returns the command name from args or binary name.
func getCommand() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return filepath.Base(os.Args[0])
}

// MustRegister is like Register but panics if a handler with the same name already exists.
func MustRegister(name string, handler Handler) {
	if _, exists := handlers[name]; exists {
		panic(fmt.Sprintf("hookshot: handler %q already registered", name))
	}
	handlers[name] = handler
}

// ClearHandlers removes all registered handlers. Useful for testing.
func ClearHandlers() {
	handlers = map[string]Handler{}
}
