# JSON-RPC API

The KoreGo daemon implements a standard JSON-RPC 2.0 interface. It listens on a Unix socket (default: `/tmp/korego.sock`).

## Request Format

```json
{
  "jsonrpc": "2.0",
  "method": "korego.ls",
  "params": {
    "sessionId": "abc123def",
    "path": "/",
    "flags": ["-l", "-a"]
  },
  "id": 1
}
```

## Response Format

```json
{
  "jsonrpc": "2.0",
  "result": { ... structured data ... },
  "id": 1
}
```

## Methods Catalog

### Session Management
- `korego.session.create`: Creates a new persistent shell/environment session.
- `korego.session.list`: Lists active sessions.
- `korego.session.setCwd`: Changes the working directory for a session.
- `korego.session.destroy`: Terminates a session.

### Execution
- `korego.shell.exec`: Evaluates a shell script string within a session.
  - **Params:** `{"script": "echo hello > file.txt"}`

### Utilities
Every utility implemented in `pkg/` is available as `korego.<utility>`.
- `korego.cat`
- `korego.ls`
- `korego.grep`
- `korego.rm`
- ...

## Error Codes
- `-32700`: Parse error (invalid JSON)
- `-32600`: Invalid Request
- `-32601`: Method not found
- `-32602`: Invalid params (or invalid sessionId)
- `-32000`: Server error / Execution error
