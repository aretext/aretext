package display

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/state"
)

func buildTextFieldState(t *testing.T, promptText, inputText string) *state.TextFieldState {
	s, err := newEditorStateWithPath("test.txt")
	require.NoError(t, err)

	emptyAction := func(_ *state.EditorState, _ string) error { return nil }
	state.ShowTextField(s, promptText, emptyAction, nil)
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
		expectContents      [][]string
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
			expectContents: [][]string{
				{"T", "e", "s", "t", " ", "p", "r", "o", "m", "p", "t", ":", " ", " ", " "},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
				{"─", "─", "─", "─", "─", "─", "─", "─", "─", "─", "─", "─", "─", "─", "─"},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
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
			expectContents: [][]string{
				{"T", "e", "s", "t", " ", "p", "r", "o", "m", "p", "t", ":", " ", " ", " "},
				{"f", "o", "o", ".", "t", "x", "t", " ", " ", " ", " ", " ", " ", " ", " "},
				{"─", "─", "─", "─", "─", "─", "─", "─", "─", "─", "─", "─", "─", "─", "─"},
				{" ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
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
			expectContents: [][]string{
				{"T", "e", "s", "t", " ", "p", "r", "o", "m", "p", "t", ":", " ", " ", " "},
				{"f", "o", "o", ".", "t", "x", "t", " ", " ", " ", " ", " ", " ", " ", " "},
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
			expectContents: [][]string{
				{"T", "e", "s", "t", " ", "p", "r", "o", "m", "p", "t", ":", " ", " ", " "},
			},
			expectCursorVisible: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withMockScreen(t, vt.MockOptSize{X:vt.Col(tc.screenWidth), Y: vt.Row(tc.screenHeight)}, func(s tcell.Screen, b vt.MockBackend) {
				palette := NewPalette()
				textFieldState := buildTextFieldState(t, tc.promptText, tc.inputText)
				DrawTextField(s, palette, textFieldState)

				s.Sync()
				assertCellContents(t, b, tc.expectContents)

				cursorStyle := b.GetCursor()
				assert.Equal(t, tc.expectCursorVisible, cursorStyle.IsVisible())
				if tc.expectCursorVisible {
					cursorPos := b.GetPosition()
					assert.Equal(t, tc.expectCursorCol, cursorPos.X)
					assert.Equal(t, tc.expectCursorRow, cursorPos.Y)
				}
			})
		})
	}
}
