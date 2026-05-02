# KoreGo — Open TODOs & Remaining Work

> **Last updated:** 2026-05-02 | **Current BusyBox pass rate:** 405/413 (~98%)

This document tracks remaining failing tests, known deviations, and future improvements.
See [10_posix_framework.md](10_posix_framework.md) for the full Phase 10 task log.

---

## Phase C: `tar` (5 remaining failures)

### C.1 — `-X` Exclude File Flag
**File:** `pkg/tar/tar.go`

Four tests fail because `korego tar` does not support the `-X <file>` flag (read exclude patterns from a file):

```
FAIL: tar-handles-empty-include-and-non-empty-exclude-list
FAIL: tar-handles-exclude-and-extract-lists
FAIL: tar-handles-multiple-X-options
FAIL: tar-handles-nested-exclude
```

**What's needed:**
- Register `-X` as a `FlagValue` type (repeatable).
- During extraction, read each `-X` file line-by-line into an exclusion set.
- Skip any archive entry whose path matches any entry in the exclusion set.
- Support multiple `-X` flags (e.g. `-X foo.exclude -X bar.exclude`).
- Support nested paths (e.g. an exclude file containing `foo/bar`).

**Test pattern:**
```bash
tar xf foo.tar -X foo.exclude        # single exclude file
tar xf foo.tar foo bar -X foo.exclude # exclude + include list
tar xf foo.tar -X foo.exclude -X bar.exclude # multiple exclude files
```

---

### C.2 — Stdin tarball (`-f -`) and zeroed-block detection
**File:** `pkg/tar/tar.go`

Three tests fail because `korego tar` errors with `"missing archive file (-f)"` instead of reading from stdin when `-f -` is specified:

```
FAIL: tar Empty file is not a tarball        → expects: "tar: short read", exit 1
FAIL: tar Two zeroed blocks is a ('truncated') empty tarball   → expects: exit 0
FAIL: tar Twenty zeroed blocks is an empty tarball             → expects: exit 0
```

**What's needed:**
- When `-f -` is specified, read the tarball from `os.Stdin` instead of erroring.
- Handle the POSIX "end-of-archive" format: two consecutive 512-byte zero blocks should be treated as a valid (empty) tarball — exit 0, extract nothing.
- An actual empty file (0 bytes) should emit `tar: short read` and exit 1.

**Test pattern:**
```bash
dd if=/dev/zero bs=512 count=2 | tar xvf -   # Two zeroed blocks → OK (exit 0)
dd if=/dev/zero bs=512 count=20 | tar xvf -  # Twenty zeroed blocks → OK (exit 0)
: | tar xvf -                                 # Empty stream → "short read" (exit 1)
```

---

## Phase D: `gzip` (1 remaining failure)

### D.1 — Numeric Compression Levels (`-1` to `-9`)
**File:** `pkg/gzip/gzip.go`

One test fails because all compression levels produce identical output sizes:

```
FAIL: gzip-compression-levels
```

**What's needed:**
- Parse `-1` through `-9` as short flags (these are single-char flags, not long options).
- Map them to Go's `compress/flate` levels:
  - `-1` → `flate.BestSpeed` (1)
  - `-9` → `flate.BestCompression` (9)
  - `-2` through `-8` → intermediate levels (Go's flate accepts 1–9 directly).
- The test asserts that `-1` output **size > `-9`** output size on a real binary (e.g. `/usr/bin/busybox`). This naturally holds when levels are wired correctly.

**Flag parsing note:** Numeric flags like `-1` look like negative numbers to some parsers. The `common.ParseFlags` custom parser must treat `-1`..`-9` as boolean flags named `"1"`..`"9"` (or a special `compressLevel` value). Verify the parser handles this without confusing them with negative numeric arguments.

---

## Known Deviations / Future Work

These are known differences from GNU/BusyBox behavior that are low-priority or by design:

| Utility | Deviation | Priority |
|---------|-----------|----------|
| `tar` | No support for `--overwrite`, pax headers, xz/bzip2 | Low |
| `tar` | Hard links and symlink mode not fully verified | Low |
| `gzip` | No `--keep` / `-k` flag | Low |
| `grep` | No `-P` (Perl regex) — Go regexp ≠ PCRE | By design |
| `awk` | Not implemented (deferred post-MVP) | Deferred |
| `patch` | Not implemented | Deferred |
| `xargs` | `-n1`, `-n2`, `-I` not passing (skipped in suite) | Medium |

---

## Session Insights (2026-05-02)

> These are hard-won lessons from the current debugging session. They augment the entries in `AGENTS.md § 8`.

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
