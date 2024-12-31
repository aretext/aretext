//go:build darwin || freebsd || netbsd || openbsd

package server

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func drainPty(pts *os.File) error {
	err = unix.IoctlSetInt(int(pts.Fd()), unix.TIOCDRAIN, 0)
	if err != nil {
		return fmt.Errorf("ioctl TIOCDRAIN failed: %w", err)
	}
	return nil
}
