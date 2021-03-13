package input

import (
	"github.com/gdamore/tcell/v2"
)

// EventMatcher matches an input key event.
type EventMatcher struct {
	Wildcard bool      // If true, matches every input key.
	Key      tcell.Key // The kind of key to match (usually tcell.KeyRune)
	Rune     rune      // If the key is tcell.KeyRune, match this rune.
}

// Matches returns whether the input event is a match.
func (em *EventMatcher) Matches(event *tcell.EventKey) bool {
	if em.Wildcard {
		return true
	}

	if event.Key() != em.Key {
		return false
	}

	if event.Key() == tcell.KeyRune && event.Rune() != em.Rune {
		return false
	}

	return true
}
