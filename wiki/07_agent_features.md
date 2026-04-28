# Phase 07 — Agent-Ready Features + Tier 5

> **Timeline:** Week 9–10 | **Depends on:** Phase 05, 06

---

## Goal

Add stateful RPC sessions, embedded shell interpreter, and the final tier of advanced utilities.

## Tasks

### 07.1 — Stateful Sessions (`internal/daemon/session.go`)

Each agent connection can create a session that remembers state across RPC calls.

```json
// Create session
{"jsonrpc":"2.0","method":"korego.session.create","id":1}
→ {"result":{"sessionId":"abc-123","cwd":"/"}}

// Set working directory
{"jsonrpc":"2.0","method":"korego.session.setCwd","params":{"sessionId":"abc-123","path":"/tmp"},"id":2}

// Run command in session context
{"jsonrpc":"2.0","method":"korego.ls","params":{"sessionId":"abc-123","path":"."},"id":3}
// ls runs relative to /tmp because of session cwd
```

- [x] Session stores: `cwd`, `env` vars, command history
- [x] Session TTL: auto-expire after 30min idle (configurable)
- [x] `korego.session.list` — list active sessions
- [x] `korego.session.destroy` — cleanup

### 07.2 — Shell Interpreter (`internal/shell/interpreter.go`)

Embed `mvdan.cc/sh/v3` to execute shell scripts via RPC.

```json
{"jsonrpc":"2.0","method":"korego.shell.exec",
 "params":{"script":"ls -la | grep txt | wc -l", "sessionId":"abc-123"},"id":4}
→ {"result":{"stdout":"3\n","stderr":"","exitCode":0}}
```

- [x] Builtins: KoreGo utilities are registered as shell builtins (no fork/exec)
- [x] Pipes: `ls | grep | wc` uses Go channels, not OS pipes
- [x] Environment: inherits from session env vars
- [x] Safety: execution timeout (default 30s), memory limit

### 07.3 — Structured Logging

All daemon operations log as structured JSON to stderr/file.

```json
{"time":"2026-04-24T20:00:00Z","level":"info","method":"korego.ls","sessionId":"abc-123","durationMs":2.1}
```

- [x] Use `log/slog` from stdlib
- [x] Log levels: debug, info, warn, error
- [x] Fields: timestamp, level, method, session_id, duration_ms, error

### 07.4 — Tier 5 Utilities

Implemented in order of complexity, each building on patterns established by earlier utilities.

> **Note:** `awk` has been moved to a dedicated document — see [07a_awk.md](07a_awk.md).

---

#### 07.4.1 — `printf` (Formatted Output) ⚠️ Partial

**Complexity:** Low — pure output, no filesystem interaction.

Implement POSIX `printf` with format specifiers (`%s`, `%d`, `%x`, `%o`, `%f`, `\n`, `\t`, `\\`, `\0NNN`).

- [x] Parse format string and argument list (basic — delegates to `fmt.Sprintf`)
- [x] Support escape sequences `\n`, `\t`, `\r`
- [ ] Support `%x`, `%o`, `%f` specifiers with width/precision (currently relies on Go's Sprintf)
- [ ] Support `\\`, `\0NNN` octal escapes
- [ ] Cycle through args if more args than specifiers
- [x] `--json` output via `common.Render`
- [ ] Unit tests
- [ ] Compliance test

---

#### 07.4.2 — `expr` (Integer Arithmetic & String Ops) ⚠️ Naive Stub

**Complexity:** Low-Medium — recursive-descent expression parser, no I/O.

Implement POSIX `expr` supporting arithmetic (`+`, `-`, `*`, `/`, `%`), comparison (`<`, `<=`, `=`, `!=`, `>=`, `>`), logical (`|`, `&`), string matching (`:` / `match`), and `substr`, `index`, `length`.

- [x] Basic `A OP B` arithmetic (`+`, `-`, `*`, `/`) with `Atoi`
- [x] Division-by-zero guard
- [x] Exit code `1` when result is zero, `2` on error
- [ ] Tokenizer for full POSIX expr grammar
- [ ] Recursive-descent evaluator with correct operator precedence
- [ ] Comparison operators (`<`, `<=`, `=`, `!=`, `>=`, `>`)
- [ ] Logical operators (`|`, `&`)
- [ ] Modulo (`%`)
- [ ] String operations: `match`, `substr`, `index`, `length`
- [x] `--json` output via `common.Render`
- [ ] Unit tests
- [ ] Compliance test

---

#### 07.4.3 — `test` / `[` (Conditional Expressions) ❌ Not Implemented

**Complexity:** Medium — must handle file tests, string tests, integer comparisons, and compound expressions.

Implement POSIX `test` (and `[` symlink form) for shell conditional evaluation.

> ⚠️ `pkg/testcmd/` directory exists but is **empty** — no source code.

- [ ] File tests: `-e`, `-f`, `-d`, `-s`, `-r`, `-w`, `-x`, `-L`, `-h`
- [ ] String tests: `-z`, `-n`, `=`, `!=`
- [ ] Integer comparisons: `-eq`, `-ne`, `-lt`, `-le`, `-gt`, `-ge`
- [ ] Logical operators: `!`, `-a`, `-o`, `(`, `)`
- [ ] `[` mode: validate closing `]` argument
- [ ] Exit code only (no stdout), `--json` → `{"result": true/false}`
- [ ] Unit tests
- [ ] Compliance test

---

#### 07.4.4 — `sha256sum` / `md5sum` (Cryptographic Hashing) ⚠️ Partial

**Complexity:** Medium — streaming file I/O with `crypto/sha256` and `crypto/md5`.

Hash one or more files, output in standard `HASH  FILENAME` format.

**`sha256sum`** — partial implementation (75 lines):
- [x] Stream file contents through `sha256.New()` via `io.Copy` (no full-file buffering)
- [x] Multi-file hashing with per-file error handling
- [x] `--json` output via `common.Render`
- [ ] Support `-c` / `--check` mode (verify hashes from a checksum file)
- [ ] Handle stdin via `-` argument (stub comment exists, not implemented)

**`md5sum`** — not implemented:
- [ ] `pkg/md5sum/` directory exists but is **empty**
- [ ] Implement `md5sum` (mirror `sha256sum` with `crypto/md5`)

**Shared:**
- [ ] Unit tests
- [ ] Compliance test

---

#### 07.4.5 — `diff` (File Comparison) ⚠️ Naive Stub

**Complexity:** Medium-High — Myers diff algorithm, hunk formatting.

Compare two files line by line, producing unified diff output.

- [x] Read two files, compare with `bytes.Equal`
- [x] Exit code: `0` (identical), `1` (different), `2` (error)
- [x] Basic `--json` output (`{"differ": true/false}`)
- [ ] Implement Myers diff algorithm (or equivalent LCS-based)
- [ ] Unified diff format with `---`/`+++` headers and `@@` hunk markers
- [ ] Context lines (default 3, configurable via `-U N`)
- [ ] `--json` structured hunks: `{"files": [...], "hunks": [...]}`
- [ ] Unit tests
- [ ] Compliance test

---

#### 07.4.6 — `gzip` / `gunzip` (Compression) ⚠️ Partial

**Complexity:** Medium-High — streaming compression with `compress/gzip`, file replacement semantics.

Compress and decompress files using gzip format.

- [x] `gzip FILE` → creates `FILE.gz`, removes original
- [x] `gunzip FILE.gz` → restores `FILE`, removes `.gz`
- [x] `-d` / `--decompress` — gzip delegates to gunzip
- [x] Stdin/stdout piping when no file argument
- [ ] `-c` / `--stdout` — write to stdout, keep original
- [ ] `-k` / `--keep` — keep original file
- [ ] `-f` / `--force` — overwrite existing output
- [ ] `--json` output with size/ratio stats (currently just `{"status": "compressed successfully"}`)
- [ ] Unit tests
- [ ] Compliance test

---

#### 07.4.7 — `tar` (Archive Create/Extract) ⚠️ Partial

**Complexity:** High — recursive directory traversal, multiple archive formats, gzip integration.

Create and extract tar archives with optional gzip compression.

- [x] `-c` — create archive (working, uses `filepath.Walk` + `tar.Writer`)
- [x] `-f FILE` — specify archive filename
- [x] `-z` — gzip compression on create (via `compress/gzip`)
- [x] `-v` — verbose listing during create
- [x] Preserve permissions and timestamps (via `tar.FileInfoHeader`)
- [ ] `-x` — extract archive (**stub only** — returns 0 but does nothing)
- [ ] `-t` — list archive contents
- [ ] `-C DIR` — change to directory before operating
- [ ] `--json` output: `{"files": [{"name": "...", "size": N, "mode": "..."}]}`
- [ ] Unit tests
- [ ] Compliance test

### 07.5 — Benchmarking Suite

```
test/benchmark/
├── bench_cli_test.go      # Cold-start CLI latency per utility
├── bench_daemon_test.go   # Warm daemon RPC latency per utility
└── bench_batch_test.go    # Batch throughput (requests/sec)
```

Targets:
- CLI cold start: < 10ms (Go runtime init)
- Daemon echo: < 1ms
- Daemon ls: < 5ms
- Batch throughput: > 1000 req/sec

## Milestone 07

- [x] Agent creates session, sets cwd, runs relative commands
- [x] `korego.shell.exec "ls | grep go | wc -l"` returns structured result
- [x] Daemon latency < 1ms for `echo`, < 5ms for `ls`
- [x] Sessions auto-expire after TTL
- [x] `sha256sum --json file` returns hash
- [x] `tar -czf archive.tar.gz dir/` creates archive
- [x] Makefile target so the user can get an interactive shell in the docker image. (docker run -it korego)

## How to Verify

```bash
# Session workflow
echo '{"jsonrpc":"2.0","method":"korego.session.create","id":1}' | socat - UNIX-CONNECT:/tmp/korego.sock
# → get sessionId, then use it in subsequent calls

# Shell exec
echo '{"jsonrpc":"2.0","method":"korego.shell.exec","params":{"script":"echo hello | tr a-z A-Z"},"id":2}' \
  | socat - UNIX-CONNECT:/tmp/korego.sock

# Benchmarks
go test -bench=. -benchmem ./test/benchmark/
```
