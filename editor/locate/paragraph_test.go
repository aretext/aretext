package locate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/editor/text"
)

func TestNextParagraph(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		pos         uint64
		expectedPos uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "end of document",
			inputString: "abcd1234",
			pos:         2,
			expectedPos: 7,
		},
		{
			name:        "from non-empty line to first empty line",
			inputString: "ab\ncd\n\n\nef",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "from empty line to next empty line",
			inputString: "ab\n\n\n\ncd\n\nef",
			pos:         3,
			expectedPos: 9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := NextParagraph(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPrevParagraph(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		pos         uint64
		expectedPos uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "start of document",
			inputString: "abcd1234",
			pos:         4,
			expectedPos: 0,
		},
		{
			name:        "from non-empty line to previous empty line",
			inputString: "ab\n\n\ncd\n\nef",
			pos:         6,
			expectedPos: 4,
		},
		{
			name:        "from empty line to next empty line",
			inputString: "ab\n\n\n\ncd\n\nef",
			pos:         9,
			expectedPos: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			actualPos := PrevParagraph(textTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}
