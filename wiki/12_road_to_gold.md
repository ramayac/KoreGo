# Phase 12 — Road to Gold

> **Status:** Planning | **Date:** 2026-05-13 | **Depends on:** All Phases 00–11 complete

---

## Context

KoreGo MVP and all post-MVP cleanup phases (00–11a) are complete or near-complete.
The project is assessed at **Strong Silver / borderline Gold**.

This phase documents the specific gaps preventing a Gold rating and the tasks required
to close them, ordered by impact. Completing items 12.1–12.4 achieves Gold. Item 12.5
(`awk`) pushes toward Platinum.

### Current Rating Summary

| Tier | Verdict |
|------|---------|
| Bronze | ✅ Cleared |
| Silver | ✅ Cleared |
| **Gold** | ⚠️ Borderline — 4 blocking gaps |
| Platinum | ❌ Requires awk + Gold first |

---

## Gap Analysis

| # | Gap | Risk | Impact |
|---|-----|------|--------|
| 12.1 | No supply chain security (SBOM, Cosign, SLSA, scanning) | High — infra tooling supply chain attacks are common | Release trust |
| 12.2 | Shell security model undocumented | High — tool explicitly targets AI agents with untrusted input | Adoption blocker |
| 12.3 | Coverage gate is informational (warns at 50%, never fails) | Medium — a patch can drop coverage to 0% and pass CI | Code quality |
| 12.4 | CI/local BusyBox test discrepancy | Medium — old-style tests pass via system BusyBox on CI, not KoreGo | Test reliability |
| 12.5 | `awk` not implemented | Low (MVP scope excluded it) — qualifies the "100% POSIX" claim | Compliance claim |

---

## 12.1 — Supply Chain Security

**Why it matters:** KoreGo is distributed as a container image and binary. Without provenance
and signing, consumers cannot verify artifacts were built from the declared source. This is
increasingly table stakes for infrastructure tooling.

**Current state:** GoReleaser produces multi-arch binaries and pushes to GHCR. No SBOM,
no image signing, no vulnerability scanning, no provenance attestation.

### Tasks

- [ ] Add SBOM generation to `.goreleaser.yml`:
  ```yaml
  sboms:
    - artifacts: archive
    - artifacts: binary
  ```
- [ ] Add container image signing via Cosign in `.github/workflows/release.yml`:
  ```yaml
  - uses: sigstore/cosign-installer@v3
  - run: cosign sign --yes ghcr.io/ramayac/korego:${{ github.ref_name }}
  ```
- [ ] Add SLSA Level 3 provenance via `slsa-framework/slsa-github-generator` in release workflow
- [ ] Add vulnerability scanning step in CI (`ci.yml`) after Docker build:
  ```yaml
  - uses: aquasecurity/trivy-action@master
    with:
      image-ref: korego:dev
      exit-code: 1
      severity: CRITICAL,HIGH
  ```
- [ ] Update `docs/SECURITY.md` with artifact verification instructions for end users

### Acceptance

```bash
# Verify a release image is signed
cosign verify ghcr.io/ramayac/korego:latest --certificate-identity-regexp='.*' --certificate-oidc-issuer='https://token.actions.githubusercontent.com'

# Inspect SBOM
docker buildx imagetools inspect ghcr.io/ramayac/korego:latest --format '{{ json .SBOM }}'
```

---

## 12.2 — Shell Security Model

**Why it matters:** `korego.shell.exec` is the highest-risk surface in the entire codebase.
The daemon is explicitly positioned for AI agent use — agents may pass partially
untrusted or model-generated shell scripts. Without a documented and tested security
contract, operators cannot safely expose the daemon.

**Current state:** `internal/shell/interpreter.go` wraps `mvdan.cc/sh` with a 30s timeout
(configurable via `KOREGO_SHELL_TIMEOUT`) and a memory limit. No tests enforce these limits.
No documentation states what is and is not accessible from a shell script.

### Tasks

- [ ] Write `docs/SECURITY.md` defining the security model:
  - Trust level: "trusted input only" vs "safe for untrusted input" — decide and state explicitly
  - What can a script access? (filesystem paths, env vars, network)
  - What resource limits are enforced? (timeout, memory, CPU)
  - What syscalls/operations are blocked (if any)?
  - Recommended deployment posture (socket permissions, daemon user)
- [ ] Add tests in `internal/shell/interpreter_test.go`:
  - `sleep 60` is killed before the timeout elapses
  - A script that attempts to read outside a configured root path fails (if sandboxing is implemented)
  - Memory-hungry script hits the memory limit and is terminated
- [ ] Verify `KOREGO_SHELL_TIMEOUT` default (30s) is documented in README and `docs/RPC_API.md`
- [ ] Consider whether `shell.exec` should require an explicit session (preventing stateless, fire-and-forget abuse)

### Acceptance

```bash
# Timeout is enforced
echo '{"jsonrpc":"2.0","method":"shell.exec","params":{"script":"sleep 60"},"id":1}' \
  | nc -U /tmp/korego.sock
# → must return error within KOREGO_SHELL_TIMEOUT seconds, not hang

go test ./internal/shell/... -v -run TestTimeout
go test ./internal/shell/... -v -run TestResourceLimits
```

---

## 12.3 — Enforce Coverage Gate

**Why it matters:** The current CI coverage step is purely informational — it prints a
warning at <50% but never fails the build. A commit that drops coverage to 0% passes CI
undetected. This defeats the purpose of the gate.

**Current state:** `.github/workflows/ci.yml` — coverage step uses `::warning::` and exits 0.
Overall coverage is around 50–60% (exact number varies per package).

### Tasks

- [ ] Change coverage check in `ci.yml` from `::warning::` to `::error:: exit 1` at the threshold:
  ```bash
  if (( $(echo "$COVERED < 60" | bc -l) )); then
    echo "::error::Coverage ${COVERED}% is below 60% threshold"
    exit 1
  fi
  ```
- [ ] Raise threshold to 60% (current state already meets this if the suite is passing)
- [ ] Add per-package coverage reporting to `make cover-pct` output so regressions are visible
- [ ] Document the coverage policy in `AGENTS.md` so contributors know the bar

### Acceptance

```bash
# Fails CI if coverage drops below threshold
make cover-pct   # must show ≥60% overall
```

---

## 12.4 — Fix BusyBox CI / Local Discrepancy

**Why it matters:** Old-style BusyBox tests (in `test/busybox_testsuite/<applet>/`) call
`busybox <applet>` which resolves to the **system BusyBox** on CI (Ubuntu), not KoreGo.
This means CI passes those tests regardless of KoreGo's behavior — a silent blind spot.

**Current state:** CI shows 100% on old-style tests because `/usr/bin/busybox` is used.
Locally, if a `busybox` symlink in `$LINKSDIR` points to KoreGo, the same tests may fail.
The `todos.md` documents this but it is not resolved.

### Tasks

- [ ] Audit which old-style tests exist and whether they cover functionality not in `.tests` files
- [ ] Option A (preferred): Convert remaining old-style tests to `.tests` format so they route to KoreGo on all platforms
- [ ] Option B (fallback): In CI, remove or shadow `/usr/bin/busybox` so old-style tests always exercise KoreGo
- [ ] Update CI baseline counter to reflect true KoreGo pass rate (not system BusyBox pass rate)
- [ ] Document the resolution in `todos.md` and close the discrepancy note

### Acceptance

```bash
# Local and CI results must be the same pass count
make testsuite 2>&1 | tail -5
# Confirmed: same output as CI
```

---

## 12.5 — `awk` Implementation (Platinum Gate)

**Why it matters:** `awk` is the last missing POSIX.2 utility. Without it, the project's
"POSIX-compliant userland" claim carries a permanent asterisk. Every serious shell script
that processes structured text uses awk. The implementation plan already exists in
[07a_awk.md](07a_awk.md).

**Current state:** Deferred since MVP. No `pkg/awk/` exists. All other 49 target utilities
are implemented.

### Tasks

> Full implementation plan is in [07a_awk.md](07a_awk.md). This task tracks execution.

- [ ] Implement `pkg/awk/awk.go` per the plan in [07a_awk.md](07a_awk.md)
- [ ] Register in multicall dispatch (`cmd/korego/main.go`)
- [ ] `--json` output: array of per-record results with matched fields
- [ ] BusyBox test suite `awk` tests must pass
- [ ] Unit tests: ≥20 cases covering patterns, fields, `BEGIN`/`END`, built-ins (`NR`, `NF`, `FS`, `OFS`)
- [ ] Add `awk.schema.json` in `test/schemas/` and `docs/schemas/`
- [ ] Update `posix_coverage.md`: change awk from ❌ to ✅
- [ ] Update README status table

### Acceptance

```bash
echo -e "a 1\nb 2\nc 3" | ./korego awk '{print $2}'
# → 1\n2\n3

./korego awk --json '{sum += $1} END {print sum}' numbers.txt
# → {"output":"42\n","exitCode":0,"schemaVersion":"1.0"}

go test ./pkg/awk/... -v
make testsuite   # awk tests now in passed count
```

---

## Milestone 12 — Gold Checklist

```
[ ] 12.1 — SBOM + Cosign + SLSA + trivy in release/CI pipeline
[ ] 12.2 — Shell security model documented + timeout/resource tests passing
[ ] 12.3 — Coverage gate hard-fails CI at ≥60%
[ ] 12.4 — CI/local BusyBox discrepancy resolved; baselines match
[ ] 12.5 — awk implemented, BusyBox awk tests pass (Platinum gate)
```

Completing 12.1–12.4 = **Gold**.
Completing 12.1–12.5 = **Platinum**.

---

## How to Verify

```bash
# Supply chain
cosign verify ghcr.io/ramayac/korego:latest ...
docker buildx imagetools inspect ghcr.io/ramayac/korego:latest --format '{{ json .SBOM }}'

# Shell security
go test ./internal/shell/... -v -run TestTimeout
go test ./internal/shell/... -v -run TestResourceLimits

# Coverage gate
make cover-pct   # ≥60% required

# BusyBox parity
make testsuite   # same result locally and in CI

# awk (Platinum)
echo "hello world" | ./korego awk '{print $1}'
go test ./pkg/awk/... -v
```
