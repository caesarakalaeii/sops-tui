// Package git provides tests for the git integration backend.
//
// Tests use t.TempDir() and go-git's PlainInit to create isolated test repos.
// All tests are safe to run in parallel.
package git_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/git"
)

// initRepo initializes a bare-minimum git repository in tmpDir with a
// fake user identity set via the git environment variables.
func initRepo(t *testing.T, tmpDir string) *gogit.Repository {
	t.Helper()
	repo, err := gogit.PlainInit(tmpDir, false)
	require.NoError(t, err)
	return repo
}

// commitFile creates (or overwrites) a file at relPath in the working tree,
// stages it, and commits with the provided message.
func commitFile(t *testing.T, repo *gogit.Repository, dir, relPath, content, message string) {
	t.Helper()

	absPath := filepath.Join(dir, relPath)
	err := os.MkdirAll(filepath.Dir(absPath), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(absPath, []byte(content), 0o644)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	_, err = wt.Add(relPath)
	require.NoError(t, err)

	sig := &object.Signature{
		Name:  "Test Author",
		Email: "test@example.com",
		When:  time.Now(),
	}
	_, err = wt.Commit(message, &gogit.CommitOptions{
		Author:    sig,
		Committer: sig,
	})
	require.NoError(t, err)
}

// TestIsGitRepo verifies that IsGitRepo returns false for non-git dirs
// and true for git repos.
func TestIsGitRepo(t *testing.T) {
	t.Run("non-git directory returns false", func(t *testing.T) {
		dir := t.TempDir()
		assert.False(t, git.IsGitRepo(dir))
	})

	t.Run("git repo returns true", func(t *testing.T) {
		dir := t.TempDir()
		initRepo(t, dir)
		assert.True(t, git.IsGitRepo(dir))
	})
}

// TestGetFileStatuses tests all status codes (modified, untracked, clean) and the
// non-git directory case.
func TestGetFileStatuses(t *testing.T) {
	t.Run("non-git directory returns empty map and nil error", func(t *testing.T) {
		dir := t.TempDir()
		statuses, err := git.GetFileStatuses(dir, []string{"anything.yaml"})
		require.NoError(t, err)
		assert.Empty(t, statuses)
	})

	t.Run("modified file returns GitStatusModified", func(t *testing.T) {
		dir := t.TempDir()
		repo := initRepo(t, dir)

		// Commit the file first
		commitFile(t, repo, dir, "secrets.yaml", "original", "initial commit")

		// Modify the file without committing
		err := os.WriteFile(filepath.Join(dir, "secrets.yaml"), []byte("modified"), 0o644)
		require.NoError(t, err)

		statuses, err := git.GetFileStatuses(dir, []string{"secrets.yaml"})
		require.NoError(t, err)
		assert.Equal(t, git.GitStatusModified, statuses["secrets.yaml"])
	})

	t.Run("untracked file returns GitStatusUntracked", func(t *testing.T) {
		dir := t.TempDir()
		repo := initRepo(t, dir)
		// Commit at least one file so the repo has a valid HEAD
		commitFile(t, repo, dir, "committed.yaml", "data", "initial commit")

		// Create an untracked file
		err := os.WriteFile(filepath.Join(dir, "untracked.yaml"), []byte("new"), 0o644)
		require.NoError(t, err)

		statuses, err := git.GetFileStatuses(dir, []string{"untracked.yaml"})
		require.NoError(t, err)
		assert.Equal(t, git.GitStatusUntracked, statuses["untracked.yaml"])
	})

	t.Run("clean committed file returns GitStatusClean", func(t *testing.T) {
		dir := t.TempDir()
		repo := initRepo(t, dir)
		// Commit two files so we can check one that is clean but has a sibling
		commitFile(t, repo, dir, "clean.yaml", "data", "initial commit")
		// Modify the sibling to have at least one dirty entry, ensuring HEAD exists
		commitFile(t, repo, dir, "other.yaml", "other", "second commit")

		statuses, err := git.GetFileStatuses(dir, []string{"clean.yaml"})
		require.NoError(t, err)
		assert.Equal(t, git.GitStatusClean, statuses["clean.yaml"])
	})
}

// TestGetFileHistory tests commit log retrieval and error behavior on non-git dirs.
func TestGetFileHistory(t *testing.T) {
	t.Run("non-git directory returns error", func(t *testing.T) {
		dir := t.TempDir()
		entries, err := git.GetFileHistory(dir, "secrets.yaml", 5)
		assert.Error(t, err)
		assert.Nil(t, entries)
	})

	t.Run("returns CommitEntry slices with expected fields", func(t *testing.T) {
		dir := t.TempDir()
		repo := initRepo(t, dir)
		commitFile(t, repo, dir, "secrets.yaml", "v1", "first commit")
		commitFile(t, repo, dir, "secrets.yaml", "v2", "second commit")

		entries, err := git.GetFileHistory(dir, "secrets.yaml", 10)
		require.NoError(t, err)
		require.NotEmpty(t, entries)

		first := entries[0]
		assert.Len(t, first.ShortHash, 7, "ShortHash must be 7 characters")
		assert.NotEmpty(t, first.Author)
		assert.NotEmpty(t, first.Subject)
		assert.NotEmpty(t, first.RelDate)
	})

	t.Run("limit parameter is respected", func(t *testing.T) {
		dir := t.TempDir()
		repo := initRepo(t, dir)
		// Create 3 commits with distinct content so go-git does not reject empty commits
		commitFile(t, repo, dir, "secrets.yaml", "version1", "commit 1")
		commitFile(t, repo, dir, "secrets.yaml", "version2", "commit 2")
		commitFile(t, repo, dir, "secrets.yaml", "version3", "commit 3")

		entries, err := git.GetFileHistory(dir, "secrets.yaml", 2)
		require.NoError(t, err)
		assert.Len(t, entries, 2, "limit=2 should return at most 2 entries")
	})
}

// TestGetLastCommitTime verifies GetLastCommitTime returns correct timestamps.
func TestGetLastCommitTime(t *testing.T) {
	t.Run("returns commit timestamp for committed file", func(t *testing.T) {
		dir := t.TempDir()
		repo := initRepo(t, dir)
		commitFile(t, repo, dir, "secrets.yaml", "v1", "initial commit")

		ts, err := git.GetLastCommitTime(dir, "secrets.yaml")
		require.NoError(t, err)
		assert.False(t, ts.IsZero(), "commit timestamp must not be zero for a committed file")
	})

	t.Run("non-git directory returns error", func(t *testing.T) {
		dir := t.TempDir()
		ts, err := git.GetLastCommitTime(dir, "secrets.yaml")
		assert.Error(t, err, "non-git directory must return an error")
		assert.True(t, ts.IsZero(), "timestamp must be zero on error")
	})

	t.Run("file with no commits returns zero time", func(t *testing.T) {
		dir := t.TempDir()
		repo := initRepo(t, dir)
		// Commit a different file so the repo has a valid HEAD, but don't commit secrets.yaml
		commitFile(t, repo, dir, "other.yaml", "data", "initial commit")

		ts, err := git.GetLastCommitTime(dir, "secrets.yaml")
		require.NoError(t, err)
		assert.True(t, ts.IsZero(), "timestamp must be zero for a file with no commits")
	})
}

// TestRelativeTime exercises the exported relativeTime function.
func TestRelativeTime(t *testing.T) {
	t.Run("just now for sub-minute duration", func(t *testing.T) {
		// 30 seconds ago
		result := git.RelativeTime(time.Now().Add(-30 * time.Second))
		assert.Equal(t, "just now", result)
	})

	t.Run("minutes ago", func(t *testing.T) {
		result := git.RelativeTime(time.Now().Add(-5 * time.Minute))
		assert.Equal(t, "5 minutes ago", result)
	})

	t.Run("1 minute ago singular", func(t *testing.T) {
		result := git.RelativeTime(time.Now().Add(-61 * time.Second))
		assert.Equal(t, "1 minute ago", result)
	})
}

// TestGetBranch verifies branch resolution and the detached-HEAD case.
// Mirrors TestGetFileStatuses 3-subtest shape (status_test.go:76-127).
// Phase 8 D-215.
func TestGetBranch(t *testing.T) {
	t.Run("non-git directory returns ErrRepositoryNotExists", func(t *testing.T) {
		dir := t.TempDir()
		branch, detached, err := git.GetBranch(dir)
		require.ErrorIs(t, err, gogit.ErrRepositoryNotExists)
		assert.Equal(t, "", branch)
		assert.False(t, detached)
	})

	t.Run("normal branch returns branch name", func(t *testing.T) {
		dir := t.TempDir()
		repo := initRepo(t, dir)
		commitFile(t, repo, dir, "first.yaml", "data", "initial commit")

		branch, detached, err := git.GetBranch(dir)
		require.NoError(t, err)
		assert.False(t, detached)
		// PlainInit creates the default branch (master in go-git v5.17.0;
		// some go-git versions / configs default to main). Accept either
		// so the test is portable across CI configurations.
		assert.Contains(t, []string{"master", "main"}, branch,
			"expected master or main, got %q", branch)
	})

	t.Run("detached HEAD returns 7-char hash with detached=true", func(t *testing.T) {
		dir := t.TempDir()
		repo := initRepo(t, dir)
		commitFile(t, repo, dir, "first.yaml", "data", "initial commit")

		// Detach HEAD by checking out the commit hash directly.
		head, err := repo.Head()
		require.NoError(t, err)
		wt, err := repo.Worktree()
		require.NoError(t, err)
		err = wt.Checkout(&gogit.CheckoutOptions{Hash: head.Hash()})
		require.NoError(t, err)

		branch, detached, err := git.GetBranch(dir)
		require.NoError(t, err)
		assert.True(t, detached, "checkout to commit hash must produce detached HEAD")
		assert.Len(t, branch, 7, "detached HEAD branch must be 7-char short hash; got %q (%d chars)", branch, len(branch))
		// Verify it is the prefix of the actual commit hash.
		assert.Equal(t, head.Hash().String()[:7], branch)
	})
}
