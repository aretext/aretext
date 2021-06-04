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
func ListDir(root string, dirNamesToHide map[string]struct{}) []string {
	// Use a semaphore to limit the number of open files.
	semaphoreChan := make(chan struct{}, runtime.NumCPU())
	return listDirRec(root, dirNamesToHide, semaphoreChan)
}

func listDirRec(root string, dirNamesToHide map[string]struct{}, semaphoreChan chan struct{}) []string {
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

		if shouldSkipDir(path, dirNamesToHide) {
			continue
		}

		// Traverse subdirectories concurrently.
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			subpaths := listDirRec(path, dirNamesToHide, semaphoreChan)
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
		return nil, errors.Wrapf(err, "os.Open")
	}
	dirs, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "f.ReadDir")
	}
	return dirs, nil
}

func shouldSkipDir(path string, dirNamesToHide map[string]struct{}) bool {
	name := filepath.Base(path)
	_, ok := dirNamesToHide[name]
	return ok
}
