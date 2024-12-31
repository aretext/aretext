//go:build darwin || freebsd || netbsd || openbsd

package server

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func drainPty(pts *os.File) error {
	_ = pts.SetReadDeadline(time.Now())

	_ = syscall.SetNonblock(int(pts.Fd()), true)

	err := unix.IoctlSetInt(int(pts.Fd()), unix.TIOCDRAIN, 0)
	if err != nil {
		return fmt.Errorf("ioctl TIOCDRAIN failed: %w", err)
	}
	return nil
}
