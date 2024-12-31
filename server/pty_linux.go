//go:build linux

package server

import (
	"os"

	"golang.org/x/sys/unix"
)

func drainPty(pts *os.File) error {
	// TCSETSW waits for all terminal events to drain, then updates attributes (no-op in this case).
	tio, err := unix.IoctlGetTermios(int(pts.Fd()), unix.TCGETS)
	if err != nil {
		return err
	}
	if err = unix.IoctlSetTermios(int(pts.Fd()), unix.TCSETSW, tio); err != nil {
		return err
	}

	return nil
}
