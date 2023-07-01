package file

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aretext/aretext/text"
)

// Save writes the text to disk and starts a new watcher to detect subsequent changes.
// This adds the POSIX end-of-file indicator (line feed at the end of the file).
func Save(path string, tree *text.Tree, watcherPollInterval time.Duration) (*Watcher, error) {
	// Compose a reader that calculates the checksum and appends the POSIX EOF indicator.
	checksummer := NewChecksummer()
	textReader := tree.ReaderAtPosition(0)
	posixEofReader := strings.NewReader("\n")
	r := io.TeeReader(io.MultiReader(&textReader, posixEofReader), checksummer)

	// Save the contents of the reader to a file.
	// On Linux, this uses renamio. On Windows, this uses the Go stdlib.
	// This implicitly calculates the checksum.
	err := platformSpecificSave(path, r)
	if err != nil {
		return nil, err
	}

	// Start a new watcher for subsequent changes to the file.
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("os.Stat: %w", err)
	}
	checksum := checksummer.Checksum()
	watcher := NewWatcher(watcherPollInterval, path, fileInfo.ModTime(), fileInfo.Size(), checksum)

	return watcher, nil
}
