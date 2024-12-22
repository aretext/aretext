package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// RunClient starts an aretext client.
// The client connects to an aretext server over a Unix Domain Socket (UDS),
// opens a pseudoterminal (pty), and sends the server one end of the pty over UDS.
//
// The client is responsible only for:
// 1. proxying input/output from pty to the client's tty.
// 2. handling SIGWINCH signals when the tty resizes.
// 3. detecting if the server has terminated and, if so, exiting.
//
// The server handles everything else.
func RunClient(ctx context.Context, config Config) error {
	// Register for SIGINT to notify server of client termination.
	// Register for SIGWINCH to detect when tty size changes.
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGWINCH)
	signal.Notify(signalCh, syscall.SIGINT)

	// Set tty to raw mode and restore on exit.
	ttyFd := int(os.Stdin.Fd())
	oldTtyState, err := term.MakeRaw(ttyFd)
	if err != nil {
		return fmt.Errorf("failed to set tty state: %w", err)
	}
	defer term.Restore(ttyFd, oldTtyState)

	// Create psuedoterminal (pty) pair.
	ptmx, pts, err := createPtyPair()
	if err != nil {
		return fmt.Errorf("failed to create pty: %w", err)
	}

	// Connect to the server over unix domain socket (UDS).
	conn, err := connectToServer(config.ServerSocketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	// Send ClientHello to the server, along with pts to delegate
	// the psuedoterminal to the server.
	err = sendClientHelloWithPty(conn, pts)
	if err != nil {
		return fmt.Errorf("failed to send ClientHello: %w", err)
	}

	// Wait for server to reply with ServerHello.
	clientId, err := waitForServerHello(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed waiting for ServerHello: %w", err)
	}

	// Close pts as it's now owned by the server.
	pts.Close()

	// Handle signals and server messages.
	go handleSignals(signalCh, ptmx, conn)
	go handleServerMessages(conn, ptmx)

	// Proxy ptmx <-> tty.
	proxyTtyUntilClosed(ptmx)

	return nil
}

func createPtyPair() (ptmx *os.File, pts *os.File, err error) {
	// Create the pty pair.
	ptmxFd, err := unix.Open("/dev/ptmx", os.O_RDWR, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open /dev/ptmx: %w", err)
	}

	// Unlock pts.
	locked := 0
	result, _, err := unix.Syscall(unix.SYS_IOCTL, uintptr(ptmxFd), unix.TIOCSPTLCK, uintptr(unsafe.Pointer(&locked)))
	if int(result) == -1 {
		return nil, nil, fmt.Errorf("could not unlock pty: %w", err)
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

func connectToServer(socketPath string) (*net.UnixConn, error) {
	// TODO
	return nil, nil
}

func sendClientHelloWithPty(conn *net.UnixConn, pts *os.File) error {
	// TODO
	return nil
}

func waitForServerHello(ctx context.Context, conn *net.UnixConn) (clientId int, err error) {
	// TODO
	return 0, nil
}

func handleSignals(signalCh chan os.Signal, ptmx *os.File, conn *net.UnixConn) error {
	// TODO
	return nil
}

func handleServerMessages(conn *net.UnixConn, ptmx *os.File) {
	// TODO
}

func proxyTtyUntilClosed(ptmx *os.File) {
	doneCh := make(chan struct{})

	// Copy pty -> tty
	go func(ptyOut io.Writer, ttyIn io.Reader) {
		_, _ = io.Copy(ptyOut, ttyIn)
		doneCh <- struct{}{}
	}(ptmx, os.Stdin)

	// Copy tty -> pty
	go func(ttyOut io.Writer, ptyIn io.Reader) {
		_, _ = io.Copy(ttyOut, ptyIn)
		doneCh <- struct{}{}
	}(os.Stdout, ptmx)

	// Block until pty closed (either by client or server).
	select {
	case <-doneCh:
	}
}
