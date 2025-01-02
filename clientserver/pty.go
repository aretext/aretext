package clientserver

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func createPtyPair(width int, height int) (ptmx *os.File, pts *os.File, err error) {
	// Create the pty pair.
	ptmxFd, err := unix.Open("/dev/ptmx", os.O_RDWR, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open /dev/ptmx: %w", err)
	}

	err = syscall.SetNonblock(ptmxFd, true)
	if err != nil {
		return nil, nil, fmt.Errorf("syscall.SetNonblock: %w", err)
	}

	// Unlock pts.
	err = unlockPts(ptmxFd)
	if err != nil {
		return nil, nil, err
	}

	// File descriptors for both ptmx and pts.
	ptmx = os.NewFile(uintptr(ptmxFd), "")
	pts, err = ptsFileFromPtmx(ptmx)
	if err != nil {
		return nil, nil, err
	}

	// Set terminal size on ptmx. On macOS this needs to happen
	// AFTER opening the pts or else we get "inappropriate ioctl for device" error.
	ws := unix.Winsize{
		Col: uint16(width),
		Row: uint16(height),
	}
	err = unix.IoctlSetWinsize(ptmxFd, unix.TIOCSWINSZ, &ws)
	if err != nil {
		return nil, nil, fmt.Errorf("unix.IoctlSetWinsize: %w", err)
	}

	return ptmx, pts, nil
}
