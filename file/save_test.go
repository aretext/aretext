package file

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
)

func TestSaveNewFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}()

	path := path.Join(tmpDir, "test.txt")
	saveAndAssertContents(t, path, "abcd1234")
}

func TestSaveModifyExistingFile(t *testing.T) {
	path, cleanup := createTestFile(t, "old contents")
	defer cleanup()
	saveAndAssertContents(t, path, "new contents")
}

func saveAndAssertContents(t *testing.T, path string, contents string) {
	tree, err := text.NewTreeFromString(contents)
	require.NoError(t, err)

	watcher, err := Save(path, tree, testWatcherPollInterval)
	require.NoError(t, err)
	assert.Equal(t, path, watcher.Path())
	defer watcher.Stop()

	fileBytes, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	expectedContents := contents + "\n" // Append POSIX EOF
	assert.Equal(t, expectedContents, string(fileBytes))
}
