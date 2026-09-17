package scanner

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// collectInfo runs git commands against repoPath and returns a populated Repo.
// All errors are non-fatal: fields are left at zero values on failure.
func collectInfo(repoPath string) *Repo {
	repo := &Repo{Path: repoPath}
	repo.Branch = branch(repoPath)
	repo.IsDirty, repo.ChangedFiles = dirtyState(repoPath)
	repo.Ahead, repo.Behind = aheadBehind(repoPath)
	repo.StashCount = stashCount(repoPath)
	repo.LastCommit, repo.LastCommitAt = lastCommitTime(repoPath)
	return repo
}

// runGit executes git with the given arguments inside dir and returns stdout.
// It returns an empty string on error.
func runGit(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	// Discard stderr to avoid noise.
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// branch returns the current branch name or "detached @ <short-sha>".
func branch(dir string) string {
	// --short returns the branch name or nothing in detached HEAD state.
	name := runGit(dir, "symbolic-ref", "--short", "HEAD")
	if name != "" {
		return name
	}
	// Detached HEAD: show short SHA.
	sha := runGit(dir, "rev-parse", "--short", "HEAD")
	if sha == "" {
		return "unknown"
	}
	return fmt.Sprintf("detached @ %s", sha)
}

// dirtyState returns whether the repo has uncommitted changes and how many files are affected.
func dirtyState(dir string) (isDirty bool, changedFiles int) {
	return parsePorcelain(runGit(dir, "status", "--porcelain"))
}

// parsePorcelain counts non-empty lines in git status --porcelain output.
// Extracted as a pure function for testability.
func parsePorcelain(output string) (isDirty bool, count int) {
	if strings.TrimSpace(output) == "" {
		return false, 0
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	count = 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			count++
		}
	}
	return count > 0, count
}

// aheadBehind returns the number of commits ahead/behind the upstream branch.
// Returns 0, 0 if there is no upstream configured.
func aheadBehind(dir string) (ahead, behind int) {
	output := runGit(dir, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	return parseAheadBehind(output)
}

// parseAheadBehind parses the output of `git rev-list --left-right --count HEAD...@{u}`.
// Extracted as a pure function for testability.
func parseAheadBehind(output string) (ahead, behind int) {
	if output == "" {
		return 0, 0
	}
	parts := strings.Fields(output)
	if len(parts) != 2 {
		return 0, 0
	}
	ahead, _ = strconv.Atoi(parts[0])
	behind, _ = strconv.Atoi(parts[1])
	return ahead, behind
}

// stashCount returns the number of stash entries.
func stashCount(dir string) int {
	output := runGit(dir, "stash", "list")
	if output == "" {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	count := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			count++
		}
	}
	return count
}

// lastCommitTime returns a human-relative string and exact timestamp for the most recent commit.
func lastCommitTime(dir string) (string, time.Time) {
	// %ct is the Unix timestamp of the commit.
	tsStr := runGit(dir, "log", "-1", "--format=%ct")
	if tsStr == "" {
		return "no commits", time.Time{}
	}
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return "unknown", time.Time{}
	}
	t := time.Unix(ts, 0)
	return humanDuration(time.Since(t)), t
}

// humanDuration returns a short, human-readable duration string.
func humanDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	case d < 30*24*time.Hour:
		weeks := int(d.Hours() / (24 * 7))
		return fmt.Sprintf("%dw ago", weeks)
	case d < 365*24*time.Hour:
		months := int(d.Hours() / (24 * 30))
		return fmt.Sprintf("%dmo ago", months)
	default:
		years := int(d.Hours() / (24 * 365))
		return fmt.Sprintf("%dy ago", years)
	}
}
