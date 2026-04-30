# test-finder

A CLI tool to discover Go modules and analyze their test coverage across directories.

## Overview

`test-finder` recursively scans directories for Go modules (`go.mod` files), runs coverage analysis on each one, and reports function-level test coverage. It supports monorepos and Go workspaces out of the box.

## Features

- **Recursive module discovery** — finds all `go.mod` and `go.work` files in a directory tree
- **Function-level coverage** — reports per-function coverage percentages, not just package-level
- **Multiple output formats** — human-readable table or machine-parseable JSON
- **Monorepo support** — handles Go workspace files (`go.work`) and nested modules
- **Ignored directories** — skips `.git`, `vendor`, `node_modules`, `bin`, and other common directories by default

## Installation

### From source

```bash
go install github.com/padiazg/test-finder@latest
```

### From binary

Download the latest release from the [GitHub releases page](https://github.com/padiazg/test-finder/releases) and place it on your `PATH`.

## Usage

### Scan the current directory

```bash
test-finder scan
```

### Scan a specific path

```bash
test-finder scan /path/to/project
```

### Output formats

#### Table output (default)

```
┌────────────────────┬────────────────────┬──────────────────┬───────────────┬──────────┐
│ PROJECT            │ PACKAGE            │ FILE             │ FUNCTION      │ COVERAGE │
├────────────────────┼────────────────────┼──────────────────┼───────────────┼──────────┤
│ github.com/my/repo │ github.com/my/repo │ user_service.go  │ CreateUser    │ 100.0%   │
│                    │                    │                  │ GetUser       │ 75.0%    │
│                    │                    │                  │ DeleteUser    │ 50.0%    │
├────────────────────┼────────────────────┼──────────────────┼───────────────┼──────────┤
│ github.com/my/repo │ github.com/my/repo │ order_handler.go │ ProcessOrder  │ 80.0%    │
│                    │                    │                  │ CancelOrder   │ 0.0%     │
└────────────────────┴────────────────────┴──────────────────┴───────────────┴──────────┘
```

#### JSON output

```bash
test-finder scan --output json
```

```json
[
  {
    "Module": "github.com/my/repo",
    "Path": "/home/user/my-repo",
    "Files": [
      {
        "Coverage": [
          {"Name": "CreateUser", "Line": 12, "Coverage": 100.0},
          {"Name": "GetUser", "Line": 25, "Coverage": 75.0}
        ],
        "FileName": "user_service.go",
        "Package": "github.com/my/repo",
        "Path": "/home/user/my-repo/user_service.go",
        "Average": 87.5
      }
    ]
  }
]
```

## Architecture

```
test-finder/
├── cmd/                     # CLI commands (Cobra)
│   ├── root.go              # Root command
│   └── scan.go              # Scan subcommand
├── internal/
│   ├── project/             # Domain types (Project, FileNode, FunctionCoverage)
│   └── scan/v2/             # Module discovery and scanning logic
├── pkg/helpers/             # Utility functions (mod parsing, path helpers)
├── main.go                  # Entry point
└── go.mod
```

### How it works

1. **Walk** the target directory recursively, looking for `go.mod` and `go.work` files
2. **Parse** each module file to extract the module path and directory
3. **Run** `go test -coverprofile=coverage.out ./...` for each module
4. **Parse** the coverage output with `go tool cover -func` to get per-function coverage
5. **Output** results in the selected format

## Ignored Directories

The following directories are automatically skipped during scanning:

| Directory    | Reason                    |
|-------------|---------------------------|
| `.git`      | Version control           |
| `vendor`    | Vendored dependencies     |
| `node_modules` | Node.js dependencies   |
| `.cache`    | Build caches              |
| `.idea`     | IDE metadata              |
| `bin`       | Compiled binaries         |
| `dist`      | Build output              |
| `build`     | Build output              |
| `.tmp`      | Temporary files           |
| `.opencode` | AI tool metadata          |
| `.vscode`   | Editor settings           |

## Requirements

- Go 1.26+
- A working Go toolchain (for `go test` and `go tool cover`)

## License

MIT
