#!/usr/bin/env bash
#
# e2e-smoke.sh — end-to-end smoke test for the pkgxray-guard hook.
#
# Exercises the REAL chain the unit tests can't: a genuine Claude Code
# PreToolUse payload on stdin → the compiled hook binary → (for registry
# installs) the real pkgxray CLI → an allow/ask/deny decision. The unit tests
# use a fake pkgxray; this proves the binary and the CLI actually integrate.
#
# Deterministic cases run offline (git-URL/non-install/local specs never call
# pkgxray) and are hard assertions. One live case runs `npm install <pkg>`
# through the real pkgxray CLI and checks that a valid decision comes back — the
# exact verdict depends on the registry, so it isn't pinned.
#
# Usage:  ./e2e-smoke.sh
# Env:    PKGXRAY_BIN  the pkgxray CLI the hook shells out to (default: pkgxray)

set -uo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo .)"
BIN="$(mktemp -d)/pkgxray-guard"
PASS=0
FAIL=0

echo "building hook binary…"
( cd "$REPO_ROOT" && go build -o "$BIN" ./examples/pkgxray-guard ) || { echo "build failed"; exit 1; }

# drive <policy> <command>  → prints the hook's Claude decision JSON
drive() {
  printf '{"session_id":"e2e","cwd":"/tmp","hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"%s"},"tool_use_id":"t1"}' "$2" \
    | PKGXRAY_HOOK_POLICY="$1" "$BIN" claude-pre-tool-use
}

# assert <name> <expected-decision> <policy> <command>
assert() {
  local name="$1" want="$2" out
  out="$(drive "$3" "$4")"
  if printf '%s' "$out" | grep -q "\"permissionDecision\":\"$want\""; then
    echo "  PASS: $name → $want"; PASS=$((PASS + 1))
  else
    echo "  FAIL: $name (want $want)"; echo "    got: $out"; FAIL=$((FAIL + 1))
  fi
}

echo "=== deterministic cases (offline, no pkgxray call) ==="
assert "git-URL install under strict blocks"  deny  strict   "npm install git+https://github.com/evil/pkg.git"
assert "git-URL install under balanced asks"  ask   balanced "npm install git+https://github.com/evil/pkg.git"
assert "non-install command passes"           allow balanced "ls -la"
assert "local path install is skipped"        allow balanced "npm install ./local-tarball.tgz"

echo "=== live case (real pkgxray) ==="
CLI="${PKGXRAY_BIN:-pkgxray}"
if command -v "$CLI" >/dev/null 2>&1; then
  out="$(drive balanced "npm install left-pad")"
  if printf '%s' "$out" | grep -qE "\"permissionDecision\":\"(allow|ask|deny)\""; then
    verdict="$(printf '%s' "$out" | grep -oE '"permissionDecision":"[a-z]+"' | head -1)"
    echo "  PASS: real pkgxray returned a decision ($verdict)"; PASS=$((PASS + 1))
  else
    echo "  FAIL: no valid decision from real pkgxray"; echo "    got: $out"; FAIL=$((FAIL + 1))
  fi
else
  echo "  SKIP: '$CLI' not on PATH — deterministic cases still ran (set PKGXRAY_BIN to enable)"
fi

echo ""
echo "e2e: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
