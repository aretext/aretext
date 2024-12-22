package file

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateCreate checks whether the user can (probably) create a file at a path.
// This is meant to catch common issues (non-existent directory, file already exists)
// but isn't 100% accurate. In particular, another process could modify the filesystem
// after the check, or the user might not have permission to create the file.
func ValidateCreate(path string) error {
	dir, filename := filepath.Split(path)

	// If the filename is empty, return an error.
	if filename == "" {
		if dir == "" {
			return fmt.Errorf("File path is empty")
		} else {
			return fmt.Errorf("File path must end with a file name")
		}
	}

	dir = filepath.Clean(dir)

	// If the directory doesn't exist, return an error.
	if f, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Directory does not exist: %s", dir)
		} else {
			return fmt.Errorf("Error checking if directory exists: %w", err)
		}
	} else if !f.IsDir() {
		return fmt.Errorf("Not a directory: %s", dir)
	}

	// If the file already exists, return an error.
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("File already exists at %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Error checking if file exists: %w", err)
	}

	return nil
}
