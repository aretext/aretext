package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/editor/text"
)

func TestScrollViewByNumLines(t *testing.T) {
	testCases := []struct {
		name               string
		inputString        string
		initialView        viewState
		direction          ScrollDirection
		numLines           uint64
		expectedtextOrigin uint64
	}{
		{
			name:               "empty, scroll up",
			inputString:        "",
			initialView:        viewState{textOrigin: 0, height: 100, width: 100},
			direction:          ScrollDirectionBackward,
			numLines:           1,
			expectedtextOrigin: 0,
		},
		{
			name:               "empty, scroll down",
			inputString:        "",
			initialView:        viewState{textOrigin: 0, height: 100, width: 100},
			direction:          ScrollDirectionForward,
			numLines:           1,
			expectedtextOrigin: 0,
		},
		{
			name:               "scroll up",
			inputString:        "ab\ncd\nef\ngh\nij\nkl\nmn",
			initialView:        viewState{textOrigin: 12, height: 2, width: 100},
			direction:          ScrollDirectionBackward,
			numLines:           3,
			expectedtextOrigin: 3,
		},
		{
			name:               "scroll down",
			inputString:        "ab\ncd\nef\ngh\nij\nkl\nmn",
			initialView:        viewState{textOrigin: 3, height: 2, width: 100},
			direction:          ScrollDirectionForward,
			numLines:           3,
			expectedtextOrigin: 12,
		},
		{
			name:               "scroll down to last line",
			inputString:        "ab\ncd\nef\ngh\nij\nkl\nmn",
			initialView:        viewState{textOrigin: 0, height: 6, width: 100},
			numLines:           10,
			expectedtextOrigin: 12,
		},
		{
			name:               "scroll down view taller than document",
			inputString:        "ab\ncd\nef\ngh",
			initialView:        viewState{textOrigin: 0, height: 100, width: 100},
			numLines:           1,
			expectedtextOrigin: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.view = tc.initialView
			ScrollViewByNumLines(state, tc.direction, tc.numLines)
			assert.Equal(t, tc.expectedtextOrigin, state.documentBuffer.view.textOrigin)
		})
	}
}
