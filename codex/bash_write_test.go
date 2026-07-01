package codex

import (
	"reflect"
	"testing"
)

func TestParseBashRedirectWrite_CatHeredocSingleQuotedDelim(t *testing.T) {
	// Captured verbatim from a Codex 0.130.0 session that produced no
	// afterFileEdit before we added this parser — the canonical
	// greenfield-write shape.
	cmd := "cat <<'EOF' > greet.txt\nhello world\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	want := []PatchFile{{
		Operation: "add",
		FilePath:  "greet.txt",
		Edits: []PatchEdit{{
			OldString: "",
			NewString: "hello world",
		}},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseBashRedirectWrite =\n  %+v\nwant\n  %+v", got, want)
	}
}

func TestParseBashRedirectWrite_CatHeredocDoubleQuotedDelim(t *testing.T) {
	cmd := "cat <<\"PYEOF\" > app.py\nprint('hi')\nPYEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "app.py" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "app.py")
	}
	if got[0].Edits[0].NewString != "print('hi')" {
		t.Errorf("NewString = %q, want %q", got[0].Edits[0].NewString, "print('hi')")
	}
}

func TestParseBashRedirectWrite_CatHeredocUnquotedDelim(t *testing.T) {
	cmd := "cat <<EOF > notes.md\nfirst line\nsecond line\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].Edits[0].NewString != "first line\nsecond line" {
		t.Errorf("NewString = %q", got[0].Edits[0].NewString)
	}
}

func TestParseBashRedirectWrite_CatHeredocAppendRedirect(t *testing.T) {
	cmd := "cat <<'EOF' >> log.txt\nappended line\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "log.txt" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "log.txt")
	}
}

func TestParseBashRedirectWrite_CdAndCatHeredoc(t *testing.T) {
	// Codex frequently prefixes the cat invocation with a `cd` so the
	// relative path resolves correctly. The regex anchors at `;`, `&`,
	// `|`, or start-of-line, so `&&` separators must still match.
	cmd := "cd /tmp/probe && cat <<'EOF' > out.txt\nbody\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "out.txt" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "out.txt")
	}
}

func TestParseBashRedirectWrite_CatHeredocTabIndent(t *testing.T) {
	// `<<-` strips leading TABs (not spaces) from each body line and
	// from the delimiter line. Verify both behaviours.
	cmd := "cat <<-EOF > indented.txt\n\tline one\n\tline two\n\tEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].Edits[0].NewString != "line one\nline two" {
		t.Errorf("NewString = %q (tabs not stripped)", got[0].Edits[0].NewString)
	}
}

func TestParseBashRedirectWrite_TeeHeredoc(t *testing.T) {
	cmd := "tee out.txt <<'EOF'\nfrom tee\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "out.txt" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "out.txt")
	}
	if got[0].Edits[0].NewString != "from tee" {
		t.Errorf("NewString = %q", got[0].Edits[0].NewString)
	}
}

func TestParseBashRedirectWrite_TeeAppend(t *testing.T) {
	cmd := "tee -a log.txt <<EOF\nappended\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "log.txt" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "log.txt")
	}
}

func TestParseBashRedirectWrite_NoHeredoc_PlainBash(t *testing.T) {
	// Must NOT fire for ordinary Bash commands.
	tests := []string{
		"pwd",
		"ls -la",
		"git diff",
		"cat README.md",
		"echo hello > greet.txt",  // simple redirect (no heredoc) — out of scope
		"printf 'hi' > greet.txt", // not handled here
		"apply_patch <<'PATCH'\n*** Begin Patch\n*** Add File: a\n+x\n*** End Patch\nPATCH",
	}
	for _, cmd := range tests {
		if _, ok := ParseBashRedirectWrite(cmd); ok {
			t.Errorf("ParseBashRedirectWrite(%q) ok = true; want false", cmd)
		}
	}
}

func TestParseBashRedirectWrite_HeredocMissingTerminator(t *testing.T) {
	// Without the closing EOF line we can't trust the body. Return ok=true
	// with no structured files so the bridge dispatches raw fallback.
	cmd := "cat <<'EOF' > greet.txt\nhello world\n"

	files, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("expected ok=true so the bridge can dispatch raw fallback")
	}
	if len(files) != 0 {
		t.Errorf("expected no structured files for heredoc missing terminator; got %+v", files)
	}
}

func TestParseBashRedirectWrite_DelimAppearingInsideBodyIsNotTerminator(t *testing.T) {
	// The delimiter must appear on a line by itself — `EOFISH` should
	// not terminate an `EOF` heredoc.
	cmd := "cat <<'EOF' > out.txt\nEOFISH is not a terminator\nstill in body\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	want := "EOFISH is not a terminator\nstill in body"
	if got[0].Edits[0].NewString != want {
		t.Errorf("NewString = %q, want %q", got[0].Edits[0].NewString, want)
	}
}

func TestParseBashRedirectWrite_PathTraversalIsPreservedVerbatim(t *testing.T) {
	// We never normalize the path — if Codex emits something
	// suspicious like `../../etc/passwd`, downstream policy needs to
	// see it as-is so it can reject it. Verify we don't accidentally
	// rewrite or strip such paths.
	cmd := "cat <<'EOF' > ../../etc/passwd\nroot:x:0:0\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "../../etc/passwd" {
		t.Errorf("FilePath = %q, want preserved verbatim", got[0].FilePath)
	}
}

func TestParseBashRedirectWrite_BodyWithEmptyLines(t *testing.T) {
	cmd := "cat <<'EOF' > out.txt\n\nline two\n\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	want := "\nline two\n"
	if got[0].Edits[0].NewString != want {
		t.Errorf("NewString = %q, want %q", got[0].Edits[0].NewString, want)
	}
}

// Regression: the original regex-based parser used FindStringSubmatchIndex
// inside matchWriteRedirect, which only returns the first match. A
// Codex Bash command chaining two heredoc writes would dispatch only
// the first to OnAfterFileEdit, silently bypassing path-deny rules,
// secret scanners, and audit logging for every subsequent write. See
// the security report on the cursor/add-codex-support branch for the
// attack-path narrative; the literal payload below is the canonical
// shape from that report.
func TestParseBashRedirectWrite_MultipleCatHeredocs_AllReported(t *testing.T) {
	cmd := "cat <<'EOF' > allowed.txt\nok\nEOF\ncat <<'EOF' > .env\nTOKEN=secret\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	want := []PatchFile{
		{Operation: "add", FilePath: "allowed.txt", Edits: []PatchEdit{{OldString: "", NewString: "ok"}}},
		{Operation: "add", FilePath: ".env", Edits: []PatchEdit{{OldString: "", NewString: "TOKEN=secret"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseBashRedirectWrite =\n  %+v\nwant\n  %+v", got, want)
	}
}

// Regression: multi-heredoc commands prefixed with the `cd … && cat …`
// shape Codex often emits — Stmts nested inside BinaryCmds must still
// be visited by the AST walk and reported alongside the top-level
// writes that follow them.
func TestParseBashRedirectWrite_MultipleCatHeredocs_WithCdPrefix(t *testing.T) {
	cmd := "cd /tmp && cat <<'EOF' > a.txt\nA\nEOF\ncat <<'EOF' > b.txt\nB\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 PatchFiles, got %d: %+v", len(got), got)
	}
	if got[0].FilePath != "a.txt" || got[1].FilePath != "b.txt" {
		t.Errorf("FilePaths = [%q, %q], want [a.txt, b.txt]", got[0].FilePath, got[1].FilePath)
	}
}

// Regression: a command that mixes cat and tee heredoc writes must
// report both. The original ParseBashRedirectWrite returned on the
// first matchWriteRedirect hit, so the catRedirectHeredocRE branch
// short-circuited the teeRedirectHeredocRE branch even when the
// command contained a tee write the cat regex didn't match.
func TestParseBashRedirectWrite_CatAndTeeMixed_BothReported(t *testing.T) {
	cmd := "cat <<'EOF' > from_cat.txt\nA\nEOF\ntee from_tee.txt <<'EOF'\nB\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 PatchFiles, got %d: %+v", len(got), got)
	}
	if got[0].FilePath != "from_cat.txt" || got[1].FilePath != "from_tee.txt" {
		t.Errorf("FilePaths = [%q, %q], want [from_cat.txt, from_tee.txt]", got[0].FilePath, got[1].FilePath)
	}
}

// Regression: redirects can legally appear before the heredoc in Bash
// (`cat > FILE <<EOF`). The original regex anchored the `>` after the
// heredoc declarator and missed this shape entirely; the AST walk
// finds both redirects regardless of order.
func TestParseBashRedirectWrite_CatReverseRedirectOrder(t *testing.T) {
	cmd := "cat > greet.txt <<'EOF'\nhello\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "greet.txt" {
		t.Errorf("FilePath = %q, want greet.txt", got[0].FilePath)
	}
	if got[0].Edits[0].NewString != "hello" {
		t.Errorf("NewString = %q, want hello", got[0].Edits[0].NewString)
	}
}

// Regression: paths that need shell quoting (spaces, special chars)
// were broken by the regex's `[^\s;&|]+` path class. The AST parser
// gives us a syntax.Word we can unquote correctly.
func TestParseBashRedirectWrite_QuotedPathWithSpaces(t *testing.T) {
	cmd := "cat <<'EOF' > \"name with spaces.txt\"\nbody\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "name with spaces.txt" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "name with spaces.txt")
	}
}

// Regression: writes nested inside an `if`/`fi` block (and other
// compound shapes) must still be visited by the walker. The regex
// implementation happened to match these because it ignored block
// structure entirely; the AST implementation has to actively recurse
// into compound commands. This locks that recursion in.
func TestParseBashRedirectWrite_HeredocInsideIfBlock(t *testing.T) {
	cmd := "if true; then cat <<'EOF' > out.txt\nbody\nEOF\nfi"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if len(got) != 1 || got[0].FilePath != "out.txt" {
		t.Fatalf("got %+v, want one PatchFile with FilePath=out.txt", got)
	}
}

// Regression: a quoted target path (single or double quotes) must be
// reported to OnAfterFileEdit with the surrounding quotes stripped —
// i.e. the value Bash would actually write to, not the surface token.
//
// The original regex implementation captured the path via
// `[^\s;&|]+`, which happily included surrounding quote characters,
// so policies running `filepath.Rel(repoRoot, ctx.FilePath)` followed
// by `strings.HasPrefix(rel, "..")` would see a string starting with
// `'` or `"` and the containment check would silently pass while the
// dangerous write proceeded. The AST-based parser unquotes
// SglQuoted/DblQuoted Word parts via wordToLiteral; this test pins
// that behavior so a future refactor can't regress it.
func TestParseBashRedirectWrite_CatSingleQuotedPathEscapesContainment(t *testing.T) {
	cmd := "cat <<'EOF' > '../../.ssh/authorized_keys'\nattacker-controlled\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "../../.ssh/authorized_keys" {
		t.Errorf("FilePath = %q, want unquoted %q (quoted-path containment bypass)",
			got[0].FilePath, "../../.ssh/authorized_keys")
	}
}

func TestParseBashRedirectWrite_CatDoubleQuotedPathEscapesContainment(t *testing.T) {
	cmd := "cat <<'EOF' > \"../../.ssh/authorized_keys\"\nattacker-controlled\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "../../.ssh/authorized_keys" {
		t.Errorf("FilePath = %q, want unquoted %q",
			got[0].FilePath, "../../.ssh/authorized_keys")
	}
}

func TestParseBashRedirectWrite_TeeQuotedPathEscapesContainment(t *testing.T) {
	cmd := "tee '../../.ssh/authorized_keys' <<'EOF'\nattacker-controlled\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "../../.ssh/authorized_keys" {
		t.Errorf("FilePath = %q, want unquoted %q", got[0].FilePath, "../../.ssh/authorized_keys")
	}
}

// Regression: an invalid Bash command (unclosed heredoc, dangling
// quote, etc.) must fail closed rather than fall back to a partial
// parse. The unified bridge surfaces the raw command through its
// fallback path when both detectors return ok=false, so the
// underlying policy still sees the event — but only after the
// detector has explicitly declined to invent a PatchFile out of
// malformed input.
func TestParseBashRedirectWrite_InvalidBash_FailsClosed(t *testing.T) {
	cmd := "cat <<'EOF' > greet.txt\nhello\n" // missing terminator

	files, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("expected ok=true so the bridge can dispatch raw fallback")
	}
	if len(files) != 0 {
		t.Errorf("expected no structured files on invalid Bash; got %+v", files)
	}
}

// Regression: when one heredoc in a multi-write command is well-formed
// and a later token makes the command unparseable, we should not
// silently return the well-formed write — the entire command is
// suspect and the unified bridge's fallback path is more appropriate.
func TestParseBashRedirectWrite_PartiallyMalformed_FailsClosed(t *testing.T) {
	// Second heredoc never terminates; parser will reject the whole
	// command. The first write must not slip through.
	cmd := "cat <<'EOF' > a.txt\nA\nEOF\ncat <<'EOF' > b.txt\nB\n"

	files, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("expected ok=true so the bridge can dispatch raw fallback")
	}
	if len(files) != 0 {
		t.Errorf("expected no structured files when later heredoc lacks terminator; got %+v", files)
	}
}

// Regression: tee writes its stdin to EVERY file operand, not just
// the first. The original teeRedirectWrite extracted only the first
// non-option positional arg and broke out of the loop, so a command
// like `tee allowed.txt ../../.ssh/authorized_keys <<EOF` would create
// both files on disk but report only allowed.txt to OnAfterFileEdit.
// Any cwd-containment or path-deny policy would never see the
// authorized_keys write — a confused-deputy bypass that needs no
// shell-expansion trick, since every target is a plain literal that
// passes wordToLiteral cleanly. This test pins the multi-target shape:
// one PatchFile per target, all sharing the heredoc body, in source
// order.
func TestParseBashRedirectWrite_TeeMultipleTargets_AllReported(t *testing.T) {
	cmd := "tee allowed.txt ../../.ssh/authorized_keys <<'EOF'\nattacker-controlled\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	want := []PatchFile{
		{Operation: "add", FilePath: "allowed.txt", Edits: []PatchEdit{{OldString: "", NewString: "attacker-controlled"}}},
		{Operation: "add", FilePath: "../../.ssh/authorized_keys", Edits: []PatchEdit{{OldString: "", NewString: "attacker-controlled"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseBashRedirectWrite =\n  %+v\nwant\n  %+v", got, want)
	}
}

func TestParseBashRedirectWrite_CatPipeTeeReported(t *testing.T) {
	cmd := "cat <<'EOF' | tee allowed.txt ../../.ssh/authorized_keys >/dev/null\nattacker-controlled\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	want := []PatchFile{
		{Operation: "add", FilePath: "allowed.txt", Edits: []PatchEdit{{OldString: "", NewString: "attacker-controlled"}}},
		{Operation: "add", FilePath: "../../.ssh/authorized_keys", Edits: []PatchEdit{{OldString: "", NewString: "attacker-controlled"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseBashRedirectWrite =\n  %+v\nwant\n  %+v", got, want)
	}
}

func TestParseBashRedirectWrite_NonLiteralTargetUsesFallback(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{
			name: "environment expansion",
			cmd:  "cat <<'EOF' > $HOME/.ssh/authorized_keys\nkey\nEOF",
		},
		{
			name: "tilde expansion",
			cmd:  "cat <<'EOF' > ~/secret.txt\nsecret\nEOF",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, ok := ParseBashRedirectWrite(tt.cmd)
			if !ok {
				t.Fatalf("expected ok=true so the bridge can dispatch raw fallback")
			}
			if len(files) != 0 {
				t.Fatalf("files = %+v, want raw fallback with no structured files", files)
			}
		})
	}
}

// Regression: `tee -a FILE1 FILE2 <<EOF` (append mode, multi-target)
// must still report both targets. Verifies that the `-a` flag is
// skipped without short-circuiting the rest of the positional walk.
func TestParseBashRedirectWrite_TeeAppendMultipleTargets(t *testing.T) {
	cmd := "tee -a log1.txt log2.txt <<'EOF'\nappended\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if len(got) != 2 || got[0].FilePath != "log1.txt" || got[1].FilePath != "log2.txt" {
		t.Fatalf("expected [log1.txt, log2.txt], got %+v", got)
	}
}

// Regression: GNU tee accepts `--` as an end-of-options marker, so
// any subsequent positional starting with `-` is a path, not a flag.
// Without this handling, a target like `-weird.txt` would be silently
// dropped from the report (or worse, an attacker could disguise a
// dangerous path as something the option-strip heuristic skips).
func TestParseBashRedirectWrite_TeeEndOfOptionsMarker(t *testing.T) {
	cmd := "tee -a -- -weird.txt other.txt <<'EOF'\nbody\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if len(got) != 2 || got[0].FilePath != "-weird.txt" || got[1].FilePath != "other.txt" {
		t.Fatalf("expected [-weird.txt, other.txt], got %+v", got)
	}
}

// Regression: a quoted flag (`"-a"`) is semantically identical to the
// unquoted form once Bash strips the quotes. The original
// implementation used Word.Lit() which returns "" for fully-quoted
// words, so a quoted `-a` would be misclassified as a path and end up
// in the targets slice. wordToLiteral handles this correctly.
func TestParseBashRedirectWrite_TeeQuotedFlagSkipped(t *testing.T) {
	cmd := "tee \"-a\" log.txt <<'EOF'\nbody\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if len(got) != 1 || got[0].FilePath != "log.txt" {
		t.Fatalf("expected single PatchFile for log.txt, got %+v", got)
	}
}

// Regression: mvdan.cc/sh represents `2> stderr.log` as
// {Op: RdrOut, N: lit("2"), …} — the same RedirOperator as a stdout
// `> file` redirect, distinguished only by the fd field N. The
// original catRedirectWrite ignored N and just picked the last
// RdrOut/AppOut entry, so a command like
// `cat <<EOF > .env 2> allowed.log` would have `out` overwritten by
// the (later) stderr entry and report `allowed.log` as the write
// target. Path-deny rules on `.env` and secret scanners would never
// see the actual write — a complete bypass with no expansion trick.
func TestParseBashRedirectWrite_CatStderrShadowsStdoutTarget(t *testing.T) {
	cmd := "cat <<'EOF' > .env 2> allowed.log\nTOKEN=secret\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 PatchFile, got %d: %+v", len(got), got)
	}
	if got[0].FilePath != ".env" {
		t.Errorf("FilePath = %q, want .env (stderr 2> redirect must not shadow stdout target)",
			got[0].FilePath)
	}
	if got[0].Edits[0].NewString != "TOKEN=secret" {
		t.Errorf("NewString = %q, want %q", got[0].Edits[0].NewString, "TOKEN=secret")
	}
}

// Regression: explicit `1>` is identical to bare `>` — both target
// stdout. Verify the fd filter accepts `1` as a synonym for "no fd".
func TestParseBashRedirectWrite_CatExplicitStdoutFd(t *testing.T) {
	cmd := "cat <<'EOF' 1> out.txt\nbody\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "out.txt" {
		t.Errorf("FilePath = %q, want out.txt", got[0].FilePath)
	}
}

// Regression: a cat statement whose ONLY `>` redirect is on a
// non-stdout fd (e.g. `cat <<EOF 2> log.txt` — stdin from heredoc,
// stderr to log.txt, stdout discarded) has no actual write target we
// can attribute to the heredoc body. The heredoc body just goes to
// stdout (which here is the default tty/pipe), so we must NOT report
// log.txt as if the heredoc landed there. The detector returns
// ok=false and the unified bridge's fallback handles the raw command.
func TestParseBashRedirectWrite_CatStderrOnlyRedirect_NotReported(t *testing.T) {
	cmd := "cat <<'EOF' 2> log.txt\nbody\nEOF"

	if files, ok := ParseBashRedirectWrite(cmd); ok {
		t.Errorf("expected ok=false when only redirect is non-stdout; got %+v", files)
	}
}

// Regression: `&>` (RdrAll) redirects both stdout and stderr to a
// single file. That file IS the heredoc's write target, so it must
// be reported. Without explicit handling of RdrAll/AppAll, a command
// like `cat <<EOF &> .env` would silently bypass detection.
func TestParseBashRedirectWrite_CatRdrAll(t *testing.T) {
	cmd := "cat <<'EOF' &> combined.log\nbody\nEOF"

	got, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("ParseBashRedirectWrite(...) ok = false; want true")
	}
	if got[0].FilePath != "combined.log" {
		t.Errorf("FilePath = %q, want combined.log", got[0].FilePath)
	}
}

// Regression: a tee statement with multiple targets where ONE target
// is unparseable (empty word offsets, etc.) must fail closed for the
// whole statement — partial coverage of a multi-target tee is the
// exact bypass we're guarding against. We trigger the empty-path
// codepath via `tee "" foo.txt <<EOF` (an empty literal target).
func TestParseBashRedirectWrite_TeeUnparseableTarget_FailsClosed(t *testing.T) {
	cmd := "tee \"\" foo.txt <<'EOF'\nbody\nEOF"

	files, ok := ParseBashRedirectWrite(cmd)
	if !ok {
		t.Fatalf("expected ok=true so the bridge can dispatch raw fallback")
	}
	if len(files) != 0 {
		t.Errorf("expected raw fallback when a tee target is empty; got %+v", files)
	}
}
