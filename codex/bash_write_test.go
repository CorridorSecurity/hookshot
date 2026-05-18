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
