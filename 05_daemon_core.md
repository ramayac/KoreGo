# Phase 05 — JSON-RPC Daemon Core

> **Timeline:** Week 6–7 | **Depends on:** Phase 03, 04

---

## Goal

Build the persistent JSON-RPC 2.0 daemon that listens on a Unix socket, routes methods to `pkg/` libraries, and handles concurrency.

## Tasks

### 05.1 — Daemon Entry Point

`korego daemon --socket /var/run/korego.sock --workers 4 --log-level info`

- [ ] PID file at `/var/run/korego.pid`
- [ ] Signal handling: `SIGTERM`/`SIGINT` → graceful shutdown (drain in-flight, close socket, remove PID)
- [ ] `SIGHUP` → reload config (future)

### 05.2 — Unix Socket Listener (`internal/daemon/server.go`)

- [ ] `net.Listen("unix", socketPath)`
- [ ] Remove stale socket file on start (`os.Remove` if exists)
- [ ] Set socket permissions to `0660`
- [ ] Accept loop with error handling

### 05.3 — JSON-RPC Router

Method naming: `korego.<utility>` (e.g., `korego.ls`, `korego.echo`)

```json
// Request
{"jsonrpc":"2.0","method":"korego.ls","params":{"path":"/tmp","flags":["-la"]},"id":1}

// Success Response
{"jsonrpc":"2.0","result":{"files":[...]},"id":1}

// Error Response
{"jsonrpc":"2.0","error":{"code":1002,"message":"not found","data":{"path":"/nope"}},"id":1}
```

- [ ] Router maps method string → handler function
- [ ] Auto-register all utilities from dispatcher registry
- [ ] Unknown method → error code `-32601`
- [ ] Invalid params → error code `-32602`

### 05.4 — Worker Pool

```go
type WorkerPool struct {
    sem chan struct{} // buffered channel as semaphore
}

func NewWorkerPool(size int) *WorkerPool {
    return &WorkerPool{sem: make(chan struct{}, size)}
}

func (wp *WorkerPool) Submit(ctx context.Context, fn func()) error {
    select {
    case wp.sem <- struct{}{}:
        go func() {
            defer func() { <-wp.sem }()
            fn()
        }()
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

- [ ] Bounded goroutine pool (default: `runtime.NumCPU()`)
- [ ] Context-based timeout for queued requests
- [ ] Metrics: active workers, queued requests, completed, errors

### 05.5 — Batch Processing

- [ ] Detect JSON array `[{...},{...}]` → batch mode
- [ ] Process each request concurrently via worker pool
- [ ] Return results array in same order as requests
- [ ] Single request failure doesn't abort batch

### 05.6 — Health Check

`korego.ping` → `{"pong":true, "uptime":"2h30m", "version":"0.3.0", "workers":{"active":2,"total":4}}`

### 05.7 — Client Library (`pkg/client/`)

```go
client, _ := koregoclient.Dial("/var/run/korego.sock")
result, _ := client.Call("korego.ls", map[string]interface{}{"path": "/tmp"})
```

- [ ] Go client for tests and future integrations
- [ ] Connection pooling
- [ ] Timeout support

## Milestone 05

- [ ] `korego daemon` starts and listens on Unix socket
- [ ] `korego.ls` over socket returns file list without binary restart
- [ ] Batch of 10 commands returns 10 results
- [ ] 100 concurrent requests complete without errors
- [ ] `korego.ping` returns health status
- [ ] Graceful shutdown completes within 5s

## How to Verify

```bash
# Start daemon
./korego daemon --socket /tmp/korego.sock &

# RPC via socat
echo '{"jsonrpc":"2.0","method":"korego.ls","params":{"path":"/"},"id":1}' \
  | socat - UNIX-CONNECT:/tmp/korego.sock

# Batch
echo '[{"jsonrpc":"2.0","method":"korego.echo","params":{"text":"a"},"id":1},
       {"jsonrpc":"2.0","method":"korego.echo","params":{"text":"b"},"id":2}]' \
  | socat - UNIX-CONNECT:/tmp/korego.sock

# Health
echo '{"jsonrpc":"2.0","method":"korego.ping","id":99}' \
  | socat - UNIX-CONNECT:/tmp/korego.sock

# Load test
go test -run TestDaemonConcurrent -timeout 60s ./test/integration/
```
