package file

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/aretext/aretext/text"
)

// Save writes the text to disk and starts a new watcher to detect subsequent changes.
// This adds the POSIX end-of-file indicator (line feed at the end of the file).
// This directly overwrites the target file and then syncs the file to disk.
func Save(path string, tree *text.Tree, watcherPollInterval time.Duration) (*Watcher, error) {
	// Open the target file, truncating it if it already exists.
	// There's a risk that the file might get corrupted if an error occurs while
	// we're writing the new contents, but it's very difficult to prevent this.
	// In particular, tricks like writing a temporary file and renaming it to the target path
	// don't work; see https://danluu.com/deconstruct-files/
	f, err := os.Create(path)
	if err != nil {
		return nil, errors.Wrapf(err, "os.Create")
	}
	defer f.Close()

	// Compose a reader that calculates the checksum and appends the POSIX EOF indicator.
	checksummer := NewChecksummer()
	textReader := tree.ReaderAtPosition(0, text.ReadDirectionForward)
	posixEofReader := strings.NewReader("\n")
	r := io.TeeReader(io.MultiReader(textReader, posixEofReader), checksummer)

	// Write to the file and calculate the checksum.
	_, err = io.Copy(f, r)
	if err != nil {
		return nil, errors.Wrapf(err, "io.Copy")
	}

	// Sync the file to disk so the watcher calculates the checksum correctly later.
	err = f.Sync()
	if err != nil {
		return nil, errors.Wrapf(err, "File.Sync")
	}

	// Start a new watcher for subsequent changes to the file.
	checksum := checksummer.Checksum()
	lastModifiedTime, size, err := lastModifiedTimeAndSize(f)
	if err != nil {
		return nil, err
	}
	watcher := NewWatcher(watcherPollInterval, path, lastModifiedTime, size, checksum)

	return watcher, nil
}
