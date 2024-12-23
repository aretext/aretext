package protocol

import "os"

// ClientId is a unique identifier for a client assigned by the server on connection.
type ClientId int

// Message represents a serializable message sent between a client and a server.
type Message interface {
	closed() // Prevent other implementations.
}

// msgType encodes the type of message.
type msgType int

const (
	invalidMsgType = msgType(iota) // Zero value is invalid.
	clientHelloMsgType
	clientGoodbyeMsgType
	serverHelloMsgType
	serverGoodbyeMsgType
	terminalResizeMsgType
)

func msgTypeForMessage(msg Message) msgType {
	switch msg.(type) {
	case *ClientHelloMsg:
		return clientHelloMsgType
	case *ClientGoodbyeMsg:
		return clientGoodbyeMsgType
	case *ServerHelloMsg:
		return serverHelloMsgType
	case *ServerGoodbyeMsg:
		return serverGoodbyeMsgType
	case *TerminalResizeMsg:
		return terminalResizeMsgType
	default:
		return invalidMsgType
	}
}

// ClientHelloMsg is sent from client to server immediately after opening a connection.
type ClientHelloMsg struct {
	// DocumentPath is the path to the initial file to open in the editor.
	// May be empty to create a new untitled document.
	DocumentPath string

	// WorkingDir is the initial working directory of the client.
	// The server uses this when searching for files or executing shell commands on behalf of a client.
	WorkingDir string

	// TerminalEnv is the environment variables for the client's terminal ($TERM, etc.)
	// The server should use these when interacting with the client's delegated tty
	// or when executing shell commands on behalf of a client.
	TerminalEnv []string

	// Pts is a pseudoterminal delegated by the client to the server.
	// This is sent as out-of-band data SCM_RIGHTS over the Unix socket.
	Pts *os.File `json:"-"`
}

func (m *ClientHelloMsg) closed() {}

var _ Message = (*ClientHelloMsg)(nil)

// ClientGoodbyeMsg is sent from a client to gracefully terminate a connection to the server.
type ClientGoodbyeMsg struct {
	// Reason indicates the reason why the client terminated the connection.
	Reason string
}

func (m *ClientGoodbyeMsg) closed() {}

var _ Message = (*ClientGoodbyeMsg)(nil)

// ServerHelloMsg is sent from a server after receiving ClientHello.
// This tells the client that it has successfully connected to a server.
type ServerHelloMsg struct {
	// ClientId is the unique ID the server has assigned to the client.
	ClientId ClientId
}

func (m *ServerHelloMsg) closed() {}

var _ Message = (*ServerHelloMsg)(nil)

// ServerGoodbyeMsg is sent from a server to gracefully terminate a connection to the client.
type ServerGoodbyeMsg struct {
	// Reason indicates the reason why the server terminated the connection.
	Reason string
}

func (m *ServerGoodbyeMsg) closed() {}

var _ Message = (*ServerGoodbyeMsg)(nil)

// TerminalResizeMsg is sent from a client to the server to signal that the terminal has been resized.
type TerminalResizeMsg struct {
	// Width is the new width of the client's terminal.
	Width int

	// Height is the new height of the client's terminal.
	Height int
}

func (m *TerminalResizeMsg) closed() {}

var _ Message = (*TerminalResizeMsg)(nil)
