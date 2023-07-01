//go:build !windows

package file

import (
	"fmt"
	"io"

	"github.com/google/renameio/v2"
)

func platformSpecificSave(path string, r io.Reader) error {
	// Use renameio to write the file to a temporary directory, then rename it to the target file.
	// This should reduce the risk of data corruption if the editor crashes mid-write,
	// but is probably not 100% reliable (see http://danluu.com/deconstruct-files/).
	// There is a good discussion of the Go libraries solving this problem in
	// this GitHub issue comment: https://github.com/golang/go/issues/22397#issuecomment-380831736
	pf, err := renameio.NewPendingFile(path, renameio.WithPermissions(0644), renameio.WithExistingPermissions())
	if err != nil {
		return fmt.Errorf("renamio.TempFile: %w", err)
	}
	defer pf.Cleanup()

	// Copy data from the reader to the file.
	_, err = io.Copy(pf, r)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	// Sync the file to disk.
	err = pf.CloseAtomicallyReplace()
	if err != nil {
		return fmt.Errorf("renamio.CloseAtomicallyReplace: %w", err)
	}

	return nil
}
