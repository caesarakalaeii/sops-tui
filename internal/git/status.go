// Package git provides git integration for sops-tui using go-git v5.
//
// Functions: IsGitRepo checks for a git repository, GetFileStatuses returns
// worktree status for specific files, GetFileHistory returns commit log entries.
// All functions are safe to call on non-git directories.
//
// Per D-16: go-git/go-git v5 provides the git backend (pure Go, no git binary).
// Per D-12: non-git repos return empty results, not errors.
package git

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// GitStatus represents the worktree status of a file in a git repository.
type GitStatus string

const (
	// GitStatusClean indicates the file is committed and unchanged.
	GitStatusClean GitStatus = ""
	// GitStatusModified indicates the file has uncommitted modifications.
	GitStatusModified GitStatus = "M"
	// GitStatusAdded indicates the file has been staged for addition.
	GitStatusAdded GitStatus = "A"
	// GitStatusUntracked indicates the file is not tracked by git.
	GitStatusUntracked GitStatus = "?"
)

// CommitEntry is a single row in the git history overlay (D-15).
type CommitEntry struct {
	// ShortHash is the first 7 characters of the full commit hash hex.
	ShortHash string
	// RelDate is a human-readable relative date, e.g. "3 days ago".
	RelDate string
	// Author is the commit author name from the git config.
	Author string
	// Subject is the first line of the commit message.
	Subject string
}

// IsGitRepo reports whether dir is inside a git repository.
// Uses DetectDotGit so subdirectories of a repo root also return true.
// Safe to call on any directory — always returns false for non-git paths.
func IsGitRepo(dir string) bool {
	_, err := gogit.PlainOpenWithOptions(dir, &gogit.PlainOpenOptions{DetectDotGit: true})
	return err == nil
}

// GetFileStatuses returns the worktree git status for each file in relPaths.
// repoRoot is the directory containing (or inside) the .git directory.
// relPaths are slash-separated paths relative to repoRoot.
//
// Non-git directories (ErrRepositoryNotExists) return an empty map and nil error
// per D-12 — the caller can treat the absence of entries as "no git".
//
// Pitfall 6: go-git uses forward slashes for all path keys; filepath.ToSlash
// ensures correct lookup on Windows paths.
func GetFileStatuses(repoRoot string, relPaths []string) (map[string]GitStatus, error) {
	repo, err := gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})
	if err == gogit.ErrRepositoryNotExists {
		// D-12: non-git directory is not an error condition — return empty map.
		return map[string]GitStatus{}, nil
	}
	if err != nil {
		return nil, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := wt.Status()
	if err != nil {
		return nil, err
	}

	result := make(map[string]GitStatus, len(relPaths))
	for _, relPath := range relPaths {
		// Pitfall 6: go-git status map uses forward slashes as path separators.
		slashPath := filepath.ToSlash(relPath)

		// Use direct map access rather than status.File() to avoid go-git's
		// auto-creation behaviour: File() inserts a new entry with Untracked/Untracked
		// for any path not already in the map, which would wrongly report committed
		// clean files as untracked.
		fs, ok := status[slashPath]
		if !ok {
			// File is not in the status map → it is committed and unmodified (clean).
			result[relPath] = GitStatusClean
			continue
		}

		var gs GitStatus
		switch {
		case fs.Worktree == gogit.Modified || fs.Staging == gogit.Modified:
			gs = GitStatusModified
		case fs.Staging == gogit.Added:
			gs = GitStatusAdded
		case fs.Worktree == gogit.Untracked:
			gs = GitStatusUntracked
		default:
			gs = GitStatusClean
		}
		result[relPath] = gs
	}

	return result, nil
}

// GetFileHistory returns up to limit commit entries for the file at relPath
// within the repository at repoRoot.
//
// relPath must be slash-separated and relative to the repository root (Pitfall 6).
// If limit <= 0, all commits are returned.
//
// Returns nil, err when repoRoot is not a git repository — unlike GetFileStatuses,
// the caller (detail history view) needs to know that history is unavailable.
func GetFileHistory(repoRoot, relPath string, limit int) ([]CommitEntry, error) {
	repo, err := gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, err
	}

	// Pitfall 6: go-git expects forward-slash paths in LogOptions.
	slashPath := filepath.ToSlash(relPath)

	iter, err := repo.Log(&gogit.LogOptions{FileName: &slashPath})
	if err != nil {
		return nil, err
	}

	var entries []CommitEntry
	err = iter.ForEach(func(c *object.Commit) error {
		if limit > 0 && len(entries) >= limit {
			return storer.ErrStop
		}
		subject := c.Message
		if idx := strings.IndexByte(subject, '\n'); idx >= 0 {
			subject = subject[:idx]
		}
		subject = strings.TrimSpace(subject)
		entries = append(entries, CommitEntry{
			ShortHash: c.Hash.String()[:7],
			RelDate:   relativeTime(c.Author.When),
			Author:    c.Author.Name,
			Subject:   subject,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// relativeTime converts a past time into a human-readable relative string.
// Examples: "just now", "3 minutes ago", "2 hours ago", "1 day ago".
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		n := int(d.Minutes())
		if n == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", n)
	case d < 24*time.Hour:
		n := int(d.Hours())
		if n == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", n)
	case d < 30*24*time.Hour:
		n := int(d.Hours() / 24)
		if n == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", n)
	case d < 365*24*time.Hour:
		n := int(d.Hours() / (24 * 30))
		if n == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", n)
	default:
		n := int(d.Hours() / (24 * 365))
		if n == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", n)
	}
}

// GetLastCommitTime returns the timestamp of the most recent commit that touched relPath
// within the repository at repoRoot.
//
// relPath should be slash-separated and relative to the repository root (Pitfall 6).
// Returns time.Time{} (zero value) if the file has never been committed.
// Returns a non-nil error if repoRoot is not a git repository.
func GetLastCommitTime(repoRoot, relPath string) (time.Time, error) {
	repo, err := gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return time.Time{}, err
	}

	// Pitfall 6: go-git expects forward-slash paths in LogOptions.
	slashPath := filepath.ToSlash(relPath)

	iter, err := repo.Log(&gogit.LogOptions{FileName: &slashPath})
	if err != nil {
		return time.Time{}, err
	}

	var commitTime time.Time
	err = iter.ForEach(func(c *object.Commit) error {
		commitTime = c.Author.When
		return storer.ErrStop // stop after first (most recent) commit
	})
	if err != nil && err != storer.ErrStop {
		return time.Time{}, err
	}

	return commitTime, nil
}

// RelativeTime exports relativeTime for testing.
var RelativeTime = relativeTime
