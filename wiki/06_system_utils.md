# Phase 06 — Tier 4: System & Process Utilities

> **Timeline:** Week 7–8 | **Depends on:** Phase 05

---

## Goal

Implement 14 system/process utilities. These require deeper OS interaction via `syscall` and `/proc`.

## Utilities

### 06.1 — `ps`
- Read `/proc/[pid]/stat`, `/proc/[pid]/cmdline`, `/proc/[pid]/status`
- Flags: `-e`/`-A` (all), `-f` (full), `-o` (custom format), `-p` (by PID)
- `--json` → `[{"pid":1, "ppid":0, "user":"root", "cmd":"init", "cpu":"0.1%", "mem":"1.2%"}]`

### 06.2 — `kill`
- `kill -TERM <pid>`, `kill -9 <pid>`, `kill -l` (list signals)
- `--json` → `{"signaled":[{"pid":123,"signal":"SIGTERM","success":true}]}`

### 06.3 — `sleep`
- `sleep 5`, `sleep 1.5`, `sleep 1m`, `sleep 1h30m`
- Support Go duration format as extension

### 06.4 — `date`
- Flags: `-u` (UTC), `+FORMAT` (strftime-style)
- `--json` → `{"iso":"2026-04-24T...","unix":1745539200,"utc":"...","timezone":"PDT"}`
- Requires timezone data in container

### 06.5 — `id` / `groups`
- `id` → `uid=1000(user) gid=1000(user) groups=...`
- `--json` → `{"uid":1000,"user":"...","gid":1000,"group":"...","groups":[...]}`

### 06.6 — `chmod` / `chown` / `chgrp`
- Symbolic (`chmod u+x`) and octal (`chmod 755`) modes
- `-R` recursive
- `--json` → `{"changed":[{"path":"...","mode":"0755"}]}`

### 06.7 — `df`
- Flags: `-h` (human), `-i` (inodes), `-T` (type)
- Read from `syscall.Statfs`
- `--json` → `[{"filesystem":"/dev/sda1","size":...,"used":...,"avail":...,"mountpoint":"/"}]`

### 06.8 — `du`
- Flags: `-h`, `-s` (summary), `-d` (max depth)
- Walk directory tree, sum sizes
- `--json` → `[{"path":"./src","size":1048576,"files":42}]`

### 06.9 — `find`
- Flags: `-name`, `-type`, `-size`, `-mtime`, `-exec`, `-maxdepth`, `-print0`
- `--json` → `[{"path":"./a/b.txt","type":"f","size":100,"mtime":"..."}]`

### 06.10 — `xargs`
- Read stdin, build and execute command lines
- Flags: `-n` (max args), `-I{}` (replace), `-P` (parallel), `-0` (null delim)
- `--json` → `[{"command":"rm file1 file2","exitCode":0}]`

## Milestone 06

- [ ] `korego ps --json` returns process list from `/proc`
- [ ] `korego find --json -name "*.go" ./` returns structured file list
- [ ] `korego df --json -h` returns filesystem usage
- [ ] `korego du --json -sh ./` returns directory size
- [ ] All accessible via both CLI and daemon RPC
- [ ] >80% test coverage on all Tier 4 utilities

## How to Verify

```bash
./korego ps --json | head -5
./korego df --json -h
./korego find --json -name "*.go" -type f ./pkg/
./korego date --json -u
./korego id --json
```
