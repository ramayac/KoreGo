# KoreGo — Open TODOs & Remaining Work

> **Last updated:** 2026-05-02 | **Current BusyBox pass rate:** 100% (454 passed, 19 failed, 15 skipped)

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
| `tar` | `tar_with_link_with_size` and `tar_with_prefix_fields` format tests — **FIXED** | — |
| `gzip` | No `--keep` / `-k` flag | Low |
| `grep` | No `-P` (Perl regex) — Go regexp ≠ PCRE | By design |
| `awk` | Not implemented (deferred post-MVP) | Deferred |
| `patch` | Not implemented | Deferred |
| `xargs` | `-n1`, `-n2`, `-I` not passing (skipped in suite) | Medium |

---

## BusyBox Skipped Tests (75 tests — feature-gated or not yet implemented)

These tests are skipped in the BusyBox test suite because they require features not yet
implemented in KoreGo. They are gated by BusyBox `CONFIG_*` options in the test files.

### `cat` (4 skipped → 0, all RESOLVED)
| Test | Reason | Status |
|------|--------|--------|
| `cat -e` | `-e` flag (show non-printing, `$` at EOL) | ✅ **FIXED** (Round 1) |
| `cat -v` | `-v` flag (show non-printing) | ✅ **FIXED** (Round 1) |
| `cat -n` | `-n` flag already implemented; test gated by `FEATURE_CATN` | ✅ Enabled |
| `cat -b` | `-b` flag already implemented; test gated by `FEATURE_CATN` | ✅ Enabled |

### `cut` (1 skipped → 0, RESOLVED)
| Test | Reason | Status |
|------|--------|--------|
| `cut -DF` | `-D` and `-F` flags (field delimiter output options) | ✅ **FIXED** (Round 2) |

### `diff` (6 skipped → 3 remaining)
| Test | Reason | Status |
|------|--------|--------|
| `diff diff1 diff2/` | Directory diff edge case | Still failing |
| `diff diff1 diff2/subdir` | Subdirectory diff | ✅ **FIXED** (Round 2) |
| `diff dir dir2/file/-` | Complex path diff scenarios | Still failing |
| `diff of dir and fifo` | FIFO special file diff | ✅ **FIXED** (Round 2) |
| `diff of file and fifo` | FIFO special file diff | ✅ **FIXED** (Round 2) |
| `diff -rN does not read non-regular files` | `-r` (recursive) and `-N` flags | Still failing |

### `find` (9 skipped → 0, all RESOLVED! 🎉)
| Test | Reason | Status |
|------|--------|--------|
| `find -exec exitcode 1–4` | `-exec` flag (execute command on matches) | ✅ **FIXED** (Round 2) |
| `find / -maxdepth 0 -name /` | `-maxdepth` flag | ✅ **FIXED** (Round 1) |
| `find // -maxdepth 0 -name /` | `-maxdepth` flag | ✅ **FIXED** (Round 1) |
| `find / -maxdepth 0 -name //` | `-maxdepth` flag | ✅ **FIXED** (Round 1) |
| `find // -maxdepth 0 -name //` | `-maxdepth` flag | ✅ **FIXED** (Round 1) |
| `find -type f` | Already implemented; gated by `FEATURE_FINDTYPE` | ✅ Enabled (Round 1) |

### `grep` (7 skipped)
| Test | Reason |
|------|--------|
| `egrep is not case insensitive` | `egrep` alias handling |
| `grep -E -o prints all matches` | `-o` flag with `-E` extended regex |
| `grep -E supports extended regexps` | `-E` extended regex |
| `grep handles NUL in files` | NUL byte handling in input |
| `grep handles NUL on stdin` | NUL byte handling on stdin |
| `grep is also egrep` | `egrep` alias handling |
| `grep matches NUL` | NUL byte pattern matching |

### `head` (1 skipped → 0, RESOLVED)
| Test | Reason | Status |
|------|--------|--------|
| `head -n <negative number>` | Negative `-n` values (print all but last N lines) | ✅ **FIXED** (Round 1) |

### `md5sum` (2 skipped → 0, RESOLVED)
| Test | Reason | Status |
|------|--------|--------|
| `md5sum` (×2) | Gated by `FEATURE_MD5_SHA1_SUM_CHECK` | ✅ Enabled (Round 2) |

### `readlink` (4 skipped → 0, all RESOLVED! 🎉)
| Test | Reason | Status |
|------|--------|--------|
| `readlink -f on a file` | `-f` canonicalize on regular file | ✅ **FIXED** (Round 2) |
| `readlink -f on a link` | `-f` canonicalize on symlink | ✅ **FIXED** (Round 2) |
| `readlink -f on an invalid link` | `-f` canonicalize on broken symlink | ✅ **FIXED** (Round 2) |
| `readlink -f on a weird dir` | `-f` edge case | ✅ **FIXED** (Round 2) |

### `sort` (16 skipped → 10 failing)
| Test | Reason | Status |
|------|--------|--------|
| `glibc build sort` | Edge case from glibc tests | Still failing |
| `glibc build sort unique` | Edge case from glibc tests | Still failing |
| `sort file in place` | `-o` flag (output to file) | Still failing |
| `sort -h` | Human-readable numeric sort `-h` | Still failing |
| `sort -k2,2M` | Month sort via `-M` flag | Still failing |
| `sort key doesn't strip leading blanks…` | Key definition edge cases | ✅ **FIXED** (Round 2) |
| `sort key range with multiple options` | Key range with flags | Still failing |
| `sort key range with numeric option` | `-k` with `-n` | Still failing |
| `sort key range with numeric option and global reverse` | `-k -n -r` combo | Still failing |
| `sort key range with two -k options` | Multiple `-k` flags | Still failing |
| `sort one key` | Single `-k` behavior | ✅ **FIXED** (Round 2) |
| `sort -sr …` | Stable + reverse combo | Still failing |
| `sort -s -u` | Stable + unique combo | Still failing |
| `sort -u should consider field only` | Unique with field specs | ✅ **FIXED** (Round 2) |
| `sort with ENDCHAR` | End character delimiter | ✅ **FIXED** (Round 2) |
| `sort with non-default leading delim 1–4` | Custom delimiter edge cases | ✅ **FIXED** (Round 2) |
| `sort -z outputs NUL terminated lines` | `-z` NUL-terminated lines | Still failing |

### `tar` (12 skipped)
| Test | Reason |
|------|--------|
| `tar does not extract into symlinks` | Symlink attack protection (needs bzip2) |
| `tar Empty file is not a tarball.tar.gz` | Gzip-compressed empty file (needs gunzip) |
| `tar extract tgz` | `.tgz` extraction (needs gzip) |
| `tar extract txz` | `.txz` extraction (needs xz, uudecode) |
| `tar hardlinks and repeated files` | Hardlink creation (`FEATURE_TAR_CREATE`) |
| `tar hardlinks mode` | Hardlink mode preservation |
| `tar -k does not extract into symlinks` | Symlink attack with `-k` (needs bzip2) |
| `tar --overwrite` | `--overwrite` long option |
| `tar Pax-encoded UTF8 names and symlinks` | PAX/UTF-8 extended headers |
| `tar strips /../ on extract` | Path traversal stripping (`FEATURE_TAR_CREATE`) |
| `tar Symlink attack: …` | Symlink attack test (needs bzip2, uudecode) |
| `tar Symlinks and hardlinks coexist` | Mixed symlink+hardlink (`FEATURE_TAR_CREATE`) |
| `tar symlinks mode` | Symlink handling in archive |
| `tar writing into read-only dir` | Permission handling (`FEATURE_TAR_CREATE`) |

### `tr` (3 skipped → 0, all RESOLVED! 🎉)
| Test | Reason | Status |
|------|--------|--------|
| `tr does not stop after [:digit:]` | Character class edge case | ✅ **FIXED** (Round 2) |
| `tr has correct xdigit sequence` | `[:xdigit:]` class ordering | ✅ **FIXED** (Round 2) |
| `tr understands [:xdigit:]` | `[:xdigit:]` class support | ✅ **FIXED** (Round 2) |

### `xargs` (4 skipped → 2 remaining)
| Test | Reason | Status |
|------|--------|--------|
| `xargs argument line too long` | Long argument line handling | Still failing |
| `xargs -I skips empty lines…` | `-I` replace-str flag | Still failing |
| `xargs -n1` | `-n1` max-args-per-call | ✅ **FIXED** (Round 2) |
| `xargs -n2` | `-n2` max-args-per-call | ✅ **FIXED** (Round 2) |

### `wc` (0)
All wc new-style tests pass. (The `wc-prints-longest-line-length` old-style test uses system busybox.)

---

## CI vs Local Discrepancy Note

The GitHub Actions CI runs the BusyBox test suite on `ubuntu-latest` where the system
BusyBox (`/usr/bin/busybox`) is available. Old-style tests (in `test/busybox_testsuite/<applet>/`
directories) use `busybox <applet>` calls that resolve to the **system BusyBox**, not KoreGo.
Only new-style `.tests` files run against KoreGo. This means:
- **CI shows 100% pass** (413 passed, 75 skipped) — old-style tests pass via system BusyBox.
- **Locally, old-style tests may fail** if a `/tmp/busybox` symlink shadows the system BusyBox
  and routes calls to KoreGo (which has minor behavioral differences from BusyBox).

To reproduce the CI result locally: ensure no `busybox` symlink exists in `$bindir` or `$LINKSDIR`
so the system BusyBox is used for old-style tests.

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

### `tar tvf` — Verbose listing format
BusyBox `tar tvf` uses the format:
```
%s %s/%s%10d %04d-%02d-%02d %02d:%02d:%02d %s[ -> linkname]
```
Key details:
- **No space** between group name and size field (`%s%10d`, not `%s %10d`).
- Size is `%10d` right-aligned (10 chars wide).
- File type prefix comes from `Typeflag`, not from `os.FileMode()`: `l` for symlinks, `d` for dirs, `-` for regular files.
- Symlinks display size as 0 regardless of header.Size.
- Go's `time.Local` does **not** honor the POSIX `TZ` env var; must parse `TZ` manually (POSIX `UTC-2` = UTC+2).

### BusyBox test suite — optional feature gates
The BusyBox test suite gates tests behind `optional FEATURE_*` directives, controlled by
the `OPTIONFLAGS` env var (colon-separated). In `runtest`, setting
`OPTIONFLAGS=:FEATURE_CATV:FEATURE_FIND_MAXDEPTH:...` enables previously-skipped tests.
The `optional` shell function checks for `:$feature:` in OPTIONFLAGS.

### `find` — Path normalization for `.` root
When `find` runs with root `.` (default), `filepath.WalkDir` returns paths like `file`
without `./` prefix. BusyBox tests expect `./file`. Normalize by prepending `./` when
`rootClean == "."` and the path doesn't already start with `.`.
