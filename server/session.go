package server

import (
	"os"

	"github.com/aretext/aretext/protocol"
)

type clientId int

type connectionState int

const (
	connectionStateRegistering = connectionState(iota)
	connectionStateActive
	connectionStateTerminating
)

type connection struct {
	clientId clientId
	state    connectionState
	pty      *os.File
	quitChan chan struct{}
}

type clientMsg struct {
	clientId clientId
	msg      protocol.Message
}
