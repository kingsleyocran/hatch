package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kingsleyocran/hatch/internal/config"
)

type App struct {
	ctx context.Context
}

type DomainInfo struct {
	Domain string `json:"domain"`
	Port   int    `json:"port"`
	Dir    string `json:"dir"`
	Alive  bool   `json:"alive"`
	HTTPS  bool   `json:"https"`
}

type StatusInfo struct {
	Running     bool   `json:"running"`
	Uptime      string `json:"uptime"`
	DomainCount int    `json:"domain_count"`
	ActiveCount int    `json:"active_count"`
}

type ConfigInfo struct {
	DefaultTLD string `json:"default_tld"`
	AutoHTTPS  bool   `json:"auto_https"`
	DaemonPort int    `json:"daemon_port"`
	HTTPSPort  int    `json:"https_port"`
	DNSPort    int    `json:"dns_port"`
}

type request struct {
	Action string `json:"action"`
	Domain string `json:"domain,omitempty"`
	Port   int    `json:"port,omitempty"`
	Dir    string `json:"dir,omitempty"`
	HTTPS  bool   `json:"https,omitempty"`
}

type response struct {
	OK       bool         `json:"ok"`
	Message  string       `json:"message,omitempty"`
	Projects []DomainInfo `json:"projects,omitempty"`
	Status   *StatusInfo  `json:"status,omitempty"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) socketPath() string {
	return config.Default().SocketPath()
}

func (a *App) send(req request) (*response, error) {
	network := "unix"
	if runtime.GOOS == "windows" {
		network = "tcp"
	}
	conn, err := net.DialTimeout(network, a.socketPath(), 5*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	json.NewEncoder(conn).Encode(req)
	var resp response
	json.NewDecoder(conn).Decode(&resp)
	return &resp, nil
}

func (a *App) GetDomains() []DomainInfo {
	resp, err := a.send(request{Action: "list"})
	if err != nil || !resp.OK {
		return []DomainInfo{}
	}
	return resp.Projects
}

func (a *App) GetStatus() *StatusInfo {
	resp, err := a.send(request{Action: "status"})
	if err != nil || !resp.OK || resp.Status == nil {
		return &StatusInfo{Running: false}
	}
	return resp.Status
}

func (a *App) AddDomain(domain string, port int, https bool) string {
	dir, _ := os.Getwd()
	resp, err := a.send(request{Action: "add", Domain: domain, Port: port, Dir: dir, HTTPS: https})
	if err != nil {
		return err.Error()
	}
	if !resp.OK {
		return resp.Message
	}
	return ""
}

func (a *App) RemoveDomain(domain string) string {
	resp, err := a.send(request{Action: "remove", Domain: domain})
	if err != nil {
		return err.Error()
	}
	if !resp.OK {
		return resp.Message
	}
	return ""
}

func (a *App) GetConfig() *ConfigInfo {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return &ConfigInfo{DefaultTLD: "test"}
	}
	return &ConfigInfo{
		DefaultTLD: cfg.DefaultTLD,
		AutoHTTPS:  cfg.AutoHTTPS,
		DaemonPort: cfg.DaemonPort,
		HTTPSPort:  cfg.HTTPSPort,
		DNSPort:    cfg.DNSPort,
	}
}

func (a *App) SetConfig(key, value string) string {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return err.Error()
	}
	switch key {
	case "default_tld":
		cfg.DefaultTLD = value
	case "auto_https":
		cfg.AutoHTTPS = value == "true"
	}
	if err := config.Save(cfg, config.DefaultPath()); err != nil {
		return err.Error()
	}
	return ""
}

type PortInfo struct {
	Port    int    `json:"port"`
	Name    string `json:"name"`
	Process string `json:"process"`
	Type    string `json:"type"`
	Dir     string `json:"dir"`
}

func (a *App) ScanPorts() []PortInfo {
	if runtime.GOOS == "windows" {
		return a.scanPortsWindows()
	}
	return a.scanPortsUnix()
}

func (a *App) scanPortsWindows() []PortInfo {
	out, err := exec.Command("netstat", "-ano", "-p", "TCP").Output()
	if err != nil {
		return []PortInfo{}
	}

	mapped := make(map[int]bool)
	for _, d := range a.GetDomains() {
		mapped[d.Port] = true
	}

	var ports []PortInfo
	seen := make(map[int]bool)
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, "LISTENING") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		addr := fields[1]
		if idx := strings.LastIndex(addr, ":"); idx >= 0 {
			port, _ := strconv.Atoi(addr[idx+1:])
			if port > 0 && !mapped[port] && !seen[port] && port != 80 && port != 443 && port != 15353 && port != 19876 {
				seen[port] = true
				ports = append(ports, PortInfo{Port: port, Name: fmt.Sprintf("pid:%s", fields[len(fields)-1]), Process: ""})
			}
		}
	}
	sort.Slice(ports, func(i, j int) bool { return ports[i].Port < ports[j].Port })
	return ports
}

func (a *App) scanPortsUnix() []PortInfo {
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P", "-F", "pcn").Output()
	if err != nil {
		return []PortInfo{}
	}

	mapped := make(map[int]bool)
	for _, d := range a.GetDomains() {
		mapped[d.Port] = true
	}

	type rawPort struct {
		port    int
		pid     int
		process string
	}

	var raw []rawPort
	seen := make(map[int]bool)
	var pid int
	var cmd string
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case 'p':
			p, _ := strconv.Atoi(line[1:])
			pid = p
			cmd = ""
		case 'c':
			cmd = line[1:]
		case 'n':
			idx := strings.LastIndex(line, ":")
			if idx >= 0 {
				port, _ := strconv.Atoi(line[idx+1:])
				if port > 0 && !mapped[port] && !seen[port] && port != 80 && port != 443 && port != 15353 {
					seen[port] = true
					raw = append(raw, rawPort{port: port, pid: pid, process: cmd})
				}
			}
		}
	}

	var ports []PortInfo
	for _, r := range raw {
		dir := a.getProcessDir(r.pid)
		name := filepath.Base(dir)
		ptype := ""
		if dir != "" {
			name, ptype = a.detectProject(dir)
		}
		ports = append(ports, PortInfo{Port: r.port, Name: name, Process: r.process, Type: ptype, Dir: dir})
	}

	sort.Slice(ports, func(i, j int) bool { return ports[i].Port < ports[j].Port })
	return ports
}

func (a *App) getProcessDir(pid int) string {
	out, err := exec.Command("lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn").Output()
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

func (a *App) detectProject(dir string) (string, string) {
	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		var pkg struct{ Name string `json:"name"` }
		if json.Unmarshal(data, &pkg) == nil && pkg.Name != "" {
			return pkg.Name, "node"
		}
	}
	if data, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "module ") {
			parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(lines[0], "module ")), "/")
			return parts[len(parts)-1], "go"
		}
	}
	return filepath.Base(dir), ""
}

func (a *App) IsDaemonRunning() bool {
	resp, err := a.send(request{Action: "ping"})
	return err == nil && resp.OK
}

func (a *App) StartDaemon() string {
	home, _ := os.UserHomeDir()
	binary := filepath.Join(home, ".hatch", "bin", "hatch")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	switch runtime.GOOS {
	case "darwin":
		plist := "/Library/LaunchDaemons/com.hatch.daemon.plist"
		if _, err := os.Stat(plist); err == nil {
			cmd := exec.Command("osascript", "-e",
				`do shell script "launchctl unload /Library/LaunchDaemons/com.hatch.daemon.plist 2>/dev/null; launchctl load /Library/LaunchDaemons/com.hatch.daemon.plist" with prompt "Hatch needs to start the daemon." with administrator privileges`)
			if out, err := cmd.CombinedOutput(); err != nil {
				return string(out)
			}
			return ""
		}

	case "linux":
		cmd := exec.Command("pkexec", "systemctl", "start", "hatch.service")
		if out, err := cmd.CombinedOutput(); err == nil {
			return ""
		} else {
			_ = out
		}

	case "windows":
		cmd := exec.Command("schtasks", "/run", "/tn", "HatchDaemon")
		if out, err := cmd.CombinedOutput(); err == nil {
			return ""
		} else {
			_ = out
		}
	}

	cmd := exec.Command(binary, "start")
	if out, err := cmd.CombinedOutput(); err != nil {
		return string(out)
	}
	return ""
}

func (a *App) StopDaemon() string {
	resp, err := a.send(request{Action: "stop"})
	if err != nil {
		return err.Error()
	}
	if !resp.OK {
		return resp.Message
	}
	return ""
}
