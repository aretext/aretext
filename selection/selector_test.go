package selection

import (
	"testing"

	"github.com/aretext/aretext/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelector(t *testing.T) {
	testCases := []struct {
		name             string
		inputString      string
		mode             Mode
		initialCursorPos uint64
		finalCursorPos   uint64
		expectedRegion   Region
	}{
		{
			name:             "no selection",
			inputString:      "abcd",
			mode:             ModeNone,
			initialCursorPos: 1,
			finalCursorPos:   3,
			expectedRegion:   Region{},
		},
		{
			name:             "charwise, no cursor movement",
			inputString:      "abcdefghijklmnop",
			mode:             ModeChar,
			initialCursorPos: 3,
			finalCursorPos:   3,
			expectedRegion: Region{
				StartPos: 3,
				EndPos:   4,
			},
		},
		{
			name:             "charwise, cursor movement forward",
			inputString:      "abcdefghijklmnop",
			mode:             ModeChar,
			initialCursorPos: 3,
			finalCursorPos:   6,
			expectedRegion: Region{
				StartPos: 3,
				EndPos:   7,
			},
		},
		{
			name:             "charwise, cursor movement forward to end of document",
			inputString:      "abcd",
			mode:             ModeChar,
			initialCursorPos: 2,
			finalCursorPos:   4,
			expectedRegion: Region{
				StartPos: 2,
				EndPos:   4,
			},
		},
		{
			name:             "charwise, cursor movement backward",
			inputString:      "abcdefghijklmnop",
			mode:             ModeChar,
			initialCursorPos: 6,
			finalCursorPos:   3,
			expectedRegion: Region{
				StartPos: 3,
				EndPos:   7,
			},
		},
		{
			name:             "charwise, cursor movement backward to start of document",
			inputString:      "abcdefghijklmnop",
			mode:             ModeChar,
			initialCursorPos: 6,
			finalCursorPos:   0,
			expectedRegion: Region{
				StartPos: 0,
				EndPos:   7,
			},
		},
		{
			name:             "charwise select last empty line",
			inputString:      "abc\n",
			mode:             ModeChar,
			initialCursorPos: 4,
			finalCursorPos:   4,
			expectedRegion: Region{
				StartPos: 4,
				EndPos:   4,
			},
		},
		{
			name:             "linewise, no cursor movement",
			inputString:      "abcd\nefgh\nijkl\nmnop",
			mode:             ModeLine,
			initialCursorPos: 6,
			finalCursorPos:   6,
			expectedRegion: Region{
				StartPos: 5,
				EndPos:   9,
			},
		},
		{
			name:             "linewise, cursor movement in same line",
			inputString:      "abcd\nefgh\nijkl\nmnop",
			mode:             ModeLine,
			initialCursorPos: 6,
			finalCursorPos:   8,
			expectedRegion: Region{
				StartPos: 5,
				EndPos:   9,
			},
		},
		{
			name:             "linewise, cursor movement to line below",
			inputString:      "abcd\nefgh\nijkl\nmnop",
			mode:             ModeLine,
			initialCursorPos: 6,
			finalCursorPos:   15,
			expectedRegion: Region{
				StartPos: 5,
				EndPos:   19,
			},
		},
		{
			name:             "linewise, cursor movement to line above",
			inputString:      "abcd\nefgh\nijkl\nmnop",
			mode:             ModeLine,
			initialCursorPos: 15,
			finalCursorPos:   6,
			expectedRegion: Region{
				StartPos: 5,
				EndPos:   19,
			},
		},
		{
			name:             "linewise, cursor movement to blank line",
			inputString:      "abcd\n\n\nefgh",
			mode:             ModeLine,
			initialCursorPos: 2,
			finalCursorPos:   5,
			expectedRegion: Region{
				StartPos: 0,
				EndPos:   5,
			},
		},
		{
			name:             "linewise select last empty line",
			inputString:      "abc\n",
			mode:             ModeLine,
			initialCursorPos: 4,
			finalCursorPos:   4,
			expectedRegion: Region{
				StartPos: 4,
				EndPos:   4,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			var s Selector
			s.Start(tc.mode, tc.initialCursorPos)
			r := s.Region(tree, tc.finalCursorPos)
			assert.Equal(t, tc.expectedRegion, r)
		})
	}
}
