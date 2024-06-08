package file

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/renameio/v2"

	"github.com/aretext/aretext/text"
)

// Save writes the text to disk and starts a new watcher to detect subsequent changes.
// This adds the POSIX end-of-file indicator (line feed at the end of the file).
func Save(path string, tree *text.Tree, appendPosixEof bool, watcherPollInterval time.Duration) (*Watcher, error) {
	// If the path is a symlink, this will return the symlink target so we save
	// over the target file instead of overwriting the symlink itself.
	targetPath, err := targetPathForSave(path)
	if err != nil {
		return nil, err
	}
	log.Printf("Saving file at target path %s", targetPath)

	// Use renameio to write the file to a temporary directory, then rename it to the target file.
	// This should reduce the risk of data corruption if the editor crashes mid-write,
	// but is probably not 100% reliable (see http://danluu.com/deconstruct-files/).
	// There is a good discussion of the Go libraries solving this problem in
	// this GitHub issue comment: https://github.com/golang/go/issues/22397#issuecomment-380831736
	pf, err := renameio.NewPendingFile(targetPath, renameio.WithPermissions(0644), renameio.WithExistingPermissions())
	if err != nil {
		return nil, fmt.Errorf("renamio.TempFile: %w", err)
	}
	defer pf.Cleanup()

	// Compose a reader that calculates the checksum and appends the POSIX EOF indicator.
	checksummer := NewChecksummer()
	textReader := tree.ReaderAtPosition(0)
	r := io.Reader(&textReader)
	if appendPosixEof {
		posixEofReader := strings.NewReader("\n")
		r = io.MultiReader(r, posixEofReader)
	}
	r = io.TeeReader(r, checksummer)

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

func targetPathForSave(path string) (string, error) {
	fileInfo, err := os.Lstat(path)
	if os.IsNotExist(err) {
		// New file, return the original path.
		return path, nil
	} else if err != nil {
		return "", fmt.Errorf("os.Lstat: %w", err)
	}

	if fileInfo.Mode()&os.ModeSymlink == 0 {
		// Not a symlink, so return the original path.
		return path, nil
	} else {
		// Symlink, so lookup the target.
		target, err := os.Readlink(path)
		if err != nil {
			return "", fmt.Errorf("os.Readlink: %w", err)
		}
		log.Printf("Resolved symlink target %s -> %s", path, target)
		return target, nil
	}
}
