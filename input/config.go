package input

import (
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
}
