# Phase 09 — Release Automation & Documentation

> **Timeline:** Week 13–14 | **Depends on:** Phase 08

---

## Goal

Automate releases, finalize all documentation, and prepare for public launch.

## Tasks

### 09.1 — Multi-Arch Builds

- [ ] Build for `linux/amd64` and `linux/arm64`
- [ ] Docker `buildx` with `--platform linux/amd64,linux/arm64`
- [ ] CI builds both architectures
- [ ] Binary size targets: < 15MB per arch (stripped)
- [ ] Image size targets: < 20MB per arch

### 09.2 — Release Automation (GoReleaser)

```yaml
# .goreleaser.yml
builds:
  - id: korego
    main: ./cmd/korego/
    binary: korego
    env: [CGO_ENABLED=0]
    goos: [linux]
    goarch: [amd64, arm64]
    ldflags: ["-s", "-w", "-X main.version={{.Version}}"]

dockers:
  - image_templates: ["ghcr.io/org/korego:{{.Version}}"]
    dockerfile: docker/Dockerfile
    build_flag_templates: ["--platform=linux/amd64"]

changelog:
  sort: asc
  filters:
    exclude: ["^docs:", "^test:", "^ci:"]
```

- [ ] Semantic versioning (v0.1.0, v0.2.0, ...)
- [ ] Git tags trigger release builds
- [ ] Docker images published to GHCR
- [ ] GitHub Release with binaries + checksums

### 09.3 — Documentation Finalization

| Document | Content | Status |
|----------|---------|--------|
| `README.md` | Project overview, quickstart, Docker usage, examples | [ ] |
| `AGENTS.md` | Coding conventions, commit format, PR rules | [ ] |
| `docs/JSON_SCHEMA.md` | `--json` output schema for every utility with examples | [ ] |
| `docs/RPC_API.md` | JSON-RPC method catalog, request/response, error codes | [ ] |
| `wiki/posix_coverage.md` | Compliance matrix: utility × flag × status | [ ] |
| `docs/ARCHITECTURE.md` | System architecture, package diagram, data flow | [ ] |
| GoDoc (inline) | Every exported function, type, and package | [ ] |

### 09.4 — End-to-End Agent Test

Simulate a real agentic workflow as a single integration test:

```go
func TestAgentWorkflow(t *testing.T) {
    // 1. Start daemon
    // 2. Create session
    // 3. List files in /
    // 4. Create temp directory
    // 5. Write a file (via shell.exec "echo content > file")
    // 6. Read the file back
    // 7. Grep for a pattern
    // 8. Get checksum
    // 9. Cleanup
    // 10. Destroy session
    // All via RPC, all returning structured JSON
}
```

- [ ] Full workflow passes end-to-end
- [ ] All RPC responses are valid JSON-RPC 2.0
- [ ] No errors, no panics, no resource leaks

### 09.5 — Performance Report

Final benchmarks published:

```
Benchmark Results (amd64, 4-core)
─────────────────────────────────
CLI cold start (echo):     ~8ms
Daemon RPC (echo):         ~0.3ms  (26x faster)
Daemon RPC (ls /):         ~1.2ms
Daemon RPC (grep -r):      ~4.5ms
Batch throughput:          ~2400 req/sec
Binary size:               12.4 MB
Docker image size:         14.8 MB
```

## Milestone 09 (Final)

- [ ] Multi-arch images published to GHCR
- [ ] GoReleaser pipeline produces tagged releases
- [ ] All 7 documentation deliverables complete
- [ ] E2E agent test passes
- [ ] Binary < 15MB, image < 20MB
- [ ] README has quickstart that works in < 2 minutes
- [ ] POSIX compliance >= 80% across all implemented utilities

## How to Verify

```bash
# Release dry run
goreleaser release --snapshot --clean

# Image size
docker images ghcr.io/org/korego

# E2E test
go test -v -run TestAgentWorkflow ./test/integration/

# Final smoke
docker run --rm ghcr.io/org/korego:latest ls --json /
docker run --rm ghcr.io/org/korego:latest grep --json -r "korego" /bin/
docker run --rm ghcr.io/org/korego:latest uname --json
```

---

## Project Complete Checklist

- [ ] **50+ POSIX utilities** implemented in pure Go
- [ ] **`--json` flag** on every utility with consistent envelope
- [ ] **JSON-RPC 2.0 daemon** over Unix socket
- [ ] **Stateful sessions** for agentic workflows
- [ ] **Shell interpreter** via mvdan/sh
- [ ] **FROM scratch** Docker image < 20MB
- [ ] **Multi-arch** (amd64 + arm64)
- [ ] **> 80% POSIX compliance** with documented coverage
- [ ] **< 1ms daemon latency** for trivial commands
- [ ] **Security hardened** (non-root, sandboxed shell, rate limits)
- [ ] **Fully documented** (JSON schemas, RPC API, architecture)
- [ ] **CI/CD** with automated releases
