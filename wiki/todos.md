# KoreGo — Open TODOs & Remaining Work

> **Last updated:** 2026-05-01 | **Current BusyBox pass rate:** ~95% (all Phase C/D targets resolved)

This document tracks remaining failing tests, known deviations, and future improvements.
See [10_posix_framework.md](10_posix_framework.md) for the full Phase 10 task log.

---

## Phase C: `tar` (7 failures) — **COMPLETED (2026-05-01)**

### C.1 — `-X` Exclude File Flag ✅
**File:** `pkg/tar/tar.go`

Implemented:
- Registered `-X` as a `FlagValue` type (repeatable) in flag spec.
- `readExcludeFile()` reads each `-X` file line-by-line into an exclusion set.
- `isExcluded()` checks archive entries against exclusion patterns (supports nested paths via prefix matching).
- Supports multiple `-X` flags via `flags.GetAll("X")`.

### C.2 — Stdin tarball (`-f -`) and zeroed-block detection ✅
**File:** `pkg/tar/tar.go`

Implemented:
- Old-style tar flag preprocessing (`xvf` → `-x -v -f`).
- `bufio.Reader.Peek(1)` to detect empty streams (0 bytes → "short read", exit 1).
- Two or more 512-byte zero blocks are treated as valid empty tarball (exit 0).
- Default to stdin when no `-f` specified for extract/list modes.
- Also added: `-O` flag (extract to stdout), `-C` archive path resolution fix, `busybox` alias in main dispatcher, and include-list normalization (stripping `./` prefix).

### C.3 — BusyBox Test Harness Integration ✅
**File:** `test/busybox_testsuite/runtest`, `cmd/korego/main.go`

Fixed:
- Added `busybox` symlink to korego in the test runner link directory.
- Added `LINKSDIR` to PATH in old-style test runner for `busybox` resolution.
- Added `busybox` alias in main dispatcher (`name == "busybox"` → subcommand mode).
- Exported `LINKSDIR` environment variable.

---

## Phase D: `gzip` (1 failure) — **COMPLETED (2026-05-01)**

### D.1 — Numeric Compression Levels (`-1` to `-9`) ✅
**File:** `pkg/gzip/gzip.go`

Implemented:
- Added `-1` through `-9` as boolean flags in flag spec.
- `getCompressionLevel()` maps detected flags to Go's `compress/flate` levels (1–9).
- `gzip.NewWriterLevel()` used instead of `gzip.NewWriter()` when level specified.
- Also added: `-` as stdin/stdout handling for gzip positional args.

---

## Known Deviations / Future Work

These are known differences from GNU/BusyBox behavior that are low-priority or by design:

| Utility | Deviation | Priority |
|---------|-----------|----------|
| `tar` | No support for `--overwrite`, pax headers, xz/bzip2 | Low |
| `tar` | Hard links and symlink mode not fully verified | Low |
| `tar` | `tar_with_link_with_size` and `tar_with_prefix_fields` format tests fail (symlink display) | Low |
| `gzip` | No `--keep` / `-k` flag | Low |
| `grep` | No `-P` (Perl regex) — Go regexp ≠ PCRE | By design |
| `awk` | Not implemented (deferred post-MVP) | Deferred |
| `patch` | Not implemented | Deferred |
| `xargs` | `-n1`, `-n2`, `-I` not passing (skipped in suite) | Medium |

---

## Session Insights (2026-05-01)

> These are hard-won lessons from implementing Phase C and D fixes. They augment the entries in `AGENTS.md § 8`.

### `tar` — Old-style flag preprocessing
Traditional `tar` accepts flags without a leading dash (e.g., `xvf` means `-x -v -f`). BusyBox tests rely on this. A preprocessing step must expand old-style flag bundles before calling `common.ParseFlags`. The mode char (c/x/t/r/u) can appear anywhere in the bundle (e.g., `Ox` = `-O -x`).

### `tar` — Empty archive detection
Go's `archive/tar` returns `io.EOF` (not `io.ErrUnexpectedEOF`) for a completely empty reader. To detect "not a tarball" early, use `bufio.Reader.Peek(1)` before creating the tar reader.

### `tar` — Archive path resolution with `-C`
When `-C dir` changes the working directory, relative archive paths must be resolved to absolute before the chdir, otherwise `os.Open(archive)` looks in the wrong directory.

### `tar` — Include list normalization
Archive entries created by `tar -C foo .` store paths like `1/10` (not `./1/10`). Include list entries like `./1/10` must be normalized by stripping the leading `./` prefix before matching.

### `tar` — Missing include file error
When an include list is provided (positional args to extract), and no archive entries match, tar must exit non-zero with an error message.

### `gzip` — `-` as stdin/stdout in positional args
When `-` appears as a positional file argument, gzip must handle it as stdin (decompress) or stdout (compress), not attempt `os.Stat("-")`.

### `wc` — No leading padding
POSIX does not mandate the `%7d` padded format for `wc` output. The BusyBox test suite does exact string comparisons, so any leading spaces in column output will cause failures. Use `%d` without width specifiers.

### `diff` — Unified hunk range edge cases
The POSIX unified diff hunk header format has these rules (also followed by GNU diff):
- **Count of 1:** omit the `,count` — write `@@ -5 +5 @@` not `@@ -5,1 +5,1 @@`
- **Count of 0:** write `start-1,0` — e.g. `@@ -3,0 +3,0 @@` for an insert-only hunk after line 3
- **No newline marker:** emit `\ No newline at end of file` on its own line immediately after the last `+` or `-` line of the last hunk if the source/dest file has no trailing newline.

### `diff -B` — Correct `differ` flag logic
When `-B` (ignore blank line changes) is active, a file is only considered to "differ" if there are non-blank-line changes remaining after filtering. The `differ` boolean must be computed **after** `filterBlankLineChanges()` runs, not before. Failing to do this causes `diff -qB` to report "Files differ" even when the only change is blank lines.

### `diff` — Stdin (`-`) argument
`os.ReadFile("-")` tries to open a file literally named `-` and fails. When a file argument is `-`, use `io.ReadAll(os.Stdin)` instead. Both file arguments can independently be `-`, but if both are `-` and they are the same string, the files are identical — short-circuit to exit 0.

### `cp` — Symlink mode semantics
GNU `cp` symlink flag semantics (from highest to lowest precedence):

| Flag | Mode | Command-line symlinks | Internal (recursive) symlinks |
|------|----|---------------------|-------------------------------|
| `-L` | `SymlinkFollow` | Dereference | Dereference |
| `-H` | `SymlinkFollowArgs` | Dereference | Preserve |
| `-P` / `-d` | `SymlinkPreserve` | Preserve | Preserve |
| *(default, no `-R`)* | `SymlinkFollow` | Dereference | N/A |
| *(default, with `-R`)* | `SymlinkPreserve` | Preserve | Preserve |

Key insight: **all positional source arguments are "command-line arguments"** for the purpose of `-H`. Mark every source as `isArg=true`, not just the first one.

### `cp` — Copying a symlink as a symlink
When preserving symlinks (mode is `SymlinkPreserve`), do **not** call `os.Open` on the symlink. Instead:
1. `os.Lstat` to confirm it's a symlink.
2. `os.Readlink` to get the target string.
3. `os.Remove` the destination (if it exists).
4. `os.Symlink(target, dst)` to create the new symlink.

Using `os.Open` on a symlink will silently follow it and copy the underlying file contents, not the symlink itself.

### `tar -X` — Repeatable flags
The `-X` flag must be registered as repeatable (accepting multiple values). When parsing, accumulate all values from multiple `-X` occurrences into a slice. This is distinct from `-f` which takes a single value.
