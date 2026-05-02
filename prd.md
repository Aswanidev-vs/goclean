📄 PRD — goclean
🧾 Overview

goclean is a Go-based CLI tool with an interactive TUI (Terminal UI) built using Bubble Tea.

It helps developers:

Analyze Go module usage across all local projects
Identify unused modules in the global Go module cache
Interactively review and selectively delete unused dependencies

Unlike go clean -modcache, goclean is safe, transparent, and user-controlled.

🎯 Goals
Prevent unnecessary deletion of active dependencies
Reduce disk usage from stale Go modules
Provide full visibility into module usage
Deliver an intuitive, keyboard-driven TUI experience
🚫 Non-Goals
Automatic deletion without user confirmation
Modifying project go.mod files
Acting as a package manager replacement
Supporting non-Go ecosystems
👤 Target Users
Go developers working on multiple projects
Backend engineers managing large workspaces
Developers with limited disk space
Power users who prefer terminal-based tools
⚙️ Core Features
1. Project Discovery
Automatically scan for Go projects using:
Default paths:
~/go/src
~/projects
~/workspace
User-defined paths via CLI flag (--paths)
Detect projects by presence of:
go.mod
Ignore directories:
.git
node_modules
vendor
2. Dependency Aggregation

For each detected project:

Execute:

go list -m all
Extract:
module name
version
Build:
map[module@version] → []projects
3. Module Cache Analysis

Retrieve module cache path:

go env GOMODCACHE
Traverse all cached modules
Extract:
module name
version
size (optional but recommended)
4. Unused Module Detection

A module is considered unused if:

It exists in the module cache
It is NOT referenced by any detected project
🖥️ TUI Experience
🌀 Screen 1: Loading
Spinner animation
Messages:
“Scanning projects…”
“Analyzing dependencies…”
“Scanning module cache…”
📊 Screen 2: Summary Dashboard

Displays:

Total projects found
Total unique modules
Unused modules count
Estimated reclaimable disk space

Controls:

Enter → View unused modules
q → Quit
📋 Screen 3: Unused Modules List (Core UI)

Interactive selectable list:

[ ] github.com/foo/bar@v1.2.3   (12 MB)
[x] github.com/old/lib@v0.5.0   (8 MB)
[ ] golang.org/x/exp@v0.1.0     (5 MB)

Controls:

↑ / ↓ → Navigate
space → Toggle selection
a → Select all
n → Deselect all
/ → Search/filter
s → Sort (size/name)
Enter → Proceed
q → Back
⚠️ Screen 4: Confirmation

Prompt:

You are about to delete 12 modules (120 MB)
Proceed? (y/n)
⚙️ Screen 5: Deletion Progress
Progress indicator (spinner or progress bar)
Current module being deleted
✅ Screen 6: Result

Displays:

Deleted 12 modules
Freed 120 MB
Any key → Exit
🧠 Data Model
type Module struct {
    Name     string
    Version  string
    Size     int64
    UsedBy   []string
    Selected bool
}
🏗️ Architecture
goclean/
├── main.go
├── tui/
│   ├── model.go
│   ├── update.go
│   ├── view.go
├── scanner/
│   ├── discover.go
│   ├── deps.go
├── cache/
│   ├── scan.go
├── cleaner/
│   ├── delete.go
🔄 Application States
loading
summary
list
confirm
deleting
done
🔐 Safety Rules
NEVER delete:
Modules used by any detected project
Go toolchain/internal dependencies
Always require explicit user confirmation
Handle:
File locks (especially on Windows)
Partial deletion failures gracefully
⚡ CLI Interface
Default command
goclean
Optional flags
--paths → custom scan directories
--dry-run → simulate without deleting
--verbose → detailed logs
🚀 Performance Considerations
Use goroutines for:
Project scanning
Cache scanning
Implement worker pools to limit concurrency
Lazy-load module sizes to avoid startup delays
✨ UX Enhancements
Color-coded UI:
Green → safe
Yellow → caution
Red → deletion candidates
Pagination for large module lists
Search/filter support
🎁 Bonus Features (Future Scope)
Show “Used By” details panel per module
Filter by:
last accessed time
size threshold
Export report as:
JSON
CSV
Interactive TUI improvements:
split-pane view
preview panel
📦 Deliverables
Fully working Go CLI tool
TUI powered by Bubble Tea
Cross-platform support (Windows prioritized)

Build instructions:

go build -o goclean
Example usage documentation
🧩 Success Criteria
Accurately detects unused modules
No accidental deletion of active dependencies
Smooth, responsive TUI experience
Clear and intuitive user interaction