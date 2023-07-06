package file

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListDirFiles(t *testing.T) {
	paths := []string{
		"a/1.txt",
		"a/2.txt",
		"b/2.txt",
		"a/b/4.txt",
		"a/.hidden/1.txt",
		"a/.hidden/2.txt",
	}
	withTmpDirPaths(t, paths, func(tmpDir string) {
		// List all file paths in tmpdir.
		ctx := context.Background()
		options := ListDirOptions{
			DirPatternsToHide: []string{filepath.FromSlash("**/.hidden")},
		}
		foundPaths := ListDir(ctx, tmpDir, options)
		relPaths := makeRelPaths(foundPaths)

		// Check that we found all the paths we created (except hidden ones), ignoring order.
		assertPathsIgnoreOrder(t, relPaths, []string{
			filepath.FromSlash("a/1.txt"),
			filepath.FromSlash("a/2.txt"),
			filepath.FromSlash("b/2.txt"),
			filepath.FromSlash("a/b/4.txt"),
		})
	})
}

func TestListDirDirectoriesOnly(t *testing.T) {
	paths := []string{
		filepath.FromSlash("a/foo.txt"),
		filepath.FromSlash("a/b/bar.txt"),
		filepath.FromSlash("a/.hidden/hidden.txt"),
		filepath.FromSlash("c/bat.txt"),
	}
	withTmpDirPaths(t, paths, func(tmpDir string) {
		// List all subdir paths in tmpdir.
		ctx := context.Background()
		options := ListDirOptions{
			DirectoriesOnly:   true,
			DirPatternsToHide: []string{filepath.FromSlash("**/.hidden")},
		}
		foundPaths := ListDir(ctx, tmpDir, options)
		relPaths := makeRelPaths(foundPaths)

		// Check that we found all the subdir paths we created (except hidden ones), ignoring order.
		assertPathsIgnoreOrder(t, relPaths, []string{
			filepath.FromSlash("a"),
			filepath.FromSlash("a/b"),
			filepath.FromSlash("c"),
		})
	})
}

func withTmpDirPaths(t *testing.T, paths []string, f func(string)) {
	// Create a tmpdir (and delete at end of test).
	tmpDir, err := filepath.Abs("./tmp")
	require.NoError(t, err)
	err = os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set the cwd to the tmpdir (and reset at end of test).
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create subdirectories and files in the tmpdir.
	for _, p := range paths {
		fullPath := filepath.Join(tmpDir, p)
		err = os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte("test"), 0644)
		require.NoError(t, err)
	}

	// Run the test
	f(tmpDir)
}

func makeRelPaths(paths []string) []string {
	relPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		relPaths = append(relPaths, RelativePathCwd(p))
	}
	return relPaths
}

func assertPathsIgnoreOrder(t *testing.T, actual []string, expected []string) {
	sort.Strings(expected)
	sort.Strings(actual)
	assert.Equal(t, expected, actual)
}
