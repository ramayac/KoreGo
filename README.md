# KoreGo

KoreGo is a 100% Go-native, POSIX-compliant userland designed to run inside a Docker `FROM scratch` container. It serves as a modern replacement for GNU Coreutils CLI tools by compiling down to a single multicall binary (like BusyBox).

Crucially, KoreGo is designed for **Agentic Runtimes**:
- Every utility supports structured machine-readable output via a `--json` flag.
- It features a persistent JSON-RPC 2.0 daemon to avoid continuous process-spawning overhead.
- Includes a fully sandboxed shell interpreter (`mvdan.cc/sh`).

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
KoreGo MVP is complete with 50+ POSIX utilities implemented.
