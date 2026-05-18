package codex

import (
	"reflect"
	"testing"
)

func TestParseApplyPatch_AddFile(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Add File: secrets/api_key.txt\n" +
		"+sk-deadbeef\n" +
		"+second line\n" +
		"*** End Patch\n"

	got := ParseApplyPatch(patch)
	want := []PatchFile{{
		Operation: "add",
		FilePath:  "secrets/api_key.txt",
		Edits: []PatchEdit{{
			OldString: "",
			NewString: "sk-deadbeef\nsecond line",
		}},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseApplyPatch(add) =\n  %+v\nwant\n  %+v", got, want)
	}
}

func TestParseApplyPatch_DeleteFile(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Delete File: old/secrets.env\n" +
		"*** End Patch\n"

	got := ParseApplyPatch(patch)
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1", len(got))
	}
	if got[0].Operation != "delete" {
		t.Errorf("Operation = %q, want %q", got[0].Operation, "delete")
	}
	if got[0].FilePath != "old/secrets.env" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "old/secrets.env")
	}
	if len(got[0].Edits) != 0 {
		t.Errorf("Edits = %+v, want empty", got[0].Edits)
	}
}

func TestParseApplyPatch_UpdateFile(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Update File: src/auth.go\n" +
		"@@ func login\n" +
		" func login() {\n" +
		"-    token := \"hardcoded\"\n" +
		"+    token := os.Getenv(\"TOKEN\")\n" +
		"     return token\n" +
		" }\n" +
		"*** End Patch\n"

	got := ParseApplyPatch(patch)
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1", len(got))
	}
	if got[0].FilePath != "src/auth.go" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "src/auth.go")
	}
	if got[0].Operation != "update" {
		t.Errorf("Operation = %q, want %q", got[0].Operation, "update")
	}
	if len(got[0].Edits) != 1 {
		t.Fatalf("Edits = %+v, want 1 edit", got[0].Edits)
	}
	wantEdit := PatchEdit{
		OldString: "    token := \"hardcoded\"",
		NewString: "    token := os.Getenv(\"TOKEN\")",
	}
	if got[0].Edits[0] != wantEdit {
		t.Errorf("Edits[0] = %+v, want %+v", got[0].Edits[0], wantEdit)
	}
}

func TestParseApplyPatch_UpdateFile_MultipleHunks(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Update File: a.go\n" +
		"@@\n" +
		"-foo\n" +
		"+bar\n" +
		"@@\n" +
		"-baz\n" +
		"+qux\n" +
		"*** End Patch\n"

	got := ParseApplyPatch(patch)
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1", len(got))
	}
	if len(got[0].Edits) != 2 {
		t.Fatalf("Edits = %+v, want 2 edits", got[0].Edits)
	}
	if got[0].Edits[0] != (PatchEdit{OldString: "foo", NewString: "bar"}) {
		t.Errorf("Edits[0] = %+v", got[0].Edits[0])
	}
	if got[0].Edits[1] != (PatchEdit{OldString: "baz", NewString: "qux"}) {
		t.Errorf("Edits[1] = %+v", got[0].Edits[1])
	}
}

func TestParseApplyPatch_MultiFile(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Add File: new.txt\n" +
		"+hello\n" +
		"*** Update File: existing.go\n" +
		"@@\n" +
		"-old\n" +
		"+new\n" +
		"*** Delete File: stale.txt\n" +
		"*** End Patch\n"

	got := ParseApplyPatch(patch)
	if len(got) != 3 {
		t.Fatalf("got %d files, want 3", len(got))
	}
	if got[0].FilePath != "new.txt" || got[0].Operation != "add" {
		t.Errorf("files[0] = %+v", got[0])
	}
	if got[1].FilePath != "existing.go" || got[1].Operation != "update" {
		t.Errorf("files[1] = %+v", got[1])
	}
	if got[2].FilePath != "stale.txt" || got[2].Operation != "delete" {
		t.Errorf("files[2] = %+v", got[2])
	}
}

func TestParseApplyPatch_MoveTo(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Update File: old/path.go\n" +
		"*** Move to: new/path.go\n" +
		"@@\n" +
		"-foo\n" +
		"+bar\n" +
		"*** End Patch\n"

	got := ParseApplyPatch(patch)
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1", len(got))
	}
	if got[0].NewFilePath != "new/path.go" {
		t.Errorf("NewFilePath = %q, want %q", got[0].NewFilePath, "new/path.go")
	}
	if got[0].FilePath != "old/path.go" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "old/path.go")
	}
}

func TestParseApplyPatch_LeadingWrapper(t *testing.T) {
	patch := "apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: foo.txt\n" +
		"+hi\n" +
		"*** End Patch\n" +
		"PATCH\n"

	got := ParseApplyPatch(patch)
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1", len(got))
	}
	if got[0].FilePath != "foo.txt" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "foo.txt")
	}
}

func TestParseApplyPatch_EmptyAndMalformed(t *testing.T) {
	if files := ParseApplyPatch(""); files != nil {
		t.Errorf("empty input should yield nil, got %+v", files)
	}
	if files := ParseApplyPatch("not a patch"); files != nil {
		t.Errorf("non-patch input should yield nil, got %+v", files)
	}
}

// ParseApplyPatchFromBash tests
// ---------------------------------------------------------------------------
// These cover the path used by the unified codex-post-tool-use bridge when
// tool_name is "Bash" rather than "apply_patch". The patches below are
// real-shape examples copied from Codex hook payloads observed in the
// dashboard (see hook_events.data for tool_name="Bash" rows).

func TestParseApplyPatchFromBash_HeredocInvocation(t *testing.T) {
	cmd := "apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: foo.txt\n" +
		"+hello\n" +
		"*** End Patch\n" +
		"PATCH"

	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("ParseApplyPatchFromBash(heredoc) returned ok=false, want true")
	}
	if len(files) != 1 || files[0].FilePath != "foo.txt" {
		t.Errorf("files = %+v, want one file with FilePath=foo.txt", files)
	}
}

func TestParseApplyPatchFromBash_AbsolutePathBinary(t *testing.T) {
	// Codex sometimes invokes a per-session apply_patch shim by absolute
	// path. This was observed in the dashboard for session
	// 019dc16e-dd6a-70c3-9e09-2dedd6bab556. Detection must still fire.
	cmd := "/Users/me/.codex/tmp/arg0/codex-arg0IuQk4E/apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Update File: app/routes.py\n" +
		"@@\n" +
		"-old\n" +
		"+new\n" +
		"*** End Patch\n" +
		"PATCH"

	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("ParseApplyPatchFromBash(absolute path) returned ok=false, want true")
	}
	if len(files) != 1 || files[0].FilePath != "app/routes.py" {
		t.Errorf("files = %+v, want one file with FilePath=app/routes.py", files)
	}
	if len(files[0].Edits) != 1 || files[0].Edits[0].OldString != "old" || files[0].Edits[0].NewString != "new" {
		t.Errorf("edits = %+v, want [{old,new}]", files[0].Edits)
	}
}

func TestParseApplyPatchFromBash_MultiFile(t *testing.T) {
	cmd := "apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: a.txt\n" +
		"+a\n" +
		"*** Delete File: b.txt\n" +
		"*** End Patch\n" +
		"PATCH"

	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("ParseApplyPatchFromBash returned ok=false, want true")
	}
	if len(files) != 2 {
		t.Fatalf("len(files) = %d, want 2", len(files))
	}
	if files[0].Operation != "add" || files[0].FilePath != "a.txt" {
		t.Errorf("files[0] = %+v, want add a.txt", files[0])
	}
	if files[1].Operation != "delete" || files[1].FilePath != "b.txt" {
		t.Errorf("files[1] = %+v, want delete b.txt", files[1])
	}
}

func TestParseApplyPatchFromBash_PlainBashCommandSkipped(t *testing.T) {
	// A garden-variety Bash command that has nothing to do with editing
	// files. Detection must return ok=false so the unified handler can
	// short-circuit without parsing.
	if _, ok := ParseApplyPatchFromBash("ls -la /tmp"); ok {
		t.Error("ls -la /tmp wrongly detected as apply_patch")
	}
	if _, ok := ParseApplyPatchFromBash("git status"); ok {
		t.Error("git status wrongly detected as apply_patch")
	}
	if _, ok := ParseApplyPatchFromBash(""); ok {
		t.Error("empty command wrongly detected as apply_patch")
	}
}

func TestParseApplyPatchFromBash_PatchTextInsideUnrelatedHeredocSkipped(t *testing.T) {
	// A Bash heredoc that writes a documentation file containing the
	// literal "*** Begin Patch" marker — but with NO apply_patch
	// invocation before the marker. This must not be classified as a
	// file edit; otherwise every docs PR about apply_patch would trigger
	// spurious afterFileEdit telemetry.
	cmd := "cat > docs/apply_patch_format.md <<'EOF'\n" +
		"This documents the apply_patch format. Example envelope:\n" +
		"\n" +
		"*** Begin Patch\n" +
		"*** Add File: example.txt\n" +
		"+hi\n" +
		"*** End Patch\n" +
		"EOF"

	if _, ok := ParseApplyPatchFromBash(cmd); ok {
		t.Error("docs heredoc wrongly detected as apply_patch invocation")
	}
}

func TestParseApplyPatchFromBash_ApplyPatchMentionAfterEnvelopeSkipped(t *testing.T) {
	// Edge case: "apply_patch" appears in the file content (inside the
	// patch body) but not before the envelope. Detection requires the
	// token to be BEFORE *** Begin Patch.
	cmd := "cat > note.txt <<'EOF'\n" +
		"*** Begin Patch (literal text describing apply_patch syntax)\n" +
		"EOF"

	if _, ok := ParseApplyPatchFromBash(cmd); ok {
		t.Error("apply_patch mention after envelope wrongly detected")
	}
}

func TestParseApplyPatchFromBash_HeredocVariants(t *testing.T) {
	// Heredoc operator variants: <<EOF (unquoted), <<-EOF (strip leading
	// tabs), and quoted '<<"EOF"'. All should be detected.
	variants := []string{
		"apply_patch <<PATCH\n*** Begin Patch\n*** Add File: a.txt\n+x\n*** End Patch\nPATCH",
		"apply_patch <<-PATCH\n*** Begin Patch\n*** Add File: a.txt\n+x\n*** End Patch\nPATCH",
		"apply_patch <<\"PATCH\"\n*** Begin Patch\n*** Add File: a.txt\n+x\n*** End Patch\nPATCH",
	}
	for i, cmd := range variants {
		if _, ok := ParseApplyPatchFromBash(cmd); !ok {
			t.Errorf("variant[%d] not detected: %q", i, cmd)
		}
	}
}
