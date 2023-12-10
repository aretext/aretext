package display

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/state"
)

func TestDrawEditor(t *testing.T) {
	testCases := []struct {
		name             string
		buildState       func() *state.EditorState
		expectedContents [][]rune
	}{
		{
			name: "normal mode",
			buildState: func() *state.EditorState {
				s := state.NewEditorState(10, 6, nil, nil)
				state.InsertRune(s, 'a')
				state.InsertRune(s, 'b')
				state.InsertRune(s, 'c')
				return s
			},
			expectedContents: [][]rune{
				{'a', 'b', 'c', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name: "menu mode",
			buildState: func() *state.EditorState {
				s := state.NewEditorState(10, 6, nil, nil)
				state.ShowMenu(s, state.MenuStyleCommand, nil)
				state.AppendRuneToMenuSearch(s, 'a')
				state.AppendRuneToMenuSearch(s, 'b')
				state.AppendRuneToMenuSearch(s, 'c')
				return s
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
			name: "search mode",
			buildState: func() *state.EditorState {
				s := state.NewEditorState(10, 6, nil, nil)
				state.StartSearch(s, state.SearchDirectionForward, state.SearchCompleteMoveCursorToMatch)
				state.AppendRuneToSearchQuery(s, 'a')
				state.AppendRuneToSearchQuery(s, 'b')
				state.AppendRuneToSearchQuery(s, 'c')
				return s
			},
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'/', 'a', 'b', 'c', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				state := tc.buildState()
				screenWidth, screenHeight := state.ScreenSize()
				s.SetSize(int(screenWidth), int(screenHeight))
				palette := NewPalette()
				DrawEditor(s, palette, state, "")
				s.Sync()
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}
