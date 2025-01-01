package clientserver

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/aretext/aretext/clientserver/protocol"
)

// Client connects to an aretext server over a Unix Domain Socket (UDS),
// opens a pseudoterminal (pty), and sends the server one end of the pty over UDS.
//
// The client is responsible only for:
// 1. proxying input/output from pty to the client's tty.
// 2. handling SIGWINCH signals when the tty resizes.
// 3. detecting if the server has terminated and, if so, exiting.
//
// The server handles everything else.
type Client struct {
	config Config
}

// NewClient creates (but does not start) a new client with the given config.
func NewClient(config Config) *Client {
	return &Client{config}
}

// Run starts an aretext client and runs until the server terminates the connection.
// The documentPath is the initial document to open for the client, can be empty for new document.
func (c *Client) Run(documentPath string) error {
	// Register for SIGWINCH to detect when tty size changes.
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGWINCH)

	// TODO: probably check if this is a tty

	// Set tty to raw mode and restore on exit.
	restoreTty, err := setTtyRaw(os.Stdin)
	if err != nil {
		fmt.Errorf("failed to set client tty raw: %w", err)
	}
	defer restoreTty()

	// Create psuedoterminal (pty) pair.
	ptmx, ptsPath, err := createPtyPair()
	if err != nil {
		return fmt.Errorf("failed to create pty: %w", err)
	}
	defer ptmx.Close()

	// Ensure ptmx matches tty terminal size.
	_, _, err = resizePtyToMatchTty(os.Stdin, ptmx)
	if err != nil {
		return fmt.Errorf("resize ptmx: %w", err)
	}

	// Connect to the server over unix domain socket (UDS).
	conn, err := connectToServer(c.config.ServerSocketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	// Handle signals (SIGWINCH) asynchronously.
	go handleSignals(signalCh, ptmx, conn)

	// Send StartSessionMsg to the server, along with pts to delegate
	// the psuedoterminal to the server.
	err = sendStartSessionMsg(conn, ptsPath, documentPath)
	if err != nil {
		return fmt.Errorf("failed to send StartSessionMsg: %w", err)
	}

	// Proxy ptmx <-> tty.
	go proxyTtyToPtmxUntilClosed(ptmx)

	// Block until the server closes the connection.
	log.Printf("waiting for server to close connection...\n")
	var buf [1]byte
	_, err = conn.Read(buf[:])
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed read on unix socket: %w", err)
	}
	log.Printf("connection closed, exiting...\n")

	return nil
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

func sendStartSessionMsg(conn *net.UnixConn, ptsPath string, documentPath string) error {
	log.Printf("constructing StartSessionMsg\n")
	log.Printf("StartSessionMsg documentPath=%q\n", documentPath)
	log.Printf("StartSessionMsg ptsPath=%q\n", ptsPath)

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}
	log.Printf("StartSessionMsg workingDir=%q\n", workingDir)

	terminalEnv := make(map[string]string)
	for _, key := range allTerminalEnvVars {
		val, found := os.LookupEnv(key)
		if found {
			terminalEnv[key] = val
			log.Printf("StartSessionMsg terminalEnv[%q]=%q\n", key, val)
		}
	}

	msg := &protocol.StartSessionMsg{
		DocumentPath: documentPath,
		WorkingDir:   workingDir,
		TerminalEnv:  terminalEnv,
		PtsPath:      ptsPath,
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
				width, height, err := resizePtyToMatchTty(os.Stdin, ptmx)
				if err != nil {
					log.Printf("could not resize pty to match tty: %s\n", err)
					return
				}

				// Notify the server that the terminal size changed.
				msg := &protocol.ResizeTerminalMsg{
					Width:  width,
					Height: height,
				}
				err = protocol.SendMessage(conn, msg)
				if err != nil {
					fmt.Printf("failed to send ResizeTerminalMsg: %s\n", err)
				}
			}
		}
	}
}
