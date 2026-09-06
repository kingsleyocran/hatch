package cmd

import (
	"fmt"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/daemon"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Hatch daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		c := daemon.NewClient(cfg.SocketPath())
		if err := c.StopDaemon(); err != nil {
			return fmt.Errorf("stop daemon: %w", err)
		}

		fmt.Println("✓ Hatch daemon stopped")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
