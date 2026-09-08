package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/platform"
	"github.com/spf13/cobra"
)

var trustCACmd = &cobra.Command{
	Use:    "trust-ca",
	Short:  "Install the Hatch CA into the system trust store",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		certPath := filepath.Join(config.Dir(), "ca-cert.pem")
		plat := platform.Current()
		if err := plat.InstallCA(certPath); err != nil {
			return fmt.Errorf("install CA: %w", err)
		}
		fmt.Println("  ✓ CA trusted by system")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(trustCACmd)
}
