package codex

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// PatchEdit captures one removed/added pair inside an apply_patch hunk.
//
// For an "add" operation the patch contains exactly one PatchEdit with
// OldString == "" and NewString set to the added contents. For an "update"
// operation each hunk in the patch becomes one PatchEdit with the removed
// lines joined as OldString and the added lines joined as NewString. For
// "delete" the Edits slice is empty (the FilePath alone is enough to drive
// policy).
type PatchEdit struct {
	OldString string
	NewString string
}

// PatchFile represents one file affected by an apply_patch tool call.
//
// Codex's apply_patch carries a unified-diff-style envelope under
// tool_input.command (see https://developers.openai.com/codex/hooks). A
// single envelope may touch multiple files, which is why ParseApplyPatch
// returns a slice rather than a single FilePath/Edits pair.
type PatchFile struct {
	// Operation is one of "add", "update", or "delete".
	Operation string
	// FilePath is the path declared in the *** {Add,Update,Delete} File:
	// header.
	FilePath string
	// NewFilePath is set for "update" operations that include a
	// "*** Move to:" line. Empty for in-place edits.
	NewFilePath string
	// Edits captures per-hunk old/new content. See PatchEdit for shape.
	Edits []PatchEdit
}

// ParseApplyPatch parses the unified-diff envelope from a Codex apply_patch
// tool call. The input is the raw value of tool_input.command for an
// apply_patch invocation.
//
// This is a convenience helper for handlers that want structured access to
// the patch (per-file paths, rename detection, per-hunk content). It is NOT
// invoked by the unified OnAfterFileEdit bridge — that path passes the raw
// patch text through unchanged so simple path-based policies cannot be
// silently bypassed by a parser desync against Codex's actual format. Call
// this helper only if you need the parsed view.
//
// The parser is intentionally tolerant: malformed input degrades
// gracefully (an affected file is still recorded even if its Edits list is
// incomplete) rather than returning an error.
func ParseApplyPatch(patch string) []PatchFile {
	// Strip a leading "apply_patch <<'EOF'\n" wrapper if present. Some
	// Codex invocations include a here-doc style wrapper around the
	// "*** Begin Patch" block; this is a defensive measure and not part
	// of the documented wire format.
	if i := strings.Index(patch, "*** Begin Patch"); i > 0 {
		patch = patch[i:]
	}

	lines := strings.Split(patch, "\n")
	var files []PatchFile
	var current *PatchFile
	var hunkOld, hunkNew []string
	inHunk := false

	flushHunk := func() {
		if current == nil {
			return
		}
		if len(hunkOld) == 0 && len(hunkNew) == 0 {
			hunkOld, hunkNew = nil, nil
			return
		}
		current.Edits = append(current.Edits, PatchEdit{
			OldString: strings.Join(hunkOld, "\n"),
			NewString: strings.Join(hunkNew, "\n"),
		})
		hunkOld, hunkNew = nil, nil
	}

	flushFile := func() {
		flushHunk()
		if current != nil {
			files = append(files, *current)
		}
		current = nil
		inHunk = false
	}

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "*** Begin Patch"),
			strings.HasPrefix(line, "*** End of File"):
		case strings.HasPrefix(line, "*** End Patch"):
			flushFile()
		case strings.HasPrefix(line, "*** Add File: "):
			flushFile()
			current = &PatchFile{
				Operation: "add",
				FilePath:  strings.TrimPrefix(line, "*** Add File: "),
			}
			inHunk = true
		case strings.HasPrefix(line, "*** Update File: "):
			flushFile()
			current = &PatchFile{
				Operation: "update",
				FilePath:  strings.TrimPrefix(line, "*** Update File: "),
			}
			inHunk = false
		case strings.HasPrefix(line, "*** Delete File: "):
			flushFile()
			current = &PatchFile{
				Operation: "delete",
				FilePath:  strings.TrimPrefix(line, "*** Delete File: "),
			}
			inHunk = false
		case strings.HasPrefix(line, "*** Move to: "):
			if current != nil {
				current.NewFilePath = strings.TrimPrefix(line, "*** Move to: ")
			}
		case strings.HasPrefix(line, "@@"):
			flushHunk()
			inHunk = true
		case current != nil:
			switch current.Operation {
			case "add":
				if strings.HasPrefix(line, "+") {
					hunkNew = append(hunkNew, line[1:])
				} else if line != "" {
					hunkNew = append(hunkNew, line)
				}
			case "update":
				if !inHunk {
					continue
				}
				if len(line) == 0 {
					flushHunk()
					continue
				}
				switch line[0] {
				case '+':
					hunkNew = append(hunkNew, line[1:])
				case '-':
					hunkOld = append(hunkOld, line[1:])
				case ' ':
					flushHunk()
				}
			case "delete":
			}
		}
	}
	flushFile()
	return files
}

// ParseApplyPatchFromBash inspects a Codex Bash tool command and, if it
// invokes apply_patch with a heredoc body, returns the parsed patch
// files. The second return value reports whether at least one
// apply_patch heredoc invocation was found, so callers can short-circuit
// on plain Bash commands without parsing the patch envelope.
//
// Codex routes file edits through Bash heredocs of the form
//
//	apply_patch <<'PATCH'
//	*** Begin Patch
//	... unified diff envelope ...
//	*** End Patch
//	PATCH
//
// at least as often as it emits a first-class tool_name="apply_patch" call.
// Codex may also invoke a per-session shim binary, so the apply_patch token
// can appear with an absolute path prefix, e.g.
// `/Users/me/.codex/tmp/arg0/codex-arg0IuQk4E/apply_patch <<'PATCH' ...`.
// Without this helper, callers that only inspect tool_name silently miss
// every heredoc-style edit (which is why Codex sessions showed up in the
// dashboard with zero SecurityScanResult rows even though file edits had
// clearly happened).
//
// Implementation note: detection runs the command through a real Bash
// parser (mvdan.cc/sh/v3/syntax) and walks the AST looking for a
// CallExpr whose first Word is the literal `apply_patch` (or
// `*/apply_patch` for the per-session shim) paired with a heredoc
// redirect (`<<DELIM` or `<<-DELIM`, quoted or unquoted) on the same
// Stmt. This replaced an earlier regex+`strings.Index` heuristic that
// had two bypasses:
//
//  1. A decoy `*** Begin Patch` marker in a preceding command (e.g.
//     `printf '*** Begin Patch\n' >/dev/null; apply_patch <<'PATCH' ...`)
//     would anchor `strings.Index` on the decoy, leaving the regex with a
//     prefix that didn't contain the real `apply_patch <<` invocation —
//     detection silently returned (nil, false) and the post-tool handler
//     skipped dispatch entirely.
//  2. The regex required at least one space between `apply_patch` and
//     `<<`, so the zero-space `apply_patch<<'PATCH'` (valid Bash) also
//     evaded detection.
//
// A real parser closes both bypasses and also rules out the legacy
// false-positive case (an `apply_patch` token appearing inside a
// docs-file heredoc body), because the parser identifies the
// surrounding command (`cat > docs/apply_patch_format.md <<'EOF'`) as
// the invocation, not the body content.
//
// If a command contains multiple apply_patch heredoc invocations (in
// pipelines, `;` chains, subshells, conditionals, etc.) every body is
// parsed and the resulting PatchFile slices are concatenated. If the
// command is not valid Bash (parser error) this returns (nil, false) —
// silently treating malformed input as an edit would be unsafe.
func ParseApplyPatchFromBash(command string) ([]PatchFile, bool) {
	file, err := syntax.NewParser().Parse(strings.NewReader(command), "")
	if err != nil {
		return nil, false
	}

	var bodies []heredocBody
	syntax.Walk(file, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		if !stmtInvokesApplyPatch(stmt) {
			return true
		}
		for _, r := range stmt.Redirs {
			body, ok := heredocBodyFromRedirect(command, r)
			if ok {
				bodies = append(bodies, body)
			}
		}
		return true
	})

	if len(bodies) == 0 {
		return nil, false
	}

	var all []PatchFile
	for _, b := range bodies {
		all = append(all, ParseApplyPatch(b.text)...)
	}
	return all, true
}

// stmtInvokesApplyPatch reports whether stmt's primary command is the
// apply_patch executable. Accepts the bare `apply_patch` token and the
// `*/apply_patch` absolute-path form Codex uses for its per-session
// shim binary. Any other shape (function definitions, command
// substitutions, builtins, etc.) returns false.
func stmtInvokesApplyPatch(stmt *syntax.Stmt) bool {
	call, ok := stmt.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) == 0 {
		return false
	}
	name := call.Args[0].Lit()
	if name == "apply_patch" {
		return true
	}
	if strings.HasSuffix(name, "/apply_patch") {
		return true
	}
	return false
}

// heredocBody is the extracted body of one here-doc redirect, plus a
// hint about whether tab-stripping (POSIX `<<-` semantics) was applied.
// We carry this rather than returning a raw string so test failures can
// report which variant a particular body came from.
type heredocBody struct {
	text       string
	stripIndent bool
}

// heredocBodyFromRedirect returns the body text for a here-doc redirect
// (`<<` or `<<-`), sliced from the original command using positional
// offsets so we preserve the bytes exactly as Codex emitted them. For
// `<<-` (DashHdoc) we strip leading TAB characters from each line, per
// POSIX, so the body handed to ParseApplyPatch matches what Bash would
// have piped to apply_patch's stdin.
//
// The Hdoc Word's End() points just past the trailing delimiter token
// (e.g. `PATCH`), so the returned body includes that trailing line.
// ParseApplyPatch tolerates the extra line (it stops dispatching state
// at `*** End Patch`) so we don't try to trim it here — keeping the
// slice byte-exact makes the helper easier to reason about than one
// that does ad-hoc post-processing.
func heredocBodyFromRedirect(command string, r *syntax.Redirect) (heredocBody, bool) {
	if r.Op != syntax.Hdoc && r.Op != syntax.DashHdoc {
		return heredocBody{}, false
	}
	if r.Hdoc == nil {
		return heredocBody{}, false
	}
	start := int(r.Hdoc.Pos().Offset())
	end := int(r.Hdoc.End().Offset())
	if start < 0 || end > len(command) || start >= end {
		return heredocBody{}, false
	}
	body := command[start:end]
	stripIndent := r.Op == syntax.DashHdoc
	if stripIndent {
		body = stripLeadingTabsPerLine(body)
	}
	return heredocBody{text: body, stripIndent: stripIndent}, true
}

// stripLeadingTabsPerLine removes leading TAB characters (not spaces)
// from each newline-delimited line of body. Mirrors Bash's `<<-DELIM`
// semantics so ParseApplyPatch can match line prefixes like
// `*** Begin Patch` even when Codex emits the patch under the
// strip-tabs heredoc operator.
func stripLeadingTabsPerLine(body string) string {
	if !strings.ContainsRune(body, '\t') {
		return body
	}
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimLeft(line, "\t")
	}
	return strings.Join(lines, "\n")
}
