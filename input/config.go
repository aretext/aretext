package input

import (
	"github.com/aretext/aretext/exec"
)

// Config controls how user input is interpreted.
type Config struct {
	// InputMode is the current input mode of the editor.
	InputMode exec.InputMode

	// ScrollLines is the number of lines to scroll up or down with Ctrl-U / Ctrl-D.
	ScrollLines uint64

	// Set of directory names to hide from file search.
	DirNamesToHide map[string]struct{}
}
