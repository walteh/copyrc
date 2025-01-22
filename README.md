# 📦 copyrc

A tool for maintaining local copies of files from remote repositories with powerful features like file replacements, status tracking, and various providers.

```ascii

                     ┌─────────────────┐
                     │    copyrc CLI   │
                     └────────┬────────┘
                              │
                     ┌────────┴────────┐
                     │  Configuration  │
                     └────────┬────────┘
                              │
           ┌──────────────────┼──────────────────┐
           │                  │                  │
    ┌──────┴──────┐   ┌──────┴──────┐   ┌──────┴──────┐
    │  Provider   │   │  Operation  │   │   Status    │
    │  Interface  │   │   Manager   │   │   Manager   │
    └──────┬──────┘   └──────┬──────┘   └──────┬──────┘
           │                  │                  │
    ┌──────┴──────┐   ┌──────┴──────┐   ┌──────┴──────┐
    │   GitHub    │   │    File     │   │    Lock     │
    │  Provider   │   │ Operations  │   │    File     │
    └─────────────┘   └─────────────┘   └─────────────┘
```

## 🚀 Features

-   📝 Configuration-based file synchronization
-   🔄 String replacements in copied files
-   🎯 File-specific replacements
-   🔍 Status tracking with lock files
-   🚫 File ignore patterns
-   ⚡️ Asynchronous file processing
-   📦 Multiple repository providers
-   🎨 Beautiful console output

## 📋 Requirements

-   Go 1.21 or later
-   GitHub token (for GitHub provider)

## 🛠️ Installation

```bash
go install github.com/walteh/copyrc/cmd/copyrc@latest
```

## 🎯 Usage

1. Create a configuration file (`.copyrc.yaml`):

```yaml
provider:
    repo: github.com/org/repo
    ref: main
    path: pkg/templates
destination: ./local/templates
copy:
    replacements:
        - old: "foo"
          new: "bar"
        - old: "baz"
          new: "qux"
          file: specific.go
    ignore_files:
        - "*.tmp"
        - "*.log"
go_embed: true
async: true
```

2. Set up your GitHub token:

```bash
export GITHUB_TOKEN=your_token_here
```

3. Run copyrc:

```bash
copyrc --config .copyrc.yaml
```

## 🔧 Configuration

### Provider Arguments

| Field  | Description                                  |
| ------ | -------------------------------------------- |
| `repo` | Repository URL (e.g., `github.com/org/repo`) |
| `ref`  | Branch or tag (default: `main`)              |
| `path` | Path within repository                       |

### Copy Arguments

| Field          | Description                     |
| -------------- | ------------------------------- |
| `replacements` | List of string replacements     |
| `ignore_files` | List of file patterns to ignore |

### Other Options

| Field           | Description                       |
| --------------- | --------------------------------- |
| `destination`   | Local destination path            |
| `go_embed`      | Generate Go embed code            |
| `clean`         | Clean destination directory       |
| `status`        | Check local status                |
| `remote_status` | Check remote status               |
| `force`         | Force update even if status is ok |
| `async`         | Process files asynchronously      |

## 🎨 Console Output

```
copyrc • syncing repository files

[syncing /local/templates]
◆ github.com/org/repo • main
    ✓ file1.go                          managed         NEW
    ⟳ file2.go                          copy [2]        UPDATED
    • file3.go                          local           no change
    ✗ file4.go                          managed         REMOVED
```

## 🧪 Testing

Run tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## 📝 License

Copyright 2025 walteh LLC

Licensed under the Apache License, Version 2.0
