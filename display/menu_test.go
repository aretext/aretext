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
		expectedContents [][]string
	}{
		{
			name: "initial state with prompt",
			buildMenu: func() *state.MenuState {
				editorState := state.NewEditorState(100, 100, nil, nil)
				state.ShowMenu(editorState, state.MenuStyleCommand, nil)
				return editorState.Menu()
			},
			expectedContents: [][]string{
				{":", "c", "o", "m", "m", "a", "n", "d", " ", " "},
				{"─", "─", "─", "─", "─", "─", "─", "─", "─", "─"},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			},
		},
		{
			name: "query with no results",
			buildMenu: func() *state.MenuState {
				editorState := state.NewEditorState(100, 100, nil, nil)
				state.ShowMenu(editorState, state.MenuStyleCommand, nil)
				state.AppendRuneToMenuSearch(editorState, 'a')
				state.AppendRuneToMenuSearch(editorState, 'b')
				state.AppendRuneToMenuSearch(editorState, 'c')
				return editorState.Menu()
			},
			expectedContents: [][]string{
				{":", "a", "b", "c", " ", " ", " ", " ", " ", " "},
				{"─", "─", "─", "─", "─", "─", "─", "─", "─", "─"},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			},
		},
		{
			name: "query with results, first selected",
			buildMenu: func() *state.MenuState {
				editorState := state.NewEditorState(100, 100, nil, nil)
				items := []menu.Item{
					{Name: "test 1"},
					{Name: "test 2"},
					{Name: "test 3"},
				}
				state.ShowMenu(editorState, state.MenuStyleCommand, items)
				state.AppendRuneToMenuSearch(editorState, 't')
				return editorState.Menu()
			},
			expectedContents: [][]string{
				{":", "t", " ", " ", " ", " ", " ", " ", " ", " "},
				{" ", " ", ">", " ", "t", "e", "s", "t", " ", "1"},
				{" ", " ", " ", " ", "t", "e", "s", "t", " ", "2"},
				{" ", " ", " ", " ", "t", "e", "s", "t", " ", "3"},
				{"─", "─", "─", "─", "─", "─", "─", "─", "─", "─"},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			},
		},
		{
			name: "query with many results, second-to-last selected",
			buildMenu: func() *state.MenuState {
				editorState := state.NewEditorState(100, 100, nil, nil)
				items := []menu.Item{
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
				state.ShowMenu(editorState, state.MenuStyleCommand, items)
				state.AppendRuneToMenuSearch(editorState, 't')
				state.MoveMenuSelection(editorState, -2)
				return editorState.Menu()
			},
			expectedContents: [][]string{
				{":", "t", " ", " ", " ", " ", " ", " ", " ", " "},
				{" ", " ", " ", " ", "t", "e", "s", "t", " ", "5"},
				{" ", " ", " ", " ", "t", "e", "s", "t", " ", "6"},
				{" ", " ", " ", " ", "t", "e", "s", "t", " ", "7"},
				{" ", " ", ">", " ", "t", "e", "s", "t", " ", "8"},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(10, 6)
				palette := NewPalette()
				menu := tc.buildMenu()
				DrawMenu(s, palette, menu)
				s.Sync()
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}
