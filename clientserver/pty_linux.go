//go:build linux

package clientserver

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func unlockPts(ptmxFd int) error {
	locked := 0
	result, _, err := unix.Syscall(unix.SYS_IOCTL, uintptr(ptmxFd), unix.TIOCSPTLCK, uintptr(unsafe.Pointer(&locked)))
	if int(result) == -1 {
		return fmt.Errorf("could not unlock pty: %w", err)
	}
	return nil
}

func ptsFileFromPtmx(ptmx *os.File) (*os.File, error) {
	var err error
	ptsFd, _, err := unix.Syscall(unix.SYS_IOCTL, ptmx.Fd(), unix.TIOCGPTPEER, unix.O_RDWR|unix.O_NOCTTY)
	if int(ptsFd) == -1 {
		if errno, isErrno := err.(syscall.Errno); !isErrno || (errno != syscall.EINVAL && errno != syscall.ENOTTY) {
			return nil, fmt.Errorf("could not retrieve pts file descriptor: %w", err)
		}
		// On EINVAL or ENOTTY, fallback to TIOCGPTN.
		ptyN, err := unix.IoctlGetInt(int(ptmx.Fd()), unix.TIOCGPTN)
		if err != nil {
			return nil, fmt.Errorf("could not find pty number: %w", err)
		}
		ptyName := fmt.Sprintf("/dev/pts/%d", ptyN)
		fd, err := unix.Open(ptyName, unix.O_RDWR|unix.O_NOCTTY, 0o620)
		if err != nil {
			return nil, fmt.Errorf("could not open pty %s: %w", ptyName, err)
		}
		ptsFd = uintptr(fd)
	}

	return os.NewFile(ptsFd, ""), nil
}

func setTtyNonblockAndDrain(ttyFd int) error {
	err := syscall.SetNonblock(ttyFd, true)
	if err != nil {
		return fmt.Errorf("syscall.SetNonblock failed: %w", err)
	}

	// TCSETSW waits for all terminal events to drain, then updates attributes.
	// Set VMIN and VTIME to 0 to avoid waiting for more characters.
	tio, err := unix.IoctlGetTermios(ttyFd, unix.TCGETS)
	if err != nil {
		return fmt.Errorf("ioctl TCGETS failed: %w", err)
	}
	tio.Cc[unix.VMIN] = 0
	tio.Cc[unix.VTIME] = 0
	if err = unix.IoctlSetTermios(ttyFd, unix.TCSETSW, tio); err != nil {
		return fmt.Errorf("ioctl TCSETSW failed: %w", err)
	}

	return nil
}
