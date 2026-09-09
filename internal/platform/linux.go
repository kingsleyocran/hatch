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
	hatchDir := filepath.Dir(sockPath)
	return fmt.Sprintf(`[Unit]
Description=Hatch Local Domain Manager
After=network.target

[Service]
Type=simple
Environment=HATCH_DIR=%s
ExecStart=%s start --foreground
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, hatchDir, binaryPath)
}

func (l *Linux) InstallDaemon(binaryPath, sockPath string) error {
	exec.Command("systemctl", "stop", "hatch.service").Run()
	exec.Command("systemctl", "--user", "disable", "--now", "hatch.service").Run()
	os.Remove(filepath.Join(os.Getenv("HOME"), ".config", "systemd", "user", "hatch.service"))

	unitPath := "/etc/systemd/system/hatch.service"
	content := l.systemdUnit(binaryPath, sockPath)
	if err := os.WriteFile(unitPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write unit: %w", err)
	}

	exec.Command("systemctl", "daemon-reload").Run()
	cmd := exec.Command("systemctl", "enable", "--now", "hatch.service")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("enable service: %s: %w", string(out), err)
	}

	if hatchDir := filepath.Dir(sockPath); hatchDir != "" {
		os.Chmod(filepath.Join(hatchDir, "hatch.sock"), 0666)
	}
	return nil
}

func (l *Linux) UninstallDaemon() error {
	exec.Command("systemctl", "disable", "--now", "hatch.service").Run()
	os.Remove("/etc/systemd/system/hatch.service")
	exec.Command("systemctl", "daemon-reload").Run()
	return nil
}

func (l *Linux) InstallCA(certPath string) error {
	destDir := "/usr/local/share/ca-certificates"
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	data, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destDir, "hatch-ca.crt"), data, 0644); err != nil {
		return err
	}
	cmd := exec.Command("update-ca-certificates")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("update-ca-certificates: %s: %w", string(out), err)
	}
	return nil
}
