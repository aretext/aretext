package file

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
)

const DefaultPollInterval = time.Second

// Watcher checks if a file's contents have changed.
type Watcher struct {
	path         string
	lastModified time.Time
	size         int64
	checksum     string
	changedChan  chan struct{}
	quitChan     chan struct{}
}

// NewWatcher returns a watcher for a file.
// lastModified is the time the file was last modified, as reported when the file was loaded.
// size is the size in bytes of the file when it was loaded.
// checksum is an MD5 hash of the file's contents when it was loaded.
func NewWatcher(pollInterval time.Duration, path string, lastModified time.Time, size int64, checksum string) *Watcher {
	w := &Watcher{
		path:         path,
		size:         size,
		lastModified: lastModified,
		checksum:     checksum,
		changedChan:  make(chan struct{}, 0),
		quitChan:     make(chan struct{}, 0),
	}
	go w.checkFileLoop(pollInterval)
	return w
}

// Path returns the path to the file being watched.
func (w *Watcher) Path() string {
	return w.path
}

// Stop stops the watcher from checking for changes.
func (w *Watcher) Stop() {
	if w.quitChan == nil {
		return
	}
	log.Printf("Stopping file watcher for %s...\n", w.path)
	close(w.quitChan)
	w.quitChan = nil
}

// ChangedChan returns a channel that is closed when the file's contents change.
// This can produce false negatives if an error occurs accessing the file (for example, if file permissions changed).
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
				close(w.changedChan)
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
		if !os.IsNotExist(err) {
			log.Printf("Error retrieving file info: %v\n", err)
		}
		return false
	}

	// It is safe to read lastModified and size because no other goroutine mutates these.
	// This check could produce a false negative if someone modifies the file immediately after loading it and doesn't change the size,
	// but it's so much cheaper than calculating the md5 checksum that we do it anyway.
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
		return "", errors.Wrapf(err, "os.Open()")
	}
	defer f.Close()

	checksummer := NewChecksummer()
	if _, err := io.Copy(checksummer, f); err != nil {
		return "", errors.Wrapf(err, "io.Copy()")
	}

	return checksummer.Checksum(), nil
}
