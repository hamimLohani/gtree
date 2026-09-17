package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// initGitRepo creates a minimal git repo at dir (must already exist).
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		// Suppress output in tests.
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	// Create an initial commit so log commands work.
	f := filepath.Join(dir, "README.md")
	if err := os.WriteFile(f, []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "init")
}

func TestScanFindsRepos(t *testing.T) {
	root := t.TempDir()

	// Create two repos: root/a and root/b/c
	aDir := filepath.Join(root, "a")
	cDir := filepath.Join(root, "b", "c")
	for _, d := range []string{aDir, cDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		initGitRepo(t, d)
	}

	opts := Options{WorkerCount: 2}
	repos, err := Scan(root, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(repos))
	}
}

func TestScanRespectsMaxDepth(t *testing.T) {
	root := t.TempDir()

	// Shallow repo: root/a (depth 1)
	aDir := filepath.Join(root, "a")
	if err := os.MkdirAll(aDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, aDir)

	// Deep repo: root/b/c/d (depth 3)
	dDir := filepath.Join(root, "b", "c", "d")
	if err := os.MkdirAll(dDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, dDir)

	opts := Options{MaxDepth: 2, WorkerCount: 2}
	repos, _ := Scan(root, opts)

	if len(repos) != 1 {
		t.Errorf("with MaxDepth=2 expected 1 repo, got %d", len(repos))
	}
	if len(repos) > 0 && filepath.Base(repos[0].Path) != "a" {
		t.Errorf("expected repo 'a', got %s", repos[0].Path)
	}
}

func TestScanIgnoresDir(t *testing.T) {
	root := t.TempDir()

	// Repo inside an ignored directory.
	nodeDir := filepath.Join(root, "node_modules", "pkg")
	if err := os.MkdirAll(nodeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, nodeDir)

	// Visible repo.
	aDir := filepath.Join(root, "a")
	if err := os.MkdirAll(aDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, aDir)

	opts := Options{
		IgnoreDirs:  []string{"node_modules"},
		WorkerCount: 2,
	}
	repos, _ := Scan(root, opts)

	if len(repos) != 1 {
		t.Errorf("expected 1 repo (node_modules ignored), got %d", len(repos))
	}
}

func TestScanOnlyDirty(t *testing.T) {
	root := t.TempDir()

	// Clean repo.
	cleanDir := filepath.Join(root, "clean")
	if err := os.MkdirAll(cleanDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, cleanDir)

	// Dirty repo: add an untracked file after init.
	dirtyDir := filepath.Join(root, "dirty")
	if err := os.MkdirAll(dirtyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, dirtyDir)
	if err := os.WriteFile(filepath.Join(dirtyDir, "new.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := Options{OnlyDirty: true, WorkerCount: 2}
	repos, _ := Scan(root, opts)

	if len(repos) != 1 {
		t.Errorf("--only-dirty: expected 1 repo, got %d", len(repos))
	}
	if len(repos) > 0 && !repos[0].IsDirty {
		t.Error("returned repo is not dirty")
	}
}

func TestSortByDirty(t *testing.T) {
	repos := []*Repo{
		{Path: "/root/a", IsDirty: false},
		{Path: "/root/b", IsDirty: true},
		{Path: "/root/c", IsDirty: false},
		{Path: "/root/d", IsDirty: true},
	}
	sortRepos(repos, "dirty")
	if !repos[0].IsDirty || !repos[1].IsDirty {
		t.Error("dirty repos should appear first")
	}
	if repos[2].IsDirty || repos[3].IsDirty {
		t.Error("clean repos should appear after dirty")
	}
}

func TestSortByRecent(t *testing.T) {
	now := time.Now()
	repos := []*Repo{
		{Path: "/root/old", LastCommitAt: now.Add(-48 * time.Hour)},
		{Path: "/root/new", LastCommitAt: now.Add(-1 * time.Hour)},
		{Path: "/root/mid", LastCommitAt: now.Add(-24 * time.Hour)},
	}
	sortRepos(repos, "recent")
	if repos[0].Path != "/root/new" {
		t.Errorf("most recent should be first, got %s", repos[0].Path)
	}
	if repos[2].Path != "/root/old" {
		t.Errorf("oldest should be last, got %s", repos[2].Path)
	}
}
