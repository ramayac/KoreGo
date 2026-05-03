# KoreGo

KoreGo is a 100% Go-native, POSIX-compliant userland designed to run inside a Docker `FROM scratch` container. It serves as a modern replacement for GNU Coreutils CLI tools by compiling down to a single multicall binary (like BusyBox).

KoreGo is designed for **Agentic Runtimes**:
- Every utility supports structured machine-readable output via a `--json` flag.
- A persistent JSON-RPC 2.0 daemon avoids continuous process-spawning overhead.
- A fully sandboxed shell interpreter (`mvdan.cc/sh`) enables portable scripting.
- **>90% compatible with the BusyBox test suite** — 479 passed, 1 failed, 10 skipped (97.9% effective pass rate out of 490 tests).

## Quickstart

### Docker
```bash
docker pull ghcr.io/ramayac/korego:latest
docker run --rm ghcr.io/ramayac/korego:latest ls --json /
```

### Build from Source
```bash
make all
./korego --list-commands
```

### Run POSIX Tests
```bash
make testsuite
```

### Start Daemon
```bash
./korego daemon --socket /tmp/korego.sock &
```

## Documentation
- [Architecture](docs/ARCHITECTURE.md)
- [JSON Schema](docs/JSON_SCHEMA.md)
- [RPC API](docs/RPC_API.md)
- [POSIX Coverage](wiki/posix_coverage.md)

## Status

KoreGo MVP is complete with **49 POSIX utilities implemented** (100% of target scope).

**BusyBox Test Suite:** 479 passed, 1 failed, 10 skipped (97.9% effective pass rate out of 490 tests)

| Utility | Status | Notes |
|---------|--------|-------|
| Core & Env (10) | ✅ | echo, env, pwd, true, false, whoami, hostname, basename, dirname, printenv |
| Filesystem (11) | ✅ | ls, cat, mkdir, rmdir, rm, cp, mv, touch, ln, stat, readlink |
| Text (10) | ✅ | head, tail, wc, sort, uniq, tr, cut, tee, grep, sed |
| System (13) | ✅ | ps, kill, sleep, date, id, groups, chmod, chown, chgrp, df, du, find, xargs |
| Agent (5) | ✅ | diff, tar, gzip, printf, expr, sha256sum, md5sum |

**All Phases Complete (00–10).** The single remaining test failure (`tar writing into read-only dir`) is umask-dependent and passes with umask 022. All 10 skipped tests require external compression tools (bzip2, xz) or PAX extended header support.

`awk` is deferred to a post-MVP release (see [POSIX FAQ](wiki/posix_faq.md)).
