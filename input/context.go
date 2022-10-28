package input

import (
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/state"
)

// Context influences how user input is interpreted.
type Context struct {
	// InputMode is the current input mode of the editor.
	InputMode state.InputMode

	// ScrollLines is the number of lines to scroll up or down with Ctrl-F / Ctrl-N.
	// Ctrl-U / Ctrl-D scroll for half of that amount.
	ScrollLines uint64

	// Glob patterns for directories to hide from file search.
	DirPatternsToHide []string

	// Information about the current selection (visual mode).
	// If not in visual mode, the mode will be selection.ModeNone
	// and the end locator will be nil.
	SelectionMode       selection.Mode
	SelectionEndLocator state.Locator
}

func ContextFromEditorState(editorState *state.EditorState) Context {
	_, screenHeight := editorState.ScreenSize()
	scrollLines := uint64(screenHeight)
	return Context{
		InputMode:           editorState.InputMode(),
		ScrollLines:         scrollLines,
		DirPatternsToHide:   editorState.DirPatternsToHide(),
		SelectionMode:       editorState.DocumentBuffer().SelectionMode(),
		SelectionEndLocator: editorState.DocumentBuffer().SelectionEndLocator(),
	}
}
