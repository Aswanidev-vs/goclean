package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Aswanidev-vs/goclean/config"
	"github.com/Aswanidev-vs/goclean/scanner"
	"github.com/Aswanidev-vs/goclean/tui"
)

var version = "v1.0.0"

func main() {
	pathsFlag := flag.String("paths", "", "Comma-separated list of directories to scan")
	dryRun := flag.Bool("dry-run", false, "Simulate deletion without removing files")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	showVersion := flag.Bool("version", false, "Show version")
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

	model := tui.NewModel(scanPaths, cfg.DryRun, *verbose, version)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
