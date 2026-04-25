# POSIX Compliance Matrix

This document tracks the implementation status of KoreGo utilities against the POSIX standard.

| Utility | Status | Flags Implemented | Flags Missing | Notes |
|---------|--------|-------------------|---------------|-------|
| ls      | ✅     | -laRhtrdSi1A      | -C -x -g -G   | |
| grep    | ⚠️     | -ivclnrEFwx       | BRE backrefs   | Go RE2 limitation |
| sed     | ⚠️     | s, d, p, q, -i    | y, H, G, N     | Incremental |
| awk     | ⚠️     | Basic patterns     | Full awk spec  | Subset only |
| echo    | ✅     | -n                | N/A           | POSIX compliant |
| cat     | ✅     | -u                | N/A           | POSIX compliant |
| pwd     | ✅     | -L -P             | N/A           | POSIX compliant |
