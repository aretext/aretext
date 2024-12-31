//go:build linux

package server

import (
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// TODO: is all of this necessary? don't think VMIN / VTIME needs to change,
// and may not even need read deadline or nonblock...
func drainPty(pts *os.File) error {
	_ = pts.SetReadDeadline(time.Now())

	_ = syscall.SetNonblock(int(pts.Fd()), true)
	tio, err := unix.IoctlGetTermios(int(pts.Fd()), unix.TCGETS)
	if err != nil {
		return err
	}
	tio.Cc[unix.VMIN] = 0
	tio.Cc[unix.VTIME] = 0
	if err = unix.IoctlSetTermios(int(pts.Fd()), unix.TCSETSW, tio); err != nil {
		return err
	}

	return nil
}
