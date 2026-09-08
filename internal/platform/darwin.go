//go:build darwin

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	hatchPfConf := "/etc/pf.anchors/com.hatch"
	if err := os.MkdirAll(filepath.Dir(hatchPfConf), 0755); err != nil {
		return err
	}

	rule := d.pfctlRule(fromPort, toPort)

	existing, _ := os.ReadFile(hatchPfConf)
	existingStr := string(existing)
	if !strings.Contains(existingStr, fmt.Sprintf("port %d ->", fromPort)) {
		if err := os.WriteFile(hatchPfConf, []byte(existingStr+rule), 0644); err != nil {
			return fmt.Errorf("write pf rules: %w", err)
		}
	}

	pfConf, _ := os.ReadFile("/etc/pf.conf")
	pfStr := string(pfConf)
	needsUpdate := false

	if !strings.Contains(pfStr, "rdr-anchor \"com.hatch\"") {
		rdrLine := "rdr-anchor \"com.hatch\""
		loadLine := "load anchor \"com.hatch\" from \"/etc/pf.anchors/com.hatch\""

		// rdr-anchor must appear with other rdr-anchors, before filter anchors
		if idx := strings.Index(pfStr, "rdr-anchor \"com.apple"); idx >= 0 {
			endOfLine := strings.Index(pfStr[idx:], "\n")
			if endOfLine >= 0 {
				insertAt := idx + endOfLine + 1
				pfStr = pfStr[:insertAt] + rdrLine + "\n" + pfStr[insertAt:]
			}
		} else {
			pfStr = pfStr + "\n" + rdrLine + "\n"
		}

		// load anchor goes at the end
		if !strings.Contains(pfStr, loadLine) {
			pfStr = pfStr + loadLine + "\n"
		}

		needsUpdate = true
	}

	if needsUpdate {
		if err := os.WriteFile("/etc/pf.conf", []byte(pfStr), 0644); err != nil {
			return fmt.Errorf("update pf.conf: %w", err)
		}
	}

	cmd := exec.Command("pfctl", "-f", "/etc/pf.conf")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pfctl reload: %s: %w", string(out), err)
	}

	cmd = exec.Command("pfctl", "-e")
	cmd.CombinedOutput()

	return nil
}

func (d *Darwin) TeardownPortForward(fromPort, toPort int) error {
	os.Remove("/etc/pf.anchors/com.hatch")

	pfConf, err := os.ReadFile("/etc/pf.conf")
	if err == nil {
		cleaned := strings.Replace(string(pfConf), "\nrdr-anchor \"com.hatch\"\nload anchor \"com.hatch\" from \"/etc/pf.anchors/com.hatch\"\n", "", 1)
		os.WriteFile("/etc/pf.conf", []byte(cleaned), 0644)
		exec.Command("pfctl", "-f", "/etc/pf.conf").Run()
	}

	return nil
}

func (d *Darwin) launchdPlist(binaryPath, sockPath string) string {
	hatchDir := filepath.Dir(sockPath)
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
    <key>EnvironmentVariables</key>
    <dict>
        <key>HATCH_DIR</key>
        <string>%s</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/hatch.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/hatch.stderr.log</string>
</dict>
</plist>`, binaryPath, hatchDir)
}

func (d *Darwin) InstallDaemon(binaryPath, sockPath string) error {
	exec.Command("launchctl", "unload", "/Library/LaunchDaemons/com.hatch.daemon.plist").CombinedOutput()
	exec.Command("launchctl", "unload", filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.hatch.daemon.plist")).CombinedOutput()
	os.Remove(filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.hatch.daemon.plist"))

	plistPath := "/Library/LaunchDaemons/com.hatch.daemon.plist"
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
	plistPath := "/Library/LaunchDaemons/com.hatch.daemon.plist"
	exec.Command("launchctl", "unload", plistPath).CombinedOutput()
	os.Remove(plistPath)
	return nil
}

func (d *Darwin) InstallCA(certPath string) error {
	// Try System keychain first (works on older macOS)
	cmd := exec.Command("security", "add-trusted-cert", "-d", "-r", "trustRoot",
		"-k", "/Library/Keychains/System.keychain", certPath)
	if _, err := cmd.CombinedOutput(); err == nil {
		return nil
	}

	// Fall back to login keychain (works on newer macOS without extra GUI prompt)
	home := os.Getenv("HOME")
	if hatchDir := os.Getenv("HATCH_DIR"); hatchDir != "" {
		home = filepath.Dir(hatchDir)
	}
	loginKeychain := filepath.Join(home, "Library", "Keychains", "login.keychain-db")
	cmd = exec.Command("security", "add-trusted-cert", "-r", "trustRoot",
		"-k", loginKeychain, certPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("security add-trusted-cert: %s: %w", string(out), err)
	}
	return nil
}
