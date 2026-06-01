package lang

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Aswanidev-vs/goclean/util"
)

type CachedPackage struct {
	Name    string
	Version string
	Size    int64
	Path    string
}

type LangCache struct {
	ID         string
	Name       string
	Icon       string
	CachePaths func() []string
	ScanFunc   func(cachePath string) []CachedPackage
}

var Registry []LangCache

func init() {
	Registry = []LangCache{
		goCache(),
		pythonCache(),
		rustCache(),
		nodeCache(),
		javaMavenCache(),
		javaGradleCache(),
		dotnetCache(),
	}
}

func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func pathExists(p string) bool {
	return util.PathExists(p)
}

func envOrDefault(key, def string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return def
}

func goCache() LangCache {
	return LangCache{
		ID:   "go",
		Name: "Go",
		Icon: "🐹",
		CachePaths: func() []string {
			cmd := exec.Command("go", "env", "GOMODCACHE")
			out, err := cmd.Output()
			if err == nil {
				p := strings.TrimSpace(string(out))
				if p != "" {
					return []string{p}
				}
			}
			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				gopath = filepath.Join(homeDir(), "go")
			}
			return []string{filepath.Join(gopath, "pkg", "mod")}
		},
		ScanFunc: scanGoCache,
	}
}

func pythonCache() LangCache {
	return LangCache{
		ID:   "python",
		Name: "Python (pip)",
		Icon: "🐍",
		CachePaths: func() []string {
			var paths []string

			cmd := exec.Command("pip", "cache", "dir")
			out, err := cmd.Output()
			if err == nil {
				p := strings.TrimSpace(string(out))
				if p != "" {
					paths = append(paths, p)
				}
			}

			if runtime.GOOS == "windows" {
				localApp := os.Getenv("LOCALAPPDATA")
				if localApp != "" {
					paths = append(paths, filepath.Join(localApp, "pip", "cache"))
					paths = append(paths, filepath.Join(localApp, "pypa", "pip", "cache"))
				}
				appData := os.Getenv("APPDATA")
				if appData != "" {
					paths = append(paths, filepath.Join(appData, "pip", "cache"))
				}
			} else {
				xdg := os.Getenv("XDG_CACHE_HOME")
				if xdg != "" {
					paths = append(paths, filepath.Join(xdg, "pip"))
				}
				paths = append(paths, filepath.Join(homeDir(), ".cache", "pip"))
			}

			return paths
		},
		ScanFunc: scanPipCache,
	}
}

func rustCache() LangCache {
	return LangCache{
		ID:   "rust",
		Name: "Rust (Cargo)",
		Icon: "🦀",
		CachePaths: func() []string {
			cargoHome := os.Getenv("CARGO_HOME")
			if cargoHome == "" {
				cargoHome = filepath.Join(homeDir(), ".cargo")
			}
			return []string{
				filepath.Join(cargoHome, "registry", "cache"),
				filepath.Join(cargoHome, "registry", "src"),
			}
		},
		ScanFunc: scanCargoCache,
	}
}

func nodeCache() LangCache {
	return LangCache{
		ID:   "node",
		Name: "Node.js (npm)",
		Icon: "📦",
		CachePaths: func() []string {
			cmd := exec.Command("npm", "config", "get", "cache")
			out, err := cmd.Output()
			if err == nil {
				p := strings.TrimSpace(string(out))
				if p != "" {
					return []string{p}
				}
			}

			if runtime.GOOS == "windows" {
				localApp := os.Getenv("LOCALAPPDATA")
				if localApp != "" {
					return []string{filepath.Join(localApp, "npm-cache")}
				}
			}

			return []string{filepath.Join(homeDir(), ".npm")}
		},
		ScanFunc: scanNpmCache,
	}
}

func javaMavenCache() LangCache {
	return LangCache{
		ID:   "maven",
		Name: "Java (Maven)",
		Icon: "☕",
		CachePaths: func() []string {
			m2 := os.Getenv("M2_HOME")
			if m2 != "" {
				return []string{filepath.Join(m2, "repository")}
			}
			localRepo := os.Getenv("MAVEN_OPTS")
			_ = localRepo
			return []string{filepath.Join(homeDir(), ".m2", "repository")}
		},
		ScanFunc: scanMavenCache,
	}
}

func javaGradleCache() LangCache {
	return LangCache{
		ID:   "gradle",
		Name: "Java (Gradle)",
		Icon: "🐘",
		CachePaths: func() []string {
			gradleHome := os.Getenv("GRADLE_HOME")
			if gradleHome != "" {
				return []string{filepath.Join(gradleHome, "caches")}
			}
			userHome := os.Getenv("GRADLE_USER_HOME")
			if userHome != "" {
				return []string{filepath.Join(userHome, "caches")}
			}
			return []string{filepath.Join(homeDir(), ".gradle", "caches")}
		},
		ScanFunc: scanGradleCache,
	}
}

func dotnetCache() LangCache {
	return LangCache{
		ID:   "dotnet",
		Name: ".NET (NuGet)",
		Icon: "🟣",
		CachePaths: func() []string {
			if runtime.GOOS == "windows" {
				localApp := os.Getenv("LOCALAPPDATA")
				if localApp != "" {
					return []string{filepath.Join(localApp, "NuGet", "Cache")}
				}
			}
			nuget := os.Getenv("NUGET_PACKAGES")
			if nuget != "" {
				return []string{nuget}
			}
			return []string{filepath.Join(homeDir(), ".nuget", "packages")}
		},
		ScanFunc: scanNuGetCache,
	}
}

func scanGoCache(cachePath string) []CachedPackage {
	if cachePath == "" || !pathExists(cachePath) {
		return nil
	}
	var pkgs []CachedPackage
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
			fullPath := filepath.Join(cachePath, name)
			pkgs = append(pkgs, CachedPackage{
				Name:    parts[0],
				Version: parts[1],
				Size:    util.DirSize(fullPath),
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
			verEntries, err := os.ReadDir(modPath)
			if err != nil {
				continue
			}
			for _, verEntry := range verEntries {
				if !verEntry.IsDir() {
					continue
				}
				version := verEntry.Name()
				if strings.Contains(version, "@") {
					continue
				}
				fullPath := filepath.Join(modPath, version)
				pkgs = append(pkgs, CachedPackage{
					Name:    name + "/" + modEntry.Name(),
					Version: version,
					Size:    util.DirSize(fullPath),
					Path:    fullPath,
				})
			}
		}
	}
	return pkgs
}

func scanPipCache(cachePath string) []CachedPackage {
	if cachePath == "" || !pathExists(cachePath) {
		return nil
	}
	var pkgs []CachedPackage
	wheelDir := filepath.Join(cachePath, "wheels")
	if pathExists(wheelDir) {
		filepath.WalkDir(wheelDir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if filepath.Ext(p) == ".whl" {
				base := filepath.Base(p)
				name := base
				version := ""
				parts := strings.Split(base, "-")
				if len(parts) >= 2 {
					name = parts[0]
					version = parts[1]
				}
				info, err := d.Info()
				var size int64
				if err == nil {
					size = info.Size()
				}
				pkgs = append(pkgs, CachedPackage{
					Name:    name,
					Version: version,
					Size:    size,
					Path:    p,
				})
			}
			return nil
		})
	}
	httpDir := filepath.Join(cachePath, "http")
	if pathExists(httpDir) {
		filepath.WalkDir(httpDir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			if info.Size() > 1024*1024 {
				pkgs = append(pkgs, CachedPackage{
					Name:    filepath.Base(p),
					Version: "(cached)",
					Size:    info.Size(),
					Path:    p,
				})
			}
			return nil
		})
	}
	return pkgs
}

func scanCargoCache(cachePath string) []CachedPackage {
	if cachePath == "" || !pathExists(cachePath) {
		return nil
	}
	var pkgs []CachedPackage
	entries, err := os.ReadDir(cachePath)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		indexDir := filepath.Join(cachePath, entry.Name())
		pkgEntries, err := os.ReadDir(indexDir)
		if err != nil {
			continue
		}
		for _, pkgEntry := range pkgEntries {
			if !pkgEntry.IsDir() {
				continue
			}
			pkgPath := filepath.Join(indexDir, pkgEntry.Name())
			verEntries, err := os.ReadDir(pkgPath)
			if err != nil {
				continue
			}
			for _, verEntry := range verEntries {
				if !verEntry.IsDir() {
					continue
				}
				fullPath := filepath.Join(pkgPath, verEntry.Name())
				pkgs = append(pkgs, CachedPackage{
					Name:    pkgEntry.Name(),
					Version: verEntry.Name(),
					Size:    util.DirSize(fullPath),
					Path:    fullPath,
				})
			}
		}
	}
	return pkgs
}

func scanNpmCache(cachePath string) []CachedPackage {
	if cachePath == "" || !pathExists(cachePath) {
		return nil
	}
	var pkgs []CachedPackage
	entries, err := os.ReadDir(cachePath)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "_cacache" || name == "_locks" || name == "_logs" {
			continue
		}
		if strings.HasPrefix(name, "@") {
			orgDir := filepath.Join(cachePath, name)
			orgEntries, err := os.ReadDir(orgDir)
			if err != nil {
				continue
			}
			for _, orgEntry := range orgEntries {
				if !orgEntry.IsDir() {
					continue
				}
				fullPath := filepath.Join(orgDir, orgEntry.Name())
				pkgs = append(pkgs, CachedPackage{
					Name:    name + "/" + orgEntry.Name(),
					Version: "",
					Size:    util.DirSize(fullPath),
					Path:    fullPath,
				})
			}
			continue
		}
		fullPath := filepath.Join(cachePath, name)
		pkgs = append(pkgs, CachedPackage{
			Name:    name,
			Version: "",
			Size:    util.DirSize(fullPath),
			Path:    fullPath,
		})
	}
	return pkgs
}

func scanMavenCache(cachePath string) []CachedPackage {
	if cachePath == "" || !pathExists(cachePath) {
		return nil
	}
	var pkgs []CachedPackage
	filepath.WalkDir(cachePath, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".jar") {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(cachePath, filepath.Dir(p))
			rel = filepath.ToSlash(rel)
			version := ""
			parts := strings.Split(rel, "/")
			if len(parts) > 0 {
				version = parts[len(parts)-1]
				rel = strings.Join(parts[:len(parts)-1], "/")
			}
			pkgs = append(pkgs, CachedPackage{
				Name:    rel,
				Version: version,
				Size:    info.Size(),
				Path:    filepath.Dir(p),
			})
		}
		return nil
	})
	return pkgs
}

func scanGradleCache(cachePath string) []CachedPackage {
	if cachePath == "" || !pathExists(cachePath) {
		return nil
	}
	var pkgs []CachedPackage
	filepath.WalkDir(cachePath, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".jar") {
			info, err := d.Info()
			if err != nil || info.Size() <= 1024 {
				return nil
			}
			rel, _ := filepath.Rel(cachePath, filepath.Dir(p))
			pkgs = append(pkgs, CachedPackage{
				Name:    filepath.ToSlash(rel),
				Version: "",
				Size:    info.Size(),
				Path:    filepath.Dir(p),
			})
		}
		return nil
	})
	return pkgs
}

func scanNuGetCache(cachePath string) []CachedPackage {
	if cachePath == "" || !pathExists(cachePath) {
		return nil
	}
	var pkgs []CachedPackage
	entries, err := os.ReadDir(cachePath)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pkgName := entry.Name()
		pkgPath := filepath.Join(cachePath, pkgName)
		verEntries, err := os.ReadDir(pkgPath)
		if err != nil {
			continue
		}
		for _, verEntry := range verEntries {
			if !verEntry.IsDir() {
				continue
			}
			fullPath := filepath.Join(pkgPath, verEntry.Name())
			pkgs = append(pkgs, CachedPackage{
				Name:    pkgName,
				Version: verEntry.Name(),
				Size:    util.DirSize(fullPath),
				Path:    fullPath,
			})
		}
	}
	return pkgs
}
