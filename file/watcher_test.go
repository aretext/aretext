package file

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWatcherPollInterval time.Duration = time.Millisecond * 50

func TestWatcherNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Start a watcher for a new (non-existent) file.
	watcher := NewWatcherForNewFile(testWatcherPollInterval, filePath)
	defer watcher.Stop()

	// Initially there should be no changes.
	select {
	case <-watcher.ChangedChan():
		assert.Fail(t, "Unexpected change reported")
	default:
		movedOrDeleted, err := watcher.CheckFileMovedOrDeleted()
		require.NoError(t, err)
		assert.False(t, movedOrDeleted)

		changed, err := watcher.CheckFileContentsChanged()
		assert.True(t, errors.Is(err, fs.ErrNotExist))
		assert.False(t, changed)
	}

	// Create a file at the path.
	f, err := os.Create(filePath)
	require.NoError(t, err)
	defer f.Close()

	// Wait for changes to be detected (or time out and fail the test).
	select {
	case <-watcher.ChangedChan():
		changed, err := watcher.CheckFileContentsChanged()
		assert.NoError(t, err)
		assert.True(t, changed)
	case <-time.After(testWatcherPollInterval * 10):
		assert.Fail(t, "Timed out waiting for change")
	}

	// Verify that the file was NOT moved or deleted.
	movedOrDeleted, err := watcher.CheckFileMovedOrDeleted()
	require.NoError(t, err)
	assert.False(t, movedOrDeleted)
}

func TestWatcherFromLoadExistingFile(t *testing.T) {
	// Create a test file in a temporary directory.
	filePath := createTestFile(t, "abcd")

	// Load the file and start a watcher.
	_, watcher, err := Load(filePath, testWatcherPollInterval)
	require.NoError(t, err)
	defer watcher.Stop()

	// Initially, there should be no changes.
	select {
	case <-watcher.ChangedChan():
		assert.Fail(t, "Unexpected change reported")
	default:
		changed, err := watcher.CheckFileContentsChanged()
		assert.NoError(t, err)
		assert.False(t, changed)
	}

	// Modify the file.
	appendToTestFile(t, filePath, "xyz")

	// Wait for changes to be detected (or time out and fail the test).
	select {
	case <-watcher.ChangedChan():
		changed, err := watcher.CheckFileContentsChanged()
		assert.NoError(t, err)
		assert.True(t, changed)
	case <-time.After(testWatcherPollInterval * 10):
		assert.Fail(t, "Timed out waiting for change")
	}

	// Verify that the file was NOT moved or deleted.
	movedOrDeleted, err := watcher.CheckFileMovedOrDeleted()
	require.NoError(t, err)
	assert.False(t, movedOrDeleted)

	// Delete the file.
	err = os.Remove(filePath)
	require.NoError(t, err)

	// Should detect that the file was moved or deleted.
	movedOrDeleted, err = watcher.CheckFileMovedOrDeleted()
	require.NoError(t, err)
	assert.True(t, movedOrDeleted)
}
