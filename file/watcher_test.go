package file

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWatcherPollInterval time.Duration = time.Millisecond * 50

func createTestFile(t *testing.T, s string) (string, func()) {
	tmpDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	filePath := path.Join(tmpDir, "test.txt")
	f, err := os.Create(filePath)
	require.NoError(t, err)
	defer f.Close()

	_, err = io.WriteString(f, s)
	require.NoError(t, err)

	cleanupFunc := func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}

	return filePath, cleanupFunc
}

func appendToTestFile(t *testing.T, path string, s string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	defer f.Close()

	_, err = io.WriteString(f, s)
	require.NoError(t, err)
}

func expectNoChange(t *testing.T, watcher *Watcher) {
	select {
	case <-watcher.ChangedChan():
		assert.Fail(t, "Unexpected change reported")
	default:
		assert.False(t, watcher.ChangedFlag())
		return
	}
}

func waitForChangeOrTimeout(t *testing.T, watcher *Watcher) {
	select {
	case <-watcher.ChangedChan():
		assert.True(t, watcher.ChangedFlag())
		return
	case <-time.After(testWatcherPollInterval * 10):
		assert.Fail(t, "Timed out waiting for change")
	}
}

func TestWatcher(t *testing.T) {
	// Create a test file in a temporary directory.
	filePath, cleanup := createTestFile(t, "abcd")
	defer cleanup()

	// Load the file and start a watcher.
	_, watcher, err := Load(filePath, testWatcherPollInterval)
	require.NoError(t, err)
	defer watcher.Stop()

	// Initially, there should be no changes.
	expectNoChange(t, watcher)

	// Modify the file.
	appendToTestFile(t, filePath, "xyz")

	// Wait for changes to be detected (or time out and fail the test).
	waitForChangeOrTimeout(t, watcher)
}
