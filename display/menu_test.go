package display

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/state"
)

func TestDrawMenu(t *testing.T) {
	testCases := []struct {
		name             string
		buildMenu        func() *state.MenuState
		expectedContents [][]rune
	}{
		{
			name: "not visible",
			buildMenu: func() *state.MenuState {
				return &state.MenuState{}
			},
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "visible, initial state with prompt",
			buildMenu: func() *state.MenuState {
				editorState := state.NewEditorState(100, 100, nil)
				loadItems := func() []menu.Item { return nil }
				state.ShowMenu(editorState, state.MenuStyleCommand, loadItems)
				return editorState.Menu()
			},
			expectedContents: [][]rune{
				{':', 'c', 'o', 'm', 'm', 'a', 'n', 'd', ' ', ' '},
				{'─', '─', '─', '─', '─', '─', '─', '─', '─', '─'},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "visible, query with no results",
			buildMenu: func() *state.MenuState {
				editorState := state.NewEditorState(100, 100, nil)
				loadItems := func() []menu.Item { return nil }
				state.ShowMenu(editorState, state.MenuStyleCommand, loadItems)
				state.AppendRuneToMenuSearch(editorState, 'a')
				state.AppendRuneToMenuSearch(editorState, 'b')
				state.AppendRuneToMenuSearch(editorState, 'c')
				return editorState.Menu()
			},
			expectedContents: [][]rune{
				{':', 'a', 'b', 'c', ' ', ' ', ' ', ' ', ' ', ' '},
				{'─', '─', '─', '─', '─', '─', '─', '─', '─', '─'},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "visible, query with results, first selected",
			buildMenu: func() *state.MenuState {
				editorState := state.NewEditorState(100, 100, nil)
				loadItems := func() []menu.Item {
					return []menu.Item{
						{Name: "test first"},
						{Name: "test second"},
						{Name: "test third"},
					}
				}
				state.ShowMenu(editorState, state.MenuStyleCommand, loadItems)
				state.AppendRuneToMenuSearch(editorState, 't')
				return editorState.Menu()
			},
			expectedContents: [][]rune{
				{':', 't', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', '>', ' ', 't', 'e', 's', 't', ' ', 'f'},
				{' ', ' ', ' ', ' ', 't', 'e', 's', 't', ' ', 's'},
				{' ', ' ', ' ', ' ', 't', 'e', 's', 't', ' ', 't'},
				{'─', '─', '─', '─', '─', '─', '─', '─', '─', '─'},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "visible, query with many results, second-to-last selected",
			buildMenu: func() *state.MenuState {
				editorState := state.NewEditorState(100, 100, nil)
				loadItems := func() []menu.Item {
					return []menu.Item{
						{Name: "test 1"},
						{Name: "test 2"},
						{Name: "test 3"},
						{Name: "test 4"},
						{Name: "test 5"},
						{Name: "test 6"},
						{Name: "test 7"},
						{Name: "test 8"},
						{Name: "test 9"},
					}
				}
				state.ShowMenu(editorState, state.MenuStyleCommand, loadItems)
				state.AppendRuneToMenuSearch(editorState, 't')
				state.MoveMenuSelection(editorState, -2)
				return editorState.Menu()
			},
			expectedContents: [][]rune{
				{':', 't', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', 't', 'e', 's', 't', ' ', '5'},
				{' ', ' ', ' ', ' ', 't', 'e', 's', 't', ' ', '6'},
				{' ', ' ', ' ', ' ', 't', 'e', 's', 't', ' ', '7'},
				{' ', ' ', '>', ' ', 't', 'e', 's', 't', ' ', '8'},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(10, 6)
				menu := tc.buildMenu()
				DrawMenu(s, menu)
				s.Sync()
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}
