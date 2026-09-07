package scanner

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Service struct {
	Port int
	PID  int
	Dir  string
	Name string
	Type string
}

func ScanPorts() ([]Service, error) {
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P", "-F", "pcn").Output()
	if err != nil {
		return nil, err
	}

	var services []Service
	var currentPID int

	for _, line := range strings.Split(string(out), "\n") {
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case 'p':
			pid, _ := strconv.Atoi(line[1:])
			currentPID = pid
		case 'n':
			addr := line[1:]
			if idx := strings.LastIndex(addr, ":"); idx >= 0 {
				port, _ := strconv.Atoi(addr[idx+1:])
				if port > 0 && strings.HasPrefix(addr, "127.0.0.1") {
					dir := findProcessDir(currentPID)
					name, ptype := DetectProject(dir)
					services = append(services, Service{
						Port: port,
						PID:  currentPID,
						Dir:  dir,
						Name: name,
						Type: ptype,
					})
				}
			}
		}
	}

	seen := make(map[int]bool)
	var unique []Service
	for _, s := range services {
		if !seen[s.Port] {
			seen[s.Port] = true
			unique = append(unique, s)
		}
	}

	return unique, nil
}

func findProcessDir(pid int) string {
	out, err := exec.Command("lsof", "-p", strconv.Itoa(pid), "-Fn", "-d", "cwd").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "n/") {
			return line[1:]
		}
	}
	return ""
}

func DetectProject(dir string) (name, projectType string) {
	if dir == "" {
		return "", ""
	}

	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		var pkg struct {
			Name string `json:"name"`
		}
		if json.Unmarshal(data, &pkg) == nil && pkg.Name != "" {
			return pkg.Name, "node"
		}
	}

	if data, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "module ") {
			mod := strings.TrimPrefix(lines[0], "module ")
			parts := strings.Split(strings.TrimSpace(mod), "/")
			return parts[len(parts)-1], "go"
		}
	}

	if data, err := os.ReadFile(filepath.Join(dir, "Cargo.toml")); err == nil {
		re := regexp.MustCompile(`(?m)^name\s*=\s*"([^"]+)"`)
		if m := re.FindSubmatch(data); len(m) > 1 {
			return string(m[1]), "rust"
		}
	}

	if data, err := os.ReadFile(filepath.Join(dir, "pyproject.toml")); err == nil {
		re := regexp.MustCompile(`(?m)^name\s*=\s*"([^"]+)"`)
		if m := re.FindSubmatch(data); len(m) > 1 {
			return string(m[1]), "python"
		}
	}

	return filepath.Base(dir), ""
}
