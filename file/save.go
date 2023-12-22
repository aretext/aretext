package file

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/renameio/v2"

	"github.com/aretext/aretext/text"
)

// Save writes the text to disk and starts a new watcher to detect subsequent changes.
// This adds the POSIX end-of-file indicator (line feed at the end of the file).
func Save(path string, tree *text.Tree, watcherPollInterval time.Duration) (*Watcher, error) {
	// Use renameio to write the file to a temporary directory, then rename it to the target file.
	// This should reduce the risk of data corruption if the editor crashes mid-write,
	// but is probably not 100% reliable (see http://danluu.com/deconstruct-files/).
	// There is a good discussion of the Go libraries solving this problem in
	// this GitHub issue comment: https://github.com/golang/go/issues/22397#issuecomment-380831736
	pf, err := renameio.NewPendingFile(path, renameio.WithPermissions(0644), renameio.WithExistingPermissions())
	if err != nil {
		return nil, fmt.Errorf("renamio.TempFile: %w", err)
	}
	defer pf.Cleanup()

	// Compose a reader that calculates the checksum and appends the POSIX EOF indicator.
	checksummer := NewChecksummer()
	textReader := tree.ReaderAtPosition(0)
	posixEofReader := strings.NewReader("\n")
	r := io.TeeReader(io.MultiReader(&textReader, posixEofReader), checksummer)

	// Write to the file and calculate the checksum.
	_, err = io.Copy(pf, r)
	if err != nil {
		return nil, fmt.Errorf("io.Copy: %w", err)
	}

	// Sync the file to disk so the watcher calculates the checksum correctly later.
	err = pf.CloseAtomicallyReplace()
	if err != nil {
		return nil, fmt.Errorf("renamio.CloseAtomicallyReplace: %w", err)
	}

	// Start a new watcher for subsequent changes to the file.
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("os.Stat: %w", err)
	}
	checksum := checksummer.Checksum()
	watcher := NewWatcherForExistingFile(watcherPollInterval, path, fileInfo.ModTime(), fileInfo.Size(), checksum)

	return watcher, nil
}
