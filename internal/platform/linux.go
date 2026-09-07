//go:build linux

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Linux struct{}

func Current() Platform {
	return &Linux{}
}

func (l *Linux) NeedsSudo() bool {
	return true
}

func (l *Linux) resolvedConf(dnsPort int) string {
	return fmt.Sprintf("[Resolve]\nDNS=127.0.0.1:%d\nDomains=~test\n", dnsPort)
}

func (l *Linux) SetupDNS(tld string, dnsPort int) error {
	confDir := "/etc/systemd/resolved.conf.d"
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return fmt.Errorf("create resolved config dir: %w", err)
	}

	path := filepath.Join(confDir, "hatch.conf")
	content := l.resolvedConf(dnsPort)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write resolved config: %w", err)
	}

	cmd := exec.Command("systemctl", "restart", "systemd-resolved")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("restart resolved: %s: %w", string(out), err)
	}

	return nil
}

func (l *Linux) TeardownDNS(tld string) error {
	path := "/etc/systemd/resolved.conf.d/hatch.conf"
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	exec.Command("systemctl", "restart", "systemd-resolved").Run()
	return nil
}

func (l *Linux) iptablesRule(fromPort, toPort int) string {
	return fmt.Sprintf("iptables -t nat -A OUTPUT -p tcp --dport %d -o lo -j REDIRECT --to-port %d", fromPort, toPort)
}

func (l *Linux) SetupPortForward(fromPort, toPort int) error {
	cmd := exec.Command("iptables", "-t", "nat", "-A", "OUTPUT",
		"-p", "tcp", "--dport", fmt.Sprintf("%d", fromPort),
		"-o", "lo", "-j", "REDIRECT", "--to-port", fmt.Sprintf("%d", toPort))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("iptables: %s: %w", string(out), err)
	}
	return nil
}

func (l *Linux) TeardownPortForward(fromPort, toPort int) error {
	cmd := exec.Command("iptables", "-t", "nat", "-D", "OUTPUT",
		"-p", "tcp", "--dport", fmt.Sprintf("%d", fromPort),
		"-o", "lo", "-j", "REDIRECT", "--to-port", fmt.Sprintf("%d", toPort))
	cmd.Run()
	return nil
}

func (l *Linux) systemdUnit(binaryPath, sockPath string) string {
	return fmt.Sprintf(`[Unit]
Description=hatch.service - Hatch Local Domain Manager
After=network.target

[Service]
Type=simple
ExecStart=%s start --foreground
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`, binaryPath)
}

func (l *Linux) InstallDaemon(binaryPath, sockPath string) error {
	unitDir := filepath.Join(os.Getenv("HOME"), ".config", "systemd", "user")
	if err := os.MkdirAll(unitDir, 0755); err != nil {
		return err
	}

	unitPath := filepath.Join(unitDir, "hatch.service")
	content := l.systemdUnit(binaryPath, sockPath)
	if err := os.WriteFile(unitPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write unit: %w", err)
	}

	exec.Command("systemctl", "--user", "daemon-reload").Run()
	cmd := exec.Command("systemctl", "--user", "enable", "--now", "hatch.service")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("enable service: %s: %w", string(out), err)
	}
	return nil
}

func (l *Linux) UninstallDaemon() error {
	exec.Command("systemctl", "--user", "disable", "--now", "hatch.service").Run()
	unitPath := filepath.Join(os.Getenv("HOME"), ".config", "systemd", "user", "hatch.service")
	os.Remove(unitPath)
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	return nil
}
