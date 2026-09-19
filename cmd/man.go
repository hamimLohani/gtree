package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var manCmd = &cobra.Command{
	Use:   "man [directory]",
	Short: "Generate man pages for gtree",
	Long: `Generate man pages for gtree into the specified directory,
or print the main man page to stdout if no directory is provided.

Examples:
  gtree man                   # Print gtree man page to stdout
  gtree man man/              # Generate all man pages into man/ directory
  gtree man | man -l -        # View man page directly`,
	Args:   cobra.MaximumNArgs(1),
	Hidden: false,
	RunE: func(cmd *cobra.Command, args []string) error {
		header := &doc.GenManHeader{
			Title:   "GTREE",
			Section: "1",
			Source:  "gtree " + Version,
			Manual:  "gtree Manual",
		}

		if len(args) == 0 {
			return doc.GenMan(rootCmd, header, os.Stdout)
		}

		dir := args[0]
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("could not create directory: %w", err)
		}

		if err := doc.GenManTree(rootCmd, header, dir); err != nil {
			return fmt.Errorf("could not generate man pages: %w", err)
		}

		fmt.Printf("Man pages generated in %s\n", dir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(manCmd)
}
