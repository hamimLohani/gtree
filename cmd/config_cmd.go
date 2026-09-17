package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hamimlohani/gtree/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or edit the gtree configuration file",
	Long: `Show the location of the gtree config file and open it in $EDITOR.

Examples:
  gtree config            Print config file path and current values
  gtree config --edit     Open config file in $EDITOR`,
	RunE: runConfig,
}

var configEditFlag bool

func init() {
	configCmd.Flags().BoolVarP(&configEditFlag, "edit", "e", false, "open config file in $EDITOR")
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadOrCreate()
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}

	cfgPath := config.FilePath()

	if configEditFlag {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			editor = "vi"
		}
		c := exec.Command(editor, cfgPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	// Pretty-print current config.
	fmt.Printf("Config file: %s\n\n", cfgPath)
	fmt.Printf("default_path:  %s\n", cfg.DefaultPath)
	fmt.Printf("default_depth: %d\n", cfg.DefaultDepth)
	fmt.Printf("default_sort:  %s\n", cfg.DefaultSort)
	fmt.Printf("theme:         %s\n", cfg.Theme)
	fmt.Printf("ignore_dirs:\n")
	for _, d := range cfg.IgnoreDirs {
		fmt.Printf("  - %s\n", d)
	}
	return nil
}
