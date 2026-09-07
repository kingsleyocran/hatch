package cmd

import (
	"fmt"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/daemon"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon health and domain summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		c := daemon.NewClient(cfg.SocketPath())
		status, err := c.Status()
		if err != nil {
			fmt.Println("✗ Hatch daemon is not running")
			return nil
		}

		fmt.Println("✓ Hatch daemon is running")
		fmt.Printf("  Uptime:  %s\n", status.Uptime)
		fmt.Printf("  Domains: %d (%d active)\n", status.DomainCount, status.ActiveCount)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
