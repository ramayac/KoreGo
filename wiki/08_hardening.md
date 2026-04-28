# Phase 08 — Production Hardening & Security

> **Timeline:** Week 11–12 | **Depends on:** Phase 07

---

## Goal

Harden the daemon and container for production use. Security audit, resource limits, and stability testing.

## Tasks

### 08.1 — Security Audit

- [ ] **Path traversal:** All file operations validate paths. No `../../etc/shadow` escapes.
- [ ] **Shell sandbox:** `korego.shell.exec` runs with:
  - No network access (no `net.Dial` exposed)
  - Restricted filesystem (configurable allowed paths)
  - Execution timeout (default 30s, max 5min)
  - Memory limit (default 128MB per script)
- [ ] **Rate limiting:** Max 100 RPC requests/sec per connection (configurable)
- [ ] **Input validation:** All RPC params validated against expected types/ranges
- [ ] **Symlink following:** Configurable — refuse to follow symlinks outside allowed paths

### 08.2 — Non-Root Container

- [ ] Create user `korego:1000` in builder stage
- [ ] Copy `/etc/passwd` and `/etc/group` to scratch image
- [ ] `USER korego` in Dockerfile
- [ ] Socket created with proper group permissions
- [ ] Test: daemon works as non-root

### 08.3 — Resource Management

- [ ] Session TTL with automatic cleanup goroutine
- [ ] Connection limits (max concurrent connections)
- [ ] Request body size limit (default 1MB)
- [ ] Response size limit for `--json` (paginate large results)
- [ ] Memory profiling with `pprof` endpoint (debug mode only)

### 08.4 — Stability Testing

- [ ] 24-hour soak test under moderate load
- [ ] Memory leak detection (`go test -race`, `pprof` heap profiles)
- [ ] Graceful shutdown under load (SIGTERM during active requests)
- [ ] Socket file cleanup on crash (stale socket detection on restart)
- [ ] Fuzz testing on JSON-RPC parser (`go test -fuzz`)

### 08.5 — POSIX Coverage Matrix (`wiki/posix_coverage.md`)

- [x] Matrix covers all 50+ utilities (Completed, see `posix_coverage.md`)
- [x] Each utility lists: implemented flags, missing flags, known deviations
- [ ] Overall compliance percentage calculated

*(Note: `awk` has been explicitly deferred from the MVP scope.)*

## Milestone 08

- [ ] Security audit complete, no critical findings
- [ ] Daemon runs as non-root in container
- [ ] 24-hour soak test passes (no memory leaks, no crashes)
- [ ] Fuzz testing on JSON-RPC parser finds no panics
- [ ] POSIX coverage matrix published, >= 80% overall

## How to Verify

```bash
# Soak test
./korego daemon --socket /tmp/korego.sock &
go test -run TestSoak -timeout 25h ./test/integration/

# Fuzz
go test -fuzz FuzzJSONRPC -fuzztime 10m ./pkg/common/

# Memory profile
curl --unix-socket /tmp/korego.sock http://localhost/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Non-root test
docker run --rm --user korego korego:prod whoami  # → korego
```
