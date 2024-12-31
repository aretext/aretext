//go:build darwin || freebsd || netbsd || openbsd

package clientserver

import (
	"fmt"
	"os"
	"syscall"
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

func setTtyNonblockAndDrain(ttyFd int) error {
	err := syscall.SetNonblock(ttyFd, true)
	if err != nil {
		return fmt.Errorf("syscall.SetNonblock failed: %w", err)
	}

	tio, err := unix.IoctlGetTermios(ttyFd, unix.TIOCGETA)
	if err != nil {
		return fmt.Errorf("ioctl TIOCGETA failed: %w", err)
	}
	tio.Cc[unix.VMIN] = 0
	tio.Cc[unix.VTIME] = 0
	if err = unix.IoctlSetTermios(ttyFd, unix.TIOCSETAW, tio); err != nil {
		return fmt.Errorf("ioctl TIOCSETAW failed: %w", err)
	}
	return nil
}

func flushPts(pts *os.File) error {
	if err := unix.IoctlSetInt(int(pts.Fd()), unix.TIOCFLUSH, 0); err != nil {
		return fmt.Errorf("failed to flush pty: %w", err)
	}
	return nil
}
