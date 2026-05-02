package cache

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CachedModule struct {
	Name    string
	Version string
	Size    int64
	Path    string
}

func GetCachePath() string {
	cmd := exec.Command("go", "env", "GOMODCACHE")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func ScanCache(cachePath string) []CachedModule {
	if cachePath == "" {
		return nil
	}

	var modules []CachedModule

	entries, err := os.ReadDir(cachePath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()

		if name == "cache" || name == "download" {
			continue
		}

		if strings.Contains(name, "@") {
			parts := strings.SplitN(name, "@", 2)
			modName := parts[0]
			version := parts[1]
			fullPath := filepath.Join(cachePath, name)
			size := dirSize(fullPath)
			modules = append(modules, CachedModule{
				Name:    modName,
				Version: version,
				Size:    size,
				Path:    fullPath,
			})
			continue
		}

		orgDir := filepath.Join(cachePath, name)
		modEntries, err := os.ReadDir(orgDir)
		if err != nil {
			continue
		}

		for _, modEntry := range modEntries {
			if !modEntry.IsDir() {
				continue
			}

			modPath := filepath.Join(orgDir, modEntry.Name())
			versionEntries, err := os.ReadDir(modPath)
			if err != nil {
				continue
			}

			for _, verEntry := range versionEntries {
				if !verEntry.IsDir() {
					continue
				}

				version := verEntry.Name()
				if strings.HasPrefix(version, "@") || strings.Contains(version, "@") {
					continue
				}

				fullPath := filepath.Join(modPath, version)
				size := dirSize(fullPath)
				modName := name + "/" + modEntry.Name()
				modules = append(modules, CachedModule{
					Name:    modName,
					Version: version,
					Size:    size,
					Path:    fullPath,
				})
			}
		}
	}

	return modules
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}
