package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Aswanidev-vs/goclean/lang"
)

func (m Model) View() string {
	switch m.screen {
	case ScreenMenu:
		return m.viewMenu()
	case ScreenPaths:
		return m.viewPaths()
	case ScreenLoading:
		return m.viewLoading("Scanning projects and dependencies...")
	case ScreenSummary:
		return m.viewSummary()
	case ScreenList:
		return m.viewList()
	case ScreenConfirm:
		return m.viewConfirm()
	case ScreenDeleting:
		return m.viewDeleting("Deleting unused modules...")
	case ScreenDone:
		return m.viewDone()
	case ScreenLangSelect:
		return m.viewLangSelect()
	case ScreenCacheLoading:
		return m.viewLoading("Loading cache...")
	case ScreenCache:
		return m.viewCache()
	case ScreenCacheConfirm:
		return m.viewCacheConfirm()
	case ScreenCacheDeleting:
		return m.viewDeleting("Deleting packages...")
	}
	return ""
}

func (m Model) viewMenu() string {
	var b strings.Builder

	b.WriteString("\n")
	if m.version != "" && m.version != "dev" {
		b.WriteString(titleStyle.Render(fmt.Sprintf("  goclean %s", m.version)))
	} else {
		b.WriteString(titleStyle.Render("  goclean"))
	}
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Package Cache Cleaner"))
	b.WriteString("\n\n")

	descriptions := []string{
		"scan & find unused Go modules",
		"browse & delete cached packages",
		"set directories to scan",
		"",
		"",
	}

	for i, item := range m.menuItems {
		prefix := "  "
		if i == m.menuCursor {
			prefix = cursorStyle.Render("▸ ")
			b.WriteString(menuActiveStyle.Render(prefix + item))
		} else {
			b.WriteString(menuItemStyle.Render(prefix + item))
		}

		if i == 3 {
			status := "OFF"
			if m.dryRun {
				status = "ON"
			}
			b.WriteString(dimStyle.Render(fmt.Sprintf("  — %s", status)))
		} else if descriptions[i] != "" {
			b.WriteString(dimStyle.Render("  — " + descriptions[i]))
		}

		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.showInfo {
		b.WriteString(infoBox.Render(fmt.Sprintf(
			"Paths: %s | Dry-run: %v",
			m.pathDisplayShort(),
			m.dryRun,
		)))
		b.WriteString("\n\n")
	}
	b.WriteString(helpStyle.Render("↑/↓ move | enter select | i info | q quit"))
	return b.String()
}

func (m Model) viewPaths() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Configure Scan Paths"))
	b.WriteString("\n\n")
	b.WriteString(m.pathInput.View())
	b.WriteString("\n\n")

	if len(m.paths) > 0 {
		b.WriteString(subtitleStyle.Render("  Current paths:"))
		b.WriteString("\n")
		for i, p := range m.paths {
			b.WriteString(selectedStyle.Render(fmt.Sprintf("    %d. %s", i+1, p)))
			b.WriteString("\n")
		}
	} else {
		b.WriteString(dimStyle.Render("  No paths configured."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter add | ctrl+d remove | esc back"))
	return b.String()
}

func (m Model) viewLoading(msg string) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.spinner.View() + " ")
	b.WriteString(warningStyle.Render(msg))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  This may take a moment for large caches."))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("↑/↓ move | space toggle | a all | n none | / search | s sort | enter delete | q back"))
	return b.String()
}

func (m Model) viewSummary() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Scan Results"))
	b.WriteString("\n\n")

	boxContent := fmt.Sprintf(
		"Projects scanned:     %s\n"+
			"Unique modules:       %s\n"+
			"Unused modules:       %s\n"+
			"Reclaimable space:    %s",
		successStyle.Render(fmt.Sprintf("%d", m.projectCount)),
		successStyle.Render(fmt.Sprintf("%d", m.totalModules)),
		warningStyle.Render(fmt.Sprintf("%d", len(m.unusedModules))),
		warningStyle.Render(formatSize(m.freedBytes)),
	)
	b.WriteString(greenBox.Render(boxContent))
	b.WriteString("\n\n")

	if len(m.unusedModules) == 0 {
		b.WriteString(successStyle.Render("  No unused modules found! Your cache is clean."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("m menu | q quit"))
	} else {
		b.WriteString(subtitleStyle.Render("  Press Enter to review unused modules"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("enter view | m menu | q quit"))
	}
	return b.String()
}

func (m Model) viewList() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Unused Modules"))
	b.WriteString("\n\n")

	if m.searchActive {
		b.WriteString("  " + m.search.View())
		b.WriteString("\n\n")
	}

	visible := m.getVisibleModules()
	if len(visible) == 0 {
		b.WriteString(dimStyle.Render("  No modules match the filter."))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("q back | / search"))
		return b.String()
	}

	pageSize := m.getPageSize()
	start := m.offset
	end := start + pageSize
	if end > len(visible) {
		end = len(visible)
	}

	for i := start; i < end; i++ {
		mod := visible[i]
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▸ ")
		}
		checkbox := unselectedStyle.Render("[ ]")
		if mod.Selected {
			checkbox = selectedStyle.Render("[✓]")
		}
		sizeStr := dimStyle.Render(fmt.Sprintf("(%s)", formatSize(mod.Size)))
		name := mod.Name
		if mod.Version != "" {
			name += "@" + mod.Version
		}
		nameStr := lipgloss.NewStyle().Render(name)
		if i == m.cursor {
			nameStr = cursorStyle.Render(name)
		}
		b.WriteString(fmt.Sprintf("  %s%s %s %s\n", cursor, checkbox, nameStr, sizeStr))
	}

	if len(visible) > pageSize {
		b.WriteString(dimStyle.Render(fmt.Sprintf("\n    Showing %d-%d of %d", start+1, end, len(visible))))
	}

	selectedCount, selectedSize := m.getSelectedCount()
	if selectedCount > 0 {
		b.WriteString("\n")
		b.WriteString(successStyle.Render(fmt.Sprintf("    %d selected (%s)", selectedCount, formatSize(selectedSize))))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("↑/↓: move | space: toggle | a: all | n: none | /: search | s: sort | enter: delete | q: back"))
	return b.String()
}

func (m Model) viewConfirm() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Confirm Deletion"))
	b.WriteString("\n\n")
	count, size := m.getSelectedCount()
	b.WriteString(yellowBox.Render(fmt.Sprintf("  Delete %d modules (%s)?", count, formatSize(size))))
	b.WriteString("\n\n")
	if m.dryRun {
		b.WriteString(dimStyle.Render("  (Dry run — no files will be deleted)"))
		b.WriteString("\n\n")
	}
	b.WriteString(warningStyle.Render("  Proceed?"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("y yes | n go back"))
	return b.String()
}

func (m Model) viewDeleting(msg string) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.spinner.View() + " ")
	b.WriteString(warningStyle.Render(msg))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  This may take a moment."))
	return b.String()
}

func (m Model) viewDone() string {
	var b strings.Builder
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render("  Error: " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press any key to exit"))
		return b.String()
	}

	b.WriteString(titleStyle.Render("  Done"))
	b.WriteString("\n\n")

	if m.dryRun {
		b.WriteString(greenBox.Render(fmt.Sprintf("  Dry run: would delete %d packages (%s)", m.deleteCount, formatSize(m.freedBytes))))
	} else {
		b.WriteString(greenBox.Render(fmt.Sprintf("  Deleted %d packages\n  Freed %s", m.deleteCount, formatSize(m.freedBytes))))
	}

	b.WriteString("\n\n")

	failedCount := 0
	for _, r := range m.deleteResults {
		if r.Error != nil {
			failedCount++
		}
	}
	if failedCount > 0 {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  %d deletions failed:", failedCount)))
		b.WriteString("\n")
		for _, r := range m.deleteResults {
			if r.Error != nil {
				b.WriteString(dimStyle.Render(fmt.Sprintf("    • %s: %s", r.Path, r.Error)))
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press any key to exit"))
	return b.String()
}

func (m Model) viewLangSelect() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Select Language"))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  Choose a language to browse its package cache"))
	b.WriteString("\n\n")

	for i, lc := range lang.Registry {
		cursor := "  "
		if i == m.langCursor {
			cursor = cursorStyle.Render("▸ ")
		}

		label := fmt.Sprintf("%s  %s", lc.Icon, lc.Name)

		paths := lc.CachePaths()
		avail := false
		for _, p := range paths {
			if pathExists(p) {
				avail = true
				break
			}
		}

		status := ""
		if !avail {
			status = dimStyle.Render("  (cache not found)")
		}

		nameStr := lipgloss.NewStyle().Render(label)
		if i == m.langCursor {
			nameStr = cursorStyle.Render(label)
		}

		b.WriteString(fmt.Sprintf("  %s%s%s\n", cursor, nameStr, status))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("↑/↓ navigate | enter select | q back"))
	return b.String()
}

func (m Model) viewCache() string {
	var b strings.Builder
	b.WriteString("\n")

	langName := m.cacheLang
	for _, lc := range lang.Registry {
		if lc.ID == m.cacheLang {
			langName = lc.Icon + " " + lc.Name
			break
		}
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("  Cache: %s", langName)))
	b.WriteString("\n\n")

	if m.cacheSearch {
		b.WriteString("  " + m.search.View())
		b.WriteString("\n\n")
	}

	totalSize := m.getCacheTotalSize()
	b.WriteString(dimStyle.Render(fmt.Sprintf("  %d packages (%s total)", len(m.cacheModules), formatSize(totalSize))))
	b.WriteString("\n\n")

	visible := m.getCacheVisible()
	if len(visible) == 0 {
		b.WriteString(dimStyle.Render("  No packages found in cache."))
		b.WriteString("\n\n")

		for _, lc := range lang.Registry {
			if lc.ID == m.cacheLang {
				paths := lc.CachePaths()
				b.WriteString(dimStyle.Render("  Expected cache locations:"))
				b.WriteString("\n")
				for _, p := range paths {
					exists := pathExists(p)
					status := selectedStyle.Render(" ✓")
					if !exists {
						status = errorStyle.Render(" ✗")
					}
					b.WriteString(dimStyle.Render(fmt.Sprintf("    %s%s", p, status)))
					b.WriteString("\n")
				}
				break
			}
		}

		b.WriteString("\n")
		b.WriteString(helpStyle.Render("q back"))
		return b.String()
	}

	pageSize := m.getPageSize()
	start := m.cacheOffset
	end := start + pageSize
	if end > len(visible) {
		end = len(visible)
	}

	for i := start; i < end; i++ {
		mod := visible[i]
		cursor := "  "
		if i == m.cacheCursor {
			cursor = cursorStyle.Render("▸ ")
		}
		checkbox := unselectedStyle.Render("[ ]")
		if mod.Selected {
			checkbox = selectedStyle.Render("[✓]")
		}
		sizeStr := dimStyle.Render(fmt.Sprintf("(%s)", formatSize(mod.Size)))
		name := mod.Name
		if mod.Version != "" {
			name += "@" + mod.Version
		}
		nameStr := lipgloss.NewStyle().Render(name)
		if i == m.cacheCursor {
			nameStr = cursorStyle.Render(name)
		}
		b.WriteString(fmt.Sprintf("  %s%s %s %s\n", cursor, checkbox, nameStr, sizeStr))
	}

	if len(visible) > pageSize {
		b.WriteString(dimStyle.Render(fmt.Sprintf("\n    Showing %d-%d of %d", start+1, end, len(visible))))
	}

	selectedCount, selectedSize := m.getCacheSelectedCount()
	if selectedCount > 0 {
		b.WriteString("\n")
		b.WriteString(successStyle.Render(fmt.Sprintf("    %d selected (%s)", selectedCount, formatSize(selectedSize))))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("↑/↓: move | space: toggle | a: all | n: none | /: search | s: sort | enter: delete | q: back"))
	return b.String()
}

func (m Model) viewCacheConfirm() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Confirm Deletion"))
	b.WriteString("\n\n")
	count, size := m.getCacheSelectedCount()
	b.WriteString(yellowBox.Render(fmt.Sprintf("  Delete %d packages from cache (%s)?", count, formatSize(size))))
	b.WriteString("\n\n")
	if m.dryRun {
		b.WriteString(dimStyle.Render("  (Dry run — no files will be deleted)"))
		b.WriteString("\n\n")
	}
	b.WriteString(warningStyle.Render("  Proceed?"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("y yes | n go back"))
	return b.String()
}
