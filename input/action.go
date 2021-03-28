package input

import (
	"github.com/aretext/aretext/state"
	"github.com/gdamore/tcell/v2"
)

// Action is a function that mutates the editor state.
type Action func(*state.EditorState)

// ActionBuilder is invoked when the input parser accepts a sequence of keypresses matching a rule.
type ActionBuilder func(inputEvents []*tcell.EventKey, count *int64, config Config) Action

// EmptyAction is an action that does nothing.
func EmptyAction(s *state.EditorState) {}
