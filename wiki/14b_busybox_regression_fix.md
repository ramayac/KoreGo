# BusyBox Test Suite Regression Fix — 2026-05-15

## Final Result: 79 → 3 failures (96.2% pass rate)

After a two-phase investigation and fix session, 76 of 79 failing BusyBox tests were resolved. Three remain, all in the `date` utility due to Go runtime limitations.

## The Big Lesson

**Nearly all failures traced back to ONE architectural mistake:** `common.ParseFlags` applied uniformly to all utilities, with no escape hatch for tools where arguments can be arbitrary text starting with `-`. When echo/printf/expr crash on "flags" that are actually data, the BusyBox test harness (the only integration test) cascades failures into downstream utilities.

### Key Architectural Findings

1. **Shared infrastructure without guardrails.** `common.ParseFlags` treats every `-` arg as a flag. Utilities like echo, printf, and expr need a "stop at first non-flag" mode. Manual flag parsing is the correct fix for free-form utilities.

2. **`-j` short flag for `--json` is dangerous.** It collides with `tar -j` (bzip2), and any utility where `-j` could be legitimate data. Use long-form `--json` only.

3. **Integration tests catch cascading failures.** The BusyBox suite chains utilities: echo creates files → diff compares → ls lists → find verifies. A bug in echo silently broke 12 diff tests. Without per-commit gating, regressions go undetected.

## Root Causes & Fixes

### 1. Echo → Diff (12 failures)

**Bug:** `echo` used `common.ParseFlags` which treated `---` as a flag group.

**Fix:** Manual flag parsing (`parseEchoFlags`). Only `-n`, `-e`, `-E`, `--json` recognized. Unknown chars in flag groups cause atomic fallback to literal text. Octal escapes now handle `\0` as marker (not a digit), supporting `\41`, `\041`, `\0041`, `\00041`.

### 2. Printf (21 → 0 failures)

**Bug:** `common.ParseFlags` treated `-5` as a flag. Format parser missing: `\c` stop, `*` width/precision, length modifiers, format reuse, `%b`, character constants, negative width/precision, error handling.

**Fix:** Manual flag parsing. Complete format parser rewrite: `\c` at Format level, `*` width/precision from args, length modifier stripping (`z`,`l`,`L`,`h`), format reuse loop, `%b` escape processing, `'"x` char constants, interleaved error output, bare `%`/`%r` abort.

### 3. Test/`[` Utility (9 → 0 failures)

**Bug:** Expression parser lacked lookahead — `!` and `(` treated as operators even when followed by binary ops (`=`, `==`). `!` alone errored; `-f` alone errored.

**Fix:** Lookahead in `parseNot` and `parsePrimary` — if next token is a binary operator, treat `!`/`(` as string literals. `!` at end of expression negates empty (false) → true.

### 4. cp (5 → 0 failures)

- Missing `-a` (archive = `-dRp`) and `--parents` flags
- **`devID()` formatted the Sys() pointer** instead of dereferencing to `Dev:Ino` → hard link tracking never matched across `Lstat` calls

### 5. du (4 → 0 failures)

- Missing `-k`, `-m`, `-l` flags
- `-h` output format: needed one-decimal formatting (`1.0M` not `1M`)
- Added inode deduplication (skip hard links unless `-l`)

### 6. ls (4 → 0 failures)

All 4 failures caused by `diff -w` not being supported. Fixes also needed:
- Added `-w`/`--ignore-all-space` flag to diff
- Removed `total` header for single-file arguments
- Preserved user-specified path in filename column (GNU ls behavior)
- Don't show directory header for file arguments

### 7. mv (4 → 0 failures)

Test failures were false positives — caused by missing features in supporting utilities:
- **touch** needed `-d` date flag
- **chmod** needed POSIX symbolic modes (`a-r`, `u+x`)
- **mv** needed `-t` target-directory flag

### 8. gunzip (3 → 0 failures)

- Hardcoded `"gzip:"` prefix in errors; needed command-name-aware messages
- Missing `.gz` suffix check before decompression
- Wrong "already exists" format for gunzip vs gzip

### 9. date (6 → 3 remaining)

Fixed: `-d` flag, `+FORMAT` strftime support, date parsing, extra arg rejection.

Remaining (3):
| Test | Root Cause | Status |
|------|-----------|--------|
| `date-@-works` | Go `time` doesn't parse POSIX TZ strings (`EET-2EEST,M3.5.0/3`) | ❌ Unfixable |
| `date-timezone` | Same Go TZ limitation | ❌ Unfixable |
| `date-works-1` | Error format: korego says `date: invalid date` but test expects BusyBox help banner | ⚠️ Cosmetic |

### 10. Single-failure utilities (all → 0)

| Utility | Bug | Fix |
|---------|-----|-----|
| basename | `TrimSuffix` stripped even when suffix == whole string | Guard: only strip if suffix ≠ base |
| cat | Stdin read twice when `-` in args | Removed `|| flags.Stdin` from default-stdin check |
| expr | `ParseFlags` on negative numbers | Manual flag parsing |
| find | Missing `-xdev` and `-mtime` flags | Added to spec |
| grep | Missing filename prefix on `-r` single-path | Set prefix when recursive |
| hostname | `-d` output `(none)`, `-f` appended extra dot | Return empty string for no-domain; don't append dot |
| tail | `-2` parsed as unknown flag | Preprocess `-N` → `-n N` |
| tar | `ls -l file` showed `total` header | Fixed `isSingleFile` detection in ls |
| touch | Missing `-c` (no-create) flag | Added flag + skip logic |
| wc | Missing `-L` (max-line-length) | Added tracking + output |
| xargs | Missing `-0` (null-delimited) flag | Added `scanNulls` scanner |

## Test Hardening

~75 new unit tests added across 17 packages, each replicating a specific BusyBox test failure.
All tests follow the pattern: understand the BusyBox test → write Go unit test that reproduces → fix → validate.

## Key Diff Changes

```diff
# echo.go: ParseFlags → manual parsing
- var spec = common.FlagSpec{
-     Defs: []common.FlagDef{{Short: "n"}, {Short: "e"}, {Short: "j", Long: "json"}},
- }
- flags, err := common.ParseFlags(args, spec)
+ jsonMode, noNewline, escape, words := parseEchoFlags(args)

# cp.go: devID pointer → struct dereference
- return fmt.Sprintf("%v", fi.Sys())           // pointer address!
+ st := fi.Sys().(*syscall.Stat_t)
+ return fmt.Sprintf("%d:%d", st.Dev, st.Ino)  // actual inode identity
```
