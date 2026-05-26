# Continuous Integration Testing for the Corridor App

**Status:** Draft
**Author:** Tanuj Vasudeva
**Linear:** [COR-5749 — Build complete e2e tests for CLI + Claude Code](https://linear.app/corridorsecurity/issue/COR-5749/build-complete-e2e-tests-for-cli-claude-code)
**Project:** [Continuous Integration Testing for the Corridor App](https://linear.app/corridorsecurity/project/continuous-integration-testing-for-the-corridor-app-9b6cc143a2ab)
**Cycles in scope:** 34 (May 25 – Jun 8, 2026), 35 (Jun 8 – Jun 22, 2026), with cycle 36 stretch

---

## 1. Context

"Application" here means the functional surface that lives on the user's device and interacts with their coding agents — the CLI binary, the IDE extension, the hooks plugins, and the MCP server registrations they write into Cursor / Claude Code / Codex / Cascade / Droid / Windsurf config trees.

### 1.1 The problem we keep hitting

As we have added agent providers (Cursor → Claude Code → Codex → Cascade → Droid → Windsurf), our mock-heavy unit tests have grown insufficient at modeling real-world install + run + uninstall behavior. Concretely:

- `cli/cmd/install.go` is a 946-LoC orchestrator (`runInstall`) that calls 18 sibling functions. Any one of them silently returning early ships a broken first-install for every new customer (COR-6093, COR-5973).
- We have **two** customer-facing uninstall paths — `cli/cmd/uninstall.go` and `scripts/uninstall.py` — each with its own hand-maintained list of IDE config locations. They are guaranteed to drift (COR-6101).
- We have **5 critical CLI flows** documented in `E2E_TEST_FLOWS.md` as **manually-tested only**. The most consequential — Codex managed-hooks plist merge (COR-6043) — has the property that a single bad release wipes out the MDM-managed Codex hook config across an entire customer fleet, and we would not catch it before merge.
- POV-customer-facing regressions in the last quarter (Etsy: stale CLI from missing auto-update; Eleven Labs: 20s `analyzePlan` calls; Factory: matcher mismatch between embedded plugin and ide-extension provider) all share the same root cause: the unit tests pass against mocked filesystems, the manual smoke is run once at release time, and the regression is observed in production.

### 1.2 What landed in cycle 33 (the foundation we're building on)

Cycle 33 (May 11 – May 25) shipped the test infrastructure that the next two cycles depend on. The relevant PRs are:

| PR | Linear | What it gave us |
|---|---|---|
| corridor#3150 (the original) | COR-4103 | First attempt at full Claude Code E2E. The architectural review of #3150 is what motivated COR-4317's principles (test the right things at the right level, prefer determinism over LLM-coupled tests, use pytest + uv instead of hand-rolled). |
| corridor#4900 | COR-5749 | **Snapshot/diff harness for installation tracking.** Captures the on-disk state of `$HOME` (MCP JSON, hook binaries, rule files, Auth0 token, install marker) before/after each `corridor install` / `corridor uninstall`, then diffs against a checked-in golden snapshot. This is the primitive everything else in this project leans on. |
| corridor#5120 | COR-5749 | **Activated the installation-tracking CI gate** with the first per-target allowlists. Any PR that touches `cli/` now has to either match the golden snapshot or explicitly update the allowlist for the affected target — install-side-effect drift is no longer silent. |
| corridor#5126 | COR-5749 | **Migrated `cli/tests/agent` off the hand-rolled diff onto the harness.** Existing agent tests now share a single diffing engine with the install-tracking gate, so we don't have two implementations of "what changed under `$HOME`?" to maintain. |
| corridor#4901 + #4912 | COR-5857 | **CLI and hooks emit structured JSONL events** to `~/.corridor/debug/*.jsonl`. This replaces free-form debug logs and gives the harness something to assert against ("the `mcp.write` event fired exactly once with target=cursor"). |
| corridor#5082 | COR-5973 | **E2E smoke for `corridor install`** (in design — first end-to-end test using the harness against the real `runInstall`, not unit-mocked sub-pieces). |

### 1.3 What this design doc is and isn't

This doc plans cycles 34 and 35 — what we test, in what order, against which agents, and where each piece of test code lives.

It is **not** a re-litigation of the architecture decisions in COR-4317 (e2e/ mono-project, deterministic-tests-for-deterministic-systems, pinned-version compatibility matrix). Those are settled; this doc adopts them.

---

## 2. Goal

> On some machine, install the Corridor CLI, do a bunch of stuff with it, and uninstall. At each step, test for the correct input and output.

Concretely, by the end of cycle 35 we want a CI job that, on every PR touching `cli/`, `vscode-extension/`, or `e2e/toolchain/`:

1. Builds the CLI binary from source.
2. Spins up a clean `$HOME` and a mock Corridor backend.
3. Runs the install + use + uninstall cycle against **each supported agent** (Claude Code, Cursor, Codex), pinned to a specific agent version.
4. Asserts both:
   - **On-disk side effects** match the snapshot (MCP JSON, hooks, rules, plist for Codex on macOS) via the COR-5749 harness.
   - **Behavioral output** matches expectations (correct JSONL events, correct exit codes, correct stdout banners) via direct event assertions.
5. Fails the build if either set of assertions fails — no flaky-test allowlists, no skip annotations without a linked follow-up ticket.

We are explicitly **not** trying to test Claude Code's / Cursor's / Codex's own behavior — we are testing the seam between those agents and our CLI. Where we need the agent to actually invoke a hook or load an MCP server, we drive the agent's own non-interactive entrypoint (`claude --print …`, Cursor's `cursor-agent`, Codex's `codex exec …`) with a fixed prompt and assert on the JSONL events Corridor emits, not on the LLM output.

---

## 3. Cycle plan

### Cycle 33 (just-completed) — Test infra + Claude Code coverage

Status: **In Review** on parent COR-5749. Carrying the install-tracking gate, JSONL events, and the first install smoke into cycle 34.

Carryover into cycle 34:
- Land the remaining sub-PRs of COR-5749.
- Land COR-5973 (`corridor install` smoke) — already in design, PR corridor#5082.
- Backfill the Claude Code half of the agent matrix below to parity with what cycle 33's harness can already assert.

### Cycle 34 — Cursor coverage + uninstall guarantees

**Theme:** Bring Cursor (CLI + extension) to the same coverage level Claude Code has at the end of cycle 33, and close the highest-impact uninstall regression risk.

| Workstream | Tickets | What "done" looks like |
|---|---|---|
| **CW-1. Cursor install/uninstall snapshot** | new sub-issue of COR-5749, alongside existing harness | Cursor target added to the install-tracking allowlist + snapshot. CI fails on any `cli/cmd/install_cursor.go` or matcher change that drifts from the golden. |
| **CW-2. Extension ↔ CLI delegation E2E** | COR-5845 (extension uses local plugin dir), follow-up of COR-4225 | Pytest under `e2e/toolchain/` that builds the VSIX, invokes `corridor install --target ide-extension` through the extension's `CliProvider`, and asserts the resulting `~/.cursor/mcp.json` + hook wrapper match the golden. Catches the "extension and CLI install paths drift" class of bug (the original motivation for COR-4225). |
| **CW-3. Curl-installer E2E** | COR-6269 (cancelled — re-scope) | Playwright spec under `e2e/tests/cli-curl-installer.spec.ts` that runs the staging install script in a tmpdir-scoped `$HOME`, asserts binary lands at `~/.corridor/bin/corridor`, `corridor --version` is a non-empty semver, `corridor install --target none --no-mcp` exits 0. The cancelled ticket's PR (corridor#5328) had reviewer notes — fold those in. Skip on Windows. |
| **CW-4. Codex managed-hooks merge automation** (macOS) | COR-6043 | Go integration test gated `//go:build darwin` that exercises the actual `defaults read/write` path against a throwaway `com.corridor.test.codex.${RANDOM}` plist key. Wired into `.github/workflows/test-cli-installer.yml` as a `runs-on: macos-latest` job. Removes the manual entry from `E2E_TEST_FLOWS.md`. **Highest single-regression-blast-radius item in the project; do this in cycle 34, not cycle 35.** |
| **CW-5. Auto-update regression guard** | COR-5715 / COR-6330 / COR-6363 | A test that pins an old CLI binary, runs the session-start hook against the mock backend, and asserts the binary at `~/.corridor/bin/corridor` is now the new version. The Etsy POV regression (CLI versions weeks out of date) was exactly this bug class. |
| **CW-6. E2E runtime budget** | replacement for cancelled COR-6124 | The existing e2e CI is too slow to add three more agent matrices to. Re-open with a concrete target (e.g. p95 ≤ 8 minutes) and the proven wins from corridor#5207 (sharding), #5208 (Playwright cache), #5209 (parallelize dep containers). Without this, cycle 35 can't actually add Codex to the matrix without timing out. |

**Cycle 34 explicit non-goals:**
- Codex agent end-to-end (planned for cycle 35).
- VS Code marketplace install of the VSIX (depends on a separate build chain — track in cycle 36 stretch).
- Windows curl-installer parity (track separately; not blocking enterprise rollout).

**Cycle 34 dependency order** (to reduce wasted work if we slip):

```
CW-6 (CI runtime) ──┐
                    ├──► CW-1 (Cursor snapshot) ──► CW-2 (Extension ↔ CLI E2E)
                    ├──► CW-3 (curl installer)
                    └──► CW-4 (Codex plist) ──► CW-5 (auto-update guard)
```

CW-6 is on the critical path because every other workstream lengthens CI; if we don't pay down runtime first, we'll be tempted to mark new tests `runs-on: only when label X is set`, which silently re-creates the "manually run" problem.

### Cycle 35 — Codex coverage + install/uninstall consolidation

**Theme:** Codex is the next provider rollout. Ship the same coverage we now have for Claude Code and Cursor, and close the dual-uninstall-path drift risk while we're touching the install code anyway.

| Workstream | Tickets | What "done" looks like |
|---|---|---|
| **CW-7. Codex install/uninstall snapshot** | new sub-issue of COR-5749 | Codex target added to the install-tracking allowlist + snapshot, including the macOS plist branch (depends on CW-4 having landed the test harness for `defaults read/write`). |
| **CW-8. Codex plugin auto-enable** | COR-5942 (already shipped — add regression test) | Test asserting that after `corridor install --target codex`, the Corridor plugin is **enabled**, not just present. The original bug was "plugin is there but user has to press Enter to install." This is the kind of regression the snapshot diff alone won't catch — needs JSONL-event assertion. |
| **CW-9. Codex E2E with `codex exec`** | new sub-issue, mirrors Claude Code's cycle-33 work | Pytest that drives `codex exec` non-interactively against a fixed prompt that triggers a Corridor hook, asserts the JSONL event sequence (`hook.fire` with the expected matcher). Pinned Codex version in the compatibility matrix. |
| **CW-10. Unify install/uninstall config registry** | COR-6101 | Single source of truth for "list of IDE config locations" shared by `cli/cmd/uninstall.go` and `scripts/uninstall.py`. Add a unit test that fails if the two paths diverge. Lower priority than CW-7/8/9 but can be parallelized; a junior eng can drive this with Cursor in the loop. |
| **CW-11. Declarative `IDETarget` manifest** | COR-6093 | Refactor `cli/cmd/install.go` so adding a new agent is a sub-50-LoC manifest entry rather than a 700-LoC imperative branch. **Stretch** — only attempt if CW-7–10 are tracking ahead of plan. The risk is that a half-finished refactor in cycle 35 destabilizes the same code path we're trying to harden. |
| **CW-12. CI gate becomes blocking** | new sub-issue | Today the install-tracking allowlist gate is advisory (PRs can override). After CW-1, CW-7 land and the false-positive rate has settled, flip it to required. This is the bit that actually prevents the "merged on Friday, regression on Monday" pattern. |

**Cycle 35 explicit non-goals:**
- Cascade, Droid, Windsurf agent E2E. Each gets its own cycle once Codex is stable; we don't have customer pressure to ship them faster than that.
- Browser-side dashboard E2E for the Compliance dashboard (COR-2438 is shipped; its tests live under `e2e/webapp/` and are the responsibility of a different track).

### Cycle 36 — stretch / contingency

If anything from cycles 33–35 slips, cycle 36 is the catch-up window. If everything ships, the stretch work is:

- Cascade + Droid + Windsurf E2E to parity with Claude/Cursor/Codex (likely one cycle each, though much smaller than cycle 33's infra build because the harness is reusable).
- VS Code Marketplace VSIX install path E2E (different build chain — needs `@vscode/test-electron` or equivalent; out of scope until then).
- Windows curl-installer parity.
- A "dogfood telemetry" CI job that spot-checks the JSONL events emitted by every team member's installed CLI on a daily cron. This is the engineering-side complement to the "every Corridork has Corridor installed" cultural commitment from §1.

---

## 4. Architecture decisions (adopted from COR-4317)

These are not new — they were settled in the architectural review of corridor#3150. Repeating them here so the cycle-34/35 PRs don't re-debate them.

1. **Test the right things at the right level.** Most of what corridor#3150 attempted is already covered by `test_cli.py` and `test_hooks.py`. The genuine gap is the install side-effect surface and the install-tracking gate — that is what cycles 33–35 target.
2. **Deterministic tests for deterministic systems.** `corridor install`, the hooks binary, and MCP config writes are all deterministic given a fixed input. Do not couple their tests to non-deterministic Claude/Codex API calls. Where we need to drive an agent, use a fixed prompt and assert on Corridor's JSONL events, not on the LLM's reply.
3. **Established tools, not hand-rolled.** pytest + uv for Python, standard `testing` package for Go, `@playwright/test` for browser. No `PASS`/`FAIL` globals + `assert_true()`. The cycle-33 migration off the hand-rolled diff (corridor#5126) is the template.
4. **e2e/ as a mono-project.**
   ```
   e2e/
   ├── toolchain/   # Python: CLI, hooks, agent integration  (pytest + uv)
   ├── webapp/      # Playwright: frontend + backend browser
   ├── extension/   # VS Code VSIX
   ├── shared/      # Docker images, scripts, CI helpers
   └── README.md
   ```
   Cycles 34/35 work lives under `toolchain/` and `extension/`. The mono-project layout (COR-4322) is a prerequisite — track separately if not yet landed.
5. **Pinned compatibility matrix.** Each agent under test pins a specific version in `e2e/toolchain/conftest.py`. We update the pin in a dedicated PR with a "I read the agent's changelog and nothing relevant changed" reviewer note. We do **not** rely on `npm install -g …` returning a stable version.

---

## 5. Test surface (what each test asserts)

For every (agent, host-OS) pair in `{Claude Code, Cursor, Codex} × {linux, macOS}`, the canonical install/use/uninstall test asserts:

**Install side**
- Exit code 0; recognizable success banner on stdout.
- MCP JSON written at the agent's expected path (`~/.cursor/mcp.json`, `~/.claude/mcp.json`, Codex plist on macOS, etc.) with the correct `corridor` server entry and api-key.
- Hook binaries copied to `~/.corridor/bin/` with the expected file mode.
- Per-target rule file written for platforms that support it (`writeCorridorRule`).
- Auth0 token persisted to `~/.corridor/config.env`.
- Install marker created.
- JSONL event log contains exactly one `install.start`, one `install.complete`, and one `mcp.write` per target, in that order.
- No prompts on stdin (the install path supports non-interactive via env vars; the test reveals any regression there).

**Use side** (one fixed prompt per agent, asserting on Corridor's behavior, not the agent's)
- Driving the agent's non-interactive entrypoint produces the expected sequence of `hook.fire` events with the right matchers.
- The matcher is the agent-specific one — Factory's `.___.*`, Cursor's `mcp__.*|Bash`, Codex's `*` — and the test fails on a regression in either matcher (the COR-5844 bug class).

**Uninstall side**
- Exit code 0 on `corridor uninstall --non-interactive`.
- All on-disk side effects from install are gone.
- Foreign entries (the seeded `/usr/local/bin/user-tool` Codex plist scenario) are preserved — this is the COR-6043 / MDM regression guard.
- JSONL event log contains exactly one `uninstall.start` and one `uninstall.complete`.

**`--target` scoping**
- Re-running with `--target=<single agent>` touches only that agent's files (regression coverage for `filterTargetsByFlag`).

---

## 6. Risks and mitigations

| Risk | Mitigation |
|---|---|
| **CI runtime balloons past the timeout as we add agents.** | CW-6 first. Hard p95 budget enforced in CI. If a new agent matrix would push us over, we shard further or split the workflow file. |
| **Allowlist becomes a rubber-stamp.** Every drift PR adds an allowlist entry, the gate becomes advisory in practice. | CW-12: flip the gate to required after the allowlists stabilize. Require a Linear ticket linked from any allowlist diff in the PR template. |
| **Codex plist test pollutes a developer's real `com.openai.codex`.** | Throwaway plist key namespaced as `com.corridor.test.codex.${RANDOM}`. Cleanup runs on `t.Cleanup` even on failure. Documented in COR-6043. |
| **Pinned agent versions go stale; we ship a regression because the matrix didn't catch a real-world version.** | Quarterly cadence for bumping the pinned versions. The bump PR is its own ticket so the changelog review is the entire PR diff. |
| **The 946-LoC `runInstall` orchestrator (COR-6093) keeps growing during cycle 34/35 because we're adding test hooks.** | Do **not** mix CW-11 (the `IDETarget` refactor) into the test-coverage PRs. The refactor lands in cycle 35 only after coverage is good enough to catch its regressions. |
| **Two-uninstall-path drift (COR-6101) becomes worse during cycle 34.** | CW-10 in cycle 35 closes this. Until then, every PR that adds an IDE config location to one path is required (PR-template checkbox) to add it to the other. |
| **POV-customer Slack reports a regression we should have caught — but the test for it is "in cycle 36 stretch."** | Treat any such report as a cycle-immediate priority that bumps a stretch item out, with a backfilled E2E test as the fix's acceptance criterion. The test-first habit is the cultural shift this project is trying to enforce. |

---

## 7. Out of scope (and why)

- **LLM-quality regression testing.** Whether Claude/Codex generate a "good" response to a Corridor finding is not a CI question; it's an evals question and lives in a different project.
- **Performance benchmarking of `analyzePlan`.** COR-5842 (20s `analyzePlan`) is a backend perf issue, not a CLI install issue. Performance work belongs in the Reliability project.
- **Auth0 sign-in browser flow.** Covered by the dashboard Playwright tests under `e2e/webapp/`. The CLI tests use a pre-provisioned token from the mock backend.
- **Tests for cosmetic CLI output.** First-install banners (COR-4081), warning suppression (COR-6145), and similar are validated by snapshot-of-stdout in the smoke test, not by per-line assertions.

---

## 8. Cultural commitments (from §1.2 of the original draft)

These are not testable in CI but are in scope for this project because they are the only way to catch the bugs CI inherently can't:

1. **Every member on the eng team has Corridor installed and uses it daily.** This is our best defense against the bugs that only show up under "real human in front of an editor" conditions — laggy hooks, confusing rule wording, stale CLI versions sitting on a laptop for two weeks. If you don't like a piece of functionality for your dev workflow, you are likely not alone — flag it, file a ticket, gate the fix behind a feature flag, ship.
2. **Manually-tested-only flows in `E2E_TEST_FLOWS.md` are tech debt.** Every entry in that file is a ticket assigned to someone, with a target cycle. The file shrinks every cycle until it's empty. Cycle 35 ends with at most 1 entry remaining (Windows curl installer, which is intentionally deferred).

---

## 9. Open questions

- **Where do extension VSIX install tests live?** `e2e/extension/` per COR-4317, but the build chain for invoking VS Code in CI (`@vscode/test-electron`) is not yet stood up. Decision needed by mid-cycle 34 if we want VSIX coverage in cycle 35.
- **Do we run the macOS Codex plist job on every PR, or only on PRs touching `cli/cmd/install_codex.go` and friends?** macOS GitHub runners are 10× the cost of Linux. Recommendation: every PR touching `cli/`, with a path-filter to skip when only `e2e/webapp/` changed.
- **Mock backend: shared fixture or per-test?** Cycle 33's harness uses a shared `httptest.NewServer` per process. As we add the agent E2E paths in cycles 34/35 we'll want per-test isolation for some tests (so a misbehaving test can't pollute the next one's `client-settings` payload). Worth a small refactor PR before CW-2 lands.

---

## 10. References

- Linear project: [Continuous Integration Testing for the Corridor App](https://linear.app/corridorsecurity/project/continuous-integration-testing-for-the-corridor-app-9b6cc143a2ab)
- Parent issue: [COR-5749](https://linear.app/corridorsecurity/issue/COR-5749/build-complete-e2e-tests-for-cli-claude-code)
- Architecture principles: [COR-4317 — Improved E2E testing framework](https://linear.app/corridorsecurity/issue/COR-4317/improved-e2e-testing-framework)
- Mono-project layout: [COR-4322](https://linear.app/corridorsecurity/issue/COR-4322/reorganize-e2e-into-mono-project-structure)
- Manually-tested flow inventory: `E2E_TEST_FLOWS.md` (corridor monorepo)
- High-leverage individual gaps: [COR-5973](https://linear.app/corridorsecurity/issue/COR-5973/) (`corridor install` smoke), [COR-6043](https://linear.app/corridorsecurity/issue/COR-6043/) (Codex plist merge), [COR-6093](https://linear.app/corridorsecurity/issue/COR-6093/) (`install.go` declarative refactor), [COR-6101](https://linear.app/corridorsecurity/issue/COR-6101/) (uninstall-path unification)
- Cycle-33 foundation PRs: corridor#3150, #4900, #5120, #5126, #4901, #4912, #5082
