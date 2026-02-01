package display

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/state"
)

func TestDrawSearchQuery(t *testing.T) {
	testCases := []struct {
		name                string
		query               string
		direction           state.SearchDirection
		expectContents      [][]string
		expectCursorVisible bool
		expectCursorCol     int
		expectCursorRow     int
	}{
		{
			name:      "empty query",
			query:     "",
			direction: state.SearchDirectionForward,
			expectContents: [][]string{
				{" ", " ", " ", " ", " ", " "},
				{"/", " ", " ", " ", " ", " "},
			},
			expectCursorVisible: true,
			expectCursorCol:     1,
			expectCursorRow:     1,
		},
		{
			name:      "non-empty query",
			query:     "abcd",
			direction: state.SearchDirectionForward,
			expectContents: [][]string{
				{" ", " ", " ", " ", " ", " "},
				{"/", "a", "b", "c", "d", " "},
			},
			expectCursorVisible: true,
			expectCursorCol:     5,
			expectCursorRow:     1,
		},
		{
			name:      "clipped query",
			query:     "abcd1234",
			direction: state.SearchDirectionForward,
			expectContents: [][]string{
				{" ", " ", " ", " ", " ", " "},
				{"/", "a", "b", "c", "d", "1"},
			},
		},
		{
			name:      "backward search",
			query:     "abcd",
			direction: state.SearchDirectionBackward,
			expectContents: [][]string{
				{" ", " ", " ", " ", " ", " "},
				{"?", "a", "b", "c", "d", " "},
			},
			expectCursorVisible: true,
			expectCursorCol:     5,
			expectCursorRow:     1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withMockScreen(t, vt.MockOptSize{X: 6, Y: 2}, func(s tcell.Screen, b vt.MockBackend) {
				palette := NewPalette()
				DrawSearchQuery(s, palette, tc.query, tc.direction)
				s.Sync()
				assertCellContents(t, b, tc.expectContents)
				cursorPos := b.GetPosition(); cursorStyle := b.GetCursor(); cursorVisible := cursorStyle.IsVisible()
				assert.Equal(t, tc.expectCursorVisible, cursorVisible)
				if tc.expectCursorVisible {
					assert.Equal(t, tc.expectCursorCol, int(cursorPos.X))
					assert.Equal(t, tc.expectCursorRow, int(cursorPos.Y))
				}
			})
		})
	}
}
