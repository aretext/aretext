//go:build linux

package client

import (
	"fmt"
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
