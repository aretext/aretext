package input

import (
	"fmt"

	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/syntax"
)

func menuItems(ctx Context) []menu.Item {
	// These items are available from both normal and visual mode.
	items := []menu.Item{
		{
			Name:    "quit",
			Aliases: []string{"q"},
			Action: func(s *state.EditorState) {
				state.AbortIfUnsavedChanges(s, state.Quit, true)
			},
		},
		{
			Name:    "force quit",
			Aliases: []string{"q!"},
			Action:  state.Quit,
		},
		{
			Name:    "save document",
			Aliases: []string{"s", "w"},
			Action: func(s *state.EditorState) {
				state.AbortIfFileExistsWithChangedContent(s, state.SaveDocument)
			},
		},
		{
			Name:    "save document and quit",
			Aliases: []string{"sq", "wq"},
			Action: func(s *state.EditorState) {
				state.AbortIfFileExistsWithChangedContent(s, func(s *state.EditorState) {
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
				state.AbortIfUnsavedChanges(s, ShowFileMenu(ctx), true)
			},
		},
		{
			Name:    "open previous document",
			Aliases: []string{"p"},
			Action: func(s *state.EditorState) {
				state.AbortIfUnsavedChanges(s, state.LoadPrevDocument, true)
			},
		},
		{
			Name:    "open next document",
			Aliases: []string{"n"},
			Action: func(s *state.EditorState) {
				state.AbortIfUnsavedChanges(s, state.LoadNextDocument, true)
			},
		},
		{
			Name:    "toggle show tabs",
			Aliases: []string{"ta"},
			Action:  state.ToggleShowTabs,
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
			Name:    "toggle auto-indent",
			Aliases: []string{"ai"},
			Action:  state.ToggleAutoIndent,
		},
	}

	for _, language := range syntax.AllLanguages {
		l := language // ensure each item refers to a different language
		items = append(items, menu.Item{
			Name: fmt.Sprintf("set syntax %s", language),
			Action: func(s *state.EditorState) {
				state.SetSyntax(s, l)
			},
		})
	}

	// User-defined macros are available only in normal mode, not visual mode.
	// This avoids problematic states where a macro gets recorded in one mode
	// and exected in another.
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
