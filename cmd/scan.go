package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/daemon"
	"github.com/kingsleyocran/hatch/internal/scanner"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan running ports and suggest domain mappings",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		services, err := scanner.ScanPorts()
		if err != nil {
			return fmt.Errorf("scan ports: %w", err)
		}

		if len(services) == 0 {
			fmt.Println("No services detected on localhost.")
			return nil
		}

		fmt.Println("Detected services:")
		for _, s := range services {
			typeStr := ""
			if s.Type != "" {
				typeStr = fmt.Sprintf(" (%s)", s.Type)
			}
			fmt.Printf("  %d  →  %s%s\n", s.Port, s.Name, typeStr)
			fmt.Printf("       %s\n", s.Dir)
		}

		fmt.Println("\nAttach domains? [enter to skip, type name to map]")
		reader := bufio.NewReader(os.Stdin)

		c := daemon.NewClient(cfg.SocketPath())

		for _, s := range services {
			fmt.Printf("  %d → [%s]: ", s.Port, s.Name)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input == "" {
				input = s.Name
			}

			if input == "-" {
				continue
			}

			domain := input
			if !strings.Contains(domain, ".") {
				domain = domain + "." + cfg.DefaultTLD
			}

			if err := c.Add(domain, s.Port, s.Dir, false); err != nil {
				fmt.Printf("  ✗ %s: %s\n", domain, err)
				continue
			}
			fmt.Printf("  ✓ %s → localhost:%d\n", domain, s.Port)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
