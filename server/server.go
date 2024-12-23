package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/aretext/aretext/protocol"
)

// RunServer starts an aretext server.
// The server listens on a Unix Domain Socket (UDS) for clients to connect.
// The client sends the server a pseudoterminal (pty), which the server uses
// for input/output from/to the client's terminal.
func RunServer(config Config) error {
	releaseLock, err := acquireLock(config.LockPath)
	if err != nil {
		return fmt.Errorf("acquireLock: %w", err)
	}
	defer releaseLock()

	ul, err := createListenSocket(config.SocketPath)
	if err != nil {
		return fmt.Errorf("createListenSocket: %w", err)
	}

	// TODO: setup some kind of channel and background thread to manage editor state?
	// or do the accept/conn dance in background?

	clientId := protocol.ClientId(0)
	for {
		conn, err := ul.AcceptUnix()
		if err != nil {
			return err
		}

		go handleConnection(conn, clientId)
		clientId++
	}
}

func createListenSocket(socketPath string) (*net.UnixListener, error) {
	err := syscall.Unlink(socketPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("syscall.Unlink: %w", err)
	}

	addr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("net.ResolveUnixAddr: %w", err)
	}

	log.Printf("listening on %s\n", addr)
	ul, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, fmt.Errorf("net.ListenUnix: %w", err)
	}

	return ul, nil
}

func handleConnection(conn *net.UnixConn, clientId protocol.ClientId) {
	log.Printf("client %d connected\n", clientId)
	defer conn.Close()

	// TODO: receive ClientHelloMsg
	//  - add client state with pty, working dir, etc.
	// TODO: send ServerHelloMsg
	// TODO: loop wating for TerminalResize or ClientGoodbye
	//   - on TerminalResize, update editor state
	//   - on ClientGoodbye, remove client state and exit
	//   - on any other error, remove client state and exit
}
