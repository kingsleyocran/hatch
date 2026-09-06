package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/platform"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "One-time setup (requires sudo)",
	Long:  "Configures DNS resolver and port forwarding. Run once after install.",
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
			return fmt.Errorf("setup port forward: %w", err)
		}
		fmt.Println("  ✓ Port forwarding configured")

		if err := config.Save(cfg, config.DefaultPath()); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		fmt.Println("  ✓ Config saved")

		fmt.Println("\n✓ Setup complete. Run 'hatch start' to begin.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
