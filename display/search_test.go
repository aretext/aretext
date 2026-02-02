package display

import (
	"testing"

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
			WithMockScreen(t, func(s *MockScreen) {
				s.SetSize(6, 2)
				palette := NewPalette()
				DrawSearchQuery(s, palette, tc.query, tc.direction)
				s.Sync()
				s.AssertCellContents(t, tc.expectContents)
				s.AssertCursor(t, tc.expectCursorVisible, tc.expectCursorCol, tc.expectCursorRow)
			})
		})
	}
}
