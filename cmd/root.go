package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "hatch",
	Short: "Map local domains to your running apps",
	Long:  "Hatch maps custom local domains (e.g. cayacart.test) to your running development servers. No more juggling localhost ports.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
}
