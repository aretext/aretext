//go:build windows

package file

import (
	"fmt"
	"io"
	"os"
)

func platformSpecificSave(path string, r io.Reader) error {
	// renameio is not available on Windows (see https://github.com/google/renameio/pull/20)
	// so we use the Go stdlib instead.
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	return nil
}
