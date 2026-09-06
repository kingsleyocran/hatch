//go:build darwin

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Darwin struct{}

func Current() Platform {
	return &Darwin{}
}

func (d *Darwin) NeedsSudo() bool {
	return true
}

func (d *Darwin) resolverContent(dnsPort int) string {
	return fmt.Sprintf("nameserver 127.0.0.1\nport %d\n", dnsPort)
}

func (d *Darwin) SetupDNS(tld string, dnsPort int) error {
	resolverDir := "/etc/resolver"
	if err := os.MkdirAll(resolverDir, 0755); err != nil {
		return fmt.Errorf("create resolver dir: %w", err)
	}

	path := filepath.Join(resolverDir, tld)
	content := d.resolverContent(dnsPort)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write resolver file: %w", err)
	}

	return nil
}

func (d *Darwin) TeardownDNS(tld string) error {
	path := filepath.Join("/etc/resolver", tld)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (d *Darwin) pfctlRule(fromPort, toPort int) string {
	return fmt.Sprintf("rdr pass on lo0 inet proto tcp from any to 127.0.0.1 port %d -> 127.0.0.1 port %d\n", fromPort, toPort)
}

func (d *Darwin) SetupPortForward(fromPort, toPort int) error {
	anchorDir := "/etc/pf.anchors"
	if err := os.MkdirAll(anchorDir, 0755); err != nil {
		return err
	}

	rule := d.pfctlRule(fromPort, toPort)
	anchorFile := filepath.Join(anchorDir, "com.hatch")

	existing, _ := os.ReadFile(anchorFile)
	content := string(existing) + rule

	if err := os.WriteFile(anchorFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write pf anchor: %w", err)
	}

	cmd := exec.Command("pfctl", "-a", "com.hatch", "-f", anchorFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pfctl load: %s: %w", string(out), err)
	}

	cmd = exec.Command("pfctl", "-e")
	cmd.CombinedOutput()

	return nil
}

func (d *Darwin) TeardownPortForward(fromPort, toPort int) error {
	cmd := exec.Command("pfctl", "-a", "com.hatch", "-F", "all")
	cmd.CombinedOutput()

	os.Remove("/etc/pf.anchors/com.hatch")
	return nil
}

func (d *Darwin) launchdPlist(binaryPath, sockPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.hatch.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>start</string>
        <string>--foreground</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/hatch.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/hatch.stderr.log</string>
</dict>
</plist>`, binaryPath)
}

func (d *Darwin) InstallDaemon(binaryPath, sockPath string) error {
	plistDir := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents")
	if err := os.MkdirAll(plistDir, 0755); err != nil {
		return err
	}

	plistPath := filepath.Join(plistDir, "com.hatch.daemon.plist")
	content := d.launchdPlist(binaryPath, sockPath)

	if err := os.WriteFile(plistPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	cmd := exec.Command("launchctl", "load", plistPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl load: %s: %w", string(out), err)
	}

	return nil
}

func (d *Darwin) UninstallDaemon() error {
	plistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.hatch.daemon.plist")

	cmd := exec.Command("launchctl", "unload", plistPath)
	cmd.CombinedOutput()

	os.Remove(plistPath)
	return nil
}
