# goclean — Development Timeline

## Phase 1 — Core Go Cleaner (v1.0.0) — Complete
- [x] Project scaffolding and Go module setup
- [x] Go project discovery (recursive `go.mod` detection with goroutine-per-directory)
- [x] Dependency aggregation via `go list -m all` (semaphore-limited to 8 concurrent)
- [x] Module cache scanning (`go env GOMODCACHE`)
- [x] Unused module detection (cache minus used modules)
- [x] Concurrent module deletion with error handling (4-worker pool)
- [x] Bubble Tea TUI with 6 screens (loading, summary, list, confirm, deleting, done)
- [x] Interactive list: toggle, select all/deselect all, search/filter, sort (name/size), pagination
- [x] CLI flags: `--paths`, `--dry-run`, `--verbose`, `--version`

## Phase 2 — Manual Controls & UX (v1.1.0) — Complete
- [x] Main menu with 6-item command selection (no auto-scan on startup)
- [x] Configure Paths screen (add/remove scan directories with text input)
- [x] Persistent config file (`~/.goclean.json`) with `AddPath`/`RemovePath` helpers
- [x] Dry-run toggle persisted across sessions
- [x] Color-coded UI: green (safe), yellow (caution), red (deletion candidates)
- [x] Cross-platform path detection using `go env GOPATH` + common directory names
- [x] Info panel toggle (`i` key) showing configured paths and dry-run status

## Phase 3 — Multi-Language Cache Browser (v1.2.0) — Complete
- [x] Language registry architecture (`lang/lang.go`) with `LangCache` struct
- [x] Language selector screen with icon, name, and cache availability detection
- [x] Go module cache scanner (handles `org/module/version` and `module@version` layouts)
- [x] Python pip cache scanner (wheels directory + http cache, files > 1MB)
- [x] Rust Cargo registry scanner (`registry/cache` + `registry/src`)
- [x] Node.js npm cache scanner (scoped `@org/pkg` + unscoped packages)
- [x] Java Maven local repository scanner (`.jar` based with version extraction)
- [x] Java Gradle cache scanner (`.jar` files > 1KB)
- [x] .NET NuGet package cache scanner (`packages/name/version` layout)
- [x] Per-language cache path auto-detection via env vars and tool commands
- [x] Cache availability indicator in language selector (checkmark / "cache not found")
- [x] Expected cache locations shown when cache is empty
- [x] Temp/cache file cleaner (`tempcache/` package)
  - [x] Windows Temp files (`%TEMP%`, `%TMP%`)
  - [x] Windows Recycle Bin (via `Clear-RecycleBin` PowerShell)
  - [x] Chrome browser cache (all profiles: Default + Profile*)
  - [x] Firefox browser cache (profile detection via `profiles.ini`)
  - [x] Edge browser cache (Chromium-based, all profiles)
  - [x] Docker build cache (`docker builder prune`)
  - [x] Docker dangling images (`docker image prune`)
  - [x] Docker container log truncation
- [x] Criticality levels: Safe (green), Moderate (yellow), Caution (red)
- [x] Item detail view (`i` key) showing source, criticality, size, description

## Phase 4 — v1.3.0 Optimizations, Bug Fixes & Features — Complete

### Bug Fixes
- [x] Register `.NET (NuGet)` in language registry — `dotnetCache()` was defined but never added to `Registry`
- [x] Fix Docker container logs size — `dockerLogsSize()` always returned 0 (`total += 0` on line 187)
- [x] Fix Docker logs cleanup — `dockerLogsClean()` used `sh -c truncate` which fails on Windows; replaced with `os.Truncate` cross-platform
- [x] Fix `computeID` increment lost — `startSizeComputation()` used value receiver so `m.computeID++` was discarded; moved increment to `Update` caller
- [x] Fix `saveConfig()` value receiver — config saves operated on a copy of the Model; changed to pointer receiver `(m *Model)`

### Performance Optimizations
- [x] `filepath.WalkDir` replaces `filepath.Walk` in all scan functions (`lang/lang.go`, `util/dirsize.go`) — avoids extra `Lstat` syscalls per entry
- [x] Consolidated duplicate utility functions into `util/` package:
  - `util.DirSize()` — single source for directory size calculation
  - `util.PathExists()` — single source for path existence check
  - Removed duplicates from `tui/model.go`, `tempcache/tempcache.go`, `lang/lang.go`
- [x] Depth-limited project discovery — `maxWalkDepth = 8` prevents runaway recursive scanning on deep filesystems
- [x] Semaphore-based concurrency in dependency aggregation (8 workers) and size computation (16 workers)

### New Features
- [x] JSON export (`--export <file>`) — new `export/` package with `Report` and `CacheReport` structs; auto-exports on scan completion when flag is set
- [x] Size threshold filter — `--min-size` CLI flag (e.g. `1MB`, `100KB`, `1GB`) + `m` key in TUI to cycle thresholds (off → 1MB → 10MB → 100MB → 1GB); filters both unused modules list and cache browser
- [x] Progress bar during deletion — `progress.Model` (already in codebase) now wired into `viewDeleting` and `viewTempCacheDeleting` screens with item count context
- [x] `deleteProgressMsg` type added for future incremental progress reporting

### Code Quality
- [x] Fixed inconsistent indentation in `lang/scanNpmCache` (lines 425-431)
- [x] Removed unused `os` import from `tempcache/tempcache.go` (replaced with `util` calls)
- [x] Added `io/fs` import to `lang/lang.go` for `fs.DirEntry` usage in `WalkDir` callbacks
- [x] Version bumped to v1.3.0

## Phase 5 — Future Enhancements
- [ ] Export report as CSV (`--export report.csv` with format detection)
- [ ] Filter by last accessed time (show modules not accessed in N days)
- [ ] "Used By" detail panel per module (show which projects depend on it)
- [ ] Split-pane view with preview (module details alongside list)
- [ ] Go project dependency graph visualization
- [ ] Automatic stale cache detection (time-based heuristics)
- [ ] Plugin system for additional languages (register custom cache scanners)
- [ ] Configuration file for custom cache paths per language
- [ ] Batch operations across multiple languages in one session
- [ ] Integration with `go clean -modcache` safety checks
- [ ] Windows long path support (`\\?\` prefix for paths > 260 chars)
- [ ] Parallel language cache scanning (scan all languages concurrently)
- [ ] Incremental deletion with per-item progress messages
- [ ] Undo last deletion (snapshot before delete, restore on undo)
- [ ] Disk usage visualization (tree map or bar chart in TUI)
