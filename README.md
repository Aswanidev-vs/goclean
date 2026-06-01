# goclean

> **v1.3.0** | Multi-language package cache cleaner with an interactive TUI.

Find unused dependencies, clean OS temp files, browser caches, and Docker build artifacts to reclaim disk space. Supports Go, Python, Rust, Node.js, Java (Maven/Gradle), and .NET (NuGet).

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
| `--export <file>` | Export scan report to a JSON file |
| `--min-size <size>` | Minimum size filter (e.g. `1MB`, `100KB`, `1GB`) |

### Examples

```bash
# Run with default paths
goclean

# Custom scan paths
goclean --paths "C:\Users\me\projects,D:\work"

# Dry run mode
goclean --dry-run --verbose

# Export scan report to JSON
goclean --export report.json

# Only show packages larger than 10MB
goclean --min-size 10MB

# Combine flags
goclean --paths "~/projects" --export report.json --min-size 1MB --dry-run
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
| .NET (NuGet) | `~/.nuget/packages` / `%LOCALAPPDATA%\NuGet\Cache` |

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

### JSON Export
Export scan results to a structured JSON file for reporting or CI integration:

```bash
goclean --export report.json
```

The report includes timestamp, project count, module counts, reclaimable space, and a full list of unused modules with sizes and paths.

### Size Filter
Filter packages by minimum size to focus on the biggest space savings. Available as a CLI flag or interactively in the TUI:

```bash
# CLI flag
goclean --min-size 10MB
```

In the TUI, press `m` to cycle through thresholds: off → 1MB → 10MB → 100MB → 1GB. Works in both the unused modules list and the cache browser.

### Configure Paths
Add or remove directories for Go project scanning. Paths are saved to `~/.goclean.json` and persist across sessions.

### Toggle Dry-Run
Switch dry-run mode on/off. When enabled, no files are deleted.

## Performance

- **Lazy size loading** — module sizes are computed concurrently in the background. The scan completes in seconds; sizes fill in progressively
- **Parallel scanning** — project discovery, dependency aggregation, and cache scanning all use goroutine worker pools with semaphore-based concurrency limiting
- **Depth-limited discovery** — project scanning is capped at 8 directory levels deep to prevent runaway traversal on large filesystems
- **Efficient directory walking** — uses `filepath.WalkDir` instead of `filepath.Walk` to avoid redundant `Lstat` syscalls
- **Efficient sorting** — uses Go's `sort.Slice` (O(n log n))

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

### Package List (Unused Modules & Cache Browser)
| Key | Action |
|-----|--------|
| `↑/↓` | Move cursor |
| `space` | Toggle selection |
| `a` | Select all |
| `n` | Deselect all |
| `/` | Search/filter |
| `s` | Sort (name/size toggle) |
| `m` | Cycle min size filter (off/1MB/10MB/100MB/1GB) |
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
- Deletions are concurrent but rate-limited (4 workers)

## Changelog

### v1.3.0
- Added `.NET (NuGet)` cache browser (was defined but not registered)
- Added `--export` flag to export scan reports as JSON
- Added `--min-size` flag and `m` key to filter by minimum package size
- Added progress bar display during deletion operations
- Fixed Docker container logs size calculation (was always returning 0)
- Fixed Docker logs cleanup to work cross-platform (was using `sh -c`)
- Fixed `computeID` increment being lost due to value receiver
- Fixed `saveConfig()` using value receiver instead of pointer
- Optimized directory walking with `filepath.WalkDir` (avoids extra syscalls)
- Optimized project discovery with depth limit (max 8 levels)
- Consolidated duplicate utility functions into `util/` package

### v1.2.0
- Added temp/cache file cleaner (Windows temp, browser caches, Docker)
- Added Docker build cache, dangling images, and container log cleanup

### v1.1.0
- Added multi-language cache browser (Python, Rust, Node.js, Java, .NET)
- Added language selector with cache availability detection
- Added persistent configuration (`~/.goclean.json`)

### v1.0.0
- Initial release with Go module cache scanning and cleanup

## License

MIT
