package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <domain>",
	Short: "Open a domain in the default browser",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		domain := args[0]
		if !strings.Contains(domain, ".") {
			domain = domain + "." + cfg.DefaultTLD
		}

		url := "http://" + domain

		var openCmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			openCmd = exec.Command("open", url)
		case "linux":
			openCmd = exec.Command("xdg-open", url)
		case "windows":
			openCmd = exec.Command("cmd", "/c", "start", url)
		default:
			return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
		}

		if err := openCmd.Run(); err != nil {
			return fmt.Errorf("open browser: %w", err)
		}

		fmt.Printf("Opening %s...\n", url)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
