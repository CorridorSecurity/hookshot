// Package droid provides types and helpers for Factory Droid hooks.
//
// Factory Droid hooks use the same API as Claude Code hooks, so this package
// re-exports all types and helpers from the claude package for convenience.
//
// # Configuration
//
// Configure hooks in ~/.factory/settings.json:
//
//	{
//	  "hooks": {
//	    "PreToolUse": [
//	      {
//	        "matcher": "*",
//	        "hooks": [{ "type": "command", "command": "/path/to/hook droid-pre-tool-use" }]
//	      }
//	    ]
//	  }
//	}
//
// # Example Usage
//
//	package main
//
//	import (
//	    "github.com/CorridorSecurity/hookshot"
//	    "github.com/CorridorSecurity/hookshot/droid"
//	)
//
//	func main() {
//	    hookshot.Register("droid-pre-tool-use", func() {
//	        hookshot.Run(func(input droid.PreToolUseInput) droid.PreToolUseOutput {
//	            if input.ToolName == "Read" {
//	                return droid.AllowSilent()
//	            }
//	            return droid.PassThrough()
//	        })
//	    })
//	    hookshot.RunCommand()
//	}
package droid
