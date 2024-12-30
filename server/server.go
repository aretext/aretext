package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
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

	// TODO: this is just POC of editor state.
	mu         sync.Mutex
	dummyState map[sessionId]int
}

// NewServer creates (but does not start) a new server with the given config.
func NewServer(config Config) *Server {
	return &Server{
		config:     config,
		quitChan:   make(chan struct{}),
		dummyState: make(map[sessionId]int),
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

	// TODO: optionally quit after N seconds if no clients connected.

	// Start listening for clients to connect.
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

	// Initialize editor state for this session.
	s.initializeEditorStateForSession(id)
	defer s.deleteEditorStateForSession(id)

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
			log.Printf("processing terminal event for sessionId=%d\n", id)
			s.processTermEvent(id, event)
		case msg := <-resizeTermMsgChan:
			log.Printf("processing resize terminal event for sessionId=%d\n", id)
			s.processResizeTerminalMsg(id, msg)
		case <-s.quitChan:
			log.Printf("terminating sessionId=%d for server quit\n", id)
			return
		}

		if s.getSessionState(id) == 0 {
			log.Printf("sessionId=%d was terminated, exiting session event loop", id)
			return
		}

		// TODO: how to broadcast draw to other clients in the same document...?
		s.draw(id, screen)
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

func (s *Server) initializeEditorStateForSession(id sessionId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dummyState[id] = 1
}

func (s *Server) deleteEditorStateForSession(id sessionId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.dummyState, id)
}

func (s *Server) processTermEvent(id sessionId, event tcell.Event) {
	eventKey, ok := event.(*tcell.EventKey)
	if !ok {
		return
	}

	if eventKey.Key() == tcell.KeyEscape {
		s.setSessionState(id, 0)
	} else if eventKey.Rune() == 'a' {
		s.setSessionState(id, 1)
	} else if eventKey.Rune() == 'b' {
		s.setSessionState(id, 2)
	}
}

func (s *Server) processResizeTerminalMsg(id sessionId, msg *protocol.ResizeTerminalMsg) {
	log.Printf("received terminal resize msg for sessionId=%d, width=%d, height=%d\n", id, msg.Width, msg.Height)
}

func (s *Server) draw(id sessionId, screen tcell.Screen) {
	ss := s.getSessionState(id)
	if ss == 1 {
		screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorRed))
	} else {
		screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlue))
	}
	screen.Clear()
}

func (s *Server) setSessionState(id sessionId, newState int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.dummyState[id]; ok {
		log.Printf("Setting state for session %d to %d", id, newState)
		s.dummyState[id] = newState
	} else {
		log.Printf("Error: cannot set state for session %d because it doesn't exist", id)
	}
}

func (s *Server) getSessionState(id sessionId) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dummyState[id]
}
