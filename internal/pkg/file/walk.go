package file

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Walk walks the file tree at root, evaluating the function for every file path.
// The file paths are absolute.
func Walk(root string, f func(path string)) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error at path '%s': %v\n", path, err)
			return err
		}

		if info.Mode().IsRegular() {
			f(path)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error in file walk: %v\n", errors.Wrapf(err, "filepath.Walk"))
	}
}
