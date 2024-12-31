package clientserver

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

func acquireLock(lockpath string) (func(), error) {
	f, err := os.Create(lockpath)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	err = syscall.Flock(int(f.Fd()), int(syscall.LOCK_EX|syscall.LOCK_NB))
	if err != nil {
		return nil, &fs.PathError{
			Op:   "lock",
			Path: f.Name(),
			Err:  err,
		}
	}

	cleanup := func() {
		syscall.Flock(int(f.Fd()), int(syscall.LOCK_UN))
		f.Close()
	}

	return cleanup, nil
}
