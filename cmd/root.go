package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version is set at build time via -ldflags.
	Version = "dev"
	// BuildDate is set at build time.
	BuildDate = "unknown"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gtree [path]",
	Short: "A git repository tree viewer",
	Long: `gtree scans a directory tree for git repositories and displays
their status as a styled, nested tree in the terminal.

Examples:
  gtree                        Scan current directory
  gtree ~/Projects             Scan a specific path
  gtree status ~/Projects      Explicit subcommand form
  gtree --depth 5              Limit recursion depth
  gtree --only-dirty           Show only dirty repos
  gtree --watch                Live-refreshing view`,
	Args:              cobra.MaximumNArgs(1),
	RunE:              runStatus,
	DisableAutoGenTag: true,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// cobra already prints the error; just exit non-zero.
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags — available to all subcommands.
	rootCmd.PersistentFlags().IntP("depth", "d", 0, "maximum recursion depth (0 = unlimited)")
	rootCmd.PersistentFlags().StringP("sort", "s", "", "sort order: dirty|recent|name (default from config)")
	rootCmd.PersistentFlags().Bool("only-dirty", false, "hide clean repositories")
	rootCmd.PersistentFlags().Bool("watch", false, "live-refreshing view (polls every few seconds)")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable color output")
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default ~/.config/gtree/config.yml)")

	// Bind flags to viper so config file values can override defaults.
	_ = viper.BindPFlag("depth", rootCmd.PersistentFlags().Lookup("depth"))
	_ = viper.BindPFlag("sort", rootCmd.PersistentFlags().Lookup("sort"))
	_ = viper.BindPFlag("only_dirty", rootCmd.PersistentFlags().Lookup("only-dirty"))
	_ = viper.BindPFlag("watch", rootCmd.PersistentFlags().Lookup("watch"))
}

// initConfig reads in the config file and ENV variables if set.
func initConfig() {
	cfgFile, _ := rootCmd.PersistentFlags().GetString("config")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not determine home directory: %v\n", err)
			return
		}
		viper.AddConfigPath(fmt.Sprintf("%s/.config/gtree", home))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Allow overriding any config key with GTREE_<KEY> env vars.
	viper.SetEnvPrefix("GTREE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Not finding the config file is fine — we'll create one.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "warning: error reading config: %v\n", err)
		}
	}
}
