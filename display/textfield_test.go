package display

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/state"
)

func buildTextFieldState(t *testing.T, promptText, inputText string) *state.TextFieldState {
	s, err := newEditorStateWithPath("test.txt")
	require.NoError(t, err)

	emptyAction := func(_ *state.EditorState, _ string) error { return nil }
	state.ShowTextField(s, promptText, emptyAction)
	for _, r := range inputText {
		state.AppendRuneToTextField(s, r)
	}

	return s.TextField()

}

func TestDrawTextField(t *testing.T) {
	testCases := []struct {
		name                string
		promptText          string
		inputText           string
		screenWidth         int
		screenHeight        int
		expectContents      [][]rune
		expectCursorVisible bool
		expectCursorCol     int
		expectCursorRow     int
	}{
		{
			name:         "prompt with no input",
			screenWidth:  15,
			screenHeight: 4,
			promptText:   "Test prompt:",
			inputText:    "",
			expectContents: [][]rune{
				{'T', 'e', 's', 't', ' ', 'p', 'r', 'o', 'm', 'p', 't', ':', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'─', '─', '─', '─', '─', '─', '─', '─', '─', '─', '─', '─', '─', '─', '─'},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
			expectCursorVisible: true,
			expectCursorCol:     0,
			expectCursorRow:     1,
		},
		{
			name:         "prompt with input",
			screenWidth:  15,
			screenHeight: 4,
			promptText:   "Test prompt:",
			inputText:    "foo.txt",
			expectContents: [][]rune{
				{'T', 'e', 's', 't', ' ', 'p', 'r', 'o', 'm', 'p', 't', ':', ' ', ' ', ' '},
				{'f', 'o', 'o', '.', 't', 'x', 't', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'─', '─', '─', '─', '─', '─', '─', '─', '─', '─', '─', '─', '─', '─', '─'},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
			expectCursorVisible: true,
			expectCursorCol:     7,
			expectCursorRow:     1,
		},
		{
			name:         "prompt with input, two rows no border",
			screenWidth:  15,
			screenHeight: 2,
			promptText:   "Test prompt:",
			inputText:    "foo.txt",
			expectContents: [][]rune{
				{'T', 'e', 's', 't', ' ', 'p', 'r', 'o', 'm', 'p', 't', ':', ' ', ' ', ' '},
				{'f', 'o', 'o', '.', 't', 'x', 't', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
			expectCursorVisible: true,
			expectCursorCol:     7,
			expectCursorRow:     1,
		},
		{
			name:         "prompt with input, single row no border or input",
			screenWidth:  15,
			screenHeight: 1,
			promptText:   "Test prompt:",
			inputText:    "foo.txt",
			expectContents: [][]rune{
				{'T', 'e', 's', 't', ' ', 'p', 'r', 'o', 'm', 'p', 't', ':', ' ', ' ', ' '},
			},
			expectCursorVisible: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(tc.screenWidth, tc.screenHeight)
				palette := NewPalette()
				textFieldState := buildTextFieldState(t, tc.promptText, tc.inputText)
				DrawTextField(s, palette, textFieldState)
				s.Sync()
				assertCellContents(t, s, tc.expectContents)
				cursorCol, cursorRow, cursorVisible := s.GetCursor()
				assert.Equal(t, tc.expectCursorVisible, cursorVisible)
				if tc.expectCursorVisible {
					assert.Equal(t, tc.expectCursorCol, cursorCol)
					assert.Equal(t, tc.expectCursorRow, cursorRow)
				}
			})
		})
	}
}
