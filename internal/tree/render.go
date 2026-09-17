package tree

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// PlainRenderer renders the tree as plain text with box-drawing characters.
// This is the Step 1 renderer — no color or lipgloss dependency.
type PlainRenderer struct {
	root string
}

// NewPlainRenderer creates a PlainRenderer that collapses root to "~" in paths.
func NewPlainRenderer(root string) *PlainRenderer {
	return &PlainRenderer{root: root}
}

// Render writes the full tree (header + body + footer) to w.
func (r *PlainRenderer) Render(w io.Writer, root *Node) {
	// Gather stats.
	var totalRepos, dirtyRepos, unpushedRepos int
	countStats(root, &totalRepos, &dirtyRepos, &unpushedRepos)

	// Header.
	displayPath := collapseTilde(r.root)
	fmt.Fprintf(w, "\ngtree — %s  (%d repos found)\n", displayPath, totalRepos)
	fmt.Fprintln(w, strings.Repeat("─", 60))

	// Tree body.
	for i, child := range root.Children {
		isLast := i == len(root.Children)-1
		r.renderNode(w, child, "", isLast)
	}

	// Footer.
	fmt.Fprintln(w, strings.Repeat("─", 60))
	fmt.Fprintf(w, "  %d repos  •  %d clean  •  %d dirty", totalRepos, totalRepos-dirtyRepos, dirtyRepos)
	if unpushedRepos > 0 {
		fmt.Fprintf(w, "  •  %d unpushed", unpushedRepos)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)
}

// renderNode recursively renders a single node and its children.
func (r *PlainRenderer) renderNode(w io.Writer, n *Node, prefix string, isLast bool) {
	// Choose the connector and the continuation prefix for children.
	connector := "├── "
	childPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		childPrefix = prefix + "    "
	}

	if n.Repo != nil {
		// Leaf: render repo row.
		fmt.Fprintf(w, "%s%s%s\n", prefix, connector, formatRepoLine(n))
	} else {
		// Interior directory node.
		fmt.Fprintf(w, "%s%s%s/\n", prefix, connector, n.Name)
	}

	for i, child := range n.Children {
		r.renderNode(w, child, childPrefix, i == len(n.Children)-1)
	}
}

// formatRepoLine builds a single repo display line.
// Step 2 will replace this with a lipgloss-styled version.
func formatRepoLine(n *Node) string {
	repo := n.Repo

	// Status icon.
	status := "✓"
	if repo.IsDirty {
		status = fmt.Sprintf("✗ ~%d", repo.ChangedFiles)
	}

	// Ahead/behind.
	ab := ""
	if repo.Ahead > 0 || repo.Behind > 0 {
		ab = fmt.Sprintf(" ↑%d↓%d", repo.Ahead, repo.Behind)
	}

	// Stash indicator.
	stash := ""
	if repo.StashCount > 0 {
		stash = fmt.Sprintf(" [$%d]", repo.StashCount)
	}

	// Error note.
	errNote := ""
	if repo.Error != nil {
		errNote = fmt.Sprintf(" [err: %v]", repo.Error)
	}

	return fmt.Sprintf("%-30s  [%s]  %s%s%s  %s%s",
		filepath.Base(repo.Path),
		repo.Branch,
		status,
		ab,
		stash,
		repo.LastCommit,
		errNote,
	)
}

// countStats tallies repo, dirty, and unpushed counts recursively.
func countStats(n *Node, total, dirty, unpushed *int) {
	if n.Repo != nil {
		*total++
		if n.Repo.IsDirty {
			*dirty++
		}
		if n.Repo.Ahead > 0 {
			*unpushed++
		}
		return
	}
	for _, c := range n.Children {
		countStats(c, total, dirty, unpushed)
	}
}

// collapseTilde replaces the home directory prefix with "~".
func collapseTilde(path string) string {
	home := homeDir()
	if home == "" {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

// homeDir returns the user's home directory. Delegating to the package-level
// osUserHomeDir variable makes this overridable in tests.
func homeDir() string {
	return osUserHomeDir()
}
