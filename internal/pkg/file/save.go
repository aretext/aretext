package file

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/text"
)

// Save writes the text to disk and starts a new watcher to detect subsequent changes.
// This adds the POSIX end-of-file indicator (line feed at the end of the file).
// We try to perform the write atomically by writing to a temp file, syncing to disk, then
// renaming to the target file path.  This should work under most, but not necessarily all, circumstances.
func Save(path string, tree *text.Tree, watcherPollInterval time.Duration) (Watcher, error) {
	// Create a temporary file to write the text.
	// This ensures that we don't corrupt the original file if an error occurs during the write.
	tmpFile, err := ioutil.TempFile(os.TempDir(), "aretext-")
	if err != nil {
		return nil, errors.Wrapf(err, "ioutil.TempFile")
	}
	defer tmpFile.Close()

	// Compose a reader that calculates the checksum and appends the POSIX EOF indicator.
	checksummer := NewChecksummer()
	textReader := tree.ReaderAtPosition(0, text.ReadDirectionForward)
	posixEofReader := strings.NewReader("\n")
	r := io.TeeReader(io.MultiReader(textReader, posixEofReader), checksummer)

	// Write to the temporary file and calculate the checksum.
	_, err = io.Copy(tmpFile, r)
	if err != nil {
		return nil, errors.Wrapf(err, "io.Copy")
	}

	// Flush the temporary file to disk to (in most cases) ensure that it's persisted.
	err = tmpFile.Sync()
	if err != nil {
		return nil, errors.Wrapf(err, "File.Sync")
	}

	// Move the temporary file to the target path.
	err = os.Rename(tmpFile.Name(), path)
	if err != nil {
		return nil, errors.Wrapf(err, "os.Rename")
	}

	// Retrieve the last modified time and size for the file.
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrapf(err, "os.Stat")
	}

	// Start a new watcher for subsequent changes to the file.
	checksum := checksummer.Checksum()
	lastModifiedTime := fileInfo.ModTime()
	size := fileInfo.Size()
	watcher := newFileWatcher(watcherPollInterval, path, lastModifiedTime, size, checksum)

	return watcher, nil
}
