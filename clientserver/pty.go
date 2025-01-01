package clientserver

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func setTtyRaw(tty *os.File) (restoreTty func(), err error) {
	// Set tty to raw mode.
	ttyFd := int(tty.Fd())
	oldTtyState, err := term.MakeRaw(ttyFd)
	if err != nil {
		return nil, fmt.Errorf("failed to set tty state: %w", err)
	}

	// Restore original state.
	restoreTty = func() {
		term.Restore(ttyFd, oldTtyState)
	}

	return restoreTty, nil
}

func createPtyPair() (ptmx *os.File, ptsPath string, err error) {
	// Create the pty pair.
	ptmxFd, err := unix.Open("/dev/ptmx", os.O_RDWR, 0o600)
	if err != nil {
		return nil, "", fmt.Errorf("could not open /dev/ptmx: %w", err)
	}

	// Unlock pts.
	err = unlockPts(ptmxFd)
	if err != nil {
		return nil, "", err
	}

	// Get path to pts device.
	ptmx = os.NewFile(uintptr(ptmxFd), "")
	ptsPath, err = ptsPathFromPtmx(ptmx)
	if err != nil {
		return nil, "", err
	}
	return ptmx, ptsPath, nil
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
}
