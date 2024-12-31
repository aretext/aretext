//go:build darwin || freebsd || netbsd || openbsd

package client

func unlockPts(ptmxFd int) error {
	// BSD doesn't require unlocking pts.
	return nil
}
