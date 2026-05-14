# Phase 15 — Coverage Ramp (50% → 75%)

> **Status:** Planning | **Depends on:** Phase 12.3 | **Date:** 2026-05-13

---

## Goal

Raise overall test coverage from 50.0% to 75% by systematically closing the `run()` gap —
the private CLI-layer function that is at 0% coverage in nearly every package — and filling
the critical zero-coverage holes in `internal/daemon` and `cmd/korego`.

---

## Current State

### Headline Numbers

| Metric | Current | Target Stage 1 | Target Stage 2 | Target Stage 3 |
|--------|---------|---------------|---------------|---------------|
| Overall coverage | 50.0% | 60% | 68% | 75% |
| Packages at 0% | 1 (`cmd/korego`) | 0 | 0 | 0 |
| Packages at <10% | 4 (`daemon`, `tee`, `grep`, `internal/daemon`) | 1 | 0 | 0 |
| Packages at <30% | 10 | 4 | 0 | 0 |

### Root Cause Analysis

Every package follows the same arch: a public `Run()` (library layer, testable with `io.Reader`/`io.Writer`) and a private `run()` (CLI glue that calls `common.ParseFlags()`, opens real files, and binds `os.Stdin`/`os.Stdout`). Tests overwhelmingly cover `Run()` and helpers but skip `run()` entirely. The ratio of `run()` LOC to total LOC dictates the headline coverage number.

```
grep:  354 total LOC, 240 in run() →  16.5%  (worst utility)
head:  181 total LOC, 130 in run() →  29.0%
tee:   66  total LOC,  44 in run() →   3.3%  (worst overall)
```

### Worst 10 Packages (by coverage)

| # | Package | Cov% | Untested LOC | Dominant gap |
|---|---------|------|-------------|--------------|
| 1 | `internal/daemon` | 3.3% | 593 | Entire server, sessions, observability |
| 2 | `cmd/korego` | 0.0% | 110 | `main()` — symlink dispatch, subcommand routing |
| 3 | `pkg/daemon` | 5.9% | 41 | `run()` — socket creation, signal handling |
| 4 | `pkg/tee` | 3.3% | 44 | `run()` — file open, append mode |
| 5 | `pkg/grep` | 16.5% | 240 | `run()` — flag parsing, pattern assembly, file I/O |
| 6 | `pkg/touch` | 20.0% | 57 | `run()` — `-r`, `-t`, `-j` branches |
| 7 | `pkg/tail` | 27.6% | 134 | `run()` — flag parsing, file open loop |
| 8 | `pkg/head` | 29.0% | 128 | `run()` — flag parsing, file open loop |
| 9 | `pkg/pwd` | 30.4% | 43 | `run()` — `-L`/`-P` logic |
| 10 | `pkg/mv` | 31.8% | 53 | `run()` + `moveFile()` error paths |

---

## Implementation Plan

### Stage 1 — Critical Foundations (50% → 60%)

Close the three highest-risk zero-coverage surfaces. These are the packages that, if
broken, take down the entire daemon, RPC client, or binary entry point.

#### 1.1 — `internal/daemon` (3.3% → 35%)

**Why first:** 613 LoC. Handles all RPC dispatch, Unix socket lifecycle, session management,
rate limiting, and observability endpoints. This is the backbone of the daemon mode and it
has essentially zero unit test coverage.

**Approach — table-driven unit tests on individual functions:**

- [ ] **`server_test.go` — `processRequest`**: Build a `Server` with a test `dispatch.Register()`-ed command. Feed it JSON-RPC requests via `bytes.Buffer` connections. Cover: valid call, invalid JSON, unknown method, batch requests, notification (no `id`), error serialization.
- [ ] **`server_test.go` — `handleSingleAsync`**: Submit a task, verify async response on the conn writer.
- [ ] **`server_test.go` — `handleBatch`**: Feed a JSON array of 3 requests, verify 3 responses in order.
- [ ] **`server_test.go` — `writeError`**: Verify JSON-RPC error envelope structure.
- [ ] **`session_test.go` — lifecycle**: `NewSessionManager` → `Create` → `Get` → `SetCwd` → `List` → `Destroy`. Test TTL expiry via `cleanupLoop` with a short TTL.
- [ ] **`ratelimit_test.go`** — already exists, verify coverage of `Allow()` edge cases.
- [ ] **`server_test.go` — `NewWorkerPool` / `Submit`**: Submit tasks, verify they execute, verify graceful shutdown.

**Estimated LOC:** ~400 test lines. **Projected gain:** +3.2% overall.

#### 1.2 — `cmd/korego` (0.0% → 50%)

**Why second:** 110 LoC. The binary entry point. Has no tests at all.

**Approach:**

- [ ] **`main_test.go`**: Extract `runMain(args []string) error` from `main()`. Test:
  - Symlink name detection: `argv[0] = "ls"` → dispatches to `ls`
  - Subcommand mode: `korego ls -la`
  - `--list-commands` output
  - `--help` output
  - `--version` output
  - Missing subcommand → error

**Estimated LOC:** ~80 test lines. **Projected gain:** +0.5% overall.

#### 1.3 — `pkg/client` (44.9% → 60%)

**Why third:** 28 RPC client helpers at 0%. The client package is the programmatic interface
to the daemon — it must work. Testing one helper exercises the pattern for all.

**Approach — integration-style with a real in-process Unix socket:**

- [ ] **`client_helpers_test.go`**: Start `internal/daemon` on a temp Unix socket. Call:
  - `Cat()`, `Echo()`, `Pwd()`, `Head()`, `Tail()` — the read-only helpers
  - `Stat()`, `Ls()` — the filesystem helpers  
  - `Ping()` — basic connectivity
  - Verify JSON-RPC responses deserialize correctly into typed structs.
  - Test `Dial()` error path (no socket), `Close()`, `WithTimeout()`, `WithMaxRetries()`.
  - Test `Notify()` (fire-and-forget) and `Batch()` (multi-request).

**Estimated LOC:** ~250 test lines. **Projected gain:** +1.5% overall.

#### 1.4 — `pkg/daemon` (5.9% → 40%)

**Why fourth:** The CLI entry point for daemon mode. Thin (43 LoC) but completely untested.

**Approach:**

- [ ] **`daemon_test.go`**: Refactor `run()` to accept a `socketPath` param. Test:
  - Socket creation and bind
  - SIGTERM/SIGINT graceful shutdown
  - `--socket` flag path

**Estimated LOC:** ~60 test lines. **Projected gain:** +0.3% overall.

---

### Stage 2 — Utility CLI Layer (60% → 68%)

The `run()` gap. For every utility where `run()` is at 0%, add 2-5 tests that call
`dispatch.Lookup(name).Run(args, &buf)` with `testdata/` fixture files. This exercises
the full flag parsing + file I/O path without refactoring the `run()` signatures.

#### 2.1 — High-impact utilities (largest `run()` functions)

| Utility | Current | Target | Tests to add | Est. gain |
|---------|---------|--------|-------------|-----------|
| `grep` | 16.5% | 45% | `-e` pattern, `-f` file, `-v` invert, `-r` recursive, `-c` count, `-l`/`-L`, `-i` case, `-w` word, `-x` line, error: missing pattern, error: bad regex | +1.2% |
| `sed` | 49.1% | 65% | `s/foo/bar/`, `-n` quiet, `d` delete, `-e` script, `-f` file, `-i` in-place, address ranges, `y` transliterate | +1.1% |
| `cat` | 37.6% | 55% | `-n` number, `-b` number-nonblank, `-s` squeeze-blank, `-v` show-nonprinting, `-E` show-ends, `-A` show-all, stdin `-` | +0.6% |
| `find` | 50.0% | 65% | `-name`, `-type`, `-size`, `-exec`, `-print`, `-delete`, `-maxdepth` | +0.6% |
| `tar` | 43.5% | 60% | `-c` create, `-x` extract, `-t` list, `-v` verbose, `-f` file, `-C` directory | +0.5% |
| `sort` | 58.0% | 70% | `-n`, `-r`, `-u`, `-k`, `-o` output, `-t` delimiter | +0.5% |

#### 2.2 — Medium-impact utilities

| Utility | Current | Target | Tests to add | Est. gain |
|---------|---------|--------|-------------|-----------|
| `ls` | 45.8% | 60% | `-l` long format, `-a` all, `-R` recursive, `-h` human-readable, `-1` one-per-line | +0.4% |
| `wc` | 32.5% | 55% | `-l`, `-w`, `-c`, `-m` chars, multi-file, stdin | +0.3% |
| `head` | 29.0% | 55% | `-n` positive, `-n` negative, `-c` bytes, multi-file headers | +0.3% |
| `tail` | 27.6% | 55% | `-n` lines, `-c` bytes, `-f` follow (timeout), multi-file headers | +0.3% |
| `cp` | 49.6% | 65% | file→file, file→dir, dir→dir `-R`, `-a` preserve, error: no-such-file | +0.3% |
| `rm` | 35.3% | 55% | single file, `-r` recursive, `-f` force, `--no-preserve-root`, error: ENOENT | +0.2% |
| `stat` | 35.9% | 55% | `-c` format, `-f` filesystem, symlink deref | +0.2% |
| `touch` | 20.0% | 50% | create new, update existing, `-r` reference, `-t` timestamp, error: bad format | +0.2% |
| `mv` | 31.8% | 50% | file rename, cross-device, dir target, error: ENOENT | +0.2% |

#### 2.3 — Low-hanging fruit (small packages, big % jumps)

| Utility | Current | Target | Tests to add | Est. gain |
|---------|---------|--------|-------------|-----------|
| `tee` | 3.3% | 50% | stdin→stdout+file, `-a` append, multi-file, error: unlinkable | +0.2% |
| `pwd` | 30.4% | 60% | default, `-L`, `-P` | +0.1% |
| `mkdir` | 32.4% | 60% | single dir, `-p` parents, `-m` mode | +0.1% |
| `rmdir` | 44.4% | 65% | single, `-p` parents, error: not-empty | +0.1% |
| `readlink` | 34.4% | 55% | default, `-f` canonical, `-e`, `-m` | +0.1% |
| `xargs` | 53.3% | 65% | `-n`, `-I`, `-t`, stdin piped | +0.1% |
| `gzip` | 54.8% | 65% | `-c` stdout, `-d` decompress, `-1`..`-9` levels | +0.1% |
| `printenv` | 44.4% | 60% | all vars, single var, missing var | +0.05% |
| `echo` | 56.9% | 70% | `-n`, `\\t`, `\\0NNN`, `\\xNN`, `\\c` | +0.1% |

---

### Stage 3 — Refactor & Harden (68% → 75%)

For the worst remaining packages, refactor `run()` to accept interfaces so the CLI layer
becomes directly testable without filesystem dependencies.

#### 3.1 — Refactor `run()` signatures

- [ ] **`pkg/grep/grep.go`**: Extract `run(args, stdout, stderr io.Writer, stdin io.Reader) int`. Test all flag combinations (~12 tests).
- [ ] **`pkg/tar/tar.go`**: Extract create/extract/list into separate testable functions. Test archive round-trip (create→extract with temp dir).
- [ ] **`pkg/sed/sed.go`**: Extract `compileScripts()` and `runEngine()` from file I/O.
- [ ] **`pkg/find/find.go`**: Extract `parsePredicates()` as a pure function from OS traversal.

#### 3.2 — Per-package coverage reporting

- [ ] Add `make cover-pkg` target that prints a sorted table of per-package coverage.
- [ ] Update CI coverage gate to 60% hard-fail (from current 45%).

#### 3.3 — Documentation

- [ ] Document the coverage policy in `AGENTS.md`: library functions must be tested via `Run()`, CLI glue must be tested via dispatch, target 80% per package.
- [ ] Add `wiki/15_coverage_ramp.md` to wiki index.

---

## Test Fixture Convention

All Stage 2 tests share a common pattern. Each package under test gets a `testdata/` directory:

```
pkg/grep/testdata/
├── alice.txt          # "Alice was beginning to get very tired..."
├── empty.txt          # zero-byte file
├── binary.bin         # contains \x00 bytes
└── patterns.txt       # one pattern per line (for -f)
```

Tests call through the dispatch registry so the full `run()` path is exercised:

```go
func TestRunViaDispatch(t *testing.T) {
    cmd := dispatch.Lookup("grep")
    if cmd == nil { t.Fatal("grep not registered") }

    var out bytes.Buffer
    exitCode := cmd.Run([]string{"-c", "Alice", "testdata/alice.txt"}, &out)
    // assert exitCode, out.String()
}
```

---

## Task Summary

| Stage | Packages | New Tests | Est. LOC | Cumulative Coverage |
|-------|----------|-----------|----------|--------------------|
| **Current** | — | — | — | **50.0%** |
| 1.1 `internal/daemon` | 1 | ~25 | ~400 | 53.2% |
| 1.2 `cmd/korego` | 1 | ~8 | ~80 | 53.7% |
| 1.3 `pkg/client` | 1 | ~15 | ~250 | 55.2% |
| 1.4 `pkg/daemon` | 1 | ~4 | ~60 | 55.5% |
| — | — | — | — | **~60% (Stage 1 done)** |
| 2.1 High-impact | 6 | ~50 | ~800 | 65.0% |
| 2.2 Medium-impact | 9 | ~35 | ~500 | 67.8% |
| 2.3 Low-hanging | 9 | ~20 | ~300 | 68.5% |
| — | — | — | — | **~68% (Stage 2 done)** |
| 3.1 Refactors | 4 | ~30 | ~400 | 72.0% |
| 3.2 Tooling | — | — | — | — |
| 3.3 Docs | — | — | — | — |
| — | — | — | — | **~75% (Stage 3 done)** |

---

## Verification

```bash
# Overall coverage check
make cover-pct                           # must show >=60% (Stage 1), >=68% (Stage 2), >=75% (Stage 3)

# Per-package minimums
make cover-pkg                           # no package below 25% (Stage 1), 40% (Stage 2), 55% (Stage 3)

# CI gate
make ci                                  # coverage gate hard-fails below threshold

# No regressions in compliance
make testsuite                           # must stay at 409+ passed
```

---

## Risks

| Risk | Mitigation |
|------|------------|
| `internal/daemon` tests are flaky due to real Unix sockets | Use `t.TempDir()` for socket paths, add `t.Cleanup()` server shutdown |
| `testdata/` fixture bloat across 30+ packages | Share a top-level `testdata/` dir with symlinks where possible |
| Dispatch-call tests are slow (real fs) | Use `t.Parallel()`; filesystem utilities are fast |
| Coverage tooling brittleness | Use `go test -coverprofile` with `go tool cover -func` (stdlib, no deps) |
| Refactoring `run()` breaks existing behavior | Refactor only to extract params; no logic changes; compliance suite catches regressions |
