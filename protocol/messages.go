package protocol

// ClientId is a unique identifier for a client assigned by the server on connection.
type ClientId int

// Message represents a serializable message sent between a client and a server.
type Message interface{}

// MsgType encodes the type of message.
type msgType int

const (
	invalidMsgType = MsgType(iota) // Zero value is invalid.
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
// Along with this message, the client sends out-of-band SCM_RIGHTS with the file descriptor for the pts.
type ClientHelloMsg struct {
	// FilePath is the path to the initial file to open in the editor.
	// May be empty to create a new untitled document.
	FilePath string

	// WorkingDir is the initial working directory of the client.
	// The server uses this when searching for files or executing shell commands on behalf of a client.
	WorkingDir string

	// TerminalEnv is the environment variables for the client's terminal ($TERM, etc.)
	// The server should use these when interacting with the client's delegated tty
	// or when executing shell commands on behalf of a client.
	TerminalEnv []string
}

func (m ClientHelloMsg) MsgType() MsgType {
	return ClientHelloMsgType
}

var _ Message = (*ClientHelloMsg)(nil)

// ClientGoodbyeMsg is sent from a client to gracefully terminate a connection to the server.
type ClientGoodbyeMsg struct {
	// Reason indicates the reason why the client terminated the connection.
	Reason string
}

func (m ClientGoodbyeMsg) MsgType() MsgType {
	return ClientGoodbyeMsg
}

var _ Message = (*ClientGoodbyeMsg)(nil)

// ServerHelloMsg is sent from a server after receiving ClientHello.
// This tells the client that it has successfully connected to a server.
type ServerHelloMsg struct {
	// ClientId is the unique ID the server has assigned to the client.
	ClientId ClientId
}

func (m ServerHelloMsg) MsgType() MsgType {
	return ServerHelloMsgType
}

var _ Message = (*ServerHelloMsg)(nil)

// ServerGoodbyeMsg is sent from a server to gracefully terminate a connection to the client.
type ServerGoodbyeMsg struct {
	// Reason indicates the reason why the server terminated the connection.
	Reason string
}

func (m ServerGoodbyeMsg) MsgType() MsgType {
	return ServerGoodbyeMsgType
}

var _ Message = (*ServerGoodbyeMsg)(nil)

// TerminalResizeMsg is sent from a client to the server to signal that the terminal has been resized.
type TerminalResizeMsg struct {
	// Width is the new width of the client's terminal.
	Width int

	// Height is the new height of the client's terminal.
	Height int
}

func (m TerminalResizeMsg) MsgType() MsgType {
	return TerminalResizeMsgType
}

var _ Message = (*TerminalResizeMsg)(nil)
