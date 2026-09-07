package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/spf13/cobra"
)

var followLogs bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View daemon logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		f, err := os.Open(cfg.LogFile)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No log file found. Is the daemon running?")
				return nil
			}
			return fmt.Errorf("open log: %w", err)
		}
		defer f.Close()

		if !followLogs {
			io.Copy(os.Stdout, f)
			return nil
		}

		io.Copy(os.Stdout, f)
		for {
			time.Sleep(500 * time.Millisecond)
			io.Copy(os.Stdout, f)
		}
	},
}

func init() {
	logsCmd.Flags().BoolVar(&followLogs, "follow", false, "Follow log output")
	rootCmd.AddCommand(logsCmd)
}
