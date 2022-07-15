package locate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

func stringWithLen(n int) string {
	s := make([]rune, n)
	for i := 0; i < n; i++ {
		s[i] = 'x'
	}
	return string(s)
}

func TestViewOriginAfterScroll(t *testing.T) {
	testCases := []struct {
		name         string
		inputString  string
		cursorPos    uint64
		viewStartPos uint64
		viewWidth    uint64
		viewHeight   uint64
		expectedPos  uint64
	}{
		{
			name:         "empty",
			inputString:  "",
			cursorPos:    0,
			viewStartPos: 0,
			viewWidth:    10,
			viewHeight:   10,
			expectedPos:  0,
		},
		{
			name:         "single line",
			inputString:  "abcdefgh",
			cursorPos:    3,
			viewStartPos: 0,
			viewWidth:    10,
			viewHeight:   10,
			expectedPos:  0,
		},
		{
			name:         "multiple hard-wrapped lines, scroll up",
			inputString:  "ab\ncd\nef\ngh\nij\nkl\nmn\nop\nqr\nst\nuv",
			cursorPos:    3,
			viewStartPos: 9,
			viewWidth:    10,
			viewHeight:   10,
			expectedPos:  0,
		},
		{
			name:         "multiple hard-wrapped lines, scroll down",
			inputString:  "ab\ncd\nef\ngh\nij\nkl\nmn\nop\nqr\nst\nuv",
			cursorPos:    26,
			viewStartPos: 0,
			viewWidth:    10,
			viewHeight:   8,
			expectedPos:  12,
		},
		{
			name:         "multiple hard-wrapped lines, no scroll",
			inputString:  "ab\ncd\nef\ngh\nij\nkl\nmn\nop\nqr\nst\nuv",
			cursorPos:    26,
			viewStartPos: 12,
			viewWidth:    10,
			viewHeight:   8,
			expectedPos:  12,
		},
		{
			name:         "multiple soft-wrapped lines, scroll up",
			inputString:  "abcdefghijklmnopqrstuv",
			cursorPos:    2,
			viewStartPos: 19,
			viewWidth:    2,
			viewHeight:   8,
			expectedPos:  0,
		},
		{
			name:         "multiple soft-wrapped lines, scroll down",
			inputString:  "abcdefghijklmnopqrstuv",
			cursorPos:    20,
			viewStartPos: 0,
			viewWidth:    2,
			viewHeight:   8,
			expectedPos:  12,
		},
		{
			name:         "multiple soft-wrapped lines, no scroll",
			inputString:  "abcdefghijklmnopqrstuv",
			cursorPos:    15,
			viewStartPos: 6,
			viewWidth:    2,
			viewHeight:   8,
			expectedPos:  6,
		},
		{
			name:         "text ends with cursor on newline, scroll down",
			inputString:  "ab\ncd\nef\ngh\nij\nkl\nmn\nop\nqr\nst\n",
			cursorPos:    29,
			viewStartPos: 0,
			viewWidth:    10,
			viewHeight:   10,
			expectedPos:  9,
		},
		{
			name:         "text ends with cursor on newline short view height, scroll down",
			inputString:  "a\nb\nc\n",
			cursorPos:    6,
			viewStartPos: 0,
			viewWidth:    10,
			viewHeight:   3,
			expectedPos:  2,
		},
		{
			name:         "view height larger than text",
			inputString:  "abcd",
			cursorPos:    4,
			viewStartPos: 0,
			viewWidth:    2,
			viewHeight:   10,
			expectedPos:  0,
		},
		{
			name:         "view height smaller than scroll margins",
			inputString:  "abcdefghijklmnopqrstuv",
			cursorPos:    11,
			viewStartPos: 0,
			viewWidth:    2,
			viewHeight:   1,
			expectedPos:  10,
		},
		{
			name:         "very long text scroll up from middle",
			inputString:  stringWithLen(1024),
			cursorPos:    12,
			viewStartPos: 400,
			viewWidth:    2,
			viewHeight:   10,
			expectedPos:  6,
		},
		{
			name:         "very long text scroll down from middle",
			inputString:  stringWithLen(1024),
			cursorPos:    1024,
			viewStartPos: 400,
			viewWidth:    2,
			viewHeight:   10,
			expectedPos:  1010,
		},
		{
			name:         "very long text no scroll",
			inputString:  stringWithLen(1024),
			cursorPos:    410,
			viewStartPos: 400,
			viewWidth:    2,
			viewHeight:   10,
			expectedPos:  400,
		},
		{
			name:         "single visible line, scroll up",
			inputString:  stringWithLen(1024),
			cursorPos:    399,
			viewStartPos: 400,
			viewWidth:    10,
			viewHeight:   1,
			expectedPos:  390,
		},
		{
			name:         "single visible line, scroll down",
			inputString:  stringWithLen(1024),
			cursorPos:    411,
			viewStartPos: 400,
			viewWidth:    10,
			viewHeight:   1,
			expectedPos:  410,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			wrapConfig := segment.LineWrapConfig{
				MaxLineWidth: tc.viewWidth,
				WidthFunc: func(gc []rune, offsetInLine uint64) uint64 {
					return cellwidth.GraphemeClusterWidth(gc, offsetInLine, 4)
				},
			}
			updatedViewStartPos := ViewOriginAfterScroll(tc.cursorPos, tree, wrapConfig, tc.viewStartPos, tc.viewHeight)
			assert.Equal(t, tc.expectedPos, updatedViewStartPos)
		})
	}
}
