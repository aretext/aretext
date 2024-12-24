package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
	"golang.org/x/term"

	"github.com/aretext/aretext/protocol"
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
func RunClient(config Config, documentPath string) error {
	// Register for SIGWINCH to detect when tty size changes.
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGWINCH)

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

	// Handle signals (SIGWINCH) asynchronously.
	go handleSignals(signalCh, ptmx, conn)

	// Send RegisterClient to the server, along with pts to delegate
	// the psuedoterminal to the server.
	err = sendRegisterClient(conn, pts, documentPath)
	if err != nil {
		return fmt.Errorf("failed to send RegisterClient: %w", err)
	}

	// Close pts as it's now owned by the server.
	pts.Close()

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
	addr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("net.ResolveUnixAddr: %w", err)
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("net.DialUnix: %w", err)
	}

	return conn, nil
}

var allTerminalEnvVars = []string{"TERM", "TERMINFO", "TERMCAP", "COLORTERM", "LINES", "COLUMNS"}

func sendRegisterClient(conn *net.UnixConn, pts *os.File, documentPath string) error {
	log.Printf("constructing RegisterClientMsg\n")
	log.Printf("RegisterClient documentPath=%q\n", documentPath)

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}
	log.Printf("RegisterClient workingDir=%q\n", workingDir)

	var terminalEnv []string
	for _, key := range allTerminalEnvVars {
		val, found := os.LookupEnv(key)
		if found {
			terminalEnv = append(terminalEnv, fmt.Sprintf("%s=%s", key, val))
		}
	}
	log.Printf("RegisterClient terminalEnv=%s\n", terminalEnv)

	msg := &protocol.RegisterClientMsg{
		DocumentPath: documentPath,
		WorkingDir:   workingDir,
		TerminalEnv:  terminalEnv,
		Pts:          pts,
	}

	return protocol.SendMessage(conn, msg)
}

func handleSignals(signalCh chan os.Signal, ptmx *os.File, conn *net.UnixConn) {
	for {
		select {
		case signal := <-signalCh:
			switch signal {
			case syscall.SIGWINCH:
				log.Printf("received SIGWINCH signal\n")
				err := resizePtmxAndNotifyServer(ptmx, conn)
				if err != nil {
					log.Printf("could not resize tty: %s\n", err)
				}
			}
		}
	}
}

func resizePtmxAndNotifyServer(ptmx *os.File, conn *net.UnixConn) error {
	// Update ptmx with the same size as client tty.
	ws, err := unix.IoctlGetWinsize(int(os.Stdin.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return fmt.Errorf("unix.IoctlGetWinsize: %w", err)
	}

	err = unix.IoctlSetWinsize(int(ptmx.Fd()), unix.TIOCSWINSZ, ws)
	if err != nil {
		return fmt.Errorf("unix.IoctlSetWinsize: %w", err)
	}

	// Notify the server that the terminal size changed.
	msg := &protocol.ResizeTerminalMsg{
		Width:  int(ws.Row),
		Height: int(ws.Col),
	}
	err = protocol.SendMessage(conn, msg)
	if err != nil {
		return fmt.Errorf("failed to send ResizeTerminalMsg: %w", err)
	}

	return nil
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
