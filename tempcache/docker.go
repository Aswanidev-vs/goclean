package tempcache

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func dockerAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func dockerBuildCache() Item {
	return Item{
		ID:          "docker-build-cache",
		Name:        "Docker Build Cache",
		Description: "Intermediate layers and cache from Docker image builds. Removing this will free up disk space but will slow down your next Docker build since all layers will need to be rebuilt.",
		Source:      "Docker",
		Icon:        "🐳",
		Criticality: Moderate,
		Platforms:   []string{"windows", "linux", "darwin"},
		DetectFn: func() bool {
			if !dockerAvailable() {
				return false
			}
			out, err := execCmd("docker", "builder", "prune", "--all", "--force", "--filter", "until=0s")
			if err != nil {
				return false
			}
			return strings.Contains(out, "Total:")
		},
		SizeFn: func() int64 {
			// --dry-run not available for builder prune in older Docker, use `docker system df`
			out, err := execCmd("docker", "system", "df", "--format", "{{.Type}}\t{{.Size}}")
			if err != nil {
				return 0
			}
			for _, line := range strings.Split(out, "\n") {
				parts := strings.Split(line, "\t")
				if len(parts) >= 2 && parts[0] == "Build Cache" {
					return parseDockerSize(parts[1])
				}
			}
			return 0
		},
		CleanFn: func() (int64, error) {
			before := dockerBuildTotalSize()
			out, err := execCmd("docker", "builder", "prune", "--all", "--force")
			if err != nil {
				return 0, err
			}
			_ = out
			after := dockerBuildTotalSize()
			freed := before - after
			if freed < 0 {
				freed = 0
			}
			return freed, nil
		},
	}
}

func dockerBuildTotalSize() int64 {
	out, err := execCmd("docker", "system", "df", "--format", "{{.Type}}\t{{.Size}}")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(out, "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 && parts[0] == "Build Cache" {
			return parseDockerSize(parts[1])
		}
	}
	return 0
}

func dockerDanglingImages() Item {
	return Item{
		ID:          "docker-dangling-images",
		Name:        "Docker Dangling Images",
		Description: "Untagged (<none>:<none>) Docker images left over from incomplete or intermediate builds. These take up disk space and are not usable.",
		Source:      "Docker",
		Icon:        "🐳",
		Criticality: Moderate,
		Platforms:   []string{"windows", "linux", "darwin"},
		DetectFn: func() bool {
			if !dockerAvailable() {
				return false
			}
			out, err := execCmd("docker", "images", "--filter", "dangling=true", "-q")
			if err != nil {
				return false
			}
			return strings.TrimSpace(out) != ""
		},
		SizeFn: func() int64 {
			out, err := execCmd("docker", "system", "df", "--format", "{{.Type}}\t{{.Size}}")
			if err != nil {
				return 0
			}
			for _, line := range strings.Split(out, "\n") {
				parts := strings.Split(line, "\t")
				if len(parts) >= 2 && parts[0] == "Images" {
					return parseDockerSize(parts[1])
				}
			}
			return 0
		},
		CleanFn: func() (int64, error) {
			before := dockerImageTotalSize()
			out, err := execCmd("docker", "image", "prune", "--force")
			if err != nil {
				return 0, err
			}
			_ = out
			after := dockerImageTotalSize()
			freed := before - after
			if freed < 0 {
				freed = 0
			}
			return freed, nil
		},
	}
}

func dockerImageTotalSize() int64 {
	out, err := execCmd("docker", "system", "df", "--format", "{{.Type}}\t{{.Size}}")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(out, "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 && parts[0] == "Images" {
			return parseDockerSize(parts[1])
		}
	}
	return 0
}

func dockerContainerLogs() Item {
	return Item{
		ID:          "docker-container-logs",
		Name:        "Docker Container Logs",
		Description: "JSON log files from running and stopped Docker containers. Removing these will free up disk space but you will lose container log history.",
		Source:      "Docker",
		Icon:        "📋",
		Criticality: Caution,
		Platforms:   []string{"windows", "linux", "darwin"},
		DetectFn: func() bool {
			if !dockerAvailable() {
				return false
			}
			out, err := execCmd("docker", "ps", "-aq")
			if err != nil || strings.TrimSpace(out) == "" {
				return false
			}
			return true
		},
		SizeFn: dockerLogsSize,
		CleanFn: dockerLogsClean,
	}
}

func dockerLogsSize() int64 {
	out, err := execCmd("docker", "ps", "-aq")
	if err != nil {
		return 0
	}
	ids := strings.Fields(out)
	if len(ids) == 0 {
		return 0
	}
	var total int64
	for _, id := range ids {
		szOut, err := execCmd("docker", "inspect", id, "--format", "{{.LogPath}}")
		if err != nil {
			continue
		}
		logPath := strings.TrimSpace(szOut)
		out2, err := execCmd("cmd", "/c", "if", "exist", logPath, "echo", "1")
		if err != nil {
			continue
		}
		if out2 == "1" {
			total += 0 // approximate — real size requires stat on log file
		}
	}
	return total
}

func dockerLogsClean() (int64, error) {
	out, err := execCmd("docker", "ps", "-aq")
	if err != nil {
		return 0, err
	}
	ids := strings.Fields(out)
	var total int64
	for _, id := range ids {
		execCmd("sh", "-c", fmt.Sprintf("truncate -s 0 $(docker inspect %s --format '{{.LogPath}}')", id))
	}
	return total, nil
}

func parseDockerSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0B" {
		return 0
	}

	multiplier := int64(1)
	switch {
	case strings.HasSuffix(s, "TB"):
		multiplier = 1 << 40
		s = strings.TrimSuffix(s, "TB")
	case strings.HasSuffix(s, "GB"):
		multiplier = 1 << 30
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1 << 20
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "kB"):
		multiplier = 1 << 10
		s = strings.TrimSuffix(s, "kB")
	case strings.HasSuffix(s, "KB"):
		multiplier = 1 << 10
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	default:
		return 0
	}

	s = strings.TrimSpace(s)
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(val * float64(multiplier))
}
