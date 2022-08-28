package file

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// RelativePathCwd converts an absolute path to a path relative to the current working directory.
// If the conversion fails, the absolute path will be returned instead.
func RelativePathCwd(p string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current working directory: %v\n", errors.Wrap(err, "os.Getwd"))
		return p
	}
	return RelativePath(p, cwd)
}

// RelativePath converts an absolute path to a path relative to a base directory.
// If the conversion fails, the absolute path will be returned instead.
func RelativePath(p string, baseDir string) string {
	relPath, err := filepath.Rel(baseDir, p)
	if err != nil {
		log.Printf("Error converting %q to relative path from base %q: %v\n", p, baseDir, errors.Wrap(err, "filepath.Rel"))
		return p
	}
	return relPath
}
