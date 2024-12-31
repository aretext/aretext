package client

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

func createPtyPair() (ptmx *os.File, pts *os.File, err error) {
	// Create the pty pair.
	ptmxFd, err := unix.Open("/dev/ptmx", os.O_RDWR, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open /dev/ptmx: %w", err)
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
	return ptmx, pts, nil
}

func resizePtyToMatchTty(tty *os.File, ptmx *os.File) (width, height int, err error) {
	// Update ptmx with the same size as client tty.
	ws, err := unix.IoctlGetWinsize(int(tty.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, fmt.Errorf("unix.IoctlGetWinsize: %w", err)
	}

	err = unix.IoctlSetWinsize(int(ptmx.Fd()), unix.TIOCSWINSZ, ws)
	if err != nil {
		return 0, 0, fmt.Errorf("unix.IoctlSetWinsize: %w", err)
	}

	return int(ws.Col), int(ws.Row), nil
}

func proxyTtyToPtmxUntilClosed(ptmx *os.File) {
	// Copy tty input -> pty (async)
	go func(ptyOut io.Writer, ttyIn io.Reader) {
		_, _ = io.Copy(ptyOut, ttyIn)
	}(ptmx, os.Stdin)

	// Copy pty output -> tty
	// This blocks until the server closes the pts.
	_, _ = io.Copy(os.Stdout, ptmx)

	// Ensure tty output drained.
	_ = drainPty(ptmx)
}
