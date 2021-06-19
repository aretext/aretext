package input

import (
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/state"
)

// Config controls how user input is interpreted.
type Config struct {
	// InputMode is the current input mode of the editor.
	InputMode state.InputMode

	// ScrollLines is the number of lines to scroll up or down with Ctrl-U / Ctrl-D.
	ScrollLines uint64

	// Glob patterns for directories to hide from file search.
	DirPatternsToHide []string

	// Information about the current selection (visual mode).
	// If not in visual mode, the mode will be selection.ModeNone
	// and the end locator will be nil.
	SelectionMode       selection.Mode
	SelectionEndLocator state.Locator
}
