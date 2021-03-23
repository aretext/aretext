package file

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Walk walks the file tree at root, evaluating the function for every file path.
// Symlinks are skipped.
// The file paths are absolute.
func Walk(root string, dirNamesToHide map[string]struct{}, walkFn func(path string)) {
	stack := []string{root}
	var dirPath string
	for len(stack) > 0 {
		dirPath, stack = stack[len(stack)-1], stack[:len(stack)-1]
		if shouldSkipDir(dirPath, dirNamesToHide) {
			continue
		}

		subdirPaths, err := processDir(dirPath, walkFn)
		if err != nil {
			log.Printf("Error processing directory at '%s': %v\n", dirPath, err)
			continue
		}
		stack = append(stack, subdirPaths...)
	}
}

func shouldSkipDir(path string, dirNamesToHide map[string]struct{}) bool {
	name := filepath.Base(path)
	_, ok := dirNamesToHide[name]
	return ok
}

func processDir(dirPath string, walkFn func(path string)) ([]string, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return nil, errors.Wrapf(err, "os.Open")
	}
	defer f.Close()

	// This is faster than calling LStat on every file, which is what filepath.Walk does.
	// See discussion here: https://github.com/golang/go/issues/16399
	fileInfos, err := f.Readdir(-1)
	if err != nil {
		return nil, errors.Wrapf(err, "File.Readdir")
	}

	subdirPaths := make([]string, 0)
	for _, fi := range fileInfos {
		absPath := filepath.Join(dirPath, fi.Name())
		if fi.Mode().IsRegular() {
			walkFn(absPath)
		} else if fi.Mode().IsDir() {
			subdirPaths = append(subdirPaths, absPath)
		}
	}

	return subdirPaths, nil
}
