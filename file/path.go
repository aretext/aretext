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
		log.Printf("Error getting current working directory: %v\n", errors.Wrapf(err, "os.Getwd"))
		return p
	}

	relPath, err := filepath.Rel(cwd, p)
	if err != nil {
		log.Printf("Error converting '%s' to relative path from base '%s': %v\n", p, cwd, errors.Wrapf(err, "filepath.Rel"))
		return p
	}

	return relPath
}
