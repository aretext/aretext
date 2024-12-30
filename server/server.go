package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/protocol"
)

type sessionId int

// Server listens on a Unix Domain Socket (UDS) for clients to connect.
// The client sends the server a pseudoterminal (pty), which the server uses
// for input/output from/to the client's terminal.
type Server struct {
	config   Config
	quitChan chan struct{}
}

// NewServer creates (but does not start) a new server with the given config.
func NewServer(config Config) *Server {
	return &Server{
		config:   config,
		quitChan: make(chan struct{}),
	}
}

// RunServer starts an aretext server.
// The server listens on a Unix Domain Socket (UDS) for clients to connect.
// The client sends the server a pseudoterminal (pty), which the server uses
// for input/output from/to the client's terminal.
func (s *Server) Run() error {
	// Acquire a filelock to ensure at most one server running at a time.
	releaseLock, err := acquireLock(s.config.LockPath)
	if err != nil {
		return fmt.Errorf("acquireLock: %w", err)
	}
	defer releaseLock()

	ul, err := createListenSocket(s.config.SocketPath)
	if err != nil {
		return fmt.Errorf("could not create listen socket: %w", err)
	}
	defer ul.Close()

	return s.listenForConnections(ul)
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

func (s *Server) listenForConnections(ul *net.UnixListener) error {
	nextSessionId := sessionId(0)
	for {
		uc, err := ul.AcceptUnix()
		if err != nil {
			log.Printf("net.AcceptUnix: %s\n", err)
			continue
		}

		go s.handleConnection(nextSessionId, uc)
		nextSessionId++
	}

	return nil
}

func (s *Server) handleConnection(id sessionId, uc *net.UnixConn) {
	log.Printf("client connected, sessionId=%d\n", id)
	defer uc.Close()

	msg, err := receiveStartSessionMsg(uc)
	if err != nil {
		log.Printf("error receiving StartSesssionMsg, sessionId=%d: %s\n", id, err)
		return
	}

	termInfo, err := tcell.LookupTerminfo(msg.TerminalEnv["TERM"])
	if err != nil {
		log.Printf("error looking up terminfo, TERM env = %q, sessionId=%d: %s\n", msg.TerminalEnv["TERM"], id, err)
		return
	}

	clientTty, err := NewTtyFromPts(msg.Pts)
	if err != nil {
		log.Printf("error constructing tty from pts, sessionId=%d: %s\n", id, err)
		return
	}
	defer clientTty.Close()

	screen, err := tcell.NewTerminfoScreenFromTtyTerminfo(clientTty, termInfo)
	if err != nil {
		log.Printf("error constructing screen for client, sessionId=%d: %s\n", id, err)
		return
	}

	err = screen.Init()
	if err != nil {
		log.Printf("error initializing screen for client, sessionId=%d: %s\n", id, err)
		return
	}
	defer screen.Fini()

	// TODO: better acquire a lock and add the session to editor state

	// Process terminal events from client tty.
	termEventChan := make(chan tcell.Event, 1024)
	screenQuitChan := make(chan struct{}, 1)
	go screen.ChannelEvents(termEventChan, screenQuitChan)
	defer func() {
		screenQuitChan <- struct{}{}
	}()

	// Process ResizeTerminalMsg from the client.
	resizeTermMsgChan := make(chan *protocol.ResizeTerminalMsg, 1)
	go func(uc *net.UnixConn) {
		for {
			msg, err := protocol.ReceiveMessage(uc)
			if errors.Is(err, io.EOF) {
				log.Printf("EOF received on client socket, sessionId=%d: %w\n", id, err)
				return
			} else if err != nil {
				log.Printf("error receiving message from client, sessionId=%d: %s\n", id, err)
				return
			}

			switch msg := msg.(type) {
			case *protocol.ResizeTerminalMsg:
				log.Printf("received ResizeTerminalMsg from client, sessionId=%d\n", id)
				resizeTermMsgChan <- msg
			default:
				log.Printf("unexpected message received from client, sessionId=%d\n", id)
			}
		}
	}(uc)

	// Main event loop.
	log.Printf("starting main event loop for sessionId=%d\n", id)
	for {
		select {
		case event := <-termEventChan:
			// TODO: process term event
			// process event -> input
			// if action, acquire lock on editor state and apply
		case msg := <-resizeTermMsgChan:
			// TODO: process resize terminal msg
			// if action, acquire lock on editor state and apply
		case <-s.quitChan:
			log.Printf("terminating sessionId=%d for server quit\n", id)
			return
		}

		// TODO: check quit flag, exit
		// TODO: redraw
	}
}

func receiveStartSessionMsg(uc *net.UnixConn) (*protocol.StartSessionMsg, error) {
	msg, err := protocol.ReceiveMessage(uc)
	if err != nil {
		return nil, fmt.Errorf("protocol.ReceiveMessage: %w", err)
	}

	startSessionMsg, ok := msg.(*protocol.StartSessionMsg)
	if !ok {
		return nil, errors.New("incorrect message type received, expected StartSessionMsg")
	}

	return startSessionMsg, nil
}
