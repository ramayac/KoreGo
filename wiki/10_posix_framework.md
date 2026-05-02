# Phase 10 — POSIX Testing Framework

> **Timeline:** Future | **Depends on:** Phase 09

---

## Goal

Integrate a formal POSIX testing framework to programmatically verify KoreGo's compliance against the IEEE Std 1003.1 spec, rather than relying solely on our internal bash compliance scripts.

## Rationale

Currently, `test/compliance/` compares KoreGo's output to the host system's GNU Coreutils. While effective, GNU Coreutils often implements extensions beyond POSIX. A dedicated POSIX test suite ensures strict adherence to the standard, avoiding GNU-specific bloat and identifying subtle deviations.

## Framework Options

### 1. The Open Group POSIX Test Suite (VSTS)
- The official, historical test suite for POSIX compliance.
- **Pros:** The gold standard.
- **Cons:** Very old, hard to compile on modern systems, proprietary/licensed portions exist.

### 2. FreeBSD / NetBSD Test Suites (`kyua` / `atf`)
- BSD implementations are generally closer to POSIX than GNU. 
- **Pros:** Open source, robust, actively maintained.
- **Cons:** Some BSD-isms may still exist.

### 3. BusyBox Test Suite
- BusyBox has an extensive `testsuite/` directory that verifies utility behavior.
- **Pros:** Designed for minimal, multi-call binaries (very similar to KoreGo).
- **Cons:** Focuses on BusyBox behavior, which sometimes deviates from strict POSIX to save space.

### 4. Custom TAP-based Test Harness
- Build a Go-based or Bash-based Test Anything Protocol (TAP) runner that executes defined POSIX inputs and asserts outputs.
- **Pros:** Full control, easy to integrate into `go test`, can easily test JSON output alongside text output.
- **Cons:** High maintenance burden, requires transcribing POSIX spec into test cases.

## Selected Approach: Hybrid (BusyBox Test Suite + Custom TAP)

Given KoreGo's architecture, we will adapt the BusyBox test suite for baseline behavior and build a custom TAP-based runner to verify our specific `--json` outputs and JSON-RPC daemon interactions.

## Tasks

### 10.1 — Framework Selection and Porting
- [x] Investigate and fork the BusyBox `testsuite/` directory.
- [x] Adapt the runner to execute against the `korego` binary instead of `busybox`.
- [x] Disable tests for utilities we haven't implemented yet.

### 10.2 — Baseline Execution
- [x] Run the adapted suite.
- [x] Document all failures in `wiki/posix_coverage.md` as known deviations.
- [x] Fix critical deviations (e.g. standard flags failing).

### 10.3 — JSON/RPC Testing Harness
- [x] Create a `test/posix-json/` directory.
- [x] Implement a Go-based TAP runner that runs POSIX commands via the JSON-RPC daemon.
- [x] Verify that the structured output semantics map correctly to POSIX expectations (e.g. exit codes, stderr vs stdout separation).

### 10.4 - Sed
- [x] Fix test/busybox_testsuite/sed.tests to pass all tests. (See [10a_sed.md](10a_sed.md) for detailed sub-tasks)

### 10.5 - busybox test step in github action
- [x] Add a step in github action to run the busybox test suite. Don't block the pipeline, just result of test that pass.

### 10.6 - Phase A: Core Utilities & Flag Parsing Test Resolution
- [x] Fixed `pkg/common/flags.go` to support concatenated values (e.g., `-s25`).
- [x] Fixed `pkg/xargs/xargs.go` to handle `-e` without a parameter.
- [x] Fixed `pkg/grep/grep.go` multiple `-e` and `-f` flags, `-f EMPTY_FILE` behavior, `-r` symlink traversal (trailing slash), and `-L` exit codes.
- [x] Fixed `pkg/find/find.go` trailing slash canonicalization by preprocessing single-dash long flags (`-name`, `-type`).
- [x] Fixed `pkg/sed/sed.go` missing labels validation and NUL bytes handling by removing the `0` EOF marker.
- [x] Fixed `pkg/echo/echo.go` to support `\0NNN` octal escapes and `\xNN` hex escapes, required by BusyBox's `sed.tests` harness.
- [x] Fixed `pkg/wc/wc.go` field formatting to align with POSIX (no leading spaces), which resolved the `tail -c +N` validation failures.

## Milestone 10
- [ ] External test suite integrated into `make test` or `make compliance`.
- [ ] CI pipeline runs the external test suite.
- [ ] `posix_coverage.md` updated with programmatic results.
