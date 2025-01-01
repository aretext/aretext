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

	"golang.org/x/term"

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
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("device is not a terminal")
	}

	// Register for SIGWINCH to detect when tty size changes.
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGWINCH)

	// Set tty to raw mode and restore on exit.
	log.Printf("set tty to raw\n")
	restoreTty, err := setTtyRaw(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to set client tty raw: %w", err)
	}
	defer restoreTty()

	// Create pipes connected to tty.
	pipeInReader, pipeInWriter, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("os.Pipe: %w", err)
	}
	defer pipeInWriter.Close()
	defer pipeInReader.Close()

	pipeOutReader, pipeOutWriter, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("os.Pipe: %w", err)
	}
	defer pipeOutWriter.Close()
	defer pipeInReader.Close()

	// Get terminal size.
	termWidth, termHeight, err := getTtySize(os.Stdin)
	if err != nil {
		return err
	}

	// Connect to the server over unix domain socket (UDS).
	log.Printf("connecting to server at %s\n", c.config.ServerSocketPath)
	conn, err := connectToServer(c.config.ServerSocketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	// Get working dir
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}

	// Send StartSessionMsg to the server.
	msg := &protocol.StartSessionMsg{
		PipeIn:         pipeInReader,
		PipeOut:        pipeOutWriter,
		TerminalWidth:  termWidth,
		TerminalHeight: termHeight,
		TerminalEnv:    getTerminalEnv(),
		DocumentPath:   documentPath,
		WorkingDir:     workingDir,
	}
	log.Printf("sending start msg to server: %v\n", msg)
	err = protocol.SendMessage(conn, msg)
	if err != nil {
		return fmt.Errorf("failed to send StartSessionMsg: %w", err)
	}

	// Close pipe file descriptors that are now owned by the server.
	pipeInReader.Close()
	pipeOutWriter.Close()

	// Handle signals (SIGWINCH) asynchronously.
	go handleSignals(signalCh, os.Stdin, conn)

	// Copy tty input -> server pipe in
	go func() {
		log.Printf("start copying tty input -> server pipe in\n")
		_, _ = io.Copy(pipeInWriter, os.Stdin)
		log.Printf("finished copying tty input -> server pipe in\n")
	}()

	// Copy server pipe out -> tty output
	go func() {
		log.Printf("start copying server pipe out -> tty out\n")
		_, _ = io.Copy(os.Stdout, pipeOutReader)
		log.Printf("finish copying server pipe out -> tty out\n")
	}()

	// Block until the server closes the connection.
	log.Printf("blocking until server closes conn\n")
	return blockUntilServerClosesConn(conn)
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

func getTerminalEnv() map[string]string {
	terminalEnv := make(map[string]string)
	for _, key := range allTerminalEnvVars {
		val, found := os.LookupEnv(key)
		if found {
			terminalEnv[key] = val
		}
	}
	return terminalEnv
}

func handleSignals(signalCh chan os.Signal, tty *os.File, conn *net.UnixConn) {
	for {
		select {
		case signal := <-signalCh:
			switch signal {
			case syscall.SIGWINCH:
				log.Printf("received SIGWINCH signal\n")
				width, height, err := getTtySize(tty)
				if err != nil {
					log.Printf("could not get tty size: %s\n", err)
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

func blockUntilServerClosesConn(conn *net.UnixConn) error {
	var buf [1]byte // at least one byte to block on read
	for {
		_, err := conn.Read(buf[:])
		if err != nil && errors.Is(err, io.EOF) {
			log.Printf("server closed connection\n")
			return nil
		} else if err != nil {
			return fmt.Errorf("error reading conn: %w", err)
		}
	}
}
