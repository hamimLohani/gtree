// Package tree builds a nested structure from a flat list of repo paths
// and renders it as a tree.
package tree

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/hamimlohani/gtree/internal/scanner"
)

// Node represents one element in the display tree.
// It is either a directory group node or a leaf (git repo).
type Node struct {
	// Name is the base name of this node (directory segment or repo name).
	Name string
	// FullPath is the absolute path this node represents.
	FullPath string
	// Repo is non-nil for leaf nodes (git repositories).
	Repo *scanner.Repo
	// Children holds nested directory nodes.
	Children []*Node
}

// Build constructs a nested Node tree from a flat list of repos.
// root is the scan root used to compute relative paths for grouping.
func Build(root string, repos []*scanner.Repo) *Node {
	rootNode := &Node{
		Name:     filepath.Base(root),
		FullPath: root,
	}

	for _, repo := range repos {
		rel, err := filepath.Rel(root, repo.Path)
		if err != nil || rel == "." {
			// Repo is the root itself — add it as a direct child.
			rootNode.Children = append(rootNode.Children, &Node{
				Name:     filepath.Base(repo.Path),
				FullPath: repo.Path,
				Repo:     repo,
			})
			continue
		}

		// Split the relative path into segments and walk/create nodes.
		segments := strings.Split(filepath.ToSlash(rel), "/")
		insert(rootNode, segments, repo)
	}

	sortNodes(rootNode)
	return rootNode
}

// insert walks segments into the tree rooted at parent, creating intermediate
// directory nodes as needed, and attaches repo as a leaf at the end.
func insert(parent *Node, segments []string, repo *scanner.Repo) {
	if len(segments) == 0 {
		return
	}

	seg := segments[0]
	rest := segments[1:]

	if len(rest) == 0 {
		// Leaf: this segment IS the repo.
		parent.Children = append(parent.Children, &Node{
			Name:     seg,
			FullPath: repo.Path,
			Repo:     repo,
		})
		return
	}

	// Interior node: find or create a directory node for seg.
	for _, child := range parent.Children {
		if child.Name == seg && child.Repo == nil {
			insert(child, rest, repo)
			return
		}
	}

	// Create a new directory node.
	dirNode := &Node{
		Name:     seg,
		FullPath: filepath.Join(parent.FullPath, seg),
	}
	parent.Children = append(parent.Children, dirNode)
	insert(dirNode, rest, repo)
}

// sortNodes recursively sorts children: directory nodes before repo leaves,
// then alphabetically within each group.
func sortNodes(n *Node) {
	sort.Slice(n.Children, func(i, j int) bool {
		ci, cj := n.Children[i], n.Children[j]
		// Directories before repos.
		if (ci.Repo == nil) != (cj.Repo == nil) {
			return ci.Repo == nil
		}
		return ci.Name < cj.Name
	})
	for _, child := range n.Children {
		sortNodes(child)
	}
}
