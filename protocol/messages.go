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
	registerClientMsgType
	resizeTerminalMsgType
)

func msgTypeForMessage(msg Message) msgType {
	switch msg.(type) {
	case *RegisterClientMsg:
		return registerClientMsgType
	case *ResizeTerminalMsg:
		return resizeTerminalMsgType
	default:
		return invalidMsgType
	}
}

// RegisterClientMsg is sent from client to server immediately after opening a connection.
type RegisterClientMsg struct {
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

func (m *RegisterClientMsg) closed() {}

var _ Message = (*RegisterClientMsg)(nil)

// ResizeTerminalMsg is sent from a client to the server to signal that the terminal has been resized.
type ResizeTerminalMsg struct {
	// Width is the new width of the client's terminal.
	Width int

	// Height is the new height of the client's terminal.
	Height int
}

func (m *ResizeTerminalMsg) closed() {}

var _ Message = (*ResizeTerminalMsg)(nil)
