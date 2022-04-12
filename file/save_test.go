package file

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
)

func TestSaveNewFile(t *testing.T) {
	tmpDir := t.TempDir()

	path := path.Join(tmpDir, "test.txt")
	saveAndAssertContents(t, path, "abcd1234", 0644)
}

func TestSaveModifyExistingFile(t *testing.T) {
	path := createTestFile(t, "old contents")
	saveAndAssertContents(t, path, "new contents", 0644)
}

func TestSaveModifyExistingFilePreservePermissions(t *testing.T) {
	path := createTestFile(t, "old contents")

	err := os.Chmod(path, 0600)
	require.NoError(t, err)
	saveAndAssertContents(t, path, "new contents", 0600)
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
	assert.Equal(t, fileInfo.Mode().Perm(), perms)
}
