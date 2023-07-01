package file

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
)

func TestSaveNewFile(t *testing.T) {
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "test.txt")
	saveAndAssertContents(t, path, "abcd1234", expectedPermForPlatform(0644))
}

func TestSaveModifyExistingFile(t *testing.T) {
	path := createTestFile(t, "old contents")
	saveAndAssertContents(t, path, "new contents", expectedPermForPlatform(0644))
}

func TestSaveModifyExistingFilePreservePermissions(t *testing.T) {
	path := createTestFile(t, "old contents")

	err := os.Chmod(path, 0600) // On Windows this is a no-op.
	require.NoError(t, err)
	saveAndAssertContents(t, path, "new contents", expectedPermForPlatform(0600))
}

func saveAndAssertContents(t *testing.T, path string, contents string, perms os.FileMode) {
	tree, err := text.NewTreeFromString(contents)
	require.NoError(t, err)

	watcher, err := Save(path, tree, testWatcherPollInterval)
	require.NoError(t, err)
	assert.Equal(t, path, watcher.Path())
	defer watcher.Stop()

	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	expectedContents := contents + "\n" // Append POSIX EOF
	assert.Equal(t, expectedContents, string(fileBytes))

	fileInfo, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, perms, fileInfo.Mode().Perm())
}

func expectedPermForPlatform(linuxExpectedPerm fs.FileMode) fs.FileMode {
	if runtime.GOOS == "windows" {
		// Windows supports only read-write and read-only, so assume we
		// created the file read-write (if we have permission to save it at all).
		return fs.FileMode(0666)
	}

	return linuxExpectedPerm
}
