package file

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/editor/text"
)

func TestSaveNewFile(t *testing.T) {
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "test.txt")
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

func TestSavePathToSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "test.txt")
	symlinkPath := filepath.Join(tmpDir, "testsymlink")

	// Create the target file.
	f, err := os.Create(targetPath)
	require.NoError(t, err)
	defer f.Close()
	_, err = io.WriteString(f, "test")
	require.NoError(t, err)

	// Create symlink to the target file.
	err = os.Symlink(targetPath, symlinkPath)
	require.NoError(t, err)

	// Save to the symlink path.
	saveAndAssertContents(t, symlinkPath, "new contents", 0644)

	// Verify that the symlink is still a symlink.
	fileInfo, err := os.Lstat(symlinkPath)
	require.NoError(t, err)
	assert.True(t, fileInfo.Mode()&os.ModeSymlink != 0)

	// Verify that the target file was modified.
	fileBytes, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, "new contents\n", string(fileBytes))
}

func TestSavePathToHardLink(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "test.txt")
	hardlinkPath := filepath.Join(tmpDir, "testhardlink")

	// Create the target file.
	f, err := os.Create(targetPath)
	require.NoError(t, err)
	defer f.Close()
	_, err = io.WriteString(f, "test")
	require.NoError(t, err)

	// Create hardlink to the target file.
	err = os.Link(targetPath, hardlinkPath)
	require.NoError(t, err)

	// Save to the hardlink path.
	saveAndAssertContents(t, hardlinkPath, "new contents", 0644)

	// Verify that the target file was modified.
	fileBytes, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, "new contents\n", string(fileBytes))
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
