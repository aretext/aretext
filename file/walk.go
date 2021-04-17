package file

import (
	"io/fs"
	"log"
	"path/filepath"
)

// Walk walks the file tree at root, evaluating the function for every file path.
// Symlinks are skipped.
// The file paths are absolute.
func Walk(root string, dirNamesToHide map[string]struct{}, walkFn func(path string)) {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if shouldSkipDir(path, dirNamesToHide) {
				return fs.SkipDir
			}
			return nil
		}
		walkFn(path)
		return nil
	})
	if err != nil {
		log.Printf("Error walking directory '%s': %v\n", root, err)
	}
}

func shouldSkipDir(path string, dirNamesToHide map[string]struct{}) bool {
	name := filepath.Base(path)
	_, ok := dirNamesToHide[name]
	return ok
}
