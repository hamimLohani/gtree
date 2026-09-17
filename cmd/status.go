package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hamimlohani/gtree/internal/config"
	"github.com/hamimlohani/gtree/internal/scanner"
	"github.com/hamimlohani/gtree/internal/tui"
)

// statusCmd implements the explicit `gtree status [path]` subcommand.
var statusCmd = &cobra.Command{
	Use:   "status [path]",
	Short: "Scan a directory and display git repository status",
	Long: `Scan a directory tree for git repositories and display their
status in a nested, styled tree view.

Examples:
  gtree status                 Scan current directory
  gtree status ~/Projects      Scan a specific path
  gtree status . --depth 3     Limit recursion depth`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

// runStatus is shared between the root command (gtree [path]) and the
// explicit status subcommand (gtree status [path]).
func runStatus(cmd *cobra.Command, args []string) error {
	// ── Resolve scan path ────────────────────────────────────────────────
	scanPath := "."
	if len(args) > 0 {
		scanPath = args[0]
	} else if viper.GetString("default_path") != "" {
		scanPath = viper.GetString("default_path")
	}

	absPath, err := filepath.Abs(scanPath)
	if err != nil {
		return fmt.Errorf("invalid path %q: %w", scanPath, err)
	}
	if _, statErr := os.Stat(absPath); statErr != nil {
		return fmt.Errorf("cannot access path %q: %w", absPath, statErr)
	}

	// ── Load / create config ─────────────────────────────────────────────
	cfg, err := config.LoadOrCreate()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// Flag overrides win over config file defaults.
	depth := viper.GetInt("depth")
	if depth == 0 {
		depth = cfg.DefaultDepth
	}

	onlyDirty := viper.GetBool("only_dirty")

	sortOrder := viper.GetString("sort")
	if sortOrder == "" {
		sortOrder = cfg.DefaultSort
	}

	watch := viper.GetBool("watch")

	// ── Build scanner options ─────────────────────────────────────────────
	opts := scanner.Options{
		MaxDepth:    depth,
		IgnoreDirs:  cfg.IgnoreDirs,
		OnlyDirty:   onlyDirty,
		SortOrder:   sortOrder,
		WorkerCount: 8,
	}

	// ── Hand off to the TUI ───────────────────────────────────────────────
	return tui.Run(absPath, opts, cfg, watch)
}
