# goclean

> **v1.2.0** | **Status: Go, Python, Rust, Node.js, Java (Maven/Gradle) cache browsing stable. OS temp + browser cache + Docker cleanup added.**

Multi-language package cache cleaner with an interactive TUI. Find unused dependencies, clean OS temp files, browser caches, and Docker build artifacts to reclaim disk space.

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

### Start Scan (Go)
Scans configured directories for Go projects, finds all dependencies via `go list -m all`, compares against the module cache, and identifies unused modules. Module sizes are computed lazily in the background — the list appears instantly and sizes fill in progressively.

### Browse Cache
Browse and delete cached packages by language:

| Language | Cache Location |
|----------|---------------|
| Go | `go env GOMODCACHE` |
| Python (pip) | `pip cache dir` / `~/.cache/pip` / `%LOCALAPPDATA%\pip\cache` |
| Rust (Cargo) | `~/.cargo/registry/cache` and `~/.cargo/registry/src` |
| Node.js (npm) | `npm config get cache` / `~/.npm` |
| Java (Maven) | `~/.m2/repository` |
| Java (Gradle) | `~/.gradle/caches` |

### Clean Temp & Cache Files
Detect and clean OS and application temp/cache files. Each item shows its source, description, and criticality level:

| Item | Source | Criticality |
|------|--------|-------------|
| Windows Temp (`%TEMP%`, `%TMP%`) | Windows OS | Safe |
| Recycle Bin | Windows Shell | Caution |
| Chrome Cache | Google Chrome | Safe |
| Firefox Cache | Mozilla Firefox | Safe |
| Edge Cache | Microsoft Edge | Safe |
| Docker Build Cache | Docker | Moderate |
| Docker Dangling Images | Docker | Moderate |
| Docker Container Logs | Docker | Caution |

Criticality levels: **Safe** (green) — will be recreated automatically, **Moderate** (yellow) — may slow next operation, **Caution** (red) — permanent data loss.

### Configure Paths
Add or remove directories for Go project scanning. Paths are saved to `~/.goclean.json` and persist across sessions.

### Toggle Dry-Run
Switch dry-run mode on/off. When enabled, no files are deleted.

## Performance

- **Lazy size loading** — module sizes are computed concurrently in the background. The scan completes in seconds; sizes fill in progressively
- **Parallel scanning** — project discovery, dependency aggregation, and cache scanning all use goroutine worker pools
- **Efficient sorting** — uses Go's `sort.Slice` (O(n log n)) instead of insertion sort

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

### Temp & Cache Files
| Key | Action |
|-----|--------|
| `↑/↓` | Move cursor |
| `space` | Toggle selection |
| `a` | Select all |
| `n` | Deselect all |
| `i` | Show item details (source, criticality, description) |
| `enter` | Clean selected items |
| `q` | Back to menu |

### Confirm
| Key | Action |
|-----|--------|
| `y` | Confirm deletion |
| `n` | Cancel |

## Cross-Platform

Works on Windows, macOS, and Linux. All paths are auto-detected dynamically using environment variables and tool commands — nothing is hardcoded per user or machine.

## Safety

- **Never** deletes modules used by detected Go projects (Start Scan mode)
- Always requires explicit `y` confirmation before deletion
- Supports `--dry-run` to preview what would be deleted
- Temp/cache cleaners show criticality level for each item
- Handles file lock errors gracefully (especially on Windows)
- Deletions are concurrent but rate-limited

## License

MIT
