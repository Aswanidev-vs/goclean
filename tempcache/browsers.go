package tempcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func chromeCache() Item {
	return Item{
		ID:          "chrome-cache",
		Name:        "Chrome Browser Cache",
		Description: "Cached web pages, images, and media files from Google Chrome. These will be re-downloaded when you visit the same sites again.",
		Source:      "Google Chrome",
		Icon:        "🌐",
		Criticality: Safe,
		Platforms:   []string{"windows", "linux", "darwin"},
		DetectFn: func() bool {
			for _, p := range chromeCachePaths() {
				if pathExists(p) {
					return true
				}
			}
			return false
		},
		SizeFn: chromeCacheSize,
		CleanFn: chromeCacheClean,
	}
}

func chromeCachePaths() []string {
	var paths []string
	switch {
	case os.Getenv("LOCALAPPDATA") != "":
		base := filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data")
		paths = addChromeProfilePaths(base, paths)
	case os.Getenv("HOME") != "":
		base := filepath.Join(os.Getenv("HOME"), ".config", "google-chrome")
		paths = addChromeProfilePaths(base, paths)
		base2 := filepath.Join(os.Getenv("HOME"), "Library", "Caches", "Google", "Chrome")
		paths = addChromeProfilePaths(base2, paths)
	}
	return paths
}

func addChromeProfilePaths(base string, paths []string) []string {
	entries, err := os.ReadDir(base)
	if err != nil {
		return paths
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "Profile") && name != "Default" {
			continue
		}
		cacheDir := filepath.Join(base, name, "Cache")
		if pathExists(cacheDir) {
			paths = append(paths, cacheDir)
		}
		codeCache := filepath.Join(base, name, "Code Cache")
		if pathExists(codeCache) {
			paths = append(paths, codeCache)
		}
	}
	return paths
}

func chromeCacheSize() int64 {
	var total int64
	for _, p := range chromeCachePaths() {
		total += dirSize(p)
	}
	return total
}

func chromeCacheClean() (int64, error) {
	var freed int64
	var errs []string
	for _, p := range chromeCachePaths() {
		sz := dirSize(p)
		if err := os.RemoveAll(p); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", p, err))
		} else {
			freed += sz
		}
	}
	if len(errs) > 0 {
		return freed, errors.New(strings.Join(errs, "; "))
	}
	return freed, nil
}

func firefoxCache() Item {
	return Item{
		ID:          "firefox-cache",
		Name:        "Firefox Browser Cache",
		Description: "Cached web pages, images, and media files from Mozilla Firefox. These will be re-downloaded when you visit the same sites again.",
		Source:      "Mozilla Firefox",
		Icon:        "🦊",
		Criticality: Safe,
		Platforms:   []string{"windows", "linux", "darwin"},
		DetectFn: func() bool {
			for _, p := range firefoxCachePaths() {
				if pathExists(p) {
					return true
				}
			}
			return false
		},
		SizeFn: firefoxCacheSize,
		CleanFn: firefoxCacheClean,
	}
}

func firefoxCachePaths() []string {
	var paths []string
	profiles := firefoxProfileDirs()
	for _, prof := range profiles {
		cache2 := filepath.Join(prof, "cache2")
		if pathExists(cache2) {
			paths = append(paths, cache2)
		}
		offline := filepath.Join(prof, "offlinedata")
		if pathExists(offline) {
			paths = append(paths, offline)
		}
	}
	return paths
}

func firefoxProfileDirs() []string {
	var dirs []string
	candidates := []string{os.Getenv("APPDATA"), os.Getenv("LOCALAPPDATA")}
	if os.Getenv("HOME") != "" {
		candidates = append(candidates,
			filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox"),
			filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox"),
			filepath.Join(os.Getenv("HOME"), "snap", "firefox", "common", ".mozilla", "firefox"),
		)
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		profilesIni := filepath.Join(c, "Mozilla", "Firefox", "profiles.ini")
		if !pathExists(profilesIni) {
			profilesIni = filepath.Join(c, "profiles.ini")
			if !pathExists(profilesIni) {
				continue
			}
		}
		base := filepath.Dir(profilesIni)
		data, _ := os.ReadFile(profilesIni)
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToLower(line), "path=") {
				rel := strings.TrimPrefix(line, "path=")
				rel = strings.TrimPrefix(rel, "Path=")
				profDir := filepath.Join(base, rel)
				if pathExists(profDir) {
					dirs = append(dirs, profDir)
				}
			}
		}
	}
	return dirs
}

func firefoxCacheSize() int64 {
	var total int64
	for _, p := range firefoxCachePaths() {
		total += dirSize(p)
	}
	return total
}

func firefoxCacheClean() (int64, error) {
	var freed int64
	var errs []string
	for _, p := range firefoxCachePaths() {
		sz := dirSize(p)
		if err := os.RemoveAll(p); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", p, err))
		} else {
			freed += sz
		}
	}
	if len(errs) > 0 {
		return freed, errors.New(strings.Join(errs, "; "))
	}
	return freed, nil
}

func edgeCache() Item {
	return Item{
		ID:          "edge-cache",
		Name:        "Edge Browser Cache",
		Description: "Cached web pages, images, and media files from Microsoft Edge (Chromium). These will be re-downloaded when you visit the same sites again.",
		Source:      "Microsoft Edge",
		Icon:        "🌐",
		Criticality: Safe,
		Platforms:   []string{"windows", "linux", "darwin"},
		DetectFn: func() bool {
			for _, p := range edgeCachePaths() {
				if pathExists(p) {
					return true
				}
			}
			return false
		},
		SizeFn: edgeCacheSize,
		CleanFn: edgeCacheClean,
	}
}

func edgeCachePaths() []string {
	var paths []string
	var base string
	if localApp := os.Getenv("LOCALAPPDATA"); localApp != "" {
		base = filepath.Join(localApp, "Microsoft", "Edge", "User Data")
	} else if home := os.Getenv("HOME"); home != "" {
		base = filepath.Join(home, ".config", "microsoft-edge")
	}
	if base == "" || !pathExists(base) {
		return paths
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return paths
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "Profile") && name != "Default" {
			continue
		}
		cacheDir := filepath.Join(base, name, "Cache")
		if pathExists(cacheDir) {
			paths = append(paths, cacheDir)
		}
		codeCache := filepath.Join(base, name, "Code Cache")
		if pathExists(codeCache) {
			paths = append(paths, codeCache)
		}
	}
	return paths
}

func edgeCacheSize() int64 {
	var total int64
	for _, p := range edgeCachePaths() {
		total += dirSize(p)
	}
	return total
}

func edgeCacheClean() (int64, error) {
	var freed int64
	var errs []string
	for _, p := range edgeCachePaths() {
		sz := dirSize(p)
		if err := os.RemoveAll(p); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", p, err))
		} else {
			freed += sz
		}
	}
	if len(errs) > 0 {
		return freed, errors.New(strings.Join(errs, "; "))
	}
	return freed, nil
}
