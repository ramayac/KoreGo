# Prepare to Goose — GoPOSIX Changes for GoPOSIXOS

> **Status:** COMPLETED | **Date:** 2026-05-16 | **Blocks:** GoPOSIXOS M0

---

## Context

GoPOSIXOS is a separate repo that imports GoPOSIX as a Go module and extends it with boot/system utilities (mount, init, getty, etc.). This document tracks what needs to change **in this repo (GoPOSIX)** before GoPOSIXOS can compose GoPOSIX cleanly.

**Good news:** the public API (`goposix.go`) already exports `WellKnownNames`, `Main()`, `Run()`, and supports `--list-commands` / `--version`. Most of the foundational work is done.

---

## Change 1: Register a `shell` Command ⚡

**Status: NEEDED** | **Effort: ~100 LOC** | **New files: 2** | **Prereq for: M0**

### Problem

`internal/shell/interpreter.go` exists and works — it wraps `mvdan.cc/sh/v3`, enforces resource limits (128MB memory/stream, 30s timeout), path confinement via `SecurePath`, and delegates builtins to GoPOSIX dispatch. But it's **not registered as a CLI command**. There's no way to run `goposixos shell /etc/rc` during boot.

Additionally, `internal/` packages are **not importable** by external Go modules — GoPOSIXOS can't use the interpreter library directly.

### What to build

Create `pkg/shell/shell.go` — a thin CLI wrapper that registers `shell` (and `sh`) as dispatch commands:

```
pkg/shell/
├── shell.go       # CLI wrapper: dispatch.Register("shell") + dispatch.Register("sh")
└── shell_test.go  # Tests: script file execution, stdin pipe, shebang handling
```

### Behavior

The `shell` command should handle three modes:

| Mode | Trigger | Behavior |
|------|---------|----------|
| Script file | `goposixos shell /etc/rc` | Read file, execute with `internal/shell.Exec()` |
| Stdin pipe | `echo "ls" \| goposixos shell` | Read stdin fully, execute |
| Interactive | stdin is a terminal (no args) | REPL: read line, execute, print, loop |

### Shebang quirk to handle

The Linux kernel has a well-known shebang limitation: everything after `#!` is passed as a **single argument** with a leading space.

```sh
#!/bin/goposixos shell
```
→ kernel calls `exec("/bin/goposixos", " shell", "/etc/rc")`

The dispatch sees argv[1] = `" shell"` (note the space). This won't match `"shell"`.

**Fix:** In `pkg/shell/shell.go`, trim leading whitespace from argv[0] before comparison:

```go
// Handle kernel shebang concatenation quirk:
//   #!/bin/goposixos shell  →  argv[1] = " shell"
cmdName := strings.TrimSpace(argv[0])
if cmdName == "shell" || cmdName == "sh" {
    // ...
}
```

**Better alternative:** Skip the shebang entirely. Have `init` invoke the shell explicitly:

```go
// In pkg/init/init.go (GoPOSIXOS):
cmd := exec.Command("goposixos", "shell", "/etc/rc")
```

No shebang quirk to worry about. Recommended approach, but support both.

### Why not move `internal/shell` to `pkg/shell`?

The `internal/shell` package has a dependency on `mvdan.cc/sh/v3`. Keeping it internal and wrapping it with a dispatch-registered command means:
1. The `mvdan.cc/sh` dependency stays contained — no public API surface exposes it
2. GoPOSIXOS gets the shell via blank import of `pkg/shell` — never touches `mvdan.cc/sh` directly
3. If we ever replace the interpreter engine, only `internal/shell` changes

---

## Change 2: Blank Import Shell in `cmd/goposix/main.go` ⚡

**Status: NEEDED** | **Effort: 1 line** | **Modifies: 1 file**

Add to `cmd/goposix/main.go`:

```go
_ "github.com/ramayac/goposix/pkg/shell"
```

This registers `shell` and `sh` in the multicall binary. Without this, GoPOSIXOS won't inherit the shell command.

---

## Change 3: Daemon Boot-Time Readiness ☑️

**Status: ALREADY DONE** | **Zero changes needed**

`pkg/daemon/daemon.go` is registered, supports `--socket`, `--workers`, and `--listen-addr`. GoPOSIXOS starts it with:

```sh
goposixos daemon --socket /run/goposix.sock &
```

Nothing to change. ✅

---

## Change 4: Ensure `chown`/`chgrp` Look Up Users by Name (not just UID) ❓

**Status: VERIFY** | **Prereq for: M1 (login/passwd)**

GoPOSIXOS will have `/etc/passwd` and `/etc/shadow`. `login` sets uid/gid, `chown` changes ownership. Verify that `chown` and `chgrp` resolve user/group **names** (not just numeric IDs) from the local passwd/group files. If they currently only accept numeric UIDs, this needs to be added.

```bash
# Quick check:
goposix chown root:root /tmp/test  # should work if name lookup exists
```

---

## Change 5: Shell Timeout Env Var ❓

**Status: VERIFY** | **Effort: docs only**

`internal/shell/interpreter.go` respects `GOPOSIX_SHELL_TIMEOUT` env var. GoPOSIXOS's init should set this in the environment before executing `/etc/rc` to prevent a hung boot. This is already implemented — no code change, just document it.

---

## Summary of Changes

| # | What | New Files | Modified Files | Effort |
|---|------|-----------|----------------|--------|
| 1 | `pkg/shell/` — CLI wrapper for shell interpreter | 2 (shell.go, shell_test.go) | — | ~100 LOC |
| 2 | Blank import shell in `cmd/goposix/main.go` | — | 1 | 1 line |
| 3 | Daemon readiness | — | — | Already done |
| 4 | chown/chgrp name lookup | — | VRFY | VRFY |
| 5 | Shell timeout docs | — | — | Already done |

**Total effort:** ~2–3 hours of coding + verification.

---

## What GoPOSIXOS Gets For Free (No GoPOSIX Changes)

These already work via blank import + `goposix.Main()`:

- All 50+ POSIX utilities (ls, cat, grep, sed, sort, etc.)
- `--json` output on every utility
- `--list-commands` for symlink generation
- `--version` + `goposix.Version` (settable via ldflags)
- JSON-RPC daemon (`goposixos daemon`)
- `WellKnownNames` extension for subcommand dispatch

---

## Verification Checklist

Before declaring GoPOSIX "Goose-ready":

- [x] `go test ./pkg/shell/...` passes
- [x] `make all` passes (build + unit tests)
- [x] `make testsuite` passes (477 passed, 3 failed — same baseline; no regressions)
- [x] `goposix shell -c "echo hello"` outputs "hello"
- [x] `goposix shell /path/to/script.sh` executes a script file
- [x] `echo "ls /bin" | goposix shell` works via stdin
- [x] `goposix --list-commands` includes `shell`
- [x] `chown root:root /tmp/test` resolves root by name (permission denied is expected for non-root)

---

## Related Documents

- [goposixos.md](goposixos.md) — GoPOSIXOS design (moving to goposixos repo)
- [goposix.go](../goposix.go) — Public API (already supports WellKnownNames, Run, Main)
- [internal/shell/interpreter.go](../internal/shell/interpreter.go) — Shell interpreter engine
- [05_daemon_core.md](05_daemon_core.md) — JSON-RPC daemon design
- [12_road_to_gold.md](12_road_to_gold.md) — Current stabilization state
