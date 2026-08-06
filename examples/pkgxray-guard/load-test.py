#!/usr/bin/env python3
"""Load / throughput harness for the pkgxray-guard hook.

Drives the compiled hook binary with real Claude PreToolUse payloads on stdin at
varying concurrency, and reports throughput + latency. High-concurrency benches
use offline paths (git-URL / non-install → no pkgxray call) or a fake pkgxray, so
this never hammers npm / OSV / GitHub. Real pkgxray is characterized at low
volume only (needs `pkgxray` on PATH; skipped otherwise).

Usage:  python3 load-test.py        # run from anywhere in the hookshot repo
"""
import json, os, subprocess, sys, tempfile, time
from collections import Counter
from concurrent.futures import ThreadPoolExecutor

ROOT = subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True,
                      text=True).stdout.strip() or "."
TMP = tempfile.mkdtemp()
BIN = os.path.join(TMP, "pkgxray-guard")
FAKE = os.path.join(TMP, "fake-pkgxray")

print("building hook binary…")
if subprocess.run(["go", "build", "-o", BIN, "./examples/pkgxray-guard"], cwd=ROOT).returncode:
    sys.exit("build failed")
with open(FAKE, "w") as f:
    f.write('#!/bin/sh\necho \'{"decision":"allow","report":{"summary":"ok (fake)","findings":[]}}\'\n')
os.chmod(FAKE, 0o755)

def payload(cmd):
    return json.dumps({"session_id": "lt", "cwd": "/tmp", "hook_event_name": "PreToolUse",
                       "tool_name": "Bash", "tool_input": {"command": cmd}, "tool_use_id": "t"})

def run_once(cmd, policy, pkgxray_bin=None):
    env = dict(os.environ, PKGXRAY_HOOK_POLICY=policy)
    if pkgxray_bin:
        env["PKGXRAY_BIN"] = pkgxray_bin
    t0 = time.perf_counter()
    p = subprocess.run([BIN, "claude-pre-tool-use"], input=payload(cmd),
                       capture_output=True, text=True, env=env, timeout=120)
    dt = time.perf_counter() - t0
    try:
        return dt, json.loads(p.stdout)["hookSpecificOutput"]["permissionDecision"]
    except Exception:
        return dt, f"<parse-fail:{p.stdout[:60]!r}>"

def bench(name, n, conc, cmd, policy, pkgxray_bin=None, want=None):
    lat, decs = [], []
    t0 = time.perf_counter()
    with ThreadPoolExecutor(max_workers=conc) as ex:
        for dt, dec in ex.map(lambda _: run_once(cmd, policy, pkgxray_bin), range(n)):
            lat.append(dt); decs.append(dec)
    wall = time.perf_counter() - t0
    lat.sort()
    pct = lambda q: lat[min(len(lat) - 1, int(q * len(lat)))] * 1000
    print(f"\n[{name}]  n={n} concurrency={conc}")
    print(f"  throughput : {n / wall:8.1f} req/s   (wall {wall:.2f}s)")
    print(f"  latency ms : p50={pct(.50):.1f}  p95={pct(.95):.1f}  max={lat[-1]*1000:.1f}")
    if want:
        ok = sum(1 for d in decs if d == want)
        print(f"  correctness: {ok}/{n} == {want}" + ("  ✓" if ok == n else "  ✗ MISMATCH"))
    else:
        print(f"  decisions  : {dict(Counter(decs))}")

print("=" * 64 + "\nA. Offline hook-overhead ceiling (no pkgxray call)")
bench("git-URL install → deny (strict)", 300, 16,
      "npm install git+https://github.com/evil/pkg.git", "strict", want="deny")
bench("non-install → allow", 300, 16, "ls -la", "balanced", want="allow")

print("\n" + "=" * 64 + "\nB. Full pipeline with FAKE pkgxray (subprocess spawn, no network)")
bench("registry install, 1 pkg", 300, 16, "npm install left-pad", "balanced",
      pkgxray_bin=FAKE, want="allow")

# A slow fake (50 ms/call) makes the per-package cost dominate so the multi-package
# concurrency (PKGXRAY_HOOK_CONCURRENCY) shows cleanly. Measured on ONE invocation
# at a time (concurrency=1 outer) so the internal parallelism isn't masked by CPU
# saturation from many binaries × many workers.
SLOW = os.path.join(TMP, "slow-fake-pkgxray")
with open(SLOW, "w") as f:
    f.write('#!/bin/sh\nsleep 0.05\necho \'{"decision":"allow","report":{"summary":"ok","findings":[]}}\'\n')
os.chmod(SLOW, 0o755)
many = "npm install " + " ".join(f"pkg{i}" for i in range(20))
print("\n" + "=" * 64 + "\nC. Multi-package concurrency — ONE `npm i <20 pkgs>` (slow fake, 50 ms/pkg)")
for w in (1, 4, 8, 16):
    os.environ["PKGXRAY_HOOK_CONCURRENCY"] = str(w)
    dt, dec = run_once(many, "balanced", pkgxray_bin=SLOW)
    tag = " (serial baseline)" if w == 1 else " (default)" if w == 8 else ""
    print(f"  PKGXRAY_HOOK_CONCURRENCY={w:<2}: {dt*1000:7.1f} ms  → {dec}{tag}")
os.environ.pop("PKGXRAY_HOOK_CONCURRENCY", None)

CLI = os.environ.get("PKGXRAY_BIN", "pkgxray")
if any(os.access(os.path.join(p, CLI), os.X_OK) for p in os.environ.get("PATH", "").split(":")):
    print("\n" + "=" * 64 + "\nD. Real pkgxray — LOW volume (network-polite), cold vs warm")
    for i in range(6):
        dt, dec = run_once("npm install left-pad", "balanced")
        print(f"  call {i+1} ({'cold' if i == 0 else 'warm'}): {dt*1000:7.1f} ms  → {dec}")
else:
    print(f"\nC. skipped — '{CLI}' not on PATH")
