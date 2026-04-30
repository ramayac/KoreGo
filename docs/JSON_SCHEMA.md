# JSON-RPC Output Schemas

All KoreGo utilities support structured machine-readable output via the `--json` flag or when invoked via the JSON-RPC daemon.

## Standard Envelope

Every successful utility execution outputs a standard JSON envelope:

```json
{
  "Data": { ... utility specific data ... },
  "Error": null
}
```

If an error occurs:

```json
{
  "Data": null,
  "Error": {
    "Code": 1,
    "Message": "file not found"
  }
}
```

## Utility Schemas

### `cat`
```json
{
  "Text": "content of the file\n..."
}
```

### `ls`
```json
{
  "Files": [
    {
      "Name": "file.txt",
      "Size": 1024,
      "Mode": "-rw-r--r--",
      "ModTime": "2023-10-27T10:00:00Z",
      "IsDir": false
    }
  ]
}
```

### `grep`
```json
{
  "Lines": [
    {
      "File": "test.txt",
      "LineNumber": 1,
      "Text": "matching line content"
    }
  ]
}
```

*(Note: See individual utility Go packages for exact struct definitions.)*
