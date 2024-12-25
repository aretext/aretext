package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/aretext/aretext/protocol"
)

// Server listens on a Unix Domain Socket (UDS) for clients to connect.
// The client sends the server a pseudoterminal (pty), which the server uses
// for input/output from/to the client's terminal.
type Server struct {
	config                      Config
	sessions                    map[sessionId]session
	sessionStartedEventChan     chan sessionStartedEvent
	clientDisconnectedEventChan chan clientDisconnectedEvent
	terminalResizedEventChan    chan terminalResizedEvent
	terminalScreenEventChan     chan terminalScreenEvent
}

// NewServer creates (but does not start) a new server with the given config.
func NewServer(config Config) *Server {
	return &Server{
		config:                      config,
		sessions:                    make(map[sessionId]session),
		sessionStartedEventChan:     make(chan sessionStartedEvent, 1024),
		clientDisconnectedEventChan: make(chan clientDisconnectedEvent, 1024),
		terminalResizedEventChan:    make(chan terminalResizedEvent, 1024),
		terminalScreenEventChan:     make(chan terminalScreenEvent, 1024),
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

	// Start listening for clients to connect.
	// This will spawn goroutines to manage each connection, each of which send
	// events to clientEventChannels for processing by the main event loop.
	go s.listenForConnections(ul)

	// Run the main event loop in the current goroutine.
	return s.runMainEventLoop()
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

func (s *Server) listenForConnections(ul *net.UnixListener) {
	nextSessionId := sessionId(0)
	for {
		uc, err := ul.AcceptUnix()
		if err != nil {
			log.Printf("net.AcceptUnix: %s\n", err)
			continue
		}

		go s.handleConnection(nextConnectionId, uc)
		nextConnectionId++
	}
}

func (s *Server) handleConnection(id sessionId, uc *net.UnixConn) {
	log.Printf("client connected, sessionId=%d\n", sessionId)
	defer uc.Close()

	msg, err := receiveRegisterClientMsg(uc)
	if err != nil {
		log.Printf("error registering client, sessionId=%d: %s\n", sessionId, err)
	}

	// TODO: construct a screen for the session
	// and defer screen.Fini()

	notifyClientDisconnected := func() {
		s.clientDisconnectedEventChan <- clientDisconnectedEvent{sessionId: id}
	}

	// Process ResizeTerminalMsg from the client.
	go func(uc *net.UnixConn) {
		defer notifyClientDisconnected()
		for {
			msg, err := protocol.ReceiveMessage(uc)
			if errors.Is(err, io.EOF) {
				log.Printf("EOF received on client socket, sessionId=%d: %w\n", sessionId, err)
				return
			} else if err != nil {
				log.Printf("error receiving message from client: %s\n", err)
				return
			}

			switch msg := msg.(type) {
			case *protocol.ResizeTerminalMsg:
				s.terminalResizedEventChan <- terminalResizedEvent{
					sessionId: id,
					width:     msg.Width,
					height:    msg.Height,
				}
			default:
				log.Printf("unexpected message received from client, sessionId=%d\n", id)
			}
		}
	}(uc)

	// Process pty terminal events from the client.
	go func() {
		defer notifyClientDisconnected()
		for {
			termEvent := screen.PollEvent()
			if termEvent == nil {
				log.Printf("screen.PollEvent returned nil, screen must be finalized, sessionId=%d\n", id)
				return
			}
			s.terminalScreenEventChan <- terminalScreenEvent{
				sessionId: id,
				termEvent: termEvent,
			}
		}
	}()

	// Notify main event loop that new session has started.
	quitChan := make(chan struct{})
	s.sessionStartedEventChan <- sessionStartedEvent{
		sessionId: id,
		screen:    screen,
		quitChan:  quitChan,
	}

	// Wait for quit signal, then cleanup.
	// Deferred cleanup will close the Unix socket and finalize the screen, which will cause
	// the above goroutines to exit.
	select {
	case <-quitChan:
		log.Printf("quitting session %d\n", sessionId)
	}
}

func receiveRegisterClientMsg(uc *net.UnixConn) (*protocol.RegisterClientMsg, error) {
	msg, err := protocol.ReceiveMessage(uc)
	if err != nil {
		return nil, fmt.Errorf("protocol.ReceiveMessage: %w", err)
	}

	registerClientMsg, ok := msg.(*protocol.RegisterClientMsg)
	if !ok {
		return nil, errors.New("incorrect message type received, expected RegisterClientMsg")
	}

	return registerClientMsg, nil
}

func (s *Server) runMainEventLoop() error {
	for {
		select {
		case event := <-s.sessionStartedEventChan:
			// TODO: register new session

		case event := <-s.clientDisconnectedEventChan:
			// TODO: remove session if it exists

		case event := <-s.terminalResizedEventChan:
			// TODO: update screen size

		case event := <-s.terminalScreenEventChan:
			// TODO: process terminal event (just log for now...)
		}
	}
}
