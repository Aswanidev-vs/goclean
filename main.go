package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Aswanidev-vs/goclean/config"
	"github.com/Aswanidev-vs/goclean/scanner"
	"github.com/Aswanidev-vs/goclean/tui"
)

var version = "v2.0.0"

func main() {
	pathsFlag := flag.String("paths", "", "Comma-separated list of directories to scan")
	dryRun := flag.Bool("dry-run", false, "Simulate deletion without removing files")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	showVersion := flag.Bool("version", false, "Show version")
	exportPath := flag.String("export", "", "Export scan report to JSON file")
	minSizeStr := flag.String("min-size", "", "Minimum size filter (e.g. 1MB, 100KB, 1GB)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("goclean %s\n", version)
		os.Exit(0)
	}

	cfg := config.Load()

	if *dryRun {
		cfg.DryRun = true
	}

	var scanPaths []string
	if *pathsFlag != "" {
		for _, p := range strings.Split(*pathsFlag, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				scanPaths = append(scanPaths, p)
			}
		}
		cfg.Paths = scanPaths
		config.Save(cfg)
	} else if len(cfg.Paths) > 0 {
		scanPaths = cfg.Paths
	} else {
		scanPaths = scanner.DefaultPaths()
		if len(scanPaths) > 0 {
			cfg.Paths = scanPaths
			config.Save(cfg)
		}
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Scan paths: %v\n", scanPaths)
	}

	var minSize int64
	if *minSizeStr != "" {
		minSize = parseSize(*minSizeStr)
	}

	model := tui.NewModel(scanPaths, cfg.DryRun, *verbose, version, *exportPath, minSize)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseSize(s string) int64 {
	s = strings.TrimSpace(strings.ToUpper(s))
	multiplier := int64(1)
	switch {
	case strings.HasSuffix(s, "GB"):
		multiplier = 1 << 30
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1 << 20
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		multiplier = 1 << 10
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}
	s = strings.TrimSpace(s)
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(val * float64(multiplier))
}
