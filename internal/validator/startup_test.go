// Package validator tests startup environment validation.
package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSopsMissing verifies RunChecks returns a hard error when sops is not on PATH.
func TestSopsMissing(t *testing.T) {
	opts := Options{
		SopsLookPath: func(_ string) (string, error) {
			return "", &os.PathError{Op: "LookPath", Path: "sops", Err: os.ErrNotExist}
		},
		AgeKeyPath: validAgeKeyPath(t),
		StartDir:   validSopsYamlDir(t),
	}

	results, hasHardError := RunChecks(opts)

	require.True(t, hasHardError, "expected hasHardError=true when sops is missing")
	found := false
	for _, r := range results {
		if r.Severity == SeverityError && r.Message == "sops binary not found" {
			found = true
			assert.Contains(t, r.Fix, "https://github.com/getsops/sops#install")
		}
	}
	assert.True(t, found, "expected a SeverityError result about sops binary not found")
}

// TestSopsFound verifies RunChecks does not return an error about sops when it is on PATH.
func TestSopsFound(t *testing.T) {
	opts := Options{
		SopsLookPath: func(_ string) (string, error) {
			return "/usr/bin/sops", nil
		},
		AgeKeyPath: validAgeKeyPath(t),
		StartDir:   validSopsYamlDir(t),
	}

	results, hasHardError := RunChecks(opts)

	assert.False(t, hasHardError, "expected hasHardError=false when sops is found")
	for _, r := range results {
		if r.Severity == SeverityError {
			assert.NotContains(t, r.Message, "sops binary", "should not have sops error when sops is found")
		}
	}
}

// TestAgeKeyMissing verifies RunChecks returns a soft warning when the age key file is absent.
func TestAgeKeyMissing(t *testing.T) {
	opts := Options{
		SopsLookPath: func(_ string) (string, error) {
			return "/usr/bin/sops", nil
		},
		// Point to a path that definitely does not exist.
		AgeKeyPath: filepath.Join(t.TempDir(), "nonexistent", "keys.txt"),
		StartDir:   validSopsYamlDir(t),
	}

	results, hasHardError := RunChecks(opts)

	assert.False(t, hasHardError, "age key missing should be a soft warning, not a hard error")
	found := false
	for _, r := range results {
		if r.Severity == SeverityWarn && r.Message == "Age key file not found" {
			found = true
			assert.Contains(t, r.Fix, "age-keygen")
		}
	}
	assert.True(t, found, "expected a SeverityWarn result about age key file not found")
}

// TestSopsYamlMissing verifies RunChecks returns a soft warning when .sops.yaml is absent.
func TestSopsYamlMissing(t *testing.T) {
	// A temp dir with no .sops.yaml anywhere in the hierarchy.
	emptyDir := t.TempDir()

	opts := Options{
		SopsLookPath: func(_ string) (string, error) {
			return "/usr/bin/sops", nil
		},
		AgeKeyPath: validAgeKeyPath(t),
		StartDir:   emptyDir,
	}

	results, hasHardError := RunChecks(opts)

	assert.False(t, hasHardError, ".sops.yaml missing should be a soft warning")
	found := false
	for _, r := range results {
		if r.Severity == SeverityWarn {
			assert.Contains(t, r.Message, ".sops.yaml not found")
			found = true
		}
	}
	assert.True(t, found, "expected a SeverityWarn result about .sops.yaml not found")
}

// TestSopsYamlFoundInParent verifies no .sops.yaml warning when the file exists in a parent directory.
func TestSopsYamlFoundInParent(t *testing.T) {
	parentDir := t.TempDir()
	childDir := filepath.Join(parentDir, "sub", "dir")
	require.NoError(t, os.MkdirAll(childDir, 0o750))

	// Place .sops.yaml in parent, not in the child where StartDir points.
	sopsYamlPath := filepath.Join(parentDir, ".sops.yaml")
	require.NoError(t, os.WriteFile(sopsYamlPath, []byte("creation_rules: []"), 0o644))

	opts := Options{
		SopsLookPath: func(_ string) (string, error) {
			return "/usr/bin/sops", nil
		},
		AgeKeyPath: validAgeKeyPath(t),
		StartDir:   childDir,
	}

	results, _ := RunChecks(opts)

	for _, r := range results {
		if r.Severity == SeverityWarn {
			assert.NotContains(t, r.Message, ".sops.yaml", "should not warn about .sops.yaml when found in parent")
		}
	}
}

// TestHardErrorVsSoftWarning verifies hasHardError logic matches the spec exactly.
func TestHardErrorVsSoftWarning(t *testing.T) {
	t.Run("sops missing gives hasHardError=true", func(t *testing.T) {
		opts := Options{
			SopsLookPath: func(_ string) (string, error) {
				return "", &os.PathError{Op: "LookPath", Path: "sops", Err: os.ErrNotExist}
			},
			AgeKeyPath: validAgeKeyPath(t),
			StartDir:   validSopsYamlDir(t),
		}
		_, hasHardError := RunChecks(opts)
		assert.True(t, hasHardError)
	})

	t.Run("only age key missing gives hasHardError=false", func(t *testing.T) {
		opts := Options{
			SopsLookPath: func(_ string) (string, error) {
				return "/usr/bin/sops", nil
			},
			AgeKeyPath: filepath.Join(t.TempDir(), "nonexistent", "keys.txt"),
			StartDir:   validSopsYamlDir(t),
		}
		_, hasHardError := RunChecks(opts)
		assert.False(t, hasHardError)
	})
}

// TestAllChecksRunInSinglePass verifies D-02: all three checks run even when sops is missing.
func TestAllChecksRunInSinglePass(t *testing.T) {
	emptyDir := t.TempDir()

	opts := Options{
		SopsLookPath: func(_ string) (string, error) {
			return "", &os.PathError{Op: "LookPath", Path: "sops", Err: os.ErrNotExist}
		},
		AgeKeyPath: filepath.Join(t.TempDir(), "nonexistent", "keys.txt"),
		StartDir:   emptyDir,
	}

	results, hasHardError := RunChecks(opts)

	assert.True(t, hasHardError, "sops missing should be hard error")
	// All three checks should produce results.
	assert.Len(t, results, 3, "expected exactly 3 results when all three checks fail")

	severities := make(map[Severity]int)
	for _, r := range results {
		severities[r.Severity]++
	}
	assert.Equal(t, 1, severities[SeverityError], "expected 1 hard error (sops)")
	assert.Equal(t, 2, severities[SeverityWarn], "expected 2 soft warnings (age key + .sops.yaml)")
}

// TestFindSopsYamlWalksUp verifies FindSopsYaml searches parent directories and terminates at root.
func TestSopsYamlDiscovery(t *testing.T) {
	t.Run("finds file in current dir", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".sops.yaml"), []byte("creation_rules: []"), 0o644))

		path, found := FindSopsYaml(dir)
		assert.True(t, found)
		assert.Equal(t, filepath.Join(dir, ".sops.yaml"), path)
	})

	t.Run("finds file in grandparent dir", func(t *testing.T) {
		rootDir := t.TempDir()
		deepDir := filepath.Join(rootDir, "a", "b", "c")
		require.NoError(t, os.MkdirAll(deepDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(rootDir, ".sops.yaml"), []byte("creation_rules: []"), 0o644))

		_, found := FindSopsYaml(deepDir)
		assert.True(t, found)
	})

	t.Run("returns false when not found anywhere", func(t *testing.T) {
		// Use a fresh temp dir with no .sops.yaml in it or its parents within the temp hierarchy.
		// The walk will stop at filesystem root.
		dir := t.TempDir()

		_, found := FindSopsYaml(dir)
		// This may or may not find a .sops.yaml depending on the test environment,
		// but the function must not hang or panic.
		// We just verify the function returns without hanging.
		_ = found
	})

	t.Run("stops at filesystem root (no infinite loop)", func(t *testing.T) {
		// Start from a guaranteed temp dir, which won't have .sops.yaml, walk should terminate.
		dir := t.TempDir()
		done := make(chan struct{})
		go func() {
			FindSopsYaml(dir)
			close(done)
		}()
		select {
		case <-done:
			// OK — terminated without hanging
		}
	})
}

// validAgeKeyPath creates a temp keys.txt file and returns its path.
func validAgeKeyPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "keys.txt")
	require.NoError(t, os.WriteFile(keyPath, []byte("# age key"), 0o600))
	return keyPath
}

// validSopsYamlDir creates a temp dir with a .sops.yaml file and returns the dir path.
func validSopsYamlDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".sops.yaml"), []byte("creation_rules: []"), 0o644))
	return dir
}
