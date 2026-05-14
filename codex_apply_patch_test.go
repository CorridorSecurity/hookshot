package hookshot

import (
	"reflect"
	"testing"
)

func TestParseCodexApplyPatch_AddFile(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Add File: secrets/api_key.txt\n" +
		"+sk-deadbeef\n" +
		"+second line\n" +
		"*** End Patch\n"

	got := parseCodexApplyPatch(patch)
	want := []codexApplyPatchFile{{
		Operation: "add",
		FilePath:  "secrets/api_key.txt",
		Edits: []FileEdit{{
			OldString: "",
			NewString: "sk-deadbeef\nsecond line",
		}},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseCodexApplyPatch(add) =\n  %+v\nwant\n  %+v", got, want)
	}
}

func TestParseCodexApplyPatch_DeleteFile(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Delete File: old/secrets.env\n" +
		"*** End Patch\n"

	got := parseCodexApplyPatch(patch)
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

func TestParseCodexApplyPatch_UpdateFile(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Update File: src/auth.go\n" +
		"@@ func login\n" +
		" func login() {\n" +
		"-    token := \"hardcoded\"\n" +
		"+    token := os.Getenv(\"TOKEN\")\n" +
		"     return token\n" +
		" }\n" +
		"*** End Patch\n"

	got := parseCodexApplyPatch(patch)
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
	wantEdit := FileEdit{
		OldString: "    token := \"hardcoded\"",
		NewString: "    token := os.Getenv(\"TOKEN\")",
	}
	if got[0].Edits[0] != wantEdit {
		t.Errorf("Edits[0] = %+v, want %+v", got[0].Edits[0], wantEdit)
	}
}

func TestParseCodexApplyPatch_UpdateFile_MultipleHunks(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Update File: a.go\n" +
		"@@\n" +
		"-foo\n" +
		"+bar\n" +
		"@@\n" +
		"-baz\n" +
		"+qux\n" +
		"*** End Patch\n"

	got := parseCodexApplyPatch(patch)
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1", len(got))
	}
	if len(got[0].Edits) != 2 {
		t.Fatalf("Edits = %+v, want 2 edits", got[0].Edits)
	}
	if got[0].Edits[0] != (FileEdit{OldString: "foo", NewString: "bar"}) {
		t.Errorf("Edits[0] = %+v", got[0].Edits[0])
	}
	if got[0].Edits[1] != (FileEdit{OldString: "baz", NewString: "qux"}) {
		t.Errorf("Edits[1] = %+v", got[0].Edits[1])
	}
}

func TestParseCodexApplyPatch_MultiFile(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Add File: new.txt\n" +
		"+hello\n" +
		"*** Update File: existing.go\n" +
		"@@\n" +
		"-old\n" +
		"+new\n" +
		"*** Delete File: stale.txt\n" +
		"*** End Patch\n"

	got := parseCodexApplyPatch(patch)
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

func TestParseCodexApplyPatch_MoveTo(t *testing.T) {
	patch := "*** Begin Patch\n" +
		"*** Update File: old/path.go\n" +
		"*** Move to: new/path.go\n" +
		"@@\n" +
		"-foo\n" +
		"+bar\n" +
		"*** End Patch\n"

	got := parseCodexApplyPatch(patch)
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

func TestParseCodexApplyPatch_LeadingWrapper(t *testing.T) {
	// Some Codex invocations wrap the patch in a here-doc style header.
	// The parser should skip everything before the Begin Patch marker.
	patch := "apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: foo.txt\n" +
		"+hi\n" +
		"*** End Patch\n" +
		"PATCH\n"

	got := parseCodexApplyPatch(patch)
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1", len(got))
	}
	if got[0].FilePath != "foo.txt" {
		t.Errorf("FilePath = %q, want %q", got[0].FilePath, "foo.txt")
	}
}

func TestParseCodexApplyPatch_EmptyAndMalformed(t *testing.T) {
	if files := parseCodexApplyPatch(""); files != nil {
		t.Errorf("empty input should yield nil, got %+v", files)
	}
	if files := parseCodexApplyPatch("not a patch"); files != nil {
		t.Errorf("non-patch input should yield nil, got %+v", files)
	}
}
