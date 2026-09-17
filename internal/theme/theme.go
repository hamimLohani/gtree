// Package theme provides a color palette for gtree's TUI, adapting to
// the user's preferred color scheme (auto / dark / light) and respecting
// NO_COLOR and non-TTY environments.
package theme

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Palette holds all named colors used across the UI.
type Palette struct {
	// Header
	HeaderBg   lipgloss.Color
	HeaderFg   lipgloss.Color
	HeaderPath lipgloss.Color

	// Tree structure
	TreeLine lipgloss.Color
	DirName  lipgloss.Color

	// Repo row elements
	RepoName   lipgloss.Color
	BranchFg   lipgloss.Color
	Clean      lipgloss.Color
	Dirty      lipgloss.Color
	Ahead      lipgloss.Color
	Behind     lipgloss.Color
	Stash      lipgloss.Color
	CommitTime lipgloss.Color
	Error      lipgloss.Color

	// Footer
	FooterFg    lipgloss.Color
	FooterDirty lipgloss.Color
	FooterClean lipgloss.Color
	Unpushed    lipgloss.Color

	// Separator
	Separator lipgloss.Color
}

var (
	darkPalette = Palette{
		HeaderBg:   "#1e1e2e",
		HeaderFg:   "#cdd6f4",
		HeaderPath: "#89b4fa",

		TreeLine: "#45475a",
		DirName:  "#6c7086",

		RepoName:   "#cdd6f4",
		BranchFg:   "#6c7086",
		Clean:      "#a6e3a1",
		Dirty:      "#f38ba8",
		Ahead:      "#f9e2af",
		Behind:     "#fab387",
		Stash:      "#89dceb",
		CommitTime: "#585b70",
		Error:      "#f38ba8",

		FooterFg:    "#6c7086",
		FooterDirty: "#f38ba8",
		FooterClean: "#a6e3a1",
		Unpushed:    "#f9e2af",

		Separator: "#313244",
	}

	lightPalette = Palette{
		HeaderBg:   "#eff1f5",
		HeaderFg:   "#4c4f69",
		HeaderPath: "#1e66f5",

		TreeLine: "#acb0be",
		DirName:  "#9ca0b0",

		RepoName:   "#4c4f69",
		BranchFg:   "#9ca0b0",
		Clean:      "#40a02b",
		Dirty:      "#d20f39",
		Ahead:      "#df8e1d",
		Behind:     "#fe640b",
		Stash:      "#04a5e5",
		CommitTime: "#bcc0cc",
		Error:      "#d20f39",

		FooterFg:    "#9ca0b0",
		FooterDirty: "#d20f39",
		FooterClean: "#40a02b",
		Unpushed:    "#df8e1d",

		Separator: "#ccd0da",
	}
)

// noColor returns true when color output should be suppressed.
func noColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return true
	}
	// termenv.HasDarkBackground checks if we're connected to a real TTY.
	if termenv.NewOutput(os.Stdout).ColorProfile() == termenv.Ascii {
		return true
	}
	return false
}

// Load returns the appropriate palette based on the theme name and
// environment. If color is disabled it returns a nil pointer — callers
// should check IsPlain() to decide whether to strip styles.
func Load(name string) (*Palette, bool) {
	if noColor() {
		return nil, true // plain mode
	}

	switch name {
	case "light":
		p := lightPalette
		return &p, false
	case "dark":
		p := darkPalette
		return &p, false
	default: // "auto"
		if termenv.HasDarkBackground() {
			p := darkPalette
			return &p, false
		}
		p := lightPalette
		return &p, false
	}
}
