package file

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/google/renameio/v2"

	"github.com/aretext/aretext/editor/text"
)

const defaultPermForNewFile fs.FileMode = 0644

// Save writes the text to disk and starts a new watcher to detect subsequent changes.
// This adds the POSIX end-of-file indicator (line feed at the end of the file).
func Save(path string, tree *text.Tree, watcherPollInterval time.Duration) (*Watcher, error) {
	// Compose a reader that calculates the checksum and appends the POSIX EOF indicator.
	checksummer := NewChecksummer()
	textReader := tree.ReaderAtPosition(0)
	posixEofReader := strings.NewReader("\n")
	r := io.TeeReader(io.MultiReader(&textReader, posixEofReader), checksummer)

	// Check if the path is a hardlink. If so, we need to save directly to this path
	// (not tmpfile / rename) to avoid changing the inode.
	isHardLink, err := checkIfPathIsHardLink(path)
	if err != nil {
		return nil, err
	}

	// Save the file.
	if isHardLink {
		err = saveDirectly(path, r)
	} else {
		err = saveWithTmpFileRename(path, r)
	}

	if err != nil {
		return nil, err
	}

	// Start a new watcher for subsequent changes to the file.
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("os.Stat: %w", err)
	}
	watcher := NewWatcherForExistingFile(watcherPollInterval, path, fileInfo.ModTime(), fileInfo.Size(), checksummer.Checksum())

	return watcher, nil
}

func saveDirectly(path string, r io.Reader) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultPermForNewFile)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}
	defer f.Close()

	// Write to the file.
	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	// Sync the file to disk so the watcher calculates the checksum correctly later.
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("file.Sync: %w", err)
	}

	return nil
}

func saveWithTmpFileRename(path string, r io.Reader) error {
	// If the path is a symlink, this will return the symlink target so we save
	// over the target file instead of overwriting the symlink itself.
	targetPath, err := targetPathForSave(path)
	if err != nil {
		return err
	}
	log.Printf("Saving file at target path %s", targetPath)

	// Use renameio to write the file to a temporary directory, then rename it to the target file.
	// This should reduce the risk of data corruption if the editor crashes mid-write,
	// but is probably not 100% reliable (see http://danluu.com/deconstruct-files/).
	// There is a good discussion of the Go libraries solving this problem in
	// this GitHub issue comment: https://github.com/golang/go/issues/22397#issuecomment-380831736
	pf, err := renameio.NewPendingFile(targetPath, renameio.WithPermissions(defaultPermForNewFile), renameio.WithExistingPermissions())
	if err != nil {
		return fmt.Errorf("renamio.TempFile: %w", err)
	}
	defer pf.Cleanup()

	// Write to the file.
	_, err = io.Copy(pf, r)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	// Sync the file to disk so the watcher calculates the checksum correctly later.
	err = pf.CloseAtomicallyReplace()
	if err != nil {
		return fmt.Errorf("renamio.CloseAtomicallyReplace: %w", err)
	}

	return nil
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

func checkIfPathIsHardLink(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil // new file
	} else if err != nil {
		return false, fmt.Errorf("os.Stat: %w", err)
	}

	if sys := fileInfo.Sys(); sys != nil {
		if stat, ok := sys.(*syscall.Stat_t); ok {
			return stat.Nlink > 1, nil
		}
	}

	return false, nil
}
