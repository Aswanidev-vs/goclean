package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var ignoreDirs = map[string]bool{
	".git":        true,
	"node_modules": true,
	"vendor":      true,
}

func DefaultPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var paths []string
	seen := map[string]bool{}

	addPath := func(p string) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return
		}
		if seen[abs] {
			return
		}
		info, err := os.Stat(abs)
		if err == nil && info.IsDir() {
			seen[abs] = true
			paths = append(paths, abs)
		}
	}

	cmd := exec.Command("go", "env", "GOPATH")
	out, err := cmd.Output()
	if err == nil {
		gopath := strings.TrimSpace(string(out))
		if gopath != "" {
			addPath(filepath.Join(gopath, "src"))
		}
	}

	commonNames := []string{
		"go/src",
		"projects",
		"workspace",
		"work",
		"code",
		"src",
		"dev",
		"repos",
		"Documents/projects",
		"Documents/workspace",
		"Documents/code",
		"Documents/dev",
	}

	for _, name := range commonNames {
		addPath(filepath.Join(home, name))
	}

	entries, err := os.ReadDir(home)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			lower := strings.ToLower(entry.Name())
			if lower == "desktop" || lower == "documents" || lower == "downloads" ||
				lower == "pictures" || lower == "music" || lower == "videos" ||
				lower == ".cache" || lower == ".config" || lower == ".local" ||
				lower == ".cargo" || lower == ".npm" || lower == ".gradle" ||
				lower == ".m2" || lower == ".nuget" || lower == ".vscode" ||
				lower == ".android" || lower == ".ssh" || lower == ".docker" {
				continue
			}
			addPath(filepath.Join(home, entry.Name()))
		}
	}

	return paths
}

func DiscoverProjects(paths []string) []string {
	var mu sync.Mutex
	var projects []string
	var wg sync.WaitGroup

	for _, p := range paths {
		wg.Add(1)
		go func(root string) {
			defer wg.Done()
			found := walkForProjects(root)
			mu.Lock()
			projects = append(projects, found...)
			mu.Unlock()
		}(p)
	}

	wg.Wait()
	return projects
}

func walkForProjects(root string) []string {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return nil
	}

	var projects []string

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	hasGoMod := false
	var subDirs []string

	for _, entry := range entries {
		if entry.Name() == "go.mod" && !entry.IsDir() {
			hasGoMod = true
		}
		if entry.IsDir() && !ignoreDirs[entry.Name()] {
			subDirs = append(subDirs, filepath.Join(root, entry.Name()))
		}
	}

	if hasGoMod {
		projects = append(projects, root)
		return projects
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, dir := range subDirs {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			found := walkForProjects(d)
			if len(found) > 0 {
				mu.Lock()
				projects = append(projects, found...)
				mu.Unlock()
			}
		}(dir)
	}

	wg.Wait()
	return projects
}
