package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aretext/aretext/text"
)

type LoadedFile struct {
	TextTree    *text.Tree
	FileWatcher *Watcher
	PosixEof    bool // Whether loaded document had POSIX EOF newline.
}

// Load reads a file from disk and starts a watcher to detect changes.
// This will remove the POSIX end-of-file indicator (line feed at end of file).
func Load(path string, watcherPollInterval time.Duration) (LoadedFile, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return LoadedFile{}, fmt.Errorf("filepath.Abs: %w", err)
	}

	f, err := os.Open(path)
	if err != nil {
		return LoadedFile{}, fmt.Errorf("os.Open: %w", err)
	}
	defer f.Close()

	lastModifiedTime, size, err := lastModifiedTimeAndSize(f)
	if err != nil {
		return LoadedFile{}, fmt.Errorf("lastModifiedTime: %w", err)
	}

	tree, checksum, err := readContentsAndChecksum(f)
	if err != nil {
		return LoadedFile{}, fmt.Errorf("readContentsAndChecksum: %w", err)
	}

	// POSIX files end with a single line feed to indicate the end of the file.
	// We remove it from the tree to simplify editor operations; we'll add it back when saving the file.
	hasPosixEof := endsWithLineFeed(tree)
	if hasPosixEof {
		lastPos := tree.NumChars() - 1
		tree.DeleteAtPosition(lastPos)
	}

	watcher := NewWatcherForExistingFile(watcherPollInterval, path, lastModifiedTime, size, checksum)

	return LoadedFile{
		TextTree:    tree,
		FileWatcher: watcher,
		PosixEof:    hasPosixEof,
	}, nil
}

func readContentsAndChecksum(f *os.File) (*text.Tree, string, error) {
	checksummer := NewChecksummer()
	r := io.TeeReader(f, checksummer)
	tree, err := text.NewTreeFromReader(r)
	if err != nil {
		return nil, "", fmt.Errorf("text.NewTreeFromReader: %w", err)
	}
	return tree, checksummer.Checksum(), nil
}

func lastModifiedTimeAndSize(f *os.File) (time.Time, int64, error) {
	fileInfo, err := f.Stat()
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("f.Stat: %w", err)
	}

	return fileInfo.ModTime(), fileInfo.Size(), nil
}

func endsWithLineFeed(tree *text.Tree) bool {
	reader := tree.ReverseReaderAtPosition(tree.NumChars())
	var buf [1]byte
	if n, err := reader.Read(buf[:]); err != nil || n == 0 {
		return false
	}
	return buf[0] == '\n'
}
