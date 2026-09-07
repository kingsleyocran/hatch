package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

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

		if !foreground {
			binary, err := os.Executable()
			if err != nil {
				return fmt.Errorf("resolve executable path: %w", err)
			}
			child := exec.Command(binary, "start", "--foreground")
			setForkAttrs(child)
			child.Stdout = nil
			child.Stderr = nil
			child.Stdin = nil
			if err := child.Start(); err != nil {
				return fmt.Errorf("start background daemon: %w", err)
			}
			fmt.Printf("✓ Hatch daemon started (pid %d)\n", child.Process.Pid)
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

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		<-sigCh
		d.Stop()

		return nil
	},
}

func init() {
	startCmd.Flags().BoolVar(&foreground, "foreground", false, "Run in foreground (don't daemonize)")
	rootCmd.AddCommand(startCmd)
}
