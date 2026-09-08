package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

		binary, _ := os.Executable()
		fmt.Println("  Installing daemon service...")
		if err := plat.InstallDaemon(binary, cfg.SocketPath()); err != nil {
			return fmt.Errorf("install daemon: %w", err)
		}
		fmt.Println("  ✓ Daemon installed (binds port 80 directly)")

		fmt.Println("  Initializing local Certificate Authority...")
		if !htls.CAExists(config.Dir()) {
			if err := htls.InitCA(config.Dir()); err != nil {
				return fmt.Errorf("init CA: %w", err)
			}
		}
		fmt.Println("  ✓ Local CA initialized")

		fmt.Println("  Installing CA into system trust store...")
		if err := plat.InstallCA(filepath.Join(config.Dir(), "ca-cert.pem")); err != nil {
			fmt.Printf("  ⚠ CA trust install failed: %v\n", err)
		} else {
			fmt.Println("  ✓ CA trusted by system")
		}

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
