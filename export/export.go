package export

import (
	"encoding/json"
	"os"
	"time"
)

type Report struct {
	Timestamp      string         `json:"timestamp"`
	ProjectCount   int            `json:"project_count"`
	TotalModules   int            `json:"total_modules"`
	UnusedCount    int            `json:"unused_count"`
	ReclaimableBytes int64        `json:"reclaimable_bytes"`
	Reclaimable    string         `json:"reclaimable_human"`
	Modules        []ModuleReport `json:"modules"`
}

type ModuleReport struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Size    int64  `json:"size_bytes"`
	SizeHR  string `json:"size_human"`
	Path    string `json:"path"`
}

type CacheReport struct {
	Timestamp    string          `json:"timestamp"`
	Language     string          `json:"language"`
	TotalCount   int             `json:"total_count"`
	TotalBytes   int64           `json:"total_bytes"`
	TotalHuman   string          `json:"total_human"`
	Packages     []PackageReport `json:"packages"`
}

type PackageReport struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Size    int64  `json:"size_bytes"`
	SizeHR  string `json:"size_human"`
	Path    string `json:"path"`
}

func SaveReport(path string, report *Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func SaveCacheReport(path string, report *CacheReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func NewReport(projectCount, totalModules, unusedCount int, reclaimableBytes int64, reclaimableHR string, modules []ModuleReport) *Report {
	return &Report{
		Timestamp:        time.Now().Format(time.RFC3339),
		ProjectCount:     projectCount,
		TotalModules:     totalModules,
		UnusedCount:      unusedCount,
		ReclaimableBytes: reclaimableBytes,
		Reclaimable:      reclaimableHR,
		Modules:          modules,
	}
}

func NewCacheReport(language string, totalCount int, totalBytes int64, totalHuman string, packages []PackageReport) *CacheReport {
	return &CacheReport{
		Timestamp:  time.Now().Format(time.RFC3339),
		Language:   language,
		TotalCount: totalCount,
		TotalBytes: totalBytes,
		TotalHuman: totalHuman,
		Packages:   packages,
	}
}
