package codex

import (
	"regexp"
	"strings"
)

// catRedirectHeredocRE matches the leading `cat <<['"]?DELIM['"]? > FILE`
// (or `>>`) shape that Codex uses for greenfield file writes. The regex
// must hold together because Codex emits this command without any
// surrounding context — it's the entire `tool_input.command` payload.
//
// Group layout (named for legibility in the implementation):
//
//	op    — the heredoc operator, "<<" or "<<-".
//	delim — the heredoc delimiter ("EOF", "PATCH", …), with surrounding
//	        quotes (if any) stripped by the caller.
//	redir — the redirection operator, ">" or ">>".
//	path  — the target file path. Anything up to the first whitespace,
//	        `&`, `;`, `|`, or end-of-line.
//
// Why this exact shape: the captured probe payloads (see
// vscode-extension/docs/HOOKS.md § "Probing Codex hooks") show Codex
// invoking writes as `cat <<'EOF' > greet.txt` for the greet.txt case
// and `cd … && cat <<'EOF' > path` for cwd-prefixed variants. Heredoc
// quoting can be single, double, or absent; the redirect operator can
// be `>` (overwrite) or `>>` (append). The optional `-` after `<<`
// covers the `<<-DELIM` shape that strips leading tabs from the body —
// uncommon in Codex output but cheap to support.
// `\s` would also greedy-match the trailing newline + any leading blank
// lines of the body. Use `[ \t]*` for header-line whitespace so the
// header match always ends at the EOL boundary and extractHeredocBody
// can preserve leading blank lines in the body exactly as written.
var catRedirectHeredocRE = regexp.MustCompile(
	`(?m)(?:^|[;&|])[ \t]*cat[ \t]+(?P<op><<-?)[ \t]*(?P<delim>'[^']+'|"[^"]+"|[A-Za-z_][A-Za-z0-9_]*)[ \t]*(?P<redir>>>?)[ \t]*(?P<path>[^\s;&|]+)[ \t]*$`,
)

// teeRedirectHeredocRE covers `tee [-a] FILE [< /dev/null]` and similar
// shapes where the body is fed from a heredoc opened earlier on the line.
// Less common than cat but inexpensive to handle, and the variant Codex
// occasionally emits when it wants tee's "also print to stdout" side
// effect. Captures match catRedirectHeredocRE's named groups so the
// caller can use a shared extractor.
var teeRedirectHeredocRE = regexp.MustCompile(
	`(?m)(?:^|[;&|])[ \t]*tee[ \t]+(?:-a[ \t]+)?(?P<path>[^\s;&|]+)[ \t]*(?:<[ \t]*/dev/null[ \t]*)?(?P<op><<-?)[ \t]*(?P<delim>'[^']+'|"[^"]+"|[A-Za-z_][A-Za-z0-9_]*)[ \t]*$`,
)

// ParseBashRedirectWrite inspects a Codex Bash tool command and, if the
// command is a heredoc-based file write (`cat <<EOF > FILE … EOF`),
// returns a synthetic PatchFile slice describing the write so the
// OnAfterFileEdit bridge can dispatch it through the same per-file
// pipeline as apply_patch invocations. The boolean return reports
// whether a write was detected, mirroring ParseApplyPatchFromBash's
// shape so callers can chain detectors:
//
//	if files, ok := ParseApplyPatchFromBash(cmd); ok { … }
//	if files, ok := ParseBashRedirectWrite(cmd); ok { … }
//
// Why this matters: Codex `0.130.0`+ routes greenfield file creates
// through plain Bash `cat <<'EOF' > FILE` invocations rather than the
// first-class `apply_patch` tool or any `Write`/`Edit` alias. Without
// this parser the corridor `codex-post-tool-use` handler short-circuits
// on every greenfield write — no afterFileEdit fires, no security scan
// runs, and the dashboard shows zero SecurityScanResult rows for any
// Codex session that only creates new files (the exact symptom that
// motivated this helper; see the regression test below).
//
// Heuristics are deliberately tight to keep false positives low:
//
//  1. The command must contain a `cat <<…> FILE` or `tee FILE <<…`
//     line that ends the logical statement (anchored to end-of-line so
//     trailing tokens like `&& echo done` don't mask the redirect).
//  2. The heredoc delimiter is extracted as-emitted (`'EOF'`, `"EOF"`,
//     or bare `EOF`); single- or double-quotes around it are stripped
//     before scanning the body — Bash treats quoted delimiters as a
//     "no variable expansion" hint, which doesn't change our body
//     extraction.
//  3. Body extraction walks from the line after the matched cat/tee
//     invocation forward until it sees a line that is exactly the
//     delimiter (no leading/trailing whitespace, except for `<<-`
//     which permits leading tabs per POSIX). If the delimiter is
//     missing the parser returns ok=false rather than guessing — the
//     command isn't a well-formed heredoc write.
//
// We never try to evaluate the file path: anything from the post-redir
// token up to the next whitespace/separator is treated as a literal
// path. That matches what Codex emits and avoids us silently rewriting
// `>` redirects to fd numbers or process substitutions, which a real
// shell parser would have to handle.
func ParseBashRedirectWrite(command string) ([]PatchFile, bool) {
	if !strings.Contains(command, "<<") {
		return nil, false
	}

	if files, ok := matchWriteRedirect(command, catRedirectHeredocRE); ok {
		return files, true
	}
	if files, ok := matchWriteRedirect(command, teeRedirectHeredocRE); ok {
		return files, true
	}
	return nil, false
}

func matchWriteRedirect(command string, re *regexp.Regexp) ([]PatchFile, bool) {
	// FindStringSubmatchIndex returns byte offsets so we can locate the
	// heredoc body that begins on the line after the matched header.
	idx := re.FindStringSubmatchIndex(command)
	if idx == nil {
		return nil, false
	}

	names := re.SubexpNames()
	groups := map[string]string{}
	for i, name := range names {
		if name == "" {
			continue
		}
		if start, end := idx[2*i], idx[2*i+1]; start >= 0 && end >= 0 {
			groups[name] = command[start:end]
		}
	}

	delim := stripDelimQuotes(groups["delim"])
	if delim == "" {
		return nil, false
	}
	path := strings.TrimSpace(groups["path"])
	if path == "" {
		return nil, false
	}

	stripIndent := groups["op"] == "<<-"
	body, ok := extractHeredocBody(command, idx[1], delim, stripIndent)
	if !ok {
		return nil, false
	}

	return []PatchFile{{
		Operation: "add",
		FilePath:  path,
		Edits: []PatchEdit{{
			OldString: "",
			NewString: body,
		}},
	}}, true
}

// stripDelimQuotes removes surrounding single or double quotes from a
// heredoc delimiter token. Bash treats `<<'EOF'`, `<<"EOF"`, and
// `<<EOF` as the same end marker; the only difference is whether
// variable/command substitution happens in the body. Body extraction
// doesn't care about substitution semantics — we record what Codex
// physically wrote — so we normalize away the quotes here.
func stripDelimQuotes(delim string) string {
	if len(delim) >= 2 {
		first, last := delim[0], delim[len(delim)-1]
		if (first == '\'' || first == '"') && first == last {
			return delim[1 : len(delim)-1]
		}
	}
	return delim
}

// extractHeredocBody returns the heredoc body that follows headerEnd
// (the byte offset of the end of the matched cat/tee invocation line)
// up to but not including the line that contains only `delim`. The
// returned body never includes the trailing delimiter line itself.
//
// stripIndent reflects the `<<-` operator, which strips leading TAB
// characters from each body line and from the delimiter line; spaces
// are NOT stripped, per POSIX. We follow the same rule so the body we
// hand to OnAfterFileEdit matches what Bash would have written to the
// file.
func extractHeredocBody(command string, headerEnd int, delim string, stripIndent bool) (string, bool) {
	// Body starts on the line after the header. Advance past the
	// optional `\r\n` or `\n` immediately following the match.
	start := headerEnd
	if start < len(command) && command[start] == '\r' {
		start++
	}
	if start < len(command) && command[start] == '\n' {
		start++
	}

	rest := command[start:]
	lines := strings.SplitAfter(rest, "\n")

	var body strings.Builder
	for _, raw := range lines {
		// Drop the trailing newline so we can compare against delim
		// without worrying about \r\n vs \n.
		line := strings.TrimRight(raw, "\r\n")
		cmpLine := line
		if stripIndent {
			cmpLine = strings.TrimLeft(cmpLine, "\t")
		}
		if cmpLine == delim {
			result := body.String()
			result = strings.TrimSuffix(result, "\n")
			return result, true
		}
		if stripIndent {
			line = strings.TrimLeft(line, "\t")
		}
		body.WriteString(line)
		if strings.HasSuffix(raw, "\n") {
			body.WriteByte('\n')
		}
	}
	// Missing terminator — bail rather than silently treating the
	// remainder of the command as body. A well-formed Codex command
	// always terminates the heredoc.
	return "", false
}
