package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Aswanidev-vs/goclean/cache"
	"github.com/Aswanidev-vs/goclean/cleaner"
	"github.com/Aswanidev-vs/goclean/config"
	"github.com/Aswanidev-vs/goclean/lang"
	"github.com/Aswanidev-vs/goclean/scanner"
)

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenPaths
	ScreenLoading
	ScreenSummary
	ScreenList
	ScreenConfirm
	ScreenDeleting
	ScreenDone
	ScreenLangSelect
	ScreenCacheLoading
	ScreenCache
	ScreenCacheConfirm
	ScreenCacheDeleting
)

type Pkg struct {
	Name     string
	Version  string
	Size     int64
	Path     string
	Selected bool
}

type SortMode int

const (
	SortByName SortMode = iota
	SortBySize
)

type Model struct {
	screen     Screen
	menuCursor int
	menuItems  []string
	version    string

	spinner   spinner.Model
	progress  progress.Model
	search    textinput.Model
	pathInput textinput.Model

	projectCount  int
	totalModules  int
	unusedModules []Pkg
	cursor        int
	sortMode      SortMode
	offset        int
	searchActive  bool
	filterText    string
	filteredIdx   []int

	langCursor int

	showInfo bool

	cacheModules []Pkg
	cacheCursor  int
	cacheOffset  int
	cacheSearch  bool
	cacheFilter  string
	cacheFIdx    []int
	cacheSort    SortMode
	cacheLang    string

	deleteResults []cleaner.DeleteResult
	freedBytes    int64
	deleteCount   int

	width  int
	height int

	err     error
	cfg     config.Config
	paths   []string
	dryRun  bool
	verbose bool
}

type scanResultMsg struct {
	projectCount  int
	totalModules  int
	unusedModules []Pkg
	freedBytes    int64
	err           error
}

type cacheLoadMsg struct {
	modules []Pkg
	err     error
}

type deleteDoneMsg struct {
	results    []cleaner.DeleteResult
	freedBytes int64
	count      int
}

func NewModel(paths []string, dryRun, verbose bool, ver string) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	ti := textinput.New()
	ti.Placeholder = "Search modules..."
	ti.CharLimit = 200
	ti.Width = 50

	pi := textinput.New()
	pi.Placeholder = "e.g. C:\\Users\\me\\projects"
	pi.CharLimit = 500
	pi.Width = 60

	cfg := config.Load()

	return Model{
		screen:    ScreenMenu,
		menuItems: []string{"Start Scan", "Browse Cache", "Configure Paths", "Toggle Dry-Run", "Quit"},
		spinner:   sp,
		progress:  p,
		search:    ti,
		pathInput: pi,
		cfg:       cfg,
		paths:     paths,
		dryRun:    dryRun,
		verbose:   verbose,
		showInfo:  true,
		version:   ver,
	}
}

func (m Model) saveConfig() {
	m.cfg.Paths = m.paths
	m.cfg.DryRun = m.dryRun
	config.Save(m.cfg)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) startScan() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		return doScan(m.paths)
	})
}

func doScan(paths []string) scanResultMsg {
	projects := scanner.DiscoverProjects(paths)
	if len(projects) == 0 {
		return scanResultMsg{projectCount: 0}
	}
	usedModules := scanner.AggregateDeps(projects)
	cachePath := cache.GetCachePath()
	cachedModules := cache.ScanCache(cachePath)

	var unused []Pkg
	var totalFreed int64
	for _, cm := range cachedModules {
		key := cm.Name + "@" + cm.Version
		if _, used := usedModules[key]; !used {
			unused = append(unused, Pkg{
				Name:    cm.Name,
				Version: cm.Version,
				Size:    cm.Size,
				Path:    cm.Path,
			})
			totalFreed += cm.Size
		}
	}
	return scanResultMsg{
		projectCount:  len(projects),
		totalModules:  len(usedModules),
		unusedModules: unused,
		freedBytes:    totalFreed,
	}
}

func loadLangCache(langID string) cacheLoadMsg {
	for _, lc := range lang.Registry {
		if lc.ID == langID {
			paths := lc.CachePaths()
			var all []Pkg
			for _, p := range paths {
				if !pathExists(p) {
					continue
				}
				pkgs := lc.ScanFunc(p)
				for _, pkg := range pkgs {
					all = append(all, Pkg{
						Name:    pkg.Name,
						Version: pkg.Version,
						Size:    pkg.Size,
						Path:    pkg.Path,
					})
				}
			}
			return cacheLoadMsg{modules: all}
		}
	}
	return cacheLoadMsg{err: fmt.Errorf("unknown language: %s", langID)}
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (m Model) startDelete() tea.Cmd {
	selected := m.getSelectedPaths()
	return func() tea.Msg {
		var totalFreed int64
		for _, mod := range m.unusedModules {
			if mod.Selected {
				totalFreed += mod.Size
			}
		}
		results := cleaner.DeleteModules(selected, 4, nil)
		return deleteDoneMsg{results: results, freedBytes: totalFreed, count: len(selected)}
	}
}

func (m Model) startCacheDelete() tea.Cmd {
	selected := m.getCacheSelectedPaths()
	return func() tea.Msg {
		var totalFreed int64
		for _, mod := range m.cacheModules {
			if mod.Selected {
				totalFreed += mod.Size
			}
		}
		results := cleaner.DeleteModules(selected, 4, nil)
		return deleteDoneMsg{results: results, freedBytes: totalFreed, count: len(selected)}
	}
}

func (m Model) getSelectedPaths() []string {
	var paths []string
	for _, mod := range m.unusedModules {
		if mod.Selected && mod.Path != "" {
			paths = append(paths, mod.Path)
		}
	}
	return paths
}

func (m Model) getCacheSelectedPaths() []string {
	var paths []string
	for _, mod := range m.cacheModules {
		if mod.Selected && mod.Path != "" {
			paths = append(paths, mod.Path)
		}
	}
	return paths
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func (m Model) pathDisplay() string {
	if len(m.paths) == 0 {
		return "(none)"
	}
	var short []string
	for _, p := range m.paths {
		home, _ := filepath.Abs(p)
		short = append(short, home)
	}
	return strings.Join(short, "\n    ")
}

func (m Model) pathDisplayShort() string {
	if len(m.paths) == 0 {
		return "(none)"
	}
	home, _ := filepath.Abs(homePath())
	var short []string
	for _, p := range m.paths {
		abs, _ := filepath.Abs(p)
		if strings.HasPrefix(abs, home) {
			short = append(short, "~"+abs[len(home):])
		} else {
			short = append(short, abs)
		}
	}
	return strings.Join(short, ", ")
}

func homePath() string {
	h, _ := os.UserHomeDir()
	return h
}

func (m Model) pathShort(p string) string {
	home, _ := filepath.Abs(p)
	return home
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	unselectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 2)

	menuActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true).
			Padding(0, 2)

	greenBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("42")).
			Padding(0, 1)

	yellowBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("214")).
			Padding(0, 1)

	redBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(0, 1)

	infoBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2)
)
