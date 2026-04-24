# Phase 03 — Tier 2: Filesystem Utilities

> **Timeline:** Week 3–4 | **Depends on:** Phase 00, 01 | **Blocks:** Phase 04, 05

---

## Goal

Implement the 13 core filesystem utilities with full POSIX flag compatibility and `--json` output.

---

## Utilities

### 03.1 — `ls` (`pkg/ls/`)

**Most complex utility in the project.** Get it right and everything else follows.

**Library returns:** `[]FileInfo{Name, Path, Size, Mode, ModTime, IsDir, Owner, Group, Inode, Links, Target}`

**Flags:** `-a`, `-A`, `-l`, `-R`, `-h`, `-t`, `-r`, `-S`, `-1`, `-d`, `-i`, `-s`

**`--json` output:**
```json
{"command":"ls","data":{"path":"/tmp","files":[{"name":"foo.txt","size":1024,"mode":"-rw-r--r--","isDir":false}],"total":3},"exitCode":0}
```

**Tests:**
- [ ] `-a` includes `.` and `..`; `-A` includes dotfiles but NOT `.`/`..`
- [ ] `-laR` grouped flags work (tests POSIX flag parser)
- [ ] Symlinks show target
- [ ] Exit code 2 on "no such file"
- [ ] Compare output against `/bin/ls` for same directory

### 03.2 — `cat` (`pkg/cat/`)

Streaming reader. Flags: `-n` (number lines), `-b` (number non-blank), `-s` (squeeze blank).

Must handle stdin (`cat -` or `cat` with no args) and pipes.

`--json` for small files: `{"lines":[...], "lineCount": N}`. For large files: NDJSON stream.

- [ ] `echo hello | cat` reads stdin
- [ ] `cat file1 file2` concatenates
- [ ] `cat nonexistent` exits 1 with error to stderr

### 03.3 — `mkdir`

- [ ] `-p` creates parents
- [ ] `-m 0755` sets permissions
- [ ] `--json` → `{"created": ["/path/a", "/path/b"]}`
- [ ] Exit 1 if exists (without `-p`)

### 03.4 — `rmdir`

- [ ] Removes empty directories only
- [ ] `-p` removes parents
- [ ] Exit 1 if not empty

### 03.5 — `rm`

- [ ] `-r` recursive, `-f` force, `-v` verbose
- [ ] **Safety: `rm -rf /` REFUSED** without `--no-preserve-root`
- [ ] `--json` → `{"removed": [...], "errors": [...]}`

### 03.6 — `cp`

- [ ] `-r` recursive, `-p` preserve permissions/timestamps
- [ ] `-i` interactive (prompt before overwrite)
- [ ] `--json` → `{"copied": [{"from":"...", "to":"..."}]}`

### 03.7 — `mv`

- [ ] Rename or move, `-f` force, `-i` interactive
- [ ] Cross-device move: copy + delete fallback
- [ ] `--json` → `{"moved": [{"from":"...", "to":"..."}]}`

### 03.8 — `touch`

- [ ] Creates file or updates mtime
- [ ] `-t` specific timestamp, `-r ref` reference file's time

### 03.9 — `ln`

- [ ] Hard link by default, `-s` for symbolic
- [ ] `-f` force (remove existing)

### 03.10 — `stat`

- [ ] Full file info display
- [ ] `--json` returns comprehensive struct (mode, size, uid, gid, atime, mtime, ctime, inode, links, blocks)

### 03.11 — `readlink`

- [ ] Prints symlink target
- [ ] `-f` canonicalize (resolve all symlinks)

### 03.12 — `basename` / `dirname`

- [ ] `basename /path/to/file.txt` → `file.txt`
- [ ] `basename file.txt .txt` → `file` (suffix removal)
- [ ] `dirname /path/to/file` → `/path/to`

---

## Compliance Testing

Create `test/compliance/` with shell scripts comparing korego vs system tools:

```bash
#!/bin/bash
# test/compliance/test_ls.sh
KOREGO=./korego
PASS=0; FAIL=0

run_test() {
    local desc="$1"; shift
    expected=$("$@" 2>&1); exp_rc=$?
    got=$($KOREGO "$@" 2>&1); got_rc=$?
    if [ "$expected" = "$got" ] && [ "$exp_rc" = "$got_rc" ]; then
        ((PASS++))
    else
        ((FAIL++))
        echo "FAIL: $desc (exit: expected=$exp_rc got=$got_rc)"
        diff <(echo "$expected") <(echo "$got")
    fi
}

run_test "ls /tmp" ls /tmp
run_test "ls -la /tmp" ls -la /tmp
run_test "ls nonexistent" ls nonexistent

echo "Results: $PASS passed, $FAIL failed"
```

- [ ] Per-utility compliance scripts in `test/compliance/`
- [ ] Checks both output content AND exit codes

---

## Milestone 03

- [ ] `korego ls --json /` returns valid JSON array of FileInfo
- [ ] `korego ls -laR /tmp` output matches `/bin/ls -laR /tmp` format
- [ ] `korego cat file | korego grep pattern` — pipes work
- [ ] `korego rm -rf /` is refused
- [ ] `korego stat --json /etc/passwd` returns full stat struct
- [ ] All 13 utilities have unit tests with > 85% coverage
- [ ] Compliance test suite passes > 80% vs system tools

## How to Verify

```bash
make test
docker run --rm korego:dev ls --json /bin
docker run --rm korego:dev stat --json /bin/korego
bash test/compliance/test_ls.sh
bash test/compliance/test_cat.sh
```
