package server

import (
	"github.com/gdamore/tcell"
)

type sessionStartedEvent struct {
	sessionId sessionId
	screen    tcell.Screen
	quitChan  chan struct{}
}

type clientDisconnectedEvent struct {
	sessionId sessionId
}

type terminalResizedEvent struct {
	sessionId sessionId
	width     int
	height    int
}

type terminalScreenEvent struct {
	sessionId sessionId
	termEvent tcell.Event
}
