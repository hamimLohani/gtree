package tree

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/hamimlohani/gtree/internal/theme"
)

// StyledRenderer renders the tree with full lipgloss styling.
// It gracefully degrades to the PlainRenderer when color is disabled.
type StyledRenderer struct {
	root    string
	palette *theme.Palette
	plain   bool
	width   int
}

// NewStyledRenderer creates a renderer that adapts to the terminal's color
// capabilities and width.
func NewStyledRenderer(root, themeName string) *StyledRenderer {
	palette, plain := theme.Load(themeName)
	width, _, err := term.GetSize(1) // fd 1 = stdout
	if err != nil || width <= 0 {
		width = 100
	}
	return &StyledRenderer{
		root:    root,
		palette: palette,
		plain:   plain,
		width:   width,
	}
}

// Render writes the full styled output (header + tree + footer) to w.
func (r *StyledRenderer) Render(w io.Writer, root *Node) {
	if r.plain {
		NewPlainRenderer(r.root).Render(w, root)
		return
	}

	var totalRepos, dirtyRepos, unpushedRepos int
	countStats(root, &totalRepos, &dirtyRepos, &unpushedRepos)

	fmt.Fprint(w, r.renderHeader(totalRepos))
	fmt.Fprintln(w)

	for i, child := range root.Children {
		isLast := i == len(root.Children)-1
		r.renderNode(w, child, "", isLast)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, r.renderSeparator())
	fmt.Fprintln(w, r.renderFooter(totalRepos, dirtyRepos, unpushedRepos))
	fmt.Fprintln(w)
}

// ── Header ───────────────────────────────────────────────────────────────────

func (r *StyledRenderer) renderHeader(repoCount int) string {
	p := r.palette

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.HeaderFg).
		Background(p.HeaderBg).
		PaddingLeft(1)

	pathStyle := lipgloss.NewStyle().
		Foreground(p.HeaderPath).
		Background(p.HeaderBg).
		PaddingLeft(2)

	countStyle := lipgloss.NewStyle().
		Foreground(p.HeaderFg).
		Background(p.HeaderBg).
		Faint(true).
		PaddingLeft(2).
		PaddingRight(1)

	bgStyle := lipgloss.NewStyle().
		Background(p.HeaderBg).
		Width(r.width)

	title := titleStyle.Render("🌳 gtree")
	path := pathStyle.Render(collapseTilde(r.root))
	count := countStyle.Render(fmt.Sprintf("%d repos", repoCount))

	line1 := bgStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, title, path, count))

	return "\n" + line1
}

// ── Separator ────────────────────────────────────────────────────────────────

func (r *StyledRenderer) renderSeparator() string {
	sep := strings.Repeat("─", r.width)
	return lipgloss.NewStyle().Foreground(r.palette.Separator).Render(sep)
}

// ── Tree nodes ───────────────────────────────────────────────────────────────

func (r *StyledRenderer) renderNode(w io.Writer, n *Node, prefix string, isLast bool) {
	connector := "├── "
	childPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		childPrefix = prefix + "    "
	}

	treeLine := lipgloss.NewStyle().Foreground(r.palette.TreeLine).Render(prefix + connector)

	if n.Repo != nil {
		row := r.formatRepoRow(n)
		fmt.Fprintf(w, "%s%s\n", treeLine, row)
	} else {
		dirStyle := lipgloss.NewStyle().
			Foreground(r.palette.DirName).
			Italic(true)
		fmt.Fprintf(w, "%s%s\n", treeLine, dirStyle.Render(n.Name+"/"))
	}

	for i, child := range n.Children {
		r.renderNode(w, child, childPrefix, i == len(n.Children)-1)
	}
}

// ── Repo row ─────────────────────────────────────────────────────────────────

const (
	colNameWidth   = 32
	colBranchWidth = 18
	colStatusWidth = 10
	colAbWidth     = 8
)

func (r *StyledRenderer) formatRepoRow(n *Node) string {
	repo := n.Repo
	p := r.palette

	// Repo name — bold, truncated if needed.
	name := filepath.Base(repo.Path)
	name = truncate(name, colNameWidth)
	nameStr := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.RepoName).
		Width(colNameWidth).
		Render(name)

	// Branch — dimmed in brackets.
	branchText := truncate(repo.Branch, colBranchWidth-2)
	branchStr := lipgloss.NewStyle().
		Foreground(p.BranchFg).
		Width(colBranchWidth).
		Render("[" + branchText + "]")

	// Status — colored icon + changed file count.
	var statusStr string
	if repo.IsDirty {
		statusStr = lipgloss.NewStyle().
			Foreground(p.Dirty).
			Bold(true).
			Width(colStatusWidth).
			Render(fmt.Sprintf("✗ ~%d", repo.ChangedFiles))
	} else {
		statusStr = lipgloss.NewStyle().
			Foreground(p.Clean).
			Width(colStatusWidth).
			Render("✓")
	}

	// Ahead / behind arrows.
	// Build the raw ANSI string first, then constrain it to a fixed width.
	var abRaw string
	if repo.Ahead > 0 {
		abRaw += lipgloss.NewStyle().Foreground(p.Ahead).Render(fmt.Sprintf("↑%d", repo.Ahead))
	}
	if repo.Behind > 0 {
		abRaw += lipgloss.NewStyle().Foreground(p.Behind).Render(fmt.Sprintf("↓%d", repo.Behind))
	}
	abStr := lipgloss.PlaceHorizontal(colAbWidth, lipgloss.Left, abRaw)

	// Stash badge.
	stashStr := ""
	if repo.StashCount > 0 {
		stashStr = lipgloss.NewStyle().
			Foreground(p.Stash).
			Render(fmt.Sprintf("[$%d] ", repo.StashCount))
	}

	// Last commit time — right-side, faint.
	timeStr := lipgloss.NewStyle().
		Foreground(p.CommitTime).
		Faint(true).
		Render(repo.LastCommit)

	// Error note.
	errStr := ""
	if repo.Error != nil {
		errStr = lipgloss.NewStyle().
			Foreground(p.Error).
			Render(fmt.Sprintf(" ⚠ %v", repo.Error))
	}

	return nameStr + branchStr + statusStr + abStr + stashStr + timeStr + errStr
}

// ── Footer ───────────────────────────────────────────────────────────────────

func (r *StyledRenderer) renderFooter(total, dirty, unpushed int) string {
	p := r.palette
	clean := total - dirty

	sep := lipgloss.NewStyle().Foreground(p.FooterFg).Render("  •  ")

	totalStr := lipgloss.NewStyle().Foreground(p.FooterFg).
		Bold(true).Render(fmt.Sprintf("  %d repos", total))
	cleanStr := lipgloss.NewStyle().Foreground(p.FooterClean).
		Render(fmt.Sprintf("%d clean", clean))
	dirtyStr := lipgloss.NewStyle().Foreground(p.FooterDirty).
		Render(fmt.Sprintf("%d dirty", dirty))

	out := totalStr + sep + cleanStr + sep + dirtyStr

	if unpushed > 0 {
		unpushedStr := lipgloss.NewStyle().
			Foreground(p.Unpushed).
			Bold(true).
			Render(fmt.Sprintf("%d unpushed ↑", unpushed))
		out += sep + unpushedStr
	}

	return out
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// truncate shortens s to maxLen runes, appending "…" if cut.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}
