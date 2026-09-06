package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/daemon"
	"github.com/kingsleyocran/hatch/internal/project"
	"github.com/spf13/cobra"
)

var foreground bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Hatch daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		sockPath := cfg.SocketPath()

		c := daemon.NewClient(sockPath)
		if err := c.Ping(); err == nil {
			fmt.Println("Hatch daemon is already running.")
			return nil
		}

		storePath := filepath.Join(config.Dir(), "projects.yaml")
		store, err := project.Load(storePath)
		if err != nil {
			return fmt.Errorf("load project store: %w", err)
		}

		d := daemon.NewDaemon(cfg, store, sockPath)

		if err := d.Start(); err != nil {
			return fmt.Errorf("start daemon: %w", err)
		}

		fmt.Println("✓ Hatch daemon started")

		if foreground {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
			fmt.Println("\nShutting down...")
			d.Stop()
		}

		return nil
	},
}

func init() {
	startCmd.Flags().BoolVar(&foreground, "foreground", false, "Run in foreground (don't daemonize)")
	rootCmd.AddCommand(startCmd)
}
