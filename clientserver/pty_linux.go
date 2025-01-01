//go:build linux

package clientserver

import (
	"fmt"
	"os"
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

func ptsPathFromPtmx(ptmx *os.File) (string, error) {
	ptyN, err := unix.IoctlGetInt(int(ptmx.Fd()), unix.TIOCGPTN)
	if err != nil {
		return "", fmt.Errorf("could not find pty number: %w", err)
	}
	ptyName := fmt.Sprintf("/dev/pts/%d", ptyN)
	return ptyName, nil
}
