package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print gtree version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gtree %s (built %s)\n", Version, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	// Also wire --version flag on the root command for `gtree --version`.
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("gtree {{.Version}}\n")
}
