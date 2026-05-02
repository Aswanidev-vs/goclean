package tui

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Aswanidev-vs/goclean/lang"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case scanResultMsg:
		m.projectCount = msg.projectCount
		m.totalModules = msg.totalModules
		m.unusedModules = msg.unusedModules
		m.freedBytes = msg.freedBytes
		if msg.err != nil {
			m.err = msg.err
			m.screen = ScreenDone
			return m, nil
		}
		m.screen = ScreenSummary
		m.updateFilteredIdx()
		return m, nil

	case cacheLoadMsg:
		if msg.err != nil {
			m.err = msg.err
			m.screen = ScreenDone
			return m, nil
		}
		m.cacheModules = msg.modules
		m.screen = ScreenCache
		m.cacheCursor = 0
		m.cacheOffset = 0
		m.cacheFilter = ""
		m.updateCacheFilteredIdx()
		return m, nil

	case deleteDoneMsg:
		m.deleteResults = msg.results
		m.freedBytes = msg.freedBytes
		m.deleteCount = msg.count
		m.screen = ScreenDone
		return m, nil

	case spinner.TickMsg:
		if isLoadingScreen(m.screen) {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

func isLoadingScreen(s Screen) bool {
	return s == ScreenLoading || s == ScreenDeleting || s == ScreenCacheLoading || s == ScreenCacheDeleting
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searchActive || m.cacheSearch {
		return m.handleSearchKey(msg)
	}

	switch m.screen {
	case ScreenMenu:
		return m.handleMenuKey(msg)
	case ScreenPaths:
		return m.handlePathsKey(msg)
	case ScreenLoading:
		return m.handleLoadingKey(msg)
	case ScreenSummary:
		return m.handleSummaryKey(msg)
	case ScreenList:
		return m.handleListKey(msg)
	case ScreenConfirm:
		return m.handleConfirmKey(msg)
	case ScreenDone:
		return m, tea.Quit
	case ScreenLangSelect:
		return m.handleLangSelectKey(msg)
	case ScreenCacheLoading:
		return m.handleLoadingKey(msg)
	case ScreenCache:
		return m.handleCacheKey(msg)
	case ScreenCacheConfirm:
		return m.handleCacheConfirmKey(msg)
	case ScreenCacheDeleting:
		return m.handleLoadingKey(msg)
	}

	return m, nil
}

func (m Model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < len(m.menuItems)-1 {
			m.menuCursor++
		}
	case "i":
		m.showInfo = !m.showInfo
	case "enter", " ":
		switch m.menuCursor {
		case 0:
			m.saveConfig()
			m.screen = ScreenLoading
			return m, m.startScan()
		case 1:
			m.screen = ScreenLangSelect
			m.langCursor = 0
			return m, nil
		case 2:
			m.screen = ScreenPaths
			m.pathInput.SetValue("")
			m.pathInput.Focus()
			return m, nil
		case 3:
			m.dryRun = !m.dryRun
			m.saveConfig()
			return m, nil
		case 4:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) handlePathsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.saveConfig()
		m.screen = ScreenMenu
		m.pathInput.Blur()
		return m, nil
	case "enter":
		val := strings.TrimSpace(m.pathInput.Value())
		if val != "" {
			abs, err := filepath.Abs(val)
			if err == nil {
				m.paths = append(m.paths, abs)
				m.cfg.AddPath(abs)
			}
			m.pathInput.SetValue("")
			m.saveConfig()
		}
		return m, nil
	case "ctrl+d":
		if len(m.paths) > 0 {
			removed := m.paths[len(m.paths)-1]
			m.paths = m.paths[:len(m.paths)-1]
			m.cfg.RemovePath(removed)
			m.saveConfig()
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.pathInput, cmd = m.pathInput.Update(msg)
	return m, cmd
}

func (m Model) handleLoadingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" || msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleSummaryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "m":
		m.screen = ScreenMenu
		return m, nil
	case "enter":
		if len(m.unusedModules) > 0 {
			m.screen = ScreenList
			m.cursor = 0
			m.offset = 0
		}
	}
	return m, nil
}

func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visible := m.getVisibleModules()
	visCount := len(visible)

	switch msg.String() {
	case "q", "ctrl+c":
		m.screen = ScreenSummary
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.offset {
				m.offset = m.cursor
			}
		}
	case "down", "j":
		if m.cursor < visCount-1 {
			m.cursor++
			pageSize := m.getPageSize()
			if m.cursor >= m.offset+pageSize {
				m.offset = m.cursor - pageSize + 1
			}
		}
	case " ":
		if visCount > 0 && m.cursor < visCount {
			idx := m.getRealIndex(m.cursor)
			m.toggleSelection(idx)
		}
	case "a":
		m.selectAll()
	case "n":
		m.deselectAll()
	case "/":
		m.searchActive = true
		m.search.Focus()
		return m, nil
	case "s":
		if m.sortMode == SortByName {
			m.sortMode = SortBySize
		} else {
			m.sortMode = SortByName
		}
		m.sortModules()
		m.updateFilteredIdx()
		m.cursor = 0
		m.offset = 0
	case "enter":
		if m.hasSelection() {
			m.screen = ScreenConfirm
		}
	}
	return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searchActive {
		switch msg.String() {
		case "enter":
			m.filterText = m.search.Value()
			m.searchActive = false
			m.search.Blur()
			m.updateFilteredIdx()
			m.cursor = 0
			m.offset = 0
			return m, nil
		case "esc":
			m.searchActive = false
			m.search.Blur()
			m.filterText = ""
			m.updateFilteredIdx()
			m.cursor = 0
			m.offset = 0
			return m, nil
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	}

	if m.cacheSearch {
		switch msg.String() {
		case "enter":
			m.cacheFilter = m.search.Value()
			m.cacheSearch = false
			m.search.Blur()
			m.updateCacheFilteredIdx()
			m.cacheCursor = 0
			m.cacheOffset = 0
			return m, nil
		case "esc":
			m.cacheSearch = false
			m.search.Blur()
			m.cacheFilter = ""
			m.updateCacheFilteredIdx()
			m.cacheCursor = 0
			m.cacheOffset = 0
			return m, nil
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.screen = ScreenDeleting
		return m, tea.Batch(m.spinner.Tick, m.startDelete())
	case "n", "q":
		m.screen = ScreenList
		return m, nil
	}
	return m, nil
}

func (m Model) handleLangSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.screen = ScreenMenu
		return m, nil
	case "up", "k":
		if m.langCursor > 0 {
			m.langCursor--
		}
	case "down", "j":
		if m.langCursor < len(lang.Registry)-1 {
			m.langCursor++
		}
	case "enter", " ":
		selected := lang.Registry[m.langCursor]
		m.cacheLang = selected.ID
		m.screen = ScreenCacheLoading
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			return loadLangCache(selected.ID)
		})
	}
	return m, nil
}

func (m Model) handleCacheKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visible := m.getCacheVisible()
	visCount := len(visible)

	switch msg.String() {
	case "q", "ctrl+c":
		m.screen = ScreenLangSelect
		return m, nil
	case "up", "k":
		if m.cacheCursor > 0 {
			m.cacheCursor--
			if m.cacheCursor < m.cacheOffset {
				m.cacheOffset = m.cacheCursor
			}
		}
	case "down", "j":
		if m.cacheCursor < visCount-1 {
			m.cacheCursor++
			pageSize := m.getPageSize()
			if m.cacheCursor >= m.cacheOffset+pageSize {
				m.cacheOffset = m.cacheCursor - pageSize + 1
			}
		}
	case " ":
		if visCount > 0 && m.cacheCursor < visCount {
			idx := m.getCacheRealIndex(m.cacheCursor)
			m.toggleCacheSelection(idx)
		}
	case "a":
		m.selectAllCache()
	case "n":
		m.deselectAllCache()
	case "/":
		m.cacheSearch = true
		m.search.Focus()
		return m, nil
	case "s":
		if m.cacheSort == SortByName {
			m.cacheSort = SortBySize
		} else {
			m.cacheSort = SortByName
		}
		m.sortCacheModules()
		m.updateCacheFilteredIdx()
		m.cacheCursor = 0
		m.cacheOffset = 0
	case "enter":
		if m.hasCacheSelection() {
			m.screen = ScreenCacheConfirm
		}
	}
	return m, nil
}

func (m Model) handleCacheConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.screen = ScreenCacheDeleting
		return m, tea.Batch(m.spinner.Tick, m.startCacheDelete())
	case "n", "q":
		m.screen = ScreenCache
		return m, nil
	}
	return m, nil
}

func (m *Model) getPageSize() int {
	if m.height > 10 {
		return m.height - 10
	}
	return 10
}

func (m *Model) getVisibleModules() []Pkg {
	if len(m.filteredIdx) == 0 {
		return m.unusedModules
	}
	var visible []Pkg
	for _, idx := range m.filteredIdx {
		if idx < len(m.unusedModules) {
			visible = append(visible, m.unusedModules[idx])
		}
	}
	return visible
}

func (m *Model) getRealIndex(visibleIdx int) int {
	if visibleIdx < len(m.filteredIdx) {
		return m.filteredIdx[visibleIdx]
	}
	return visibleIdx
}

func (m *Model) updateFilteredIdx() {
	m.filteredIdx = nil
	for i, mod := range m.unusedModules {
		if m.filterText == "" || strings.Contains(strings.ToLower(mod.Name), strings.ToLower(m.filterText)) {
			m.filteredIdx = append(m.filteredIdx, i)
		}
	}
}

func (m *Model) toggleSelection(idx int) {
	if idx < len(m.unusedModules) {
		m.unusedModules[idx].Selected = !m.unusedModules[idx].Selected
	}
}

func (m *Model) selectAll() {
	for i := range m.unusedModules {
		m.unusedModules[i].Selected = true
	}
}

func (m *Model) deselectAll() {
	for i := range m.unusedModules {
		m.unusedModules[i].Selected = false
	}
}

func (m *Model) hasSelection() bool {
	for _, mod := range m.unusedModules {
		if mod.Selected {
			return true
		}
	}
	return false
}

func (m *Model) getSelectedCount() (int, int64) {
	var count int
	var size int64
	for _, mod := range m.unusedModules {
		if mod.Selected {
			count++
			size += mod.Size
		}
	}
	return count, size
}

func (m *Model) sortModules() {
	switch m.sortMode {
	case SortByName:
		sortByName(m.unusedModules)
	case SortBySize:
		sortBySize(m.unusedModules)
	}
}

func (m *Model) getCacheVisible() []Pkg {
	if len(m.cacheFIdx) == 0 {
		return m.cacheModules
	}
	var visible []Pkg
	for _, idx := range m.cacheFIdx {
		if idx < len(m.cacheModules) {
			visible = append(visible, m.cacheModules[idx])
		}
	}
	return visible
}

func (m *Model) getCacheRealIndex(visibleIdx int) int {
	if visibleIdx < len(m.cacheFIdx) {
		return m.cacheFIdx[visibleIdx]
	}
	return visibleIdx
}

func (m *Model) updateCacheFilteredIdx() {
	m.cacheFIdx = nil
	for i, mod := range m.cacheModules {
		if m.cacheFilter == "" || strings.Contains(strings.ToLower(mod.Name), strings.ToLower(m.cacheFilter)) {
			m.cacheFIdx = append(m.cacheFIdx, i)
		}
	}
}

func (m *Model) toggleCacheSelection(idx int) {
	if idx < len(m.cacheModules) {
		m.cacheModules[idx].Selected = !m.cacheModules[idx].Selected
	}
}

func (m *Model) selectAllCache() {
	for i := range m.cacheModules {
		m.cacheModules[i].Selected = true
	}
}

func (m *Model) deselectAllCache() {
	for i := range m.cacheModules {
		m.cacheModules[i].Selected = false
	}
}

func (m *Model) hasCacheSelection() bool {
	for _, mod := range m.cacheModules {
		if mod.Selected {
			return true
		}
	}
	return false
}

func (m *Model) getCacheSelectedCount() (int, int64) {
	var count int
	var size int64
	for _, mod := range m.cacheModules {
		if mod.Selected {
			count++
			size += mod.Size
		}
	}
	return count, size
}

func (m *Model) getCacheTotalSize() int64 {
	var size int64
	for _, mod := range m.cacheModules {
		size += mod.Size
	}
	return size
}

func (m *Model) sortCacheModules() {
	switch m.cacheSort {
	case SortByName:
		sortByName(m.cacheModules)
	case SortBySize:
		sortBySize(m.cacheModules)
	}
}

func sortByName(mods []Pkg) {
	for i := 1; i < len(mods); i++ {
		for j := i; j > 0 && mods[j].Name < mods[j-1].Name; j-- {
			mods[j], mods[j-1] = mods[j-1], mods[j]
		}
	}
}

func sortBySize(mods []Pkg) {
	for i := 1; i < len(mods); i++ {
		for j := i; j > 0 && mods[j].Size > mods[j-1].Size; j-- {
			mods[j], mods[j-1] = mods[j-1], mods[j]
		}
	}
}
