# Wiki Log

Append-only timeline of wiki maintenance activity.

## [2026-05-13] plan | 15 — Coverage Ramp plan (50% → 75%)

Created `wiki/15_coverage_ramp.md` — 3-stage plan targeting 75% overall coverage.
Stage 1 targets `internal/daemon` (3.3%), `cmd/korego` (0%), `pkg/client` (44.9%),
and `pkg/daemon` (5.9%) to reach 60%. Stage 2 closes the `run()` gap across 24
utilities via dispatch-call tests with `testdata/` fixtures to reach 68%. Stage 3
refactors `run()` signatures to accept interfaces, pushing to 75%. Includes per-package
coverage targets, test fixture conventions, and verification commands.

## [2026-05-13] implement | 12.3b — Coverage push 46.2% → 50.0%

Added/enhanced unit tests across 15+ packages: sed (31.6%→49.1% — parser/compiler/
BRE tests), diff (42.8%→52.9% — Diff, normalizeSpace, filterBlankLines, edge cases),
grep (16.0%→16.5% — regex/invert/line-regex bounds), find (buildExecArgs),
dirname/basename/hostname/whoami/uname (run() CLI tests), dispatch (ListAll empty).

Overall coverage: **41.6% → 50.0%** (+8.4%). All 57 packages pass. CI enforces ≥45%.

## [2026-05-13] implement | 12.3 — Coverage gate enforcement + unit test expansion

Changed CI coverage step from `::warning::` (exits 0) to hard failure at 45% (`exit 1`).
Added/enhanced unit tests across 9 packages: head (11.2%→29.0%), tail (10.3%→27.6%),
grep (12.4%→16.0%), cat (21.8%→37.6%), wc (15.8%→32.5%), echo (20.8%→56.9%),
sort (23.4%→58.0%), uniq (37.2%→higher), cut (56.2%→higher), touch (20.0%→higher).

Overall coverage: **41.6% → 46.2%**. All 57 packages pass. CI enforces ≥45%.

Updated wiki: 11a_lower_priority.md (status, CI gate, remaining work table),
11_post_mvp_priorities.md (status, remaining work table).

## [2026-05-13] milestone | Promote agent architecture to Phase 14

Renamed `wiki/agent_architecture.md` → `wiki/14_agent_architecture.md`. Added to
wiki index and phases.md as Phase 14 (status: DESIGN). Updated ARCHITECTURE.md to
acknowledge pkg/agent and go-git dependency as planned additions.

## [2026-05-13] design | Agent Architecture design document

Created `wiki/agent_architecture.md` — a detailed design for an autonomous coding
agent compiled into the KoreGo binary. Covers: ReAct agent loop, go-git integration,
LLM provider interface (OpenAI / Anthropic / local), CLI + JSON-RPC dual interface,
workspace management, security model, Docker compose integration, and state machine.
No code changes; design-only phase. (Later promoted to Phase 14.)

## [2026-05-12] update | Document .goreleaser.yml file location convention

Added explanation to `09_release_docs.md` for why `.goreleaser.yml` lives at the repo
root rather than `.github/`: GoReleaser is a tool-level config, not a GitHub-specific
feature. The root is its conventional default location.

## [2026-05-12] annotate | Add source links to utility docs

Added `[pkg/<name>/](../pkg/<name>/)` links to every utility header in phase
pages (01, 03, 04, 06, 07). Also linked infrastructure packages in phases 00
and 05. All 55 utility packages now have clickable source links from their
wiki documentation.

## [2026-05-13] consolidate | Merge redundant wiki content, resolve inconsistencies

Consolidated awk content: 07a_awk.md is now the canonical awk document. Removed
duplicated task lists from 11_post_mvp_priorities.md (11.4), 12_road_to_gold.md
(12.5), and 13_code_audit.md (13.5) — all now point to 07a_awk.md with short
descriptions. Added cross-cutting deliverables (schema, BusyBox, posix_coverage,
README update) to 07a_awk.md.

Resolved shell security contradictions: 08_hardening.md claimed sandbox complete
but code had hardcoded timeout and no tests/docs. Added cross-references between
08 (design done), 11a.3 (deferred), and 12.2 (remaining work). Updated 12.2 with
code audit finding (KOREGO_SHELL_TIMEOUT not env-driven).

Resolved coverage gate inconsistency: 11a.4 claimed gate was done but 12.3/13.3
showed it's informational only (warns, exits 0). Clarified 11a.4 as "step added,
informational only; hard-fail tracked in 12.3."

Merged 13_code_audit.md execution plan into 12_road_to_gold.md. Added macOS build
breakage (13.0) as 12.0 in the gap analysis. 13 now focuses on code evidence and
wiki discrepancies, not task planning.

Updated 8 wiki files: index.md, phases.md (v2.2), 07a_awk.md, 11_post_mvp_priorities.md,
11a_lower_priority.md, 12_road_to_gold.md, 13_code_audit.md, 08_hardening.md.

Gold formula: 12.0–12.4 = Gold, +12.5 = Platinum.

## [2026-05-13] implement | 12.0 — macOS build fix

Split pkg/uname/uname.go and pkg/stat/stat.go into platform-specific files:
- uname_linux.go (syscall.Uname, [65]int8 fields)
- uname_darwin.go (unix.Uname, [256]byte fields, separate bytesToString helper)
- stat_linux.go (sys.Atim/sys.Ctim from syscall.Timespec)
- stat_darwin.go (sys.Atimespec/sys.Ctimespec)

Verified: GOOS=darwin CGO_ENABLED=0 go build ./... exits 0. Full test suite passes.

## [2026-05-13] implement | 12.2 — Shell security model

Wired KOREGO_SHELL_TIMEOUT env var in internal/shell/interpreter.go (was hardcoded
30s). Created internal/shell/interpreter_test.go with 10 tests: TestExecBasic,
TestTimeout (×2 env var scenarios), TestOutputWithinLimits, TestPathEscape,
TestPathEscapeBlocked (via shell redirection), TestEnvVarInjection, TestStderrCapture,
TestNonZeroExit, TestSyntaxError. Created docs/SECURITY.md: trust model, accessible
resources, resource limits, RPC-level protections, deployment posture, artifact
verification.

## [2026-05-13] implement | 12.4 — Fix BusyBox CI/local discrepancy

Old-style BusyBox tests called 'busybox <applet>' which resolved to system BusyBox
on CI (Ubuntu), not KoreGo — inflating the pass rate from 83.5% real to 97.9% fake.
Fixed test/busybox_testsuite/runtest: added global BBDIR temp directory with
busybox→korego symlink prepended to PATH. Removed per-applet case block (tar/gzip).
Simplified old-style test PATH to a single shared block. Updated CI baseline in
.github/workflows/ci.yml from 479 to 409. Updated todos.md discrepancy note to
RESOLVED. True baseline: 409 passed, 71 failed, 10 skipped (83.5%).

## [2026-05-13] implement | 12.1 — Supply chain security

Added sboms: stanza to .goreleaser.yml (archive + binary). Added Cosign keyless
signing (OIDC) to release.yml with id-token: write permission. Added SLSA Level 3
provenance job via slsa-framework/slsa-github-generator@v2.1.0. Added Trivy
vulnerability scan to ci.yml (CRITICAL,HIGH severity, exits 1). Updated
docs/SECURITY.md artifact verification section with cosign verify, SBOM inspect,
and slsa-verifier commands. Updated 09_release_docs.md with supply chain section
(09.2b) and milestone items.

## [2026-05-13] update | Phase 11a complete, Gold 4/5 resolved

11a_lower_priority.md milestone: 8/8 complete (11a.3 → 12.2, 11a.4 → 12.4,
11a.7 → 12.1). 12_road_to_gold.md: 4/5 Gold gaps resolved, only 12.3 (coverage
gate) remains. 13_code_audit.md: 4/6 fixed. phases.md status updated. All
deferred work from 11a now resolved via Phase 12.
