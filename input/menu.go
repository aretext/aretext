package input

import (
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/state"
)

func menuItems(ctx Context) []menu.Item {
	// These items are available from both normal and visual mode.
	items := []menu.Item{
		{
			Name:    "quit",
			Aliases: []string{"q"},
			Action: func(s *state.EditorState) {
				abortMsg := `Document has unsaved changes. Either save them ("force save") or quit without saving ("force quit")`
				state.AbortIfUnsavedChanges(s, abortMsg, state.Quit)
			},
		},
		{
			Name:    "force quit",
			Aliases: []string{"q!"},
			Action:  state.Quit,
		},
		{
			Name:   "new document",
			Action: ShowNewDocumentTextField,
		},
		{
			Name:   "move or rename document",
			Action: ShowMoveOrRenameDocumentTextField,
		},
		{
			Name:    "save document",
			Aliases: []string{"s", "w"},
			Action: func(s *state.EditorState) {
				state.AbortIfFileChanged(s, state.SaveDocument)
			},
		},
		{
			Name:    "save document and quit",
			Aliases: []string{"sq", "wq", "x"},
			Action: func(s *state.EditorState) {
				state.AbortIfFileChanged(s, func(s *state.EditorState) {
					state.SaveDocument(s)
					state.Quit(s)
				})
			},
		},
		{
			Name:    "force save document",
			Aliases: []string{"s!", "w!"},
			Action:  state.SaveDocument,
		},
		{
			Name:    "force save document and quit",
			Aliases: []string{"sq!", "wq!"},
			Action: func(s *state.EditorState) {
				state.SaveDocument(s)
				state.Quit(s)
			},
		},
		{
			Name:    "force reload",
			Aliases: []string{"r!"},
			Action:  state.ReloadDocument,
		},
		{
			Name:    "find and open",
			Aliases: []string{"f"},
			Action: func(s *state.EditorState) {
				state.AbortIfUnsavedChanges(s, state.DefaultUnsavedChangesAbortMsg, ShowFileMenu(ctx))
			},
		},
		{
			Name:    "open previous document",
			Aliases: []string{"p"},
			Action: func(s *state.EditorState) {
				state.AbortIfUnsavedChanges(s, state.DefaultUnsavedChangesAbortMsg, state.LoadPrevDocument)
			},
		},
		{
			Name:    "open next document",
			Aliases: []string{"n"},
			Action: func(s *state.EditorState) {
				state.AbortIfUnsavedChanges(s, state.DefaultUnsavedChangesAbortMsg, state.LoadNextDocument)
			},
		},
		{
			Name:    "child directory",
			Aliases: []string{"cd"},
			Action: func(s *state.EditorState) {
				state.ShowChildDirsMenu(s, ctx.DirPatternsToHide)
			},
		},
		{
			Name:    "parent directory",
			Aliases: []string{"pd"},
			Action:  state.ShowParentDirsMenu,
		},
		{
			Name:    "toggle show tabs",
			Aliases: []string{"ta"},
			Action:  state.ToggleShowTabs,
		},
		{
			Name:    "toggle show spaces",
			Aliases: []string{"sp"},
			Action:  state.ToggleShowSpaces,
		},
		{
			Name:    "toggle tab expand",
			Aliases: []string{"te"},
			Action:  state.ToggleTabExpand,
		},
		{
			Name:    "toggle line numbers",
			Aliases: []string{"nu"},
			Action:  state.ToggleShowLineNumbers,
		},
		{
			Name:    "toggle line number mode (relative/absolute)",
			Aliases: []string{"nur"},
			Action:  state.ToggleLineNumberMode,
		},
		{
			Name:    "toggle auto-indent",
			Aliases: []string{"ai"},
			Action:  state.ToggleAutoIndent,
		},
	}

	// User-defined macros are available only in normal mode, not visual mode.
	// This avoids problematic states where a macro gets recorded in one mode
	// and executed in another.
	if ctx.InputMode == state.InputModeNormal {
		items = append(items, []menu.Item{
			{
				Name:    "start/stop recording macro",
				Aliases: []string{"m"},
				Action:  state.ToggleUserMacroRecording,
			},
			{
				Name:    "replay macro",
				Aliases: []string{"r"},
				Action:  state.ReplayRecordedUserMacro,
			},
		}...)
	}

	return items
}
