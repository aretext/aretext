package display

import (
	"testing"

	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/text"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestDrawSearchQuery(t *testing.T) {
	testCases := []struct {
		name                string
		inputMode           state.InputMode
		query               string
		direction           text.ReadDirection
		expectContents      [][]rune
		expectCursorVisible bool
		expectCursorCol     int
		expectCursorRow     int
	}{
		{
			name:      "normal mode hides search query",
			inputMode: state.InputModeNormal,
			query:     "abcd1234",
			direction: text.ReadDirectionForward,
			expectContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:      "search mode with empty query",
			inputMode: state.InputModeSearch,
			query:     "",
			direction: text.ReadDirectionForward,
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
			inputMode: state.InputModeSearch,
			query:     "abcd",
			direction: text.ReadDirectionForward,
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
			inputMode: state.InputModeSearch,
			query:     "abcd1234",
			direction: text.ReadDirectionForward,
			expectContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' '},
				{'/', 'a', 'b', 'c', 'd', '1'},
			},
		},
		{
			name:      "search mode for backward search",
			inputMode: state.InputModeSearch,
			query:     "abcd",
			direction: text.ReadDirectionBackward,
			expectContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' '},
				{'?', 'a', 'b', 'c', 'd', ' '},
			},
			expectCursorVisible: true,
			expectCursorCol:     5,
			expectCursorRow:     1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(6, 2)
				DrawSearchQuery(s, tc.inputMode, tc.query, tc.direction)
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
