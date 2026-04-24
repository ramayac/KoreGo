# Phase 01 — Multicall Dispatcher + Tier 1 Utilities

> **Timeline:** Week 1–2 | **Depends on:** Phase 00 | **Blocks:** Phase 02

---

## Goal

Create the single-binary multicall dispatcher and implement the 10 simplest POSIX utilities. These are trivial to implement but prove the entire architecture works end to end.

---

## Tasks

### 01.1 — Command Registry (`internal/dispatch/dispatch.go`)

**Design:**
```go
type Command struct {
    Name    string
    Run     func(args []string) int  // returns exit code
    Usage   string
}

var registry = map[string]Command{}

func Register(cmd Command) {
    registry[cmd.Name] = cmd
}

func Lookup(name string) (Command, bool) {
    cmd, ok := registry[name]
    return cmd, ok
}
```

- [ ] Registry is a `map[string]Command`
- [ ] Each utility calls `dispatch.Register()` in its `init()` function
- [ ] `Lookup()` returns the command or false

---

### 01.2 — Multicall Entry Point (`cmd/korego/main.go`)

**Logic flow:**
```
1. name = filepath.Base(os.Args[0])
2. if name == "korego" AND len(os.Args) > 1:
       name = os.Args[1]
       os.Args = os.Args[1:]  // shift
3. cmd, ok = dispatch.Lookup(name)
4. if !ok → print "unknown command", exit 127
5. exitCode = cmd.Run(os.Args[1:])
6. os.Exit(exitCode)
```

- [ ] Symlink dispatch: `/bin/ls` → `/bin/korego` → runs `ls`
- [ ] Subcommand dispatch: `korego ls -la` → runs `ls`
- [ ] Unknown command → exit code 127 (POSIX standard)
- [ ] `korego --help` lists all registered commands
- [ ] `korego --version` prints version string

**Test cases:**
```bash
# Symlink test (manual)
ln -s ./korego ./echo
./echo hello       # prints "hello"

# Subcommand test
./korego echo hello  # prints "hello"

# Unknown command
./korego nonexist    # exit 127
```

---

### 01.3 — Tier 1 Utilities

Each utility follows this pattern:
1. **Library** in `pkg/<name>/` — pure function returning a struct
2. **CLI wrapper** — registered in dispatcher, calls library, uses `Render()`

| Utility | Library Return | `--json` Output | Notes |
|---------|---------------|-----------------|-------|
| `echo` | `EchoResult{Text string}` | `{"data":{"text":"hello"}}` | Support `-n` (no newline), `-e` (escapes) |
| `true` | — | — | Just `os.Exit(0)` |
| `false` | — | — | Just `os.Exit(1)` |
| `yes` | — | — | Infinite loop printing "y" or arg. No `--json`. |
| `whoami` | `WhoamiResult{User string, UID int}` | `{"data":{"user":"root","uid":0}}` | `os/user.Current()` |
| `hostname` | `HostnameResult{Name string}` | `{"data":{"hostname":"abc"}}` | `os.Hostname()` |
| `uname` | `UnameResult{Sysname,Node,Release,Version,Machine}` | Full struct | `syscall.Uname()` |
| `pwd` | `PwdResult{Path string}` | `{"data":{"path":"/home"}}` | `os.Getwd()`. Flag: `-P` (physical, resolve symlinks) |
| `printenv` | `PrintenvResult{Vars map[string]string}` | Full env map | Specific var: `printenv HOME` |
| `env` | `EnvResult{Vars map[string]string}` | Full env map | Like printenv but also supports `-i` (ignore environment) |

**Per-utility checklist (repeat for each):**
- [ ] `pkg/<name>/<name>.go` — library function
- [ ] `pkg/<name>/<name>_test.go` — unit tests
- [ ] CLI wrapper registered via `init()`
- [ ] `--json` flag works via `common.Render()`
- [ ] Exit codes match POSIX spec
- [ ] `--help` prints usage

---

## Milestone 01

- [ ] `korego echo hello` prints `hello`
- [ ] `korego echo --json hello` prints `{"command":"echo","data":{"text":"hello"},...}`
- [ ] `korego true` exits 0
- [ ] `korego false` exits 1
- [ ] `korego --help` lists all 10 commands
- [ ] Symlink dispatch works (`ln -s korego echo && ./echo hi`)
- [ ] Unknown commands exit 127
- [ ] All Tier 1 utilities have unit tests passing

## How to Verify

```bash
# Build
CGO_ENABLED=0 go build -o korego ./cmd/korego/

# Unit tests
go test -v -cover ./pkg/echo/ ./pkg/whoami/ ./pkg/uname/ ...

# Manual smoke tests
./korego true ; echo $?                  # 0
./korego false ; echo $?                 # 1
./korego echo hello                      # hello
./korego echo --json hello               # JSON envelope
./korego uname --json                    # structured uname
./korego whoami --json                   # {"user":"...","uid":...}
./korego nonexist ; echo $?             # 127

# Symlink test
ln -sf ./korego ./pwd && ./pwd           # prints cwd
```
