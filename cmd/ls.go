package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/daemon"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all domain mappings",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		c := daemon.NewClient(cfg.SocketPath())
		projects, err := c.List()
		if err != nil {
			return err
		}

		if len(projects) == 0 {
			fmt.Println("No domains mapped. Use 'hatch add <domain> <port>' to get started.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "DOMAIN\tPORT\tSTATUS\tHTTPS")

		for _, p := range projects {
			status := "○ stopped"
			if p.Alive {
				status = "● active"
			}
			https := ""
			if p.HTTPS {
				https = "✓"
			}
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", p.Domain, p.Port, status, https)
		}

		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
