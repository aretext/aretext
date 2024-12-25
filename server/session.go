package server

import "github.com/gdamore/tcell"

// sessionId uniquely identifies a session.
type sessionId int

// session is the data associated with a connected client.
type session struct {
	id       sessionId
	screen   tcell.Screen
	quitChan chan struct{}
}
