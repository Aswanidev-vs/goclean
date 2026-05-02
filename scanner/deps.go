package scanner

import (
	"os/exec"
	"strings"
	"sync"
)

type Module struct {
	Name    string
	Version string
}

func AggregateDeps(projects []string) map[string][]string {
	modules := make(map[string][]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 8)

	for _, proj := range projects {
		wg.Add(1)
		go func(projectPath string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			deps := listModules(projectPath)
			mu.Lock()
			for _, dep := range deps {
				key := dep.Name + "@" + dep.Version
				projects := modules[key]
				if !contains(projects, projectPath) {
					modules[key] = append(projects, projectPath)
				}
			}
			mu.Unlock()
		}(proj)
	}

	wg.Wait()
	return modules
}

func listModules(projectPath string) []Module {
	cmd := exec.Command("go", "list", "-m", "all")
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var modules []Module
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			modules = append(modules, Module{
				Name:    parts[0],
				Version: parts[1],
			})
		} else if len(parts) == 1 {
			modules = append(modules, Module{
				Name:    parts[0],
				Version: "(devel)",
			})
		}
	}
	return modules
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
