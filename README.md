# KoreGo

A Go-native, single-binary POSIX userland (>97% BusyBox test compatibility). KoreGo replaces
GNU Coreutils in Docker `FROM scratch` containers, featuring native `--json` structured output
in every tool (`--xml` in progress — see [Phase 14](wiki/14_xml_output.md)).

Key Features:
 - Machine-Readable by Default: Every utility supports a `--json` flag for structured output (`--xml` planned — [Phase 14](wiki/14_xml_output.md)).
 - Low-Overhead Execution: A persistent JSON-RPC 2.0 daemon eliminates continuous process-spawning overhead.
 - Portable Scripting: Includes a fully sandboxed shell interpreter (mvdan.cc/sh).
 - High Compatibility: 97.9% pass rate against the BusyBox test suite (479 passed, 1 failed, 10 skipped).

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

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `KOREGO_SHELL_TIMEOUT` | `30s` | Shell execution timeout (Go duration format, e.g. `60s`, `5m`) |

> `KOREGO_SHELL_TIMEOUT` controls how long `korego.shell.exec` RPC calls will run before being terminated. See [Security Model](docs/SECURITY.md) for details.

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
| Agent (7) | ✅ | diff, tar, gzip, printf, expr, sha256sum, md5sum |

**All Phases Complete (00–10).** The single remaining test failure (`tar writing into read-only dir`) is umask-dependent and passes with umask 022. All 10 skipped tests require external compression tools (bzip2, xz) or PAX extended header support.

`awk` is deferred to a post-MVP release (see [POSIX FAQ](wiki/posix_faq.md)).
