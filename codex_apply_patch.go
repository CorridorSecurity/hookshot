package hookshot

import (
	"strings"
)

// codexApplyPatchFile represents one file affected by a Codex apply_patch
// tool call. Codex's apply_patch carries a "command" field containing a
// unified-diff-style envelope (see https://developers.openai.com/codex/hooks).
// Each envelope may reference multiple files, which is why we model this as
// a slice on the parser output rather than a single FilePath/Edits pair.
type codexApplyPatchFile struct {
	// Operation is one of "add", "update", or "delete".
	Operation string
	// FilePath is the path declared in the *** {Add,Update,Delete} File: header.
	FilePath string
	// NewFilePath is set for "update" operations that include a *** Move to: line.
	NewFilePath string
	// Edits captures the per-hunk old/new content. For "add" the patch
	// contains exactly one edit with OldString=="" and NewString set to the
	// added contents. For "delete" Edits is empty (the file path is enough
	// for policy evaluation). For "update", each hunk in the patch becomes
	// one FileEdit with the removed lines joined as OldString and the added
	// lines joined as NewString.
	Edits []FileEdit
}

// parseCodexApplyPatch parses the patch envelope from a Codex apply_patch
// tool call. The input is the raw value of tool_input.command for an
// apply_patch invocation. The parser is intentionally tolerant: malformed
// input degrades gracefully (the affected file is still recorded, even if
// its Edits list is incomplete) rather than returning an error, because
// hook handlers should never silently disappear on bad input.
func parseCodexApplyPatch(patch string) []codexApplyPatchFile {
	// Strip a leading "apply_patch <<'EOF'\n" wrapper if present. Some Codex
	// builds include a here-doc style wrapper around the *** Begin Patch
	// block; this is a defensive measure and not part of the documented
	// wire format.
	if i := strings.Index(patch, "*** Begin Patch"); i > 0 {
		patch = patch[i:]
	}

	lines := strings.Split(patch, "\n")
	var files []codexApplyPatchFile
	var current *codexApplyPatchFile
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
		current.Edits = append(current.Edits, FileEdit{
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
			// Markers we don't act on; "End of File" is sometimes emitted
			// inside an Update section.
		case strings.HasPrefix(line, "*** End Patch"):
			flushFile()
		case strings.HasPrefix(line, "*** Add File: "):
			flushFile()
			current = &codexApplyPatchFile{
				Operation: "add",
				FilePath:  strings.TrimPrefix(line, "*** Add File: "),
			}
			inHunk = true
		case strings.HasPrefix(line, "*** Update File: "):
			flushFile()
			current = &codexApplyPatchFile{
				Operation: "update",
				FilePath:  strings.TrimPrefix(line, "*** Update File: "),
			}
			inHunk = false
		case strings.HasPrefix(line, "*** Delete File: "):
			flushFile()
			current = &codexApplyPatchFile{
				Operation: "delete",
				FilePath:  strings.TrimPrefix(line, "*** Delete File: "),
			}
			inHunk = false
		case strings.HasPrefix(line, "*** Move to: "):
			if current != nil {
				current.NewFilePath = strings.TrimPrefix(line, "*** Move to: ")
			}
		case strings.HasPrefix(line, "@@"):
			// Hunk separator inside an Update section. Flush the previous
			// hunk (if any) and start collecting a fresh one.
			flushHunk()
			inHunk = true
		case current != nil:
			switch current.Operation {
			case "add":
				// In an Add section every content line is prefixed with "+".
				// Be lenient: accept unprefixed lines as well so we still
				// capture the new file contents when Codex omits the
				// prefix.
				if strings.HasPrefix(line, "+") {
					hunkNew = append(hunkNew, line[1:])
				} else if line != "" {
					hunkNew = append(hunkNew, line)
				}
			case "update":
				if !inHunk {
					// Anything before the first @@ in an Update section is
					// metadata we don't care about.
					continue
				}
				if len(line) == 0 {
					// Treat blank lines as context inside an Update.
					flushHunk()
					continue
				}
				switch line[0] {
				case '+':
					hunkNew = append(hunkNew, line[1:])
				case '-':
					hunkOld = append(hunkOld, line[1:])
				case ' ':
					// Context line — flush so adjacent ±/context groups
					// don't get fused into a single edit.
					flushHunk()
				}
			case "delete":
				// Delete sections have no content lines; ignore stray text.
			}
		}
	}
	flushFile()
	return files
}
