package testutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/caesarakalaeii/sops-tui/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequireGoldenStructure_WritesOnUpdateEnv verifies that GOLDEN_UPDATE=1
// creates the fixture file on disk.
func TestRequireGoldenStructure_WritesOnUpdateEnv(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	require.NoError(t, os.Chdir(tmp))

	t.Setenv("GOLDEN_UPDATE", "1")
	testutil.RequireGoldenStructure(t, "sample", "hello world\n")

	b, err := os.ReadFile(filepath.Join(tmp, "testdata", "sample.golden"))
	require.NoError(t, err, "GOLDEN_UPDATE=1 must write the fixture")
	assert.Equal(t, "hello world\n", string(b))
}

// TestRequireGoldenStructure_ComparesWhenUnset verifies that default mode
// compares against an existing fixture and passes on a match.
func TestRequireGoldenStructure_ComparesWhenUnset(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	require.NoError(t, os.Chdir(tmp))

	require.NoError(t, os.MkdirAll("testdata", 0o755))
	require.NoError(t, os.WriteFile(filepath.Join("testdata", "match.golden"),
		[]byte("expected output"), 0o644))

	t.Setenv("GOLDEN_UPDATE", "")
	testutil.RequireGoldenStructure(t, "match", "expected output")
}

// TestRequireGoldenStructure_ANSIStrip verifies ANSI sequences are removed
// before compare/write.
func TestRequireGoldenStructure_ANSIStrip(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	require.NoError(t, os.Chdir(tmp))

	// Red "error": \x1b[31merror\x1b[0m
	ansiInput := "\x1b[31merror\x1b[0m"

	t.Setenv("GOLDEN_UPDATE", "1")
	testutil.RequireGoldenStructure(t, "ansi", ansiInput)

	b, err := os.ReadFile(filepath.Join(tmp, "testdata", "ansi.golden"))
	require.NoError(t, err)
	assert.Equal(t, "error", string(b), "ANSI sequences must be stripped before write")
}

// TestRequireGoldenColors_Empty verifies nil and empty wantColors slices
// are no-ops that do not mark the test failed.
func TestRequireGoldenColors_Empty(t *testing.T) {
	testutil.RequireGoldenColors(t, "empty-nil", "\x1b[31mred\x1b[0m", nil)
	testutil.RequireGoldenColors(t, "empty-slice", "\x1b[31mred\x1b[0m", []string{})
	assert.False(t, t.Failed(), "empty wantColors must be a no-op")
}

// TestRequireGoldenColors_Missing verifies that a missing color sequence
// is detected by the underlying missingColors helper. The helper is
// exercised rather than RequireGoldenColors directly because *testing.T
// subtests propagate failure to the parent and cannot be "caught", so
// asserting on the missing-list is the idiomatic way to verify the
// detection logic without failing the outer test.
func TestRequireGoldenColors_Missing(t *testing.T) {
	missing := testutil.MissingColorsForTest("plain text", []string{"\x1b[99m", "present"})
	assert.Equal(t, []string{"\x1b[99m", "present"}, missing,
		"all colors absent from output must be reported as missing")

	missing = testutil.MissingColorsForTest("has \x1b[31mred\x1b[0m", []string{"\x1b[31m", "\x1b[99m"})
	assert.Equal(t, []string{"\x1b[99m"}, missing,
		"only absent colors must be reported")

	missing = testutil.MissingColorsForTest("fully present", []string{"fully", "present"})
	assert.Empty(t, missing, "all-present input must return empty missing list")
}

// TestNormalise verifies trailing-whitespace stripping and line-ending
// normalisation.
func TestNormalise(t *testing.T) {
	in := "hello   \r\nworld\t\t\n  indented  \n"
	want := "hello\nworld\n  indented\n"
	assert.Equal(t, want, testutil.NormaliseForTest(in))
}
