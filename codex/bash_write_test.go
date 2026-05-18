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
		"echo hello > greet.txt",     // simple redirect (no heredoc) — out of scope
		"printf 'hi' > greet.txt",    // not handled here
		"apply_patch <<'PATCH'\n*** Begin Patch\n*** Add File: a\n+x\n*** End Patch\nPATCH",
	}
	for _, cmd := range tests {
		if _, ok := ParseBashRedirectWrite(cmd); ok {
			t.Errorf("ParseBashRedirectWrite(%q) ok = true; want false", cmd)
		}
	}
}

func TestParseBashRedirectWrite_HeredocMissingTerminator(t *testing.T) {
	// Without the closing EOF line we can't trust the body — bail.
	cmd := "cat <<'EOF' > greet.txt\nhello world\n"

	if _, ok := ParseBashRedirectWrite(cmd); ok {
		t.Errorf("expected ok=false for heredoc missing terminator")
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

// Regression: an invalid Bash command (unclosed heredoc, dangling
// quote, etc.) must fail closed rather than fall back to a partial
// parse. The unified bridge surfaces the raw command through its
// fallback path when both detectors return ok=false, so the
// underlying policy still sees the event — but only after the
// detector has explicitly declined to invent a PatchFile out of
// malformed input.
func TestParseBashRedirectWrite_InvalidBash_FailsClosed(t *testing.T) {
	cmd := "cat <<'EOF' > greet.txt\nhello\n" // missing terminator

	if files, ok := ParseBashRedirectWrite(cmd); ok {
		t.Errorf("expected ok=false on invalid Bash; got %+v", files)
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

	if files, ok := ParseBashRedirectWrite(cmd); ok {
		t.Errorf("expected ok=false when later heredoc lacks terminator; got %+v", files)
	}
}
