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

| Utility | Notes |
|---------|-------|
| `test`/`[` | Conditional expressions. Returns exit code 0/1. |
| `printf` | Formatted output (POSIX format specifiers) |
| `expr` | Integer arithmetic and string operations |
| `sha256sum`/`md5sum` | Hash files. `--json` → `[{"file":"f","hash":"abc..."}]` |
| `tar` | Create/extract archives. `-c`, `-x`, `-f`, `-z` (gzip), `-v` |
| `gzip`/`gunzip` | Compress/decompress. Uses `compress/gzip` stdlib |
| `diff` | Compare files line by line. `--json` → structured hunks |
| `awk` | Start with basic: pattern matching, field splitting, print |

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
