// Package scanner finds and inspects git repositories concurrently.
package scanner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"golang.org/x/sync/semaphore"
)

// Options controls scanner behaviour.
type Options struct {
	// MaxDepth is the maximum directory depth to recurse into.
	// 0 means unlimited.
	MaxDepth int
	// IgnoreDirs is a set of directory base-names to skip entirely.
	IgnoreDirs []string
	// OnlyDirty, if true, excludes clean repositories from results.
	OnlyDirty bool
	// SortOrder is one of "name", "dirty", "recent".
	SortOrder string
	// WorkerCount is the size of the goroutine pool used for git calls.
	WorkerCount int
}

// Repo holds all status information for a single git repository.
type Repo struct {
	// Path is the absolute path to the repository root.
	Path string
	// Branch is the current branch name, or "detached @ <sha>".
	Branch string
	// IsDirty is true when the working tree has uncommitted changes.
	IsDirty bool
	// ChangedFiles is the count of modified/untracked files.
	ChangedFiles int
	// Ahead is the number of commits ahead of upstream.
	Ahead int
	// Behind is the number of commits behind upstream.
	Behind int
	// StashCount is the number of stash entries.
	StashCount int
	// LastCommit is a human-relative time string (e.g. "2h ago").
	LastCommit string
	// Error records any non-fatal error encountered while collecting info.
	Error error
}

// Scan walks root recursively, finds all git repositories, and collects
// their status concurrently. It returns partial results even on error.
func Scan(root string, opts Options) ([]*Repo, error) {
	if opts.WorkerCount <= 0 {
		opts.WorkerCount = 4
	}

	ignoredSet := make(map[string]bool, len(opts.IgnoreDirs))
	for _, d := range opts.IgnoreDirs {
		ignoredSet[d] = true
	}

	// Phase 1: walk the directory tree and collect repo paths.
	var repoPaths []string
	var walkErrors []error

	err := walkRepos(root, opts.MaxDepth, ignoredSet, &repoPaths, &walkErrors)
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", root, err)
	}

	// Phase 2: collect git info concurrently using a semaphore worker pool.
	ctx := context.Background()
	sem := semaphore.NewWeighted(int64(opts.WorkerCount))

	var mu sync.Mutex
	var repos []*Repo

	var wg sync.WaitGroup
	for _, p := range repoPaths {
		p := p // capture
		wg.Add(1)
		if err := sem.Acquire(ctx, 1); err != nil {
			wg.Done()
			continue
		}
		go func() {
			defer wg.Done()
			defer sem.Release(1)

			repo := collectInfo(p)
			mu.Lock()
			repos = append(repos, repo)
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Phase 3: filter and sort.
	if opts.OnlyDirty {
		filtered := repos[:0]
		for _, r := range repos {
			if r.IsDirty {
				filtered = append(filtered, r)
			}
		}
		repos = filtered
	}

	sortRepos(repos, opts.SortOrder)

	// Combine walk errors into a single non-nil error if any occurred.
	var combinedErr error
	if len(walkErrors) > 0 {
		msgs := make([]string, len(walkErrors))
		for i, e := range walkErrors {
			msgs[i] = e.Error()
		}
		combinedErr = fmt.Errorf("%d path(s) skipped due to errors (first: %s)", len(walkErrors), walkErrors[0])
	}

	return repos, combinedErr
}

// walkRepos recursively walks root, appending git repo paths to out.
// It skips ignored directories and respects the max depth.
func walkRepos(root string, maxDepth int, ignored map[string]bool, out *[]string, errs *[]error) error {
	return walkDir(root, root, 0, maxDepth, ignored, out, errs)
}

func walkDir(root, dir string, depth, maxDepth int, ignored map[string]bool, out *[]string, errs *[]error) error {
	if maxDepth > 0 && depth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		*errs = append(*errs, fmt.Errorf("reading %s: %w", dir, err))
		return nil // skip; don't abort the whole scan
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()

		// Skip hidden non-.git directories and ignored names.
		if name != ".git" && name[0] == '.' {
			continue
		}
		if ignored[name] {
			continue
		}

		fullPath := filepath.Join(dir, name)

		if name == ".git" {
			// Found a git repo: record the parent directory.
			repoRoot := dir
			// Skip submodule .git files (they're files, not dirs).
			info, err := e.Info()
			if err != nil || info.Mode()&fs.ModeSymlink != 0 {
				continue
			}
			// Skip bare repos (no working tree, .git is the repo root itself).
			if isBareRepo(dir) {
				continue
			}
			*out = append(*out, repoRoot)
			// Don't recurse into nested repos' contents.
			return nil
		}

		// Check for symlinks to avoid loops.
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			continue
		}

		if err := walkDir(root, fullPath, depth+1, maxDepth, ignored, out, errs); err != nil {
			return err
		}
	}
	return nil
}

// isBareRepo returns true if dir looks like a bare git repository
// (contains HEAD, objects, refs directly rather than a .git subdir).
func isBareRepo(dir string) bool {
	// A bare repo has HEAD and objects/ directly inside.
	for _, marker := range []string{"HEAD", "objects", "refs"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err != nil {
			return false
		}
	}
	// A non-bare repo has a .git subdirectory.
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return false
	}
	return true
}

// sortRepos sorts repos in-place according to order.
func sortRepos(repos []*Repo, order string) {
	switch order {
	case "dirty":
		sort.Slice(repos, func(i, j int) bool {
			if repos[i].IsDirty != repos[j].IsDirty {
				return repos[i].IsDirty // dirty first
			}
			return repos[i].Path < repos[j].Path
		})
	case "recent":
		// LastCommit is a human string; for now fall back to name sort.
		// Step 3 will store a time.Time for proper sorting.
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Path < repos[j].Path
		})
	default: // "name"
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Path < repos[j].Path
		})
	}
}
