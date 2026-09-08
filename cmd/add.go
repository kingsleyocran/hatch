package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/daemon"
	"github.com/kingsleyocran/hatch/internal/platform"
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
			ensureCATrusted()
		}
		fmt.Printf("✓ %s → localhost:%d (%s)\n", domain, port, proto)
		return nil
	},
}

func ensureCATrusted() {
	certPath := filepath.Join(config.Dir(), "ca-cert.pem")
	if _, err := os.Stat(certPath); err != nil {
		return
	}

	plat := platform.Current()
	if err := plat.InstallCA(certPath); err != nil {
		if plat.NeedsSudo() && os.Geteuid() != 0 {
			fmt.Println("  Installing CA certificate (requires admin)...")
			binary, _ := os.Executable()
			sudoCmd := exec.Command("sudo", binary, "trust-ca")
			sudoCmd.Stdin = os.Stdin
			sudoCmd.Stdout = os.Stdout
			sudoCmd.Stderr = os.Stderr
			sudoCmd.Run()
		}
	}
}

func init() {
	addCmd.Flags().BoolVar(&httpsFlag, "https", false, "Enable HTTPS for this domain")
	rootCmd.AddCommand(addCmd)
}
