package client

import (
	"fmt"
	"io"
	"os"
	"syscall"

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

	// Retrieve pts file descriptor.
	ptsFd, _, err := unix.Syscall(unix.SYS_IOCTL, uintptr(ptmxFd), unix.TIOCGPTPEER, unix.O_RDWR|unix.O_NOCTTY)
	if int(ptsFd) == -1 {
		if errno, isErrno := err.(syscall.Errno); !isErrno || (errno != syscall.EINVAL && errno != syscall.ENOTTY) {
			return nil, nil, fmt.Errorf("could not retrieve pts file descriptor: %w", err)
		}
		// On EINVAL or ENOTTY, fallback to TIOCGPTN.
		ptyN, err := unix.IoctlGetInt(ptmxFd, unix.TIOCGPTN)
		if err != nil {
			return nil, nil, fmt.Errorf("could not find pty number: %w", err)
		}
		ptyName := fmt.Sprintf("/dev/pts/%d", ptyN)
		fd, err := unix.Open(ptyName, unix.O_RDWR|unix.O_NOCTTY, 0o620)
		if err != nil {
			return nil, nil, fmt.Errorf("could not open pty %s: %w", ptyName, err)
		}
		ptsFd = uintptr(fd)
	}

	ptmx = os.NewFile(uintptr(ptmxFd), "")
	pts = os.NewFile(ptsFd, "")
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
	doneCh := make(chan struct{})

	// Copy tty input -> pty
	go func(ptyOut io.Writer, ttyIn io.Reader) {
		_, _ = io.Copy(ptyOut, ttyIn)
		doneCh <- struct{}{}
	}(ptmx, os.Stdin)

	// Copy pty output -> tty
	go func(ttyOut io.Writer, ptyIn io.Reader) {
		_, _ = io.Copy(ttyOut, ptyIn)
		doneCh <- struct{}{}
	}(os.Stdout, ptmx)

	// Block until pty closed (either by client or server).
	select {
	case <-doneCh:
	}
}
