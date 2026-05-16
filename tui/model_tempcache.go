package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Aswanidev-vs/goclean/tempcache"
)

func (m Model) initTempCache() (Model, tea.Cmd) {
	items := tempcache.DetectAvailable()
	m.tcItems = items
	m.tcCursor = 0
	m.tcSelected = make(map[int]bool)
	m.tcSizes = make([]int64, len(items))
	for i := range m.tcSizes {
		m.tcSizes[i] = -1
	}
	m.tcComputing = true
	m.tcResults = nil
	m.tcTotalFreed = 0
	m.screen = ScreenTempCache
	return m, m.startTCSizeComputation()
}

func (m Model) startTCSizeComputation() tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.tcItems {
		idx, item := i, m.tcItems[i]
		cmds = append(cmds, func() tea.Msg {
			sizeSem <- struct{}{}
			defer func() { <-sizeSem }()
			return tcSizeMsg{index: idx, size: item.SizeFn()}
		})
	}
	cmds = append(cmds, func() tea.Msg {
		return tcSizesDoneMsg{}
	})
	return tea.Batch(cmds...)
}

func (m Model) startTCClean() tea.Cmd {
	return func() tea.Msg {
		var results []tcCleanResult
		var totalFreed int64
		for i, item := range m.tcItems {
			if !m.tcSelected[i] {
				continue
			}
			freed, err := item.CleanFn()
			results = append(results, tcCleanResult{
				name:  item.Name,
				freed: freed,
				err:   err,
			})
			if err == nil {
				totalFreed += freed
			}
		}
		return tcCleanDoneMsg{
			results:    results,
			totalFreed: totalFreed,
		}
	}
}

func (m *Model) getTCSelectedCount() (int, int64) {
	var count int
	var total int64
	for i := range m.tcItems {
		if m.tcSelected[i] {
			count++
			if m.tcSizes[i] > 0 {
				total += m.tcSizes[i]
			}
		}
	}
	return count, total
}

func (m *Model) getTCFreedCount() int64 {
	var total int64
	for _, r := range m.tcResults {
		if r.err == nil {
			total += r.freed
		}
	}
	return total
}
