package display

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/editor/state"
)

func TestDrawSearchQuery(t *testing.T) {
	testCases := []struct {
		name                string
		query               string
		direction           state.SearchDirection
		expectContents      [][]rune
		expectCursorVisible bool
		expectCursorCol     int
		expectCursorRow     int
	}{
		{
			name:      "empty query",
			query:     "",
			direction: state.SearchDirectionForward,
			expectContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' '},
				{'/', ' ', ' ', ' ', ' ', ' '},
			},
			expectCursorVisible: true,
			expectCursorCol:     1,
			expectCursorRow:     1,
		},
		{
			name:      "non-empty query",
			query:     "abcd",
			direction: state.SearchDirectionForward,
			expectContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' '},
				{'/', 'a', 'b', 'c', 'd', ' '},
			},
			expectCursorVisible: true,
			expectCursorCol:     5,
			expectCursorRow:     1,
		},
		{
			name:      "clipped query",
			query:     "abcd1234",
			direction: state.SearchDirectionForward,
			expectContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' '},
				{'/', 'a', 'b', 'c', 'd', '1'},
			},
		},
		{
			name:      "backward search",
			query:     "abcd",
			direction: state.SearchDirectionBackward,
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
				palette := NewPalette()
				DrawSearchQuery(s, palette, tc.query, tc.direction)
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
