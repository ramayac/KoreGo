# Agent Context & Directives for KoreGo

**Hello AI Assistant!** If you are reading this, you are working on **KoreGo** (formerly CoreGoLinux). This document provides the critical context, architectural invariants, and workflow rules required to contribute successfully to this project. 

## 1. Project Identity & Goal

KoreGo is a 100% Go-native, POSIX-compliant userland designed to run inside a Docker `FROM scratch` container. It serves as a modern replacement for GNU Coreutils CLI tools by compiling down to a single multicall binary (like BusyBox).

Crucially, KoreGo is designed for **Agentic Runtimes**:
1. Every utility supports structured machine-readable output via a `--json` flag.
2. It will eventually feature a persistent JSON-RPC daemon to avoid continuous process-spawning overhead.

## 2. Strict Architectural Invariants

Whenever you write or modify code in this repository, you **MUST** adhere to the following rules:

- **No CGO:** The project must compile completely statically to run in a scratch container. Always use `CGO_ENABLED=0`.
- **Zero Dependencies:** Avoid external Go modules unless absolutely necessary (e.g., a complex shell interpreter later on). Do not use external libraries for flag parsing, colors, or utility logic.
- **Unified Flag Parsing:** Use the custom POSIX-compliant parser in `pkg/common/flags.go` (`common.ParseFlags`). **Do not use the standard library `flag` package** or `pflag`. Our parser supports short flag grouping (`-laR`) and standard POSIX conventions.
- **Standardized Output:** Use the `common.Render()` and `common.RenderError()` functions in `pkg/common/output.go` to handle both standard text output and `--json` structured output.
- **Multicall Dispatch:** Every utility lives in its own package under `pkg/` (e.g., `pkg/ls`, `pkg/echo`). Utilities register themselves automatically by calling `dispatch.Register()` in their `init()` function.

## 3. Component Structure

- `cmd/korego/main.go`: The multicall entry point. Handles symlink invocation (e.g., `/bin/ls -> /bin/korego`) and subcommand invocation (`korego ls`).
- `internal/dispatch/`: The command registry.
- `pkg/common/`: Foundation libraries (flags, JSON envelope, JSON-RPC types).
- `pkg/<utility>/`: Implementation of specific POSIX utilities (e.g., `pkg/cat/`, `pkg/ls/`).
- `test/compliance/`: Bash scripts that compare KoreGo's output and exit codes against the host OS (GNU/Linux) equivalents.
- `docker/`: Dockerfiles for the production `scratch` image and the `alpine` debug image.

## 4. Development Workflow

When implementing a new utility or feature, follow this checklist:

1. **Implement the Logic:** Write the utility in `pkg/<name>/<name>.go`.
2. **Library Layer vs CLI Layer:** Separate the core logic from the CLI parsing/printing so the core logic can be tested and reused easily by the JSON-RPC daemon.
3. **Unit Tests:** Write robust unit tests in `pkg/<name>/<name>_test.go` targeting > 85% coverage.
4. **Compliance Tests:** Add a test script in `test/compliance/test_<name>.sh`. Use `set -uo pipefail` (do NOT use `set -e`, as non-zero exit codes from utilities are expected and should be captured).
5. **Registration:** 
   - Add a blank import for the package in `cmd/korego/main.go`.
   - Add the package to the `PKG_DIRS` variable in the `Makefile`.
   - Add the compliance script to the `compliance` target in the `Makefile`.
6. **Verification:**
   - Run `make all` to build and run unit tests.
   - Run `make compliance` to verify POSIX behavior against the system.
   - Run `make ci` to run the full pipeline including Docker builds.
7. **Documentation:** Update the corresponding Phase plan in the `wiki/` directory (e.g., check off the task list).

## 5. Security & Safety

- **Root Protection:** Utilities that perform destructive operations (like `rm`) must include guards against destroying the root filesystem (e.g., `rm -rf /` must be refused without `--no-preserve-root`).
- **Permissions:** Default to secure permissions. The Docker image runs as a non-root user (`korego:1000`).

## 6. Current State & Progression

Refer to the Phase documents in `wiki/` (e.g., `wiki/plan_updated.md`) to understand the current task.
- **Phase 00 & 01:** Foundation & Tier 1 (echo, true, false, env, pwd, etc.) — **COMPLETED**
- **Phase 02:** Docker CI & Scratch pipeline — **COMPLETED**
- **Phase 03:** Filesystem Utils (ls, cat, rm, cp, mv, etc.) — **COMPLETED**
- **Phase 04+:** See project roadmap.

**Always read the active Phase document before writing code!**
