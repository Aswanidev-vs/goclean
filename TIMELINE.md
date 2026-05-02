# goclean — Development Timeline

## Phase 1 — Core Go Cleaner (Complete)
- [x] Project scaffolding and Go module setup
- [x] Go project discovery (recursive `go.mod` detection)
- [x] Dependency aggregation via `go list -m all`
- [x] Module cache scanning (`go env GOMODCACHE`)
- [x] Unused module detection (cache minus used)
- [x] Concurrent module deletion with error handling
- [x] Bubble Tea TUI with 6 screens (loading, summary, list, confirm, deleting, done)
- [x] Interactive list: toggle, select all, search, sort, pagination
- [x] CLI flags: `--paths`, `--dry-run`, `--verbose`

## Phase 2 — Manual Controls & UX (Complete)
- [x] Main menu with manual command selection (no auto-scan)
- [x] Configure Paths screen (add/remove scan directories)
- [x] Persistent config file (`~/.goclean.json`)
- [x] Dry-run toggle persisted across sessions
- [x] Color-coded UI (green/yellow/red)
- [x] Cross-platform path detection (Windows/macOS/Linux)

## Phase 3 — Multi-Language Cache Browser (Complete)
- [x] Language registry architecture (`lang/lang.go`)
- [x] Language selector screen with icon, name, cache detection
- [x] Go module cache scanner (handles `org/module/version` and `module@version`)
- [x] Python pip cache scanner (wheels + http cache)
- [x] Rust Cargo registry scanner (cache + src)
- [x] Node.js npm cache scanner (scoped + unscoped packages)
- [x] Java Maven local repository scanner (`.jar` based)
- [x] Java Gradle cache scanner
- [x] .NET NuGet package cache scanner
- [x] Per-language cache path auto-detection
- [x] Cache availability indicator in language selector
- [x] Expected cache locations shown when cache is empty

## Phase 4 — Future Enhancements
- [ ] Export report as JSON/CSV
- [ ] Filter by last accessed time
- [ ] Filter by size threshold
- [ ] "Used By" detail panel per module
- [ ] Split-pane view with preview
- [ ] Go project dependency graph visualization
- [ ] Automatic stale cache detection (time-based)
- [ ] Plugin system for additional languages
- [ ] Configuration file for custom cache paths per language
- [ ] Batch operations across multiple languages
- [ ] Integration with `go clean -modcache` safety checks
- [ ] Windows long path support (`\\?\` prefix)
- [ ] Progress bar during cache scanning
- [ ] Parallel language cache scanning
