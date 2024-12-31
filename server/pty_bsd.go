//go:build darwin || freebsd || netbsd || openbsd

package server

import (
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func drainPty(pts *os.File) error {
	_ = pts.SetReadDeadline(time.Now())

	_ = syscall.SetNonblock(int(pts.Fd()), true)
	tio, err := unix.IoctlGetTermios(int(pts.Fd()), unix.TIOCGETA)
	if err != nil {
		return err
	}
	tio.Cc[unix.VMIN] = 0
	tio.Cc[unix.VTIME] = 0
	if err = unix.IoctlSetTermios(int(pts.Fd()), unix.TIOCSETAW, tio); err != nil {
		return err
	}
	return nil
}
