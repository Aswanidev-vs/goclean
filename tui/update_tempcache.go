package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleTempCacheKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.screen = ScreenMenu
		return m, nil
	case "up", "k":
		if m.tcCursor > 0 {
			m.tcCursor--
		}
	case "down", "j":
		if m.tcCursor < len(m.tcItems)-1 {
			m.tcCursor++
		}
	case " ":
		if len(m.tcItems) > 0 && m.tcCursor < len(m.tcItems) {
			m.tcSelected[m.tcCursor] = !m.tcSelected[m.tcCursor]
		}
	case "a":
		for i := range m.tcItems {
			m.tcSelected[i] = true
		}
	case "n":
		for i := range m.tcItems {
			m.tcSelected[i] = false
		}
	case "i":
		if len(m.tcItems) > 0 {
			m.screen = ScreenTempCacheDetail
		}
	case "enter":
		if m.hasTCSelection() {
			m.screen = ScreenTempCacheConfirm
		}
	}
	return m, nil
}

func (m Model) hasTCSelection() bool {
	for _, v := range m.tcSelected {
		if v {
			return true
		}
	}
	return false
}

func (m Model) handleTempCacheDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" || msg.String() == "ctrl+c" || msg.String() == "esc" {
		m.screen = ScreenTempCache
	}
	return m, nil
}

func (m Model) handleTempCacheConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.screen = ScreenTempCacheDeleting
		return m, tea.Batch(m.spinner.Tick, m.startTCClean())
	case "n", "q":
		m.screen = ScreenTempCache
		return m, nil
	}
	return m, nil
}
