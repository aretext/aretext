package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToggleShowLineNumbers(t *testing.T) {
	testCases := []struct {
		name                string
		screenWidth         uint64
		numLines            int
		numToggles          int
		expectLineNumMargin uint64
	}{
		{
			name:                "empty document, no toggle",
			screenWidth:         100,
			numLines:            0,
			numToggles:          0,
			expectLineNumMargin: 0,
		},
		{
			name:                "empty document, single toggle",
			screenWidth:         100,
			numLines:            0,
			numToggles:          1,
			expectLineNumMargin: 3,
		},
		{
			name:                "single line, one toggle",
			screenWidth:         100,
			numLines:            1,
			numToggles:          1,
			expectLineNumMargin: 3,
		},
		{
			name:                "single line, two toggles",
			screenWidth:         100,
			numLines:            1,
			numToggles:          2,
			expectLineNumMargin: 0,
		},
		{
			name:                "nine lines, one toggle",
			screenWidth:         100,
			numLines:            9,
			numToggles:          1,
			expectLineNumMargin: 3,
		},
		{
			name:                "ten lines, one toggle",
			screenWidth:         100,
			numLines:            10,
			numToggles:          1,
			expectLineNumMargin: 3,
		},
		{
			name:                "eleven lines, one toggle",
			screenWidth:         100,
			numLines:            11,
			numToggles:          1,
			expectLineNumMargin: 3,
		},
		{
			name:                "eleven lines, screen width 4",
			screenWidth:         4,
			numLines:            11,
			numToggles:          1,
			expectLineNumMargin: 3,
		},
		{
			name:                "eleven lines, screen width 3",
			screenWidth:         3,
			numLines:            11,
			numToggles:          1,
			expectLineNumMargin: 0,
		},
		{
			name:                "eleven lines, screen width 2",
			screenWidth:         2,
			numLines:            11,
			numToggles:          1,
			expectLineNumMargin: 0,
		},
		{
			name:                "eleven lines, screen width 1",
			screenWidth:         1,
			numLines:            11,
			numToggles:          1,
			expectLineNumMargin: 0,
		},
		{
			name:                "ninety nine lines, one toggle",
			screenWidth:         100,
			numLines:            99,
			numToggles:          1,
			expectLineNumMargin: 3,
		},
		{
			name:                "one hundred lines, one toggle",
			screenWidth:         100,
			numLines:            100,
			numToggles:          1,
			expectLineNumMargin: 4,
		},
		{
			name:                "one hundred lines, screen width 5",
			screenWidth:         5,
			numLines:            100,
			numToggles:          1,
			expectLineNumMargin: 4,
		},
		{
			name:                "one hundred lines, screen width 4",
			screenWidth:         4,
			numLines:            100,
			numToggles:          1,
			expectLineNumMargin: 0,
		},
		{
			name:                "one hundred lines, screen width 3",
			screenWidth:         3,
			numLines:            100,
			numToggles:          1,
			expectLineNumMargin: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			state := NewEditorState(tc.screenWidth, 100, nil)
			for i := 0; i < tc.numLines-1; i++ {
				InsertNewline(state)
			}
			for i := 0; i < tc.numToggles; i++ {
				ToggleShowLineNumbers(state)
			}
			lineNumMargin := state.DocumentBuffer().LineNumMarginWidth()
			assert.Equal(t, tc.expectLineNumMargin, lineNumMargin)
		})
	}
}
