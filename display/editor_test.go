package display

import (
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/require"

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
				s, err := newEditorStateWithPath("test.txt")
				require.NoError(t, err)
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
				{'t', 'e', 's', 't', '.', 't', 'x', 't', ' ', ' '},
			},
		},
		{
			name: "menu mode",
			buildState: func() *state.EditorState {
				s, err := newEditorStateWithPath("test.txt")
				require.NoError(t, err)
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
				{'t', 'e', 's', 't', '.', 't', 'x', 't', ' ', ' '},
			},
		},
		{
			name: "search mode",
			buildState: func() *state.EditorState {
				s, err := newEditorStateWithPath("test.txt")
				require.NoError(t, err)
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
		{
			name: "textfield mode",
			buildState: func() *state.EditorState {
				s, err := newEditorStateWithPath("test.txt")
				require.NoError(t, err)
				emptyAction := func(_ *state.EditorState, _ string) error { return nil }
				state.ShowTextField(s, "Test:", emptyAction)
				state.AppendRuneToTextField(s, 'a')
				state.AppendRuneToTextField(s, 'b')
				state.AppendRuneToTextField(s, 'c')
				return s
			},
			expectedContents: [][]rune{
				{'T', 'e', 's', 't', ':', ' ', ' ', ' ', ' ', ' '},
				{'a', 'b', 'c', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'─', '─', '─', '─', '─', '─', '─', '─', '─', '─'},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'t', 'e', 's', 't', '.', 't', 'x', 't', ' ', ' '},
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

func newEditorStateWithPath(path string) (*state.EditorState, error) {
	s := state.NewEditorState(10, 6, nil, nil)
	cursorLoc := func(p state.LocatorParams) uint64 { return 0 }
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	state.LoadDocument(s, absPath, false, cursorLoc)
	state.SetStatusMsg(s, state.StatusMsg{})
	return s, nil
}
