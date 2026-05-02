# goclean

> **v1.1.0** | **Status: Go support is fully working. Python, Rust, Node.js, Java (Maven/Gradle) cache browsing is implemented but still being tested and refined.**

Multi-language package cache cleaner with an interactive TUI. Find unused dependencies and reclaim disk space.

## Install

```bash
go install github.com/Aswanidev-vs/goclean@latest
```

Or build from source:

```bash
git clone https://github.com/Aswanidev-vs/goclean.git
cd goclean
go build -o goclean .
```

## Usage

```bash
goclean
```

### Flags

| Flag | Description |
|------|-------------|
| `--paths` | Comma-separated directories to scan for Go projects |
| `--dry-run` | Simulate deletion without removing files |
| `--verbose` | Show detailed logs |
| `--version` | Show version |

### Examples

```bash
# Run with default paths
goclean

# Custom scan paths
goclean --paths "C:\Users\me\projects,D:\work"

# Dry run mode
goclean --dry-run --verbose
```

## Features

### Start Scan (Go — fully working)
Scans configured directories for Go projects, finds all dependencies via `go list -m all`, compares against the module cache, and identifies unused modules.

### Browse Cache (multi-language — Go stable, others in progress)
Browse and delete cached packages by language:

| Language | Status | Cache Location |
|----------|--------|---------------|
| Go | Stable | `go env GOMODCACHE` |
| Python (pip) | In progress | `pip cache dir` / `~/.cache/pip` / `%LOCALAPPDATA%\pip\cache` |
| Rust (Cargo) | In progress | `~/.cargo/registry/cache` and `~/.cargo/registry/src` |
| Node.js (npm) | In progress | `npm config get cache` / `~/.npm` |
| Java (Maven) | In progress | `~/.m2/repository` |
| Java (Gradle) | In progress | `~/.gradle/caches` |

### Configure Paths
Add or remove directories for Go project scanning. Paths are saved to `~/.goclean.json` and persist across sessions.

### Toggle Dry-Run
Switch dry-run mode on/off. When enabled, no files are deleted.

## Keyboard Controls

### Menu
| Key | Action |
|-----|--------|
| `↑/↓` | Navigate |
| `enter/space` | Select |
| `i` | Toggle info panel |
| `q` | Quit |

### Language Selector
| Key | Action |
|-----|--------|
| `↑/↓` | Navigate languages |
| `enter/space` | Browse that language's cache |
| `q` | Back to menu |

### Package List
| Key | Action |
|-----|--------|
| `↑/↓` | Move cursor |
| `space` | Toggle selection |
| `a` | Select all |
| `n` | Deselect all |
| `/` | Search/filter |
| `s` | Sort (name/size toggle) |
| `enter` | Delete selected |
| `q` | Back |

### Confirm
| Key | Action |
|-----|--------|
| `y` | Confirm deletion |
| `n` | Cancel |

## Cross-Platform

Works on Windows, macOS, and Linux. All paths are auto-detected dynamically using environment variables and tool commands — nothing is hardcoded per user or machine.



## Safety

- Never deletes modules used by detected Go projects (Start Scan mode)
- Always requires explicit `y` confirmation before deletion
- Supports `--dry-run` to preview what would be deleted
- Handles file lock errors gracefully (especially on Windows)
- Deletions are concurrent but rate-limited

## License

MIT
