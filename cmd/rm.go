package cmd

import (
	"fmt"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/daemon"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <domain>",
	Short: "Remove a domain mapping",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		c := daemon.NewClient(cfg.SocketPath())
		if err := c.Remove(args[0]); err != nil {
			return err
		}

		fmt.Printf("✓ Removed %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
