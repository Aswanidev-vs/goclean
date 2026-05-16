package tempcache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Criticality int

const (
	Safe Criticality = iota
	Moderate
	Caution
)

func (c Criticality) String() string {
	switch c {
	case Safe:
		return "Safe"
	case Moderate:
		return "Moderate"
	case Caution:
		return "Caution"
	}
	return "Unknown"
}

type Item struct {
	ID          string
	Name        string
	Description string
	Source      string
	Icon        string
	Criticality Criticality
	Platforms   []string
	DetectFn    func() bool
	SizeFn      func() int64
	CleanFn     func() (int64, error)
}

var Registry []Item

func init() {
	Registry = []Item{
		windowsTemp(),
		recycleBin(),
		chromeCache(),
		firefoxCache(),
		edgeCache(),
		dockerBuildCache(),
		dockerDanglingImages(),
		dockerContainerLogs(),
	}
}

func DetectAvailable() []Item {
	var avail []Item
	for _, item := range Registry {
		ok := false
		for _, p := range item.Platforms {
			if p == runtime.GOOS {
				ok = true
				break
			}
		}
		if !ok {
			continue
		}
		if item.DetectFn() {
			avail = append(avail, item)
		}
	}
	return avail
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
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

func execCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}
