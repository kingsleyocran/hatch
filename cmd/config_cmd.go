package cmd

import (
	"fmt"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or update configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Printf("Config file: %s\n\n%s", config.DefaultPath(), string(data))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		key, value := args[0], args[1]
		switch key {
		case "default_tld":
			cfg.DefaultTLD = value
		case "auto_https":
			cfg.AutoHTTPS = value == "true"
		case "log_level":
			cfg.LogLevel = value
		default:
			return fmt.Errorf("unknown config key: %s (available: default_tld, auto_https, log_level)", key)
		}

		if err := config.Save(cfg, config.DefaultPath()); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		fmt.Printf("✓ Set %s = %s\n", key, value)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
