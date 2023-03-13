package file

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type ListDirOptions struct {
	DirPatternsToHide []string // Glob patterns for directories to skip.
	DirectoriesOnly   bool     // If true, return directories (not files) in results.
}

// ListDir lists every file in a root directory and its subdirectories.
// The returned paths are relative to the root directory.
// The order of the returned paths is non-deterministic.
// Symbolic links are not followed.
// If an error occurs while accessing a directory, ListDir will skip that
// directory and log the error.
func ListDir(ctx context.Context, root string, options ListDirOptions) []string {
	// Use a semaphore to limit the number of open files.
	semaphoreChan := make(chan struct{}, runtime.NumCPU())
	return listDirRec(ctx, root, options, semaphoreChan)
}

func listDirRec(ctx context.Context, root string, options ListDirOptions, semaphoreChan chan struct{}) []string {
	select {
	case <-ctx.Done():
		log.Printf("Context done channel closed while listing subdirectories in %q: %s\n", root, ctx.Err())
		return nil
	default:
		break
	}

	semaphoreChan <- struct{}{} // Block until open file count decreases.
	dirEntries, err := listDir(root)
	<-semaphoreChan // Decrease open file count.

	if err != nil {
		log.Printf("Error listing subdirectories in %q: %s\n", root, err)
		return nil
	}

	var mu sync.Mutex
	var results []string
	var wg sync.WaitGroup
	for _, d := range dirEntries {
		path := filepath.Join(root, d.Name())

		if !d.IsDir() {
			if !options.DirectoriesOnly {
				mu.Lock()
				results = append(results, path)
				mu.Unlock()
			}
			continue
		}

		if shouldSkipDir(path, options.DirPatternsToHide) {
			continue
		}

		if options.DirectoriesOnly {
			mu.Lock()
			results = append(results, path)
			mu.Unlock()
		}

		// Traverse subdirectories concurrently.
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			subpaths := listDirRec(ctx, path, options, semaphoreChan)
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
		return nil, fmt.Errorf("os.Open: %w", err)
	}
	dirs, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		return nil, fmt.Errorf("f.ReadDir: %w", err)
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
