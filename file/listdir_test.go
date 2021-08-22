package file

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListDir(t *testing.T) {
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
	paths := []string{
		"a/1.txt",
		"a/2.txt",
		"b/2.txt",
		"a/b/4.txt",
		"a/.hidden/1.txt",
		"a/.hidden/2.txt",
	}

	for _, p := range paths {
		fullPath := filepath.Join(tmpDir, p)
		err = os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte("test"), 0644)
		require.NoError(t, err)
	}

	// List all paths in tmpdir.
	foundPaths := ListDir(tmpDir, []string{"**/.hidden"})
	relPaths := make([]string, 0, len(foundPaths))
	for _, p := range foundPaths {
		relPaths = append(relPaths, RelativePathCwd(p))
	}

	// Check that we found all the paths we created (except hidden ones), ignoring order.
	expectedPaths := []string{
		"a/1.txt",
		"a/2.txt",
		"b/2.txt",
		"a/b/4.txt",
	}
	sort.Strings(expectedPaths)
	sort.Strings(relPaths)
	assert.Equal(t, expectedPaths, relPaths)
}
