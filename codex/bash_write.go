package codex

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// ParseBashRedirectWrite inspects a Codex Bash tool command and, for
// every heredoc-based file write detected anywhere in the command,
// returns a synthetic PatchFile so the OnAfterFileEdit bridge can
// dispatch each write through the same per-file pipeline as
// apply_patch invocations. The boolean return reports whether at least
// one write was detected, mirroring ParseApplyPatchFromBash's shape so
// callers can chain detectors:
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
// motivated this helper; see the regression tests below).
//
// Implementation note: detection runs the command through a real Bash
// parser (mvdan.cc/sh/v3/syntax) and walks the AST looking for
// CallExpr statements whose primary command is `cat` or `tee` paired
// with a heredoc redirect (`<<` or `<<-`, quoted or unquoted) on the
// same Stmt. This replaced an earlier regex-based implementation that
// reported only the first match in a command — a multi-heredoc shape
// like
//
//	cat <<'EOF' > allowed.txt
//	ok
//	EOF
//	cat <<'EOF' > .env
//	TOKEN=secret
//	EOF
//
// dispatched only the allowed.txt write to OnAfterFileEdit, silently
// bypassing path-deny rules and secret scanners for the .env write. A
// real parser also closes a handful of adjacent gaps the regex never
// covered: redirects in reverse order (`cat > FILE <<EOF`), quoted
// paths (`> "name with spaces.txt"`), writes nested inside compound
// commands (`cd … && cat …`, `if … then cat … fi`, pipelines),
// commands that mix `cat` and `tee` heredoc writes, multi-target tee
// (`tee a.txt b.txt <<EOF` writes BOTH files and emits two PatchFiles),
// and stderr-shadowed cat writes (`cat <<EOF > .env 2> log` correctly
// identifies .env as the write target rather than letting the later 2>
// entry overwrite our pick). Every recognised write becomes a
// PatchFile in the order it appears in the command.
//
// We deliberately fail closed in three places. (1) If the command is
// not valid Bash (parser error) this returns (nil, false) — silently
// treating malformed input as a no-op write would let an attacker
// construct a heredoc the parser rejects but Codex's actual shell
// accepts. (2) If a cat statement has a heredoc but no extractable
// target path (e.g. `cat <<EOF > $(some_cmd)`), that single write is
// dropped from the result but the rest of the command's writes are
// still reported, and ok stays true as long as at least one well-formed
// write was found. (3) If a tee statement has a heredoc but ANY of its
// target paths can't be cleanly extracted, the entire statement is
// dropped — partial coverage of a multi-target tee is the exact
// confused-deputy bypass we're trying to prevent, so we'd rather hand
// the raw command to the unified bridge's fallback than report a
// subset of writes. Callers that need a stricter posture should layer
// their own checks; the unified bridge then surfaces the raw command
// through the existing fallback path.
func ParseBashRedirectWrite(command string) ([]PatchFile, bool) {
	// Cheap pre-check: every heredoc-style write contains `<<`. This
	// keeps the parser allocation off the hot path for plain Bash
	// commands (`pwd`, `ls`, `git diff`, …) which dominate real Codex
	// traffic.
	if !strings.Contains(command, "<<") {
		return nil, false
	}

	file, err := syntax.NewParser().Parse(strings.NewReader(command), "")
	if err != nil {
		return nil, false
	}

	var files []PatchFile
	syntax.Walk(file, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		if patches, ok := bashRedirectWriteFromStmt(command, stmt); ok {
			files = append(files, patches...)
		}
		return true
	})
	if len(files) == 0 {
		return nil, false
	}
	return files, true
}

// bashRedirectWriteFromStmt extracts one PatchFile per write target from
// stmt if its primary command is a recognised heredoc-style file write
// (cat or tee). Most statements yield a single PatchFile, but `tee
// FILE1 FILE2 …` produces one PatchFile per target so every write goes
// through the per-file policy pipeline. Any other shape returns
// ok=false so syntax.Walk can quietly skip it.
func bashRedirectWriteFromStmt(command string, stmt *syntax.Stmt) ([]PatchFile, bool) {
	call, ok := stmt.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) == 0 {
		return nil, false
	}
	switch call.Args[0].Lit() {
	case "cat":
		return catRedirectWrite(command, stmt)
	case "tee":
		return teeRedirectWrite(command, stmt, call)
	}
	return nil, false
}

// catRedirectWrite handles `cat <<EOF > FILE` and `cat > FILE <<EOF`.
// Both shapes carry the heredoc and the `>`/`>>` target as separate
// entries on stmt.Redirs; we pick the last `>`/`>>` on fd 1 (Bash
// semantics — a later redirect on the same fd shadows an earlier one)
// and require exactly one heredoc on the statement. Two heredocs on a
// single cat is invalid Bash and never emitted by Codex; we bail
// rather than guess which one feeds the file.
//
// Crucially we filter out redirects on file descriptors other than
// stdout (fd 1). mvdan.cc/sh's parser represents `2> stderr.log` as
// {Op: RdrOut, N: lit("2"), …} — the same RedirOperator as `> file`,
// distinguished only by the fd field. Without this filter, a command
// like `cat <<EOF > .env 2> allowed.log` would have `out` overwritten
// by the (later) stderr entry and report `allowed.log` to
// OnAfterFileEdit — leaving the actual .env write completely
// unmonitored. `&>` (RdrAll) and `&>>` (AppAll) combine stdout+stderr
// to one file; that file IS a real write target, so we treat them
// like stdout-bound RdrOut/AppOut.
func catRedirectWrite(command string, stmt *syntax.Stmt) ([]PatchFile, bool) {
	var hdoc, out *syntax.Redirect
	for _, r := range stmt.Redirs {
		switch r.Op {
		case syntax.Hdoc, syntax.DashHdoc:
			if hdoc != nil {
				return nil, false
			}
			hdoc = r
		case syntax.RdrOut, syntax.AppOut:
			if !redirectsStdout(r) {
				continue
			}
			out = r
		case syntax.RdrAll, syntax.AppAll:
			out = r
		}
	}
	if hdoc == nil || out == nil {
		return nil, false
	}
	path := wordText(command, out.Word)
	if path == "" {
		return nil, false
	}
	body, ok := heredocBodyContent(command, hdoc)
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

// teeRedirectWrite handles `tee [-a] FILE… [< /dev/null] <<EOF`. tee
// writes its stdin to *every* file operand on the command line, so we
// emit one PatchFile per non-option positional argument; the heredoc
// lives on stmt.Redirs alongside any stdin redirects.
//
// Option detection uses wordToLiteral so quoted flags (`"-a"`) are
// still recognised as flags, and `--` is honoured as the GNU
// end-of-options marker (anything after `--` is a path, even if it
// starts with `-`). Returning only the first target — as the prior
// implementation did — was a confused-deputy bug: a command like
// `tee allowed.txt ../../.ssh/authorized_keys <<EOF` would write both
// files but the policy handler would only ever see `allowed.txt`, so
// any cwd-containment or path-deny rule was silently bypassed for the
// second (and every subsequent) target. We also fail closed if any
// target word doesn't yield a usable path string — partial coverage is
// worse than no coverage, because callers reasonably assume that a
// successful parse means every write was reported.
func teeRedirectWrite(command string, stmt *syntax.Stmt, call *syntax.CallExpr) ([]PatchFile, bool) {
	var paths []string
	endOfOptions := false
	for _, arg := range call.Args[1:] {
		if !endOfOptions {
			if lit, ok := wordToLiteral(arg); ok {
				if lit == "--" {
					endOfOptions = true
					continue
				}
				if strings.HasPrefix(lit, "-") {
					continue
				}
			}
		}
		path := wordText(command, arg)
		if path == "" {
			return nil, false
		}
		paths = append(paths, path)
	}
	if len(paths) == 0 {
		return nil, false
	}
	var hdoc *syntax.Redirect
	for _, r := range stmt.Redirs {
		if r.Op != syntax.Hdoc && r.Op != syntax.DashHdoc {
			continue
		}
		if hdoc != nil {
			return nil, false
		}
		hdoc = r
	}
	if hdoc == nil {
		return nil, false
	}
	body, ok := heredocBodyContent(command, hdoc)
	if !ok {
		return nil, false
	}
	files := make([]PatchFile, 0, len(paths))
	for _, path := range paths {
		files = append(files, PatchFile{
			Operation: "add",
			FilePath:  path,
			Edits: []PatchEdit{{
				OldString: "",
				NewString: body,
			}},
		})
	}
	return files, true
}

// redirectsStdout reports whether r targets file descriptor 1 (stdout).
// mvdan.cc/sh stores the optional explicit fd in r.N (e.g. `2>` →
// N.Value == "2"); a nil/empty N means the default, which for `>` and
// `>>` is fd 1. Anything other than nil/empty/"1" is some other fd
// (stderr, a custom log fd, …) and must NOT be treated as the
// command's primary write target.
func redirectsStdout(r *syntax.Redirect) bool {
	if r.N == nil {
		return true
	}
	switch r.N.Value {
	case "", "1":
		return true
	}
	return false
}

// wordText returns a usable string representation of a syntax.Word —
// preferring the unquoted literal value when all parts reduce to plain
// text, otherwise falling back to the verbatim source slice. The
// fallback matters for paths like `$HOME/.env` where downstream policy
// still needs to see the surface text even though Bash would expand it.
func wordText(command string, w *syntax.Word) string {
	if w == nil {
		return ""
	}
	if lit, ok := wordToLiteral(w); ok {
		return lit
	}
	start := int(w.Pos().Offset())
	end := int(w.End().Offset())
	if start < 0 || end > len(command) || start >= end {
		return ""
	}
	return command[start:end]
}

// wordToLiteral returns w's effective literal text if every part is a
// plain literal or quoted literal (no expansions). The second return
// is false when the word requires shell evaluation; callers fall back
// to the verbatim source slice via wordText.
func wordToLiteral(w *syntax.Word) (string, bool) {
	var b strings.Builder
	for _, part := range w.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			b.WriteString(p.Value)
		case *syntax.SglQuoted:
			b.WriteString(p.Value)
		case *syntax.DblQuoted:
			for _, inner := range p.Parts {
				lit, ok := inner.(*syntax.Lit)
				if !ok {
					return "", false
				}
				b.WriteString(lit.Value)
			}
		default:
			return "", false
		}
	}
	return b.String(), true
}

// heredocBodyContent returns the on-disk content a heredoc redirect
// would have produced — the body bytes with the trailing delimiter
// line removed and POSIX `<<-` tab-stripping applied. mvdan.cc/sh's
// Hdoc Word spans `body…\nDELIM`, so we strip from the last newline
// onward to drop the terminator. An entirely empty heredoc
// (`cat <<EOF\nEOF`) appears as r.Hdoc == nil — that's still a write
// (Bash would create an empty file), so we surface it as "".
func heredocBodyContent(command string, r *syntax.Redirect) (string, bool) {
	if r == nil {
		return "", false
	}
	if r.Hdoc == nil {
		return "", true
	}
	body, ok := heredocBodyFromRedirect(command, r)
	if !ok {
		return "", false
	}
	text := body.text
	if idx := strings.LastIndexByte(text, '\n'); idx >= 0 {
		return text[:idx], true
	}
	return "", true
}
