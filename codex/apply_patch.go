package codex

import "strings"

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
