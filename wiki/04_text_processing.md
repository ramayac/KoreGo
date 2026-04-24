# Phase 04 — Tier 3: Text Processing Utilities

> **Timeline:** Week 4–5 | **Depends on:** Phase 03

---

## Goal

Implement the 10 text processing utilities. These are critical for agentic pipelines.

## Utilities

### 04.1 — `head` / `tail`

- `head -n 10`, `tail -n 10`, `tail -f` (follow)
- `--json` → `{"lines":[...], "lineCount": N}`

### 04.2 — `wc`

- Flags: `-l` (lines), `-w` (words), `-c` (bytes), `-m` (chars)
- `--json` → `{"lines":N, "words":N, "bytes":N, "chars":N}`
- Multiple files: per-file + total

### 04.3 — `sort`

- Flags: `-r` (reverse), `-n` (numeric), `-u` (unique), `-k` (key field), `-t` (delimiter)
- `--json` → `{"lines":[...], "count":N}`

### 04.4 — `uniq`

- Flags: `-c` (count), `-d` (duplicates only), `-u` (unique only), `-i` (case insensitive)
- `--json` → `[{"line":"text", "count":N}]`

### 04.5 — `tr`

- `tr 'a-z' 'A-Z'` — character translation
- Flags: `-d` (delete), `-s` (squeeze), `-c` (complement)
- No `--json` (streaming character transform)

### 04.6 — `cut`

- Flags: `-f` (fields), `-d` (delimiter), `-c` (characters), `-b` (bytes)
- `--json` → `{"lines":[{"fields":["a","b","c"]}]}`

### 04.7 — `tee`

- Read stdin, write to stdout AND files
- Flags: `-a` (append)
- No `--json` (passthrough tool)

### 04.8 — `grep`

**Second most complex utility.**

- Flags: `-i`, `-v`, `-c`, `-n`, `-l`, `-r`, `-E` (ERE), `-F` (fixed), `-w`, `-x`, `-A`/`-B`/`-C` (context)
- `--json` → `[{"file":"f", "line":N, "text":"match", "matches":["substr"]}]`
- **Known risk:** Go `regexp` uses RE2 (no backreferences). Document difference from POSIX BRE/ERE.

### 04.9 — `sed`

**Most complex text utility.** Implement incrementally:
1. First: `s/pattern/replacement/flags` (substitute)
2. Then: address ranges (`1,5s/...`, `/pattern/s/...`)
3. Then: `d` (delete), `p` (print), `q` (quit)
4. Later: hold space, multi-line (`N`, `H`, `G`)

- Flags: `-i` (in-place), `-n` (suppress default print), `-e` (expression)
- **Known risk:** POSIX BRE vs Go RE2. May need custom BRE parser.

### 04.10 — `paste` / `join` (bonus)

- `paste -d',' file1 file2` — merge lines
- Lower priority, add if time allows

## Milestone 04

- [x] `echo "hello world" | korego wc --json` → `{"lines":1,"words":2,"bytes":12}`
- [x] `korego grep --json -rn "TODO" ./` returns structured matches
- [x] `korego sort -rn file.txt` sorts numerically in reverse
- [x] `korego sed 's/old/new/g' file.txt` performs substitution
- [x] All Tier 3 utilities have >80% test coverage
- [x] Pipes work: `korego cat f | korego grep pat | korego wc -l`

## How to Verify

```bash
# Unit tests
go test -v -cover ./pkg/grep/ ./pkg/sed/ ./pkg/wc/ ...

# Pipe chain
echo -e "foo\nbar\nfoo\nbaz" | ./korego sort | ./korego uniq --json
# → {"lines":["bar","baz","foo"],"count":3}

# grep JSON
./korego grep --json -rn "func " ./pkg/
```
