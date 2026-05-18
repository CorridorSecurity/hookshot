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

// ---------------------------------------------------------------------------
// Regression tests for two bypasses of the previous regex+strings.Index
// detector. The AST-based implementation closes both. See the doc comment
// on ParseApplyPatchFromBash for the full attack description.

func TestParseApplyPatchFromBash_DecoyMarkerBeforeRealInvocation(t *testing.T) {
	// Previously: strings.Index anchored on the decoy `*** Begin Patch`
	// inside the printf string, so the regex's prefix was just
	// `printf '` and detection returned ok=false — the real apply_patch
	// heredoc below was silently dispatched as a plain Bash command,
	// bypassing OnAfterFileEdit policies entirely.
	cmd := "printf '*** Begin Patch\\n' >/dev/null; apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: /sensitive/path\n" +
		"+malicious content\n" +
		"*** End Patch\n" +
		"PATCH"

	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("decoy-marker bypass: ParseApplyPatchFromBash returned ok=false, want true")
	}
	if len(files) != 1 || files[0].FilePath != "/sensitive/path" || files[0].Operation != "add" {
		t.Errorf("files = %+v, want one add for /sensitive/path", files)
	}
}

func TestParseApplyPatchFromBash_ZeroSpaceBeforeHeredoc(t *testing.T) {
	// Previously: the regex required `apply_patch[ \t]+<<`, so the
	// zero-space `apply_patch<<'PATCH'` (valid Bash) evaded detection.
	cmd := "apply_patch<<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: /tmp/zero-space.txt\n" +
		"+hi\n" +
		"*** End Patch\n" +
		"PATCH"

	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("zero-space heredoc: ParseApplyPatchFromBash returned ok=false, want true")
	}
	if len(files) != 1 || files[0].FilePath != "/tmp/zero-space.txt" {
		t.Errorf("files = %+v, want one add for /tmp/zero-space.txt", files)
	}
}

func TestParseApplyPatchFromBash_ZeroSpaceDashHeredoc(t *testing.T) {
	// Combine the two edge cases: no space before <<-, and DashHdoc
	// (tab-stripping) semantics. The detector must apply POSIX
	// tab-stripping to the body before handing it to ParseApplyPatch,
	// otherwise the leading tabs would prevent the `*** Begin Patch`
	// line prefix match.
	cmd := "apply_patch<<-PATCH\n" +
		"\t*** Begin Patch\n" +
		"\t*** Add File: a.txt\n" +
		"\t+x\n" +
		"\t*** End Patch\n" +
		"\tPATCH"

	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("zero-space <<-PATCH: ParseApplyPatchFromBash returned ok=false, want true")
	}
	if len(files) != 1 || files[0].FilePath != "a.txt" {
		t.Errorf("files = %+v, want one add for a.txt", files)
	}
}

func TestParseApplyPatchFromBash_MultipleInvocations(t *testing.T) {
	// Two apply_patch invocations chained with `;` — both should be
	// parsed and their files concatenated in source order. This guards
	// against a future refactor that "first-match wins" silently drops
	// edits from later invocations.
	cmd := "apply_patch <<'A'\n" +
		"*** Begin Patch\n" +
		"*** Add File: first.txt\n" +
		"+1\n" +
		"*** End Patch\n" +
		"A\n" +
		"apply_patch <<'B'\n" +
		"*** Begin Patch\n" +
		"*** Add File: second.txt\n" +
		"+2\n" +
		"*** End Patch\n" +
		"B"

	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("multi-invocation: ParseApplyPatchFromBash returned ok=false, want true")
	}
	if len(files) != 2 {
		t.Fatalf("len(files) = %d, want 2 (one per invocation)", len(files))
	}
	if files[0].FilePath != "first.txt" {
		t.Errorf("files[0].FilePath = %q, want first.txt", files[0].FilePath)
	}
	if files[1].FilePath != "second.txt" {
		t.Errorf("files[1].FilePath = %q, want second.txt", files[1].FilePath)
	}
}

func TestParseApplyPatchFromBash_InvocationInsideSubshell(t *testing.T) {
	// `( ... )` runs the inner commands in a subshell. The AST walker
	// must descend into the subshell so the apply_patch invocation is
	// still detected — otherwise an agent could nest its real
	// invocation inside `()` to evade the detector.
	cmd := "(apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: nested.txt\n" +
		"+x\n" +
		"*** End Patch\n" +
		"PATCH\n)"

	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("subshell-nested apply_patch: returned ok=false, want true")
	}
	if len(files) != 1 || files[0].FilePath != "nested.txt" {
		t.Errorf("files = %+v, want one add for nested.txt", files)
	}
}

func TestParseApplyPatchFromBash_ApplyPatchTokenInsideStringSkipped(t *testing.T) {
	// `apply_patch` appears only inside a single-quoted argument to
	// echo — not as a command. The AST identifies the invocation as
	// `echo`, so detection must return ok=false.
	cmd := "echo 'apply_patch <<EOF *** Begin Patch *** End Patch EOF'"
	if _, ok := ParseApplyPatchFromBash(cmd); ok {
		t.Error("apply_patch inside echo string wrongly detected as invocation")
	}
}

func TestParseApplyPatchFromBash_MalformedBashRejected(t *testing.T) {
	// A command that fails to parse as Bash (unterminated quote) must
	// be rejected rather than fall back to a less-safe heuristic. The
	// safe default is "no detection" — the post-tool handler will
	// short-circuit to PostToolOK and the policy gets to decide based
	// on other signals.
	cmd := "apply_patch <<'PATCH\n*** Begin Patch\n+x\n*** End Patch\nPATCH"
	if _, ok := ParseApplyPatchFromBash(cmd); ok {
		t.Error("malformed Bash wrongly accepted by detector")
	}
}

func TestParseApplyPatchFromBash_NonHeredocInvocationSkipped(t *testing.T) {
	// `apply_patch` invoked with a file argument (no heredoc) is not
	// the shape Codex uses for in-band patches and we can't extract a
	// body, so it must not be classified as a heredoc invocation. The
	// first-class tool_name="apply_patch" path handles other shapes.
	if _, ok := ParseApplyPatchFromBash("apply_patch ./some-file.patch"); ok {
		t.Error("apply_patch without heredoc wrongly detected as heredoc invocation")
	}
}

// ---------------------------------------------------------------------------
// Structured-construct regression tests. The previous regex+strings.Index
// detector only looked at byte-level neighborhood of `apply_patch`; nested
// invocations were detected by accident. The AST walker visits every Stmt,
// so any control-flow construct that contains an apply_patch heredoc must
// still be detected. These tests pin that behavior so a future refactor
// that limits the walk (e.g. "only top-level Stmts" or "stop at first
// match") can't silently regress detection coverage.

func TestParseApplyPatchFromBash_InvocationInsideIfThen(t *testing.T) {
	cmd := "if true; then apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: ifthen.txt\n" +
		"+x\n" +
		"*** End Patch\n" +
		"PATCH\n" +
		"fi"
	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("if/then-nested apply_patch: returned ok=false, want true")
	}
	if len(files) != 1 || files[0].FilePath != "ifthen.txt" {
		t.Errorf("files = %+v, want one add for ifthen.txt", files)
	}
}

func TestParseApplyPatchFromBash_InvocationInsideFunctionBody(t *testing.T) {
	// A function defined and called in one block. Both the definition
	// body (which is parsed by the AST walker) and the call site live
	// inside the same Bash command. The invocation inside the function
	// body must still trigger detection — otherwise a Codex prompt
	// that wraps its writes in `do_patch(){ ... }; do_patch` would
	// silently bypass the hook. The `}` and `do_patch` lines are
	// formatted on separate lines because Bash requires the heredoc
	// terminator to occupy a line by itself; placing `}` on the same
	// line as `PATCH` would make the closing brace part of the body.
	cmd := "do_patch() {\n" +
		"apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: fnbody.txt\n" +
		"+x\n" +
		"*** End Patch\n" +
		"PATCH\n" +
		"}\n" +
		"do_patch"
	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("function-body apply_patch: returned ok=false, want true")
	}
	if len(files) != 1 || files[0].FilePath != "fnbody.txt" {
		t.Errorf("files = %+v, want one add for fnbody.txt", files)
	}
}

func TestParseApplyPatchFromBash_InvocationAfterAndChain(t *testing.T) {
	// Short-circuit `&&` chaining is the most common shape Codex
	// emits for "do precondition then patch", e.g. `cd repo &&
	// apply_patch <<…`. The walker must descend into BinaryCmd nodes.
	cmd := "cd /tmp && apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: andchain.txt\n" +
		"+x\n" +
		"*** End Patch\n" +
		"PATCH"
	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("&& chained apply_patch: returned ok=false, want true")
	}
	if len(files) != 1 || files[0].FilePath != "andchain.txt" {
		t.Errorf("files = %+v, want one add for andchain.txt", files)
	}
}

func TestParseApplyPatchFromBash_InvocationInPipeline(t *testing.T) {
	// apply_patch as the sink of a pipeline. Unusual (Codex normally
	// feeds the patch through a heredoc, not stdin from another
	// command), but valid Bash: the body still arrives via heredoc on
	// the same Stmt as the apply_patch CallExpr. The walker must
	// descend through the pipeline operator.
	cmd := "echo unused | apply_patch <<'PATCH'\n" +
		"*** Begin Patch\n" +
		"*** Add File: pipeline.txt\n" +
		"+x\n" +
		"*** End Patch\n" +
		"PATCH"
	files, ok := ParseApplyPatchFromBash(cmd)
	if !ok {
		t.Fatal("pipeline-sink apply_patch: returned ok=false, want true")
	}
	if len(files) != 1 || files[0].FilePath != "pipeline.txt" {
		t.Errorf("files = %+v, want one add for pipeline.txt", files)
	}
}

// ---------------------------------------------------------------------------
// False-positive guards on lookalike command names. The implementation
// matches the literal `apply_patch` or the suffix `*/apply_patch` to
// allow absolute paths to per-session shim binaries. Anything else must
// NOT be detected — otherwise a legitimate command like
// `apply_patcher <<EOF ... EOF` (a hypothetical wrapper script) would
// be misclassified and have its body interpreted as a patch envelope.

func TestParseApplyPatchFromBash_LookalikeCommandNamesSkipped(t *testing.T) {
	bodyTail := " <<'X'\n*** Begin Patch\n*** Add File: a.txt\n+x\n*** End Patch\nX"
	lookalikes := []string{
		"apply_patcher",                    // trailing letter
		"apply_patch_v2",                   // trailing token
		"myapply_patch",                    // leading text, no slash separator
		"not_apply_patch",                  // looks like apply_patch but isn't
		"/opt/myapply_patch",               // suffix doesn't start with "/apply_patch"
		"/opt/tools/apply_patch_wrapper",   // path suffix isn't "/apply_patch"
	}
	for _, name := range lookalikes {
		cmd := name + bodyTail
		if files, ok := ParseApplyPatchFromBash(cmd); ok {
			t.Errorf("lookalike %q wrongly detected as apply_patch (files=%+v)", name, files)
		}
	}
}

func TestParseApplyPatchFromBash_AbsolutePathSuffixMatches(t *testing.T) {
	// The complement of the lookalike test above: any absolute or
	// relative path whose final segment is exactly `apply_patch` must
	// be detected. This covers both the Codex per-session shim
	// (`/Users/.../codex-arg0.../apply_patch`) and developer-installed
	// binaries (`/usr/local/bin/apply_patch`, `./apply_patch`).
	paths := []string{
		"/usr/local/bin/apply_patch",
		"./apply_patch",
		"bin/apply_patch",
	}
	for _, p := range paths {
		cmd := p + " <<'PATCH'\n" +
			"*** Begin Patch\n" +
			"*** Add File: a.txt\n" +
			"+x\n" +
			"*** End Patch\n" +
			"PATCH"
		if _, ok := ParseApplyPatchFromBash(cmd); !ok {
			t.Errorf("path %q not detected as apply_patch invocation", p)
		}
	}
}

func TestParseApplyPatchFromBash_DecoyCommandWithApplyPatchArgSkipped(t *testing.T) {
	// `cat` called with arguments that contain the literal string
	// `apply_patch` (filename, comment, etc.) must not trip detection.
	// The AST identifies the call as `cat`, and arguments are not
	// walked as Stmts, so this should naturally be skipped — but the
	// test pins the behavior in case a future refactor starts
	// searching argument text.
	cmds := []string{
		"cat docs/apply_patch_reference.md",
		"ls -la /opt/tools/apply_patch",
		"grep apply_patch /var/log/codex.log",
	}
	for _, c := range cmds {
		if _, ok := ParseApplyPatchFromBash(c); ok {
			t.Errorf("decoy command %q wrongly detected", c)
		}
	}
}
