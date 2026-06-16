package tempcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func windowsTemp() Item {
	return Item{
		ID:          "windows-temp",
		Name:        "Windows Temp Files",
		Description: "Temporary files created by Windows and applications while running. These files are safe to delete and help free up disk space.",
		Source:      "Windows OS & Applications",
		Icon:        "🗑️",
		Criticality: Safe,
		Platforms:   []string{"windows"},
		DetectFn: func() bool {
			return os.Getenv("TEMP") != "" || os.Getenv("TMP") != ""
		},
		SizeFn:  windowsTempSize,
		CleanFn: windowsTempClean,
	}
}

func windowsTempPaths() []string {
	var paths []string
	if t := os.Getenv("TEMP"); t != "" {
		paths = append(paths, t)
	}
	if t := os.Getenv("TMP"); t != "" {
		paths = append(paths, t)
	}
	return paths
}

func windowsTempSize() int64 {
	var total int64
	for _, p := range windowsTempPaths() {
		total += dirSize(p)
	}
	return total
}

func windowsTempClean() (int64, error) {
	var freed int64
	var errs []string
	for _, p := range windowsTempPaths() {
		entries, err := os.ReadDir(p)
		if err != nil {
			errs = append(errs, fmt.Sprintf("can't read %s: %v", p, err))
			continue
		}
		for _, entry := range entries {
			full := filepath.Join(p, entry.Name())
			info, err := os.Stat(full)
			if err != nil {
				continue
			}
			freed += info.Size()
			if err := os.RemoveAll(full); err != nil {
				freed -= info.Size()
			}
		}
	}
	if len(errs) > 0 {
		return freed, errors.New(strings.Join(errs, "; "))
	}
	return freed, nil
}

func recycleBin() Item {
	return Item{
		ID:          "recycle-bin",
		Name:        "Recycle Bin",
		Description: "Files you previously deleted that are still in the Recycle Bin. Permanently removes these files to free up disk space. This cannot be undone.",
		Source:      "Windows Shell",
		Icon:        "♻️",
		Criticality: Caution,
		Platforms:   []string{"windows"},
		DetectFn: func() bool {
			drives := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
			for _, d := range drives {
				rb := fmt.Sprintf("%c:\\$Recycle.Bin", d)
				if pathExists(rb) {
					return true
				}
			}
			return false
		},
		SizeFn:  recycleBinSize,
		CleanFn: recycleBinClean,
	}
}

func recycleBinPaths() []string {
	var paths []string
	drives := "CDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, d := range drives {
		rb := fmt.Sprintf("%c:\\$Recycle.Bin", d)
		if pathExists(rb) {
			paths = append(paths, rb)
		}
	}
	return paths
}

func recycleBinSize() int64 {
	var total int64
	for _, p := range recycleBinPaths() {
		total += dirSize(p)
	}
	return total
}

func recycleBinClean() (int64, error) {
	if runtime.GOOS != "windows" {
		return 0, fmt.Errorf("not supported on this platform")
	}
	_, err := execCmd("powershell", "-Command", "Clear-RecycleBin -Force")
	if err != nil {
		_, err = execCmd("pwsh", "-Command", "Clear-RecycleBin -Force")
		if err != nil {
			return 0, fmt.Errorf("failed to clear recycle bin (try running as admin): %w", err)
		}
	}
	sizes := recycleBinSize()
	return sizes, nil
}
