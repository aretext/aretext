package file

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"sync"
	"time"
)

const DefaultPollInterval = time.Second

// Watcher checks if a file's contents have changed.
type Watcher struct {
	// These fields are immutable, so they can be read safely from any goroutine.
	path      string
	isNewFile bool
	size      int64
	checksum  string

	// After the watcher is constructed, this field is read and written
	// only by the watcher goroutine.
	lastModified time.Time

	changedChan chan struct{}
	quitChan    chan struct{}
	stopOnce    sync.Once
}

// NewWatcherForNewFile returns a watcher for a file that does not yet exist on disk.
func NewWatcherForNewFile(pollInterval time.Duration, path string) *Watcher {
	w := &Watcher{
		path:        path,
		isNewFile:   true,
		changedChan: make(chan struct{}),
		quitChan:    make(chan struct{}),
	}
	go w.checkFileLoop(pollInterval)
	return w
}

// NewWatcherForExistingFile returns a watcher for a file that exists on disk.
// lastModified is the time the file was last modified, as reported when the file was loaded.
// size is the size in bytes of the file when it was loaded.
// checksum is an MD5 hash of the file's contents when it was loaded.
func NewWatcherForExistingFile(pollInterval time.Duration, path string, lastModified time.Time, size int64, checksum string) *Watcher {
	w := &Watcher{
		path:         path,
		size:         size,
		lastModified: lastModified,
		checksum:     checksum,
		changedChan:  make(chan struct{}),
		quitChan:     make(chan struct{}),
	}
	go w.checkFileLoop(pollInterval)
	return w
}

// NewEmptyWatcher returns a watcher that has an empty path and never triggers.
func NewEmptyWatcher() *Watcher {
	return &Watcher{changedChan: make(chan struct{})}
}

// Path returns the path to the file being watched.
func (w *Watcher) Path() string {
	return w.path
}

// Stop stops the watcher from checking for changes.
func (w *Watcher) Stop() {
	w.stopOnce.Do(func() {
		if w.quitChan != nil {
			log.Printf("Stopping file watcher for %s...\n", w.path)
			close(w.quitChan)
		}
	})
}

// CheckFileMovedOrDeleted checks whether the file used to exist
// at the path but has since been moved or deleted.
func (w *Watcher) CheckFileMovedOrDeleted() (bool, error) {
	if w.isNewFile {
		// File has not been created yet, so it can't have been moved or deleted.
		return false, nil
	}

	_, err := os.Stat(w.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// File used to exist, but no longer exists at the path,
			// so it must have been moved or deleted.
			return true, nil
		}

		return false, fmt.Errorf("os.Stat: %w", err)
	}

	// File still exists at the path.
	return false, nil
}

// CheckFileContentsChanged checks whether the file's checksum has changed.
// If the file no longer exists, this will return an error.
func (w *Watcher) CheckFileContentsChanged() (bool, error) {
	checksum, err := w.calculateChecksum()
	if err != nil {
		return false, err
	}
	changed := checksum != w.checksum
	return changed, nil
}

// ChangedChan returns a channel that receives a message when the file's contents change.
// This can produce false negatives if an error occurs accessing the file (for example, if file permissions changed).
// The channel will receive at most one message.
// This method is thread-safe.
func (w *Watcher) ChangedChan() chan struct{} {
	return w.changedChan
}

func (w *Watcher) checkFileLoop(pollInterval time.Duration) {
	log.Printf("Started file watcher for %s\n", w.path)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if w.checkFileChanged() {
				log.Printf("File change detected in %s\n", w.path)
				w.changedChan <- struct{}{}
				return
			}
		case <-w.quitChan:
			log.Printf("Quit channel closed, exiting check file loop for %s\n", w.path)
			return
		}
	}
}

func (w *Watcher) checkFileChanged() bool {
	fileInfo, err := os.Stat(w.path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			log.Printf("Error retrieving file info: %v\n", err)
		}
		return false
	}

	// If neither mtime or size changed since the last check or file load, the contents probably haven't changed.
	// This check could produce a false negative if someone modifies the file immediately after loading it (within mtime granularity)
	// and replaces bytes without changing the size, but it's so much cheaper than calculating the md5 checksum that we do it anyway.
	// It is safe to read lastModified and size because no other goroutine mutates these.
	if w.lastModified.Equal(fileInfo.ModTime()) && w.size == fileInfo.Size() {
		return false
	}

	// It is possible for someone to update the file's last modified time without changing the contents.
	// If that happens, we want to avoid calculating the checksum on every poll, so update the watcher's lastModified time.
	// It is safe to modify lastModified because the check file loop goroutine is the only reader.
	w.lastModified = fileInfo.ModTime()

	checksum, err := w.calculateChecksum()
	if err != nil {
		log.Printf("Could not checksum file: %v\n", err)
		return false
	}

	return checksum != w.checksum
}

func (w *Watcher) calculateChecksum() (string, error) {
	f, err := os.Open(w.path)
	if err != nil {
		return "", fmt.Errorf("os.Open: %w", err)
	}
	defer f.Close()

	checksummer := NewChecksummer()
	if _, err := io.Copy(checksummer, f); err != nil {
		return "", fmt.Errorf("io.Copy: %w", err)
	}

	return checksummer.Checksum(), nil
}
