package protocol

// ClientId is a unique identifier for a client assigned by the server on connection.
type ClientId int

// MsgType encodes the type of message.
type MsgType int

const (
	InvalidMsgType = MsgType(iota) // Zero value is invalid.
	ClientHelloMsgType
	ClientGoodbyeMsgType
	ServerHelloMsgType
	ServerGoodbyeMsgType
	TerminalResizeMsgType
)

// Message is a serializable message sent between a client and a server.
type Message struct {
	MsgType MsgType
}

// ClientHelloMsg is sent from client to server immediately after opening a connection.
type ClientHelloMsg struct {
	Message

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

// ClientGoodbyeMsg is sent from a client to gracefully terminate a connection to the server.
type ClientGoodbyeMsg struct {
	Message

	// Reason indicates the reason why the client terminated the connection.
	Reason string
}

// ServerHelloMsg is sent from a server after receiving ClientHello.
// This tells the client that it has successfully connected to a server.
type ServerHelloMsg struct {
	Message

	// ClientId is the unique ID the server has assigned to the client.
	ClientId ClientId
}

// ServerGoodbyeMsg is sent from a server to gracefully terminate a connection to the client.
type ServerGoodbyeMsg struct {
	Message

	// Reason indicates the reason why the server terminated the connection.
	Reason string
}

// TerminalResizeMsg is sent from a client to the server to signal that the terminal has been resized.
type TerminalResizeMsg struct {
	Message

	// Width is the new width of the client's terminal.
	Width int

	// Height is the new height of the client's terminal.
	Height int
}
