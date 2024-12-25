package server

import (
	"net"
	"os"
)

// sessionId uniquely identifies a session.
type sessionId int

// session is the data associated with a connected client.
type session struct {
	id  sessionId
	uc  *net.UnixConn
	pty *os.File
}
