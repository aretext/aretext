package display

import (
	"testing"

	"github.com/aretext/aretext/exec"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestDrawSearchQuery(t *testing.T) {
	testCases := []struct {
		name                string
		inputMode           exec.InputMode
		query               string
		expectContents      [][]rune
		expectCursorVisible bool
		expectCursorCol     int
		expectCursorRow     int
	}{
		{
			name:      "normal mode hides search query",
			inputMode: exec.InputModeNormal,
			query:     "abcd1234",
			expectContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:      "search mode with empty query",
			inputMode: exec.InputModeSearch,
			query:     "",
			expectContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' '},
				{'/', ' ', ' ', ' ', ' ', ' '},
			},
			expectCursorVisible: true,
			expectCursorCol:     1,
			expectCursorRow:     1,
		},
		{
			name:      "search mode with non-empty query",
			inputMode: exec.InputModeSearch,
			query:     "abcd",
			expectContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' '},
				{'/', 'a', 'b', 'c', 'd', ' '},
			},
			expectCursorVisible: true,
			expectCursorCol:     5,
			expectCursorRow:     1,
		},
		{
			name:      "search mode with clipped query",
			inputMode: exec.InputModeSearch,
			query:     "abcd1234",
			expectContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' '},
				{'/', 'a', 'b', 'c', 'd', '1'},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(6, 2)
				DrawSearchQuery(s, tc.inputMode, tc.query)
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
