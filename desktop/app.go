package main

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
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
	if v := os.Getenv("HATCH_DIR"); v != "" {
		return filepath.Join(v, "hatch.sock")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hatch", "hatch.sock")
}

func (a *App) send(req request) (*response, error) {
	conn, err := net.DialTimeout("unix", a.socketPath(), 5*time.Second)
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

func (a *App) IsDaemonRunning() bool {
	resp, err := a.send(request{Action: "ping"})
	return err == nil && resp.OK
}

func (a *App) StartDaemon() string {
	cmd := exec.Command("osascript", "-e",
		`do shell script "launchctl kickstart system/com.hatch.daemon 2>/dev/null || launchctl start com.hatch.daemon" with prompt "Hatch needs to start the daemon service." with administrator privileges`)
	if out, err := cmd.CombinedOutput(); err != nil {
		binary := filepath.Join(os.Getenv("HOME"), ".hatch", "bin", "hatch")
		fallback := exec.Command(binary, "start")
		if out2, err2 := fallback.CombinedOutput(); err2 != nil {
			return string(out) + " " + string(out2)
		}
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
