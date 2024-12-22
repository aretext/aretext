package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

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
func RunClient(ctx context.Context) error {
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
	conn, err := connectToServer() // TODO: configure path
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
	doneCh := make(chan struct{})
	go func(ptyOut io.Writer, ttyIn io.Reader) {
		_, _ = io.Copy(ptyOut, ttyIn)
		doneCh <- struct{}{}
	}(ptmx, os.Stdin)

	go func(ttyOut io.Writer, ptyIn io.Reader) {
		_, _ = io.Copy(ttyOut, ptyIn)
		doneCh <- struct{}{}
	}(os.Stdout, ptmx)

	// Block until pty closed (either by client or server).
	select {
	case <-doneCh:
	}

	return nil
}

func createPtyPair() (ptmx *os.File, pts *os.File, err error) {
	return nil, nil, nil
}

func connectToServer() (*net.UnixConn, error) {
	return nil, nil
}

func sendClientHelloWithPty(conn *net.UnixConn, pts *os.File) error {
	return nil
}

func waitForServerHello(ctx context.Context, conn *net.UnixConn) (clientId int, err error) {
	return 0, nil
}

func handleSignals(signalCh chan os.Signal, ptmx *os.File, conn *net.UnixConn) error {
	return nil
}

func handleServerMessages(conn *net.UnixConn, ptmx *os.File) {
}
