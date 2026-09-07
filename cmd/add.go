package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/daemon"
	"github.com/spf13/cobra"
)

var httpsFlag bool

var addCmd = &cobra.Command{
	Use:   "add <domain|name> <port>",
	Short: "Map a local domain to a port",
	Long:  "Maps a domain (e.g. cayacart.test) to localhost:<port>. If no TLD is given, uses the default.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		domain := args[0]
		if !strings.Contains(domain, ".") {
			domain = domain + "." + cfg.DefaultTLD
		}

		port, err := strconv.Atoi(args[1])
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("invalid port: %s (must be 1-65535)", args[1])
		}

		useHTTPS := httpsFlag || cfg.AutoHTTPS

		dir, _ := os.Getwd()

		c := daemon.NewClient(cfg.SocketPath())
		if err := c.Add(domain, port, dir, useHTTPS); err != nil {
			return err
		}

		proto := "http"
		if useHTTPS {
			proto = "https"
		}
		fmt.Printf("✓ %s → localhost:%d (%s)\n", domain, port, proto)
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVar(&httpsFlag, "https", false, "Enable HTTPS for this domain")
	rootCmd.AddCommand(addCmd)
}
