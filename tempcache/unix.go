package tempcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func unixTemp() Item {
	return Item{
		ID:          "unix-temp",
		Name:        "System Temp Files",
		Description: "Temporary files in /tmp and system cache directories. Safe to delete — apps recreate them as needed.",
		Source:      "Operating System",
		Icon:        "🗑️",
		Criticality: Safe,
		Platforms:   []string{"linux", "darwin"},
		DetectFn: func() bool {
			return len(unixTempPaths()) > 0
		},
		SizeFn:  unixTempSize,
		CleanFn: unixTempClean,
	}
}

func unixTempPaths() []string {
	var paths []string

	if runtime.GOOS == "linux" {
		candidates := []string{"/tmp", "/var/tmp"}
		if home := homeDir(); home != "" {
			candidates = append(candidates, filepath.Join(home, ".cache"))
		}
		for _, p := range candidates {
			if pathExists(p) {
				paths = append(paths, p)
			}
		}
	}

	if runtime.GOOS == "darwin" {
		candidates := []string{"/tmp", "/private/tmp"}
		if home := homeDir(); home != "" {
			candidates = append(candidates,
				filepath.Join(home, "Library", "Caches"),
			)
		}
		for _, p := range candidates {
			if pathExists(p) {
				paths = append(paths, p)
			}
		}
	}

	return paths
}

func unixTempSize() int64 {
	var total int64
	for _, p := range unixTempPaths() {
		total += dirSize(p)
	}
	return total
}

func unixTempClean() (int64, error) {
	var freed int64
	var errs []string
	for _, p := range unixTempPaths() {
		entries, err := os.ReadDir(p)
		if err != nil {
			errs = append(errs, fmt.Sprintf("can't read %s: %v", p, err))
			continue
		}
		for _, entry := range entries {
			full := filepath.Join(p, entry.Name())
			sz := dirSize(full)
			if err := os.RemoveAll(full); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", full, err))
			} else {
				freed += sz
			}
		}
	}
	if len(errs) > 0 {
		return freed, errors.New(strings.Join(errs, "; "))
	}
	return freed, nil
}
