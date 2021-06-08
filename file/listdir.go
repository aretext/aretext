package file

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/pkg/errors"
)

// ListDir lists every file in a root directory and its subdirectories.
// The returned paths are relative to the root directory.
// The order of the returned paths is non-deterministic.
// Symbolic links are not followed.
// If an error occurs while accessing a directory, ListDir will skip that
// directory and log the error.
func ListDir(root string, dirPatternsToHide []string) []string {
	// Use a semaphore to limit the number of open files.
	semaphoreChan := make(chan struct{}, runtime.NumCPU())
	return listDirRec(root, dirPatternsToHide, semaphoreChan)
}

func listDirRec(root string, dirPatternsToHide []string, semaphoreChan chan struct{}) []string {
	semaphoreChan <- struct{}{} // Block until open file count decreases.
	dirEntries, err := listDir(root)
	<-semaphoreChan // Decrease open file count.

	if err != nil {
		log.Printf("Error listing subdirectories in '%s': %s\n", root, err)
		return nil
	}

	var mu sync.Mutex
	var results []string
	var wg sync.WaitGroup
	for _, d := range dirEntries {
		path := filepath.Join(root, d.Name())

		if !d.IsDir() {
			results = append(results, path)
			continue
		}

		if shouldSkipDir(path, dirPatternsToHide) {
			continue
		}

		// Traverse subdirectories concurrently.
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			subpaths := listDirRec(path, dirPatternsToHide, semaphoreChan)
			mu.Lock()
			results = append(results, subpaths...)
			mu.Unlock()
		}(path)
	}
	wg.Wait()

	return results
}

func listDir(path string) ([]fs.DirEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "os.Open")
	}
	dirs, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		return nil, errors.Wrap(err, "f.ReadDir")
	}
	return dirs, nil
}

func shouldSkipDir(path string, dirPatternsToHide []string) bool {
	for _, pattern := range dirPatternsToHide {
		if GlobMatch(pattern, path) {
			return true
		}
	}
	return false
}
