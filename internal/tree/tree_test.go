package tree

import (
	"testing"

	"github.com/hamimlohani/gtree/internal/scanner"
)

func TestBuildFlat(t *testing.T) {
	repos := []*scanner.Repo{
		{Path: "/root/a"},
		{Path: "/root/b"},
	}
	root := Build("/root", repos)

	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Children))
	}
	names := []string{root.Children[0].Name, root.Children[1].Name}
	if names[0] != "a" || names[1] != "b" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestBuildNested(t *testing.T) {
	repos := []*scanner.Repo{
		{Path: "/root/group/a"},
		{Path: "/root/group/b"},
		{Path: "/root/c"},
	}
	root := Build("/root", repos)

	// Expected structure:
	//  root/
	//  ├── group/    (dir node)
	//  │   ├── a    (repo)
	//  │   └── b    (repo)
	//  └── c        (repo)

	if len(root.Children) != 2 {
		t.Fatalf("expected 2 top-level children, got %d: %v", len(root.Children), childNames(root))
	}

	// First child should be the "group" directory node (dirs before repos).
	groupNode := root.Children[0]
	if groupNode.Name != "group" {
		t.Errorf("expected first child to be 'group', got %q", groupNode.Name)
	}
	if groupNode.Repo != nil {
		t.Error("'group' node should be a directory, not a repo")
	}
	if len(groupNode.Children) != 2 {
		t.Errorf("expected 2 children under 'group', got %d", len(groupNode.Children))
	}

	// Second child is the lone repo "c".
	cNode := root.Children[1]
	if cNode.Name != "c" || cNode.Repo == nil {
		t.Errorf("expected repo leaf 'c', got name=%q repo=%v", cNode.Name, cNode.Repo)
	}
}

func TestBuildRoot(t *testing.T) {
	// Repo at the scan root itself.
	repos := []*scanner.Repo{
		{Path: "/root"},
	}
	root := Build("/root", repos)
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children))
	}
	if root.Children[0].Repo == nil {
		t.Error("expected leaf repo node")
	}
}

func childNames(n *Node) []string {
	names := make([]string, len(n.Children))
	for i, c := range n.Children {
		names[i] = c.Name
	}
	return names
}
