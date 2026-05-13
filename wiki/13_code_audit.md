# Phase 13 — Code Audit: Validated Gap Analysis

> **Status:** Review | **Date:** 2026-05-13 | **Method:** Code-first exploration, then validated against wiki/12_road_to_gold.md

---

## Context

This document records the findings of a code-first audit conducted against the live repository.
The approach was: explore code → validate against the wiki → report discrepancies.

The wiki's `12_road_to_gold.md` is accurate in its gap identification. This document adds
code-level specifics, line-number references, and two gaps not captured in the wiki.

**Overall coverage at time of audit: 41.6%** (`go test -coverprofile` across `./pkg/...` `./internal/...`)

---

## Summary

| # | Gap | In Wiki? | Code Evidence | Effort |
|---|-----|----------|---------------|--------|
| 13.0 | macOS build breakage (`uname`, `stat`, `client`) | No | `syscall.Utsname`, `sys.Atim` undefined on darwin | 2h |
| 13.1 | No supply chain security (SBOM, Cosign, SLSA, Trivy) | Yes (12.1) | `.goreleaser.yml` has no `sboms:`; `release.yml` has no signing step | 1d |
| 13.2 | Shell security model undocumented, untested | Yes (12.2) | `interpreter_test.go` does not exist; `KOREGO_SHELL_TIMEOUT` hardcoded, not env-driven | 1d |
| 13.3 | Coverage gate is informational only; real coverage 41.6% | Yes (12.3) | `ci.yml` emits `::warning::`, exits 0; worst packages at 3–14% | 3d |
| 13.4 | BusyBox CI/local discrepancy | Yes (12.4) | `runtest` creates KoreGo symlink only for `tar`/`gzip`; all other old-style tests hit system BusyBox | 1–2d |
| 13.5 | `awk` not implemented (Platinum gate) | Yes (12.5) | `pkg/awk/` does not exist | 5–10d |

---

## 13.0 — macOS Build Breakage (NOT in wiki — discovered during audit)

**Severity: Blocks local development on macOS.**

Three packages fail to build on `GOOS=darwin`:

### `pkg/uname/uname.go`
```
pkg/uname/uname.go:50:16: undefined: syscall.Utsname
pkg/uname/uname.go:51:20: undefined: syscall.Uname
```
`syscall.Utsname` and `syscall.Uname()` are Linux-only. The Darwin equivalent is
`golang.org/x/sys/unix.Uname`. The file needs a `//go:build linux` tag with a Darwin
stub alongside it.

### `pkg/stat/stat.go:58`
```
pkg/stat/stat.go:58:28: sys.Atim undefined (type *syscall.Stat_t has no field or method Atim)
pkg/stat/stat.go:59:28: sys.Ctim undefined (type *syscall.Stat_t has no field or method Ctim)
```
Linux `syscall.Stat_t` uses `Atim`/`Ctim` (`syscall.Timespec`); Darwin uses `Atimespec`/`Ctimespec`.
Fix: build tags splitting `stat_linux.go` and `stat_darwin.go` for the `Sys()` cast.

### `pkg/client`
Build fails as a dependency of the above. Resolves once `uname` and `stat` are fixed.

**Fix template for `uname`:**
```
pkg/uname/
  uname.go          # shared types, flag spec, CLI layer
  uname_linux.go    # //go:build linux — syscall.Uname impl
  uname_darwin.go   # //go:build darwin — golang.org/x/sys/unix.Uname impl
```

---

## 13.1 — Supply Chain Security

**Confirmed: zero implementation.**

### `.goreleaser.yml` — no SBOM
Current file has `builds`, `dockers`, `docker_manifests`, `changelog`. No `sboms:` stanza.

Required addition:
```yaml
sboms:
  - artifacts: archive
  - artifacts: binary
```

### `.github/workflows/release.yml` — no signing, no provenance, no scanning
Current steps: `checkout` → `setup-go` → `setup-qemu` → `setup-buildx` → `docker/login-action` → `goreleaser`.

Missing:
- `sigstore/cosign-installer@v3` + `cosign sign` after image push
- `slsa-framework/slsa-github-generator` for SLSA Level 3 provenance
- `aquasecurity/trivy-action` in `ci.yml` after `make docker`

### `docs/SECURITY.md` — does not exist
No instructions for end users to verify artifact signatures or inspect SBOMs.

---

## 13.2 — Shell Security Model

**Confirmed: partial implementation, zero tests, no docs.**

### What exists (`internal/shell/interpreter.go`)
- `context.WithTimeout(context.Background(), 30*time.Second)` — timeout is enforced
- `common.LimitWriter{Limit: 128 * 1024 * 1024}` — 128 MB per stream cap
- `openHandler` calls `common.SecurePath(path, base)` — path confinement via cwd

### What is missing

**`KOREGO_SHELL_TIMEOUT` is hardcoded, not env-driven.** The wiki states it is configurable;
the code does not read any environment variable. The timeout is always 30 seconds.

Fix (single line):
```go
timeout := 30 * time.Second
if s := os.Getenv("KOREGO_SHELL_TIMEOUT"); s != "" {
    if d, err := time.ParseDuration(s); err == nil {
        timeout = d
    }
}
ctx, cancel := context.WithTimeout(context.Background(), timeout)
```

**`internal/shell/interpreter_test.go` does not exist.** The wiki requires:
- `TestTimeout` — `sleep 60` must error within the timeout
- `TestResourceLimits` — output-heavy script must be truncated by `LimitWriter`
- `TestPathEscape` — script attempting to open `../../../etc/passwd` must be blocked

**`docs/SECURITY.md` does not exist.** Operators have no written contract for safe deployment.
Required content: trust model, accessible resources (filesystem, env, network), enforced limits,
socket permission recommendations.

---

## 13.3 — Coverage Gate

**Confirmed: gate exits 0 regardless of coverage. Actual coverage is 41.6%.**

### The gate (`ci.yml` coverage step)
```bash
if (( $(echo "$COVERED < 50" | bc -l) )); then
  echo "::warning::Coverage ${COVERED}% is below 50% threshold"
fi
# ← no exit 1
```
A commit that drops coverage to 0% passes CI. The warning is invisible in the default
GitHub Actions UI unless the step is expanded.

### Actual per-package coverage

| Package | Coverage |
|---------|---------|
| `internal/daemon` | 3.3% |
| `pkg/tee` | 3.3% |
| `pkg/tail` | 10.3% |
| `pkg/head` | 11.2% |
| `pkg/grep` | 12.4% |
| `pkg/dirname` | 14.3% |
| `pkg/echo` | 20.8% |
| `pkg/touch` | 20.0% |
| `pkg/basename` | 23.8% |
| `pkg/sort` | 23.4% |
| `pkg/whoami` | 26.3% |
| `pkg/pwd` | 30.4% |
| `pkg/sed` | 31.6% |
| `pkg/rm` | 35.3% |
| ... | ... |
| `pkg/chown` | 89.7% |
| `pkg/printf` | 89.6% |
| `pkg/truefalse` | 100.0% |
| **Overall** | **41.6%** |

### Minimum fix (gate enforcement)
```bash
# ci.yml — change warning to hard failure
if (( $(echo "$COVERED < 60" | bc -l) )); then
  echo "::error::Coverage ${COVERED}% is below 60% threshold"
  exit 1
fi
```

Note: at 41.6% current, enforcing 60% immediately would break CI. A staged approach:
enforce 45% now, raise to 60% after targeted test additions (see tasks below).

### Highest-ROI packages to test first
`pkg/head`, `pkg/tail`, `pkg/grep`, `pkg/dirname`, `pkg/echo` — all have trivial
input/output contracts and can reach 70%+ with a few table-driven test cases each.

---

## 13.4 — BusyBox CI/Local Discrepancy

**Confirmed: most old-style tests run against system BusyBox on CI, not KoreGo.**

`test/busybox_testsuite/runtest` creates a `busybox → korego` symlink only inside the
`tar` and `gzip`/`gunzip` test tempdir blocks (lines ~22–36). All other old-style applet
dirs (`basename/`, `cat/`, `cp/`, `ln/`, `mv/`, `mkdir/`, `rm/`, etc.) rely on `busybox`
being found in `$PATH`, which on the CI Ubuntu runner resolves to `/usr/bin/busybox` —
the system binary, not KoreGo.

This means CI's "479 passed" count includes tests that prove system BusyBox works, not
that KoreGo's behavior matches. The true KoreGo pass rate for those cases is unknown.

### Preferred fix (Option A)
At the top of `runtest`, after `bindir` and `LINKSDIR` are set, prepend a block that
creates `busybox → korego` symlinks in a temp bin dir and prepends it to `PATH`:

```sh
# Route ALL 'busybox <applet>' calls to korego
BBDIR=$(mktemp -d)
ln -s "$bindir/korego" "$BBDIR/busybox"
export PATH="$BBDIR:$LINKSDIR:$PATH"
trap 'rm -rf "$BBDIR"' EXIT
```

### Fallback fix (Option B — CI only)
In `ci.yml`, before `make testsuite`:
```yaml
- name: Shadow system busybox with korego
  run: |
    sudo ln -sf "$PWD/korego" /usr/local/bin/busybox
```

---

## 13.5 — `awk` (Platinum Gate)

**Confirmed: `pkg/awk/` does not exist. No code at all.**

Implementation plan exists at [07a_awk.md](07a_awk.md). This is the only remaining
POSIX.2 utility. Without it the "100% POSIX userland" claim carries a permanent asterisk.

Required deliverables:
- `pkg/awk/awk.go` — core interpreter (patterns, `$N` fields, `BEGIN`/`END`, builtins: `NR`, `NF`, `FS`, `OFS`, `print`, `printf`, `sub`, `gsub`, `split`, `length`)
- `--json` output: array of per-record results
- `test/schemas/awk.schema.json`
- BusyBox `awk.tests` passing
- ≥20 unit test cases

---

## Prioritized Execution Plan

Tasks ordered by impact/effort ratio:

| Step | Task | Effort | Result |
|------|------|--------|--------|
| 1 | Add `//go:build linux` tags to `uname`, `stat`; add Darwin stubs | 2h | `make test` works on macOS |
| 2 | Wire `KOREGO_SHELL_TIMEOUT` env var in `interpreter.go` | 30min | Wiki/code parity restored |
| 3 | Write `internal/shell/interpreter_test.go` (3–5 tests) | 4h | 12.2 test requirement met |
| 4 | Write `docs/SECURITY.md` | 2h | 12.2 doc requirement met |
| 5 | Change coverage gate from `::warning::` to `exit 1` at 45% | 15min | Gate is real; raise to 60% after step 6 |
| 6 | Add unit tests to `head`, `tail`, `grep`, `dirname`, `echo`, `tee` | 2d | Coverage > 60% overall |
| 7 | Raise coverage gate threshold to 60% | 5min | 12.3 closed |
| 8 | Fix `runtest` to route all `busybox` calls to KoreGo | 4h | 12.4 closed |
| 9 | Add Trivy to `ci.yml` after Docker build | 30min | Vulnerability scanning live |
| 10 | Add SBOM stanza to `.goreleaser.yml` | 30min | Provenance on binaries |
| 11 | Add Cosign signing + SLSA provenance to `release.yml` | 4h | 12.1 closed |
| 12 | Write `pkg/awk/` | 5–10d | Platinum gate |

**Steps 1–11 = Gold. Steps 1–12 = Platinum.**

---

## How to Verify (Acceptance Criteria)

```bash
# macOS builds cleanly
GOOS=darwin CGO_ENABLED=0 go build ./...

# Shell timeout is configurable
KOREGO_SHELL_TIMEOUT=5s go test ./internal/shell/... -v -run TestTimeout

# Coverage gate fails when below threshold
make cover-pct   # must show ≥60% overall

# BusyBox results are same locally and on CI
make testsuite 2>&1 | tail -5

# Supply chain: image is signed
cosign verify ghcr.io/ramayac/korego:latest \
  --certificate-identity-regexp='.*' \
  --certificate-oidc-issuer='https://token.actions.githubusercontent.com'

# awk works (Platinum)
echo -e "a 1\nb 2\nc 3" | ./korego awk '{print $2}'
```
