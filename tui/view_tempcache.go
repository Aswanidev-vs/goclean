package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func criticalityColor(criticality int) lipgloss.Color {
	switch criticality {
	case 0:
		return lipgloss.Color("42")
	case 1:
		return lipgloss.Color("214")
	case 2:
		return lipgloss.Color("196")
	default:
		return lipgloss.Color("241")
	}
}

func criticalityLabel(criticality int) string {
	switch criticality {
	case 0:
		return "Safe"
	case 1:
		return "Moderate"
	case 2:
		return "Caution"
	default:
		return "Unknown"
	}
}

func (m Model) viewTempCache() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Clean Temp & Cache Files"))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  Select items to clean:"))
	b.WriteString("\n\n")

	if len(m.tcItems) == 0 {
		b.WriteString(dimStyle.Render("  No temp or cache files detected."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("q quit"))
		return b.String()
	}

	for i, item := range m.tcItems {
		prefix := "  "
		if i == m.tcCursor {
			prefix = cursorStyle.Render("▸ ")
		}

		checkbox := unselectedStyle.Render("[ ]")
		if m.tcSelected[i] {
			checkbox = selectedStyle.Render("[✓]")
		}

		critColor := criticalityColor(int(item.Criticality))
		critLabel := criticalityLabel(int(item.Criticality))
		critStyle := lipgloss.NewStyle().Foreground(critColor).Bold(true)
		critBadge := critStyle.Render(fmt.Sprintf("[%s]", critLabel))

		sizeStr := dimStyle.Render("...")
		if !m.tcComputing && m.tcSizes[i] >= 0 {
			sizeStr = dimStyle.Render(fmt.Sprintf("(%s)", formatSize(m.tcSizes[i])))
		} else if m.tcSizes[i] >= 0 {
			sizeStr = dimStyle.Render(fmt.Sprintf("(%s)", formatSize(m.tcSizes[i])))
		}

		label := fmt.Sprintf("%s  %s", item.Icon, item.Name)
		nameStr := lipgloss.NewStyle().Render(label)
		if i == m.tcCursor {
			nameStr = cursorStyle.Render(label)
		}

		b.WriteString(fmt.Sprintf("  %s%s %s %s %s\n", prefix, checkbox, nameStr, critBadge, sizeStr))

		if i == m.tcCursor && m.tcSizes[i] >= 0 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("     %s", item.Description)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	selectedCount, selectedSize := m.getTCSelectedCount()
	if selectedCount > 0 {
		b.WriteString(successStyle.Render(fmt.Sprintf("    %d selected (%s)", selectedCount, formatSize(selectedSize))))
		b.WriteString("\n")
	}

	if m.tcComputing {
		b.WriteString(dimStyle.Render("  Computing sizes..."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: move | space: toggle | a: all | n: none | i: info | enter: clean selected | q: back"))
	return b.String()
}

func (m Model) viewTempCacheDetail() string {
	var b strings.Builder
	if m.tcCursor >= len(m.tcItems) {
		return ""
	}
	item := m.tcItems[m.tcCursor]

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("  %s %s", item.Icon, item.Name)))
	b.WriteString("\n\n")

	critColor := criticalityColor(int(item.Criticality))
	critLabel := criticalityLabel(int(item.Criticality))
	critStyle := lipgloss.NewStyle().Foreground(critColor).Bold(true)

	b.WriteString(greenBox.Render(fmt.Sprintf(
		"Source:      %s\n"+
			"Criticality: %s\n"+
			"Size:        %s\n\n"+
			"%s",
		item.Source,
		critStyle.Render(critLabel),
		formatSize(m.tcSizes[m.tcCursor]),
		item.Description,
	)))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("esc | q back"))
	return b.String()
}

func (m Model) viewTempCacheConfirm() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Confirm Clean"))
	b.WriteString("\n\n")

	count, size := m.getTCSelectedCount()
	b.WriteString(yellowBox.Render(fmt.Sprintf("  Clean %d items (%s)?", count, formatSize(size))))
	b.WriteString("\n\n")

	for i, item := range m.tcItems {
		if !m.tcSelected[i] {
			continue
		}
		critColor := criticalityColor(int(item.Criticality))
		critLabel := criticalityLabel(int(item.Criticality))
		critStyle := lipgloss.NewStyle().Foreground(critColor).Bold(true)
		line := fmt.Sprintf("  %s %s  %s  %s",
			item.Icon, item.Name, formatSize(m.tcSizes[i]), critStyle.Render(fmt.Sprintf("[%s]", critLabel)))
		b.WriteString(line)
		b.WriteString("\n")
	}

	if m.dryRun {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  (Dry run — no files will be deleted)"))
	}

	b.WriteString("\n\n")
	b.WriteString(warningStyle.Render("  Proceed?"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("y yes | n go back"))
	return b.String()
}

func (m Model) viewTempCacheDeleting() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.spinner.View() + " ")
	b.WriteString(warningStyle.Render("Cleaning selected items..."))
	b.WriteString("\n\n")
	selectedCount, _ := m.getTCSelectedCount()
	if selectedCount > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  Cleaning %d items...", selectedCount)))
		b.WriteString("\n")
		bar := m.progress.ViewAs(0.0)
		b.WriteString("  " + bar)
	} else {
		b.WriteString(dimStyle.Render("  This may take a moment."))
	}
	return b.String()
}

func (m Model) viewTempCacheDone() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Clean Results"))
	b.WriteString("\n\n")

	totalFreed := m.getTCFreedCount()
	b.WriteString(greenBox.Render(fmt.Sprintf("  Freed %s", formatSize(totalFreed))))
	b.WriteString("\n\n")

	for _, r := range m.tcResults {
		if r.err != nil {
			b.WriteString(errorStyle.Render(fmt.Sprintf("  ✗ %s: %v", r.name, r.err)))
		} else if r.freed > 0 {
			b.WriteString(selectedStyle.Render(fmt.Sprintf("  ✓ %s (%s)", r.name, formatSize(r.freed))))
		} else {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  - %s (nothing to clean)", r.name)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Press any key to return to menu"))
	return b.String()
}
