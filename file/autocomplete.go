package file

import (
	"os"
	"path/filepath"
	"strings"
)

// AutocompleteDirectory autocompletes the last subdirectory in a path.
func AutocompleteDirectory(path string) ([]string, error) {
	baseDir, subdirPrefix := filepath.Split(path)
	if baseDir == "" {
		baseDir = "."
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	var suffixes []string
	for _, e := range entries {
		if e.IsDir() {
			name := e.Name()
			if strings.HasPrefix(name, subdirPrefix) && len(subdirPrefix) < len(name) {
				suffixes = append(suffixes, name[len(subdirPrefix):])
			}
		}
	}

	return suffixes, nil
}
