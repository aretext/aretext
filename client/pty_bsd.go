//go:build darwin || freebsd || netbsd || openbsd

package client

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

func unlockPts(ptmxFd int) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(ptmxFd), uintptr(unix.TIOCPTYGRANT), 0)
	if errno != 0 {
		return fmt.Errorf("ioctl TIOCPTYGRANT failed: %w", errno)
	}

	_, _, errno = unix.Syscall(unix.SYS_IOCTL, uintptr(ptmxFd), uintptr(unix.TIOCPTYUNLK), 0)
	if errno != 0 {
		return fmt.Errorf("ioctl TIOCPTYUNLK failed: %w", errno)
	}

	return nil
}

func ptsFileFromPtmx(ptmx *os.File) (*os.File, error) {
	buf := make([]byte, unix.PathMax)

	// Retrieve name of pts device.
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(ptmx.Fd()), uintptr(unix.TIOCPTYGNAME), uintptr(unsafe.Pointer(&buf[0])))
	if errno != 0 {
		return nil, fmt.Errorf("ioctl TIOCPTYGNAME failed: %w", errno)
	}

	// Convert C-string (NULL terminated) to Go string.
	n := 0
	for i, b := range buf {
		if b == 0 {
			n = i
			break
		}
	}
	ptsName := string(buf[:n])

	// Open the pts device.
	pts, err := os.OpenFile(ptsName, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("os.OpenFile failed for pts at %q: %w", ptsName, err)
	}

	return pts, nil
}

// TODO: consolidate this with duplicate in client.
// Also is this even the right way to do it? look at pkg/term it's a different ioctl
func drainPty(pts *os.File) error {
	_ = pts.SetReadDeadline(time.Now())

	_ = syscall.SetNonblock(int(pts.Fd()), true)
	tio, err := unix.IoctlGetTermios(int(pts.Fd()), unix.TIOCGETA)
	if err != nil {
		return err
	}
	tio.Cc[unix.VMIN] = 0
	tio.Cc[unix.VTIME] = 0

	err = unix.IoctlSetInt(int(pts.Fd()), unix.TIOCDRAIN, 0)
	if err != nil {
		return fmt.Errorf("ioctl TIOCDRAIN failed: %w", err)
	}

	return nil
}
