package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/platform"
	htls "github.com/kingsleyocran/hatch/internal/tls"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "One-time setup (requires sudo)",
	Long:  "Configures DNS resolver and daemon service. Run once after install.",
	RunE: func(cmd *cobra.Command, args []string) error {
		plat := platform.Current()

		if plat.NeedsSudo() && os.Geteuid() != 0 {
			fmt.Println("Setup requires elevated privileges. Re-running with sudo...")
			binary, _ := os.Executable()
			sudoCmd := exec.Command("sudo", binary, "setup")
			sudoCmd.Stdin = os.Stdin
			sudoCmd.Stdout = os.Stdout
			sudoCmd.Stderr = os.Stderr
			return sudoCmd.Run()
		}

		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		fmt.Printf("Setting up Hatch for %s...\n", runtime.GOOS)

		if err := os.MkdirAll(config.Dir(), 0755); err != nil {
			return fmt.Errorf("create hatch dir: %w", err)
		}

		fmt.Printf("  Configuring DNS resolver for .%s domains...\n", cfg.DefaultTLD)
		if err := plat.SetupDNS(cfg.DefaultTLD, cfg.DNSPort); err != nil {
			return fmt.Errorf("setup DNS: %w", err)
		}
		fmt.Println("  ✓ DNS resolver configured")

		fmt.Println("  Setting up port forwarding (80 → daemon)...")
		if err := plat.SetupPortForward(80, cfg.DaemonPort); err != nil {
			fmt.Printf("  ⚠ Port forwarding failed: %v (use http://domain:%d instead)\n", err, cfg.DaemonPort)
		} else {
			fmt.Println("  ✓ Port forwarding configured")
		}

		fmt.Println("  Setting up HTTPS port forwarding (443 → daemon)...")
		if err := plat.SetupPortForward(443, cfg.HTTPSPort); err != nil {
			fmt.Printf("  ⚠ HTTPS port forwarding failed: %v\n", err)
		} else {
			fmt.Println("  ✓ HTTPS port forwarding configured")
		}

		fmt.Println("  Initializing local Certificate Authority...")
		if !htls.CAExists(config.Dir()) {
			if err := htls.InitCA(config.Dir()); err != nil {
				return fmt.Errorf("init CA: %w", err)
			}
		}
		fmt.Println("  ✓ Local CA initialized")

		if err := config.Save(cfg, config.DefaultPath()); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		fmt.Println("  ✓ Config saved")

		fmt.Println("\n✓ Setup complete. The daemon will start automatically.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
