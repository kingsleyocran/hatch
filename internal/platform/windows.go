//go:build windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Windows struct{}

func Current() Platform {
	return &Windows{}
}

func (w *Windows) NeedsSudo() bool {
	return true
}

func (w *Windows) hostsEntry(domain string) string {
	return fmt.Sprintf("127.0.0.1\t%s\n", domain)
}

func (w *Windows) SetupDNS(tld string, dnsPort int) error {
	// Windows lacks per-TLD resolver; fall back to hosts file.
	// Individual domains are added in the daemon's Add handler.
	return nil
}

func (w *Windows) TeardownDNS(tld string) error {
	return nil
}

func (w *Windows) netshArgs(fromPort, toPort int) []string {
	return []string{
		"interface", "portproxy", "add", "v4tov4",
		fmt.Sprintf("listenport=%d", fromPort),
		"listenaddress=127.0.0.1",
		fmt.Sprintf("connectport=%d", toPort),
		"connectaddress=127.0.0.1",
	}
}

func (w *Windows) SetupPortForward(fromPort, toPort int) error {
	args := w.netshArgs(fromPort, toPort)
	cmd := exec.Command("netsh", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("netsh: %s: %w", string(out), err)
	}
	return nil
}

func (w *Windows) TeardownPortForward(fromPort, toPort int) error {
	cmd := exec.Command("netsh", "interface", "portproxy", "delete", "v4tov4",
		fmt.Sprintf("listenport=%d", fromPort), "listenaddress=127.0.0.1")
	cmd.Run()
	return nil
}

func (w *Windows) InstallDaemon(binaryPath, sockPath string) error {
	cmd := exec.Command("schtasks", "/create",
		"/tn", "HatchDaemon",
		"/tr", fmt.Sprintf(`"%s" start --foreground`, binaryPath),
		"/sc", "onlogon",
		"/rl", "highest",
		"/f")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks: %s: %w", string(out), err)
	}
	return nil
}

func (w *Windows) UninstallDaemon() error {
	cmd := exec.Command("schtasks", "/delete", "/tn", "HatchDaemon", "/f")
	cmd.Run()
	return nil
}

func hostsFilePath() string {
	return filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "drivers", "etc", "hosts")
}
