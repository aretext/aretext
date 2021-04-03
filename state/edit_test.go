package state

import (
	"testing"

	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertRune(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		insertRune     rune
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "insert into empty string",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			insertRune:     'x',
			expectedCursor: cursorState{position: 1},
			expectedText:   "x",
		},
		{
			name:           "insert in middle of string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 1},
			insertRune:     'x',
			expectedCursor: cursorState{position: 2},
			expectedText:   "axbcd",
		},
		{
			name:           "insert at end of string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 4},
			insertRune:     'x',
			expectedCursor: cursorState{position: 5},
			expectedText:   "abcdx",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			InsertRune(state, tc.insertRune)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestDeleteRunes(t *testing.T) {
	testCases := []struct {
		name                 string
		inputString          string
		initialCursor        cursorState
		locator              func(LocatorParams) uint64
		expectedCursor       cursorState
		expectedText         string
		expectUnsavedChanges bool
	}{
		{
			name:          "delete from empty string",
			inputString:   "",
			initialCursor: cursorState{position: 0},
			locator: func(params LocatorParams) uint64 {
				return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
			},
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:          "delete next character at start of string",
			inputString:   "abcd",
			initialCursor: cursorState{position: 0},
			locator: func(params LocatorParams) uint64 {
				return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
			},
			expectedCursor:       cursorState{position: 0},
			expectedText:         "bcd",
			expectUnsavedChanges: true,
		},
		{
			name:          "delete from end of text",
			inputString:   "abcd",
			initialCursor: cursorState{position: 3},
			locator: func(params LocatorParams) uint64 {
				return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
			},
			expectedCursor:       cursorState{position: 3},
			expectedText:         "abc",
			expectUnsavedChanges: true,
		},
		{
			name:          "delete multiple characters",
			inputString:   "abcd",
			initialCursor: cursorState{position: 1},
			locator: func(params LocatorParams) uint64 {
				return locate.NextCharInLine(params.TextTree, 10, true, params.CursorPos)
			},
			expectedCursor:       cursorState{position: 1},
			expectedText:         "a",
			expectUnsavedChanges: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			DeleteRunes(state, tc.locator)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
			assert.Equal(t, tc.expectUnsavedChanges, state.documentBuffer.undoLog.HasUnsavedChanges())
		})
	}
}

func TestInsertNewline(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		autoIndent        bool
		cursorPos         uint64
		tabExpand         bool
		expectedCursorPos uint64
		expectedText      string
	}{
		{
			name:              "empty document, autoindent disabled",
			inputString:       "",
			cursorPos:         0,
			expectedCursorPos: 1,
			expectedText:      "\n",
		},
		{
			name:              "single line, autoindent disabled, no indentation",
			inputString:       "abcd",
			cursorPos:         2,
			expectedCursorPos: 3,
			expectedText:      "ab\ncd",
		},
		{
			name:              "single line, autoindent disabled, with indentation",
			inputString:       "\tabcd",
			cursorPos:         3,
			expectedCursorPos: 4,
			expectedText:      "\tab\ncd",
		},
		{
			name:              "single line, autoindent enabled, no indentation",
			inputString:       "abcd",
			autoIndent:        true,
			cursorPos:         2,
			expectedCursorPos: 3,
			expectedText:      "ab\ncd",
		},
		{
			name:              "single line, autoindent enabled, tab indentation",
			inputString:       "\tabcd",
			autoIndent:        true,
			cursorPos:         3,
			expectedCursorPos: 5,
			expectedText:      "\tab\n\tcd",
		},
		{
			name:              "single line, autoindent enabled, space indentation",
			inputString:       "    abcd",
			autoIndent:        true,
			cursorPos:         6,
			expectedCursorPos: 8,
			expectedText:      "    ab\n\tcd",
		},
		{
			name:              "single line, autoindent enabled, mixed tabs and spaces aligned indentation",
			inputString:       " \tabcd",
			autoIndent:        true,
			cursorPos:         4,
			expectedCursorPos: 6,
			expectedText:      " \tab\n\tcd",
		},
		{
			name:              "single line, autoindent enabled, mixed tabs and spaces misaligned indentation",
			inputString:       "\t abcd",
			autoIndent:        true,
			cursorPos:         4,
			expectedCursorPos: 7,
			expectedText:      "\t ab\n\t cd",
		},
		{
			name:              "expand tab inserts spaces",
			inputString:       "    abcd",
			autoIndent:        true,
			tabExpand:         true,
			cursorPos:         8,
			expectedCursorPos: 13,
			expectedText:      "    abcd\n    ",
		},
		{
			name:              "dedent if extra whitespace at end of current line",
			inputString:       "    abcd        xyz",
			autoIndent:        true,
			tabExpand:         true,
			cursorPos:         8,
			expectedCursorPos: 13,
			expectedText:      "    abcd\n    xyz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = cursorState{position: tc.cursorPos}
			state.documentBuffer.autoIndent = tc.autoIndent
			state.documentBuffer.tabSize = 4
			state.documentBuffer.tabExpand = tc.tabExpand
			InsertNewline(state)
			assert.Equal(t, cursorState{position: tc.expectedCursorPos}, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestInsertTab(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		expectedText   string
		expectedCursor cursorState
		tabExpand      bool
	}{
		{
			name:           "insert tab, no expand",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 2},
			expectedText:   "ab\tcd",
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "insert tab, expand full width",
			tabExpand:      true,
			inputString:    "abcd",
			initialCursor:  cursorState{position: 0},
			expectedText:   "    abcd",
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "insert tab, partial width",
			tabExpand:      true,
			inputString:    "abcd",
			initialCursor:  cursorState{position: 2},
			expectedText:   "ab  cd",
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "insert tab, expand with mixed tabs/spaces",
			tabExpand:      true,
			inputString:    "\t\tab",
			initialCursor:  cursorState{position: 2},
			expectedText:   "\t\t    ab",
			expectedCursor: cursorState{position: 6},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			state.documentBuffer.tabSize = 4
			state.documentBuffer.tabExpand = tc.tabExpand
			InsertTab(state)
			assert.Equal(t, tc.expectedText, textTree.String())
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
		})
	}
}

func TestDeleteLines(t *testing.T) {
	testCases := []struct {
		name                       string
		inputString                string
		initialCursor              cursorState
		targetLineLocator          func(LocatorParams) uint64
		abortIfTargetIsCurrentLine bool
		expectedCursor             cursorState
		expectedText               string
		expectedUnsavedChanges     bool
	}{
		{
			name:          "empty",
			inputString:   "",
			initialCursor: cursorState{position: 0},
			targetLineLocator: func(params LocatorParams) uint64 {
				return locate.StartOfLineBelow(params.TextTree, 1, params.CursorPos)
			},
			expectedCursor:         cursorState{position: 0},
			expectedText:           "",
			expectedUnsavedChanges: false,
		},
		{
			name:          "delete single line",
			inputString:   "abcd",
			initialCursor: cursorState{position: 2},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			expectedCursor:         cursorState{position: 0},
			expectedText:           "",
			expectedUnsavedChanges: true,
		},
		{
			name:          "delete single line, abort if same line",
			inputString:   "abcd",
			initialCursor: cursorState{position: 2},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			abortIfTargetIsCurrentLine: true,
			expectedCursor:             cursorState{position: 2},
			expectedText:               "abcd",
			expectedUnsavedChanges:     false,
		},
		{
			name:          "delete single line, first line",
			inputString:   "abcd\nefgh\nijk",
			initialCursor: cursorState{position: 2},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			expectedCursor:         cursorState{position: 0},
			expectedText:           "efgh\nijk",
			expectedUnsavedChanges: true,
		},
		{
			name:          "delete single line, interior line",
			inputString:   "abcd\nefgh\nijk",
			initialCursor: cursorState{position: 6},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			expectedCursor:         cursorState{position: 5},
			expectedText:           "abcd\nijk",
			expectedUnsavedChanges: true,
		},
		{
			name:          "delete single line, last line",
			inputString:   "abcd\nefgh\nijk",
			initialCursor: cursorState{position: 12},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			expectedCursor:         cursorState{position: 5},
			expectedText:           "abcd\nefgh",
			expectedUnsavedChanges: true,
		},
		{
			name:          "delete empty line",
			inputString:   "abcd\n\nefgh",
			initialCursor: cursorState{position: 5},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			expectedCursor:         cursorState{position: 5},
			expectedText:           "abcd\nefgh",
			expectedUnsavedChanges: true,
		},
		{
			name:          "delete multiple lines down",
			inputString:   "abcd\nefgh\nijk\nlmnop",
			initialCursor: cursorState{position: 0},
			targetLineLocator: func(params LocatorParams) uint64 {
				return locate.StartOfLineBelow(params.TextTree, 2, params.CursorPos)
			},
			expectedCursor:         cursorState{position: 0},
			expectedText:           "lmnop",
			expectedUnsavedChanges: true,
		},
		{
			name:          "delete multiple lines up",
			inputString:   "abcd\nefgh\nijk\nlmnop",
			initialCursor: cursorState{position: 16},
			targetLineLocator: func(params LocatorParams) uint64 {
				return locate.StartOfLineAbove(params.TextTree, 2, params.CursorPos)
			},
			expectedCursor:         cursorState{position: 0},
			expectedText:           "abcd",
			expectedUnsavedChanges: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			DeleteLines(state, tc.targetLineLocator, tc.abortIfTargetIsCurrentLine)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
			assert.Equal(t, tc.expectedUnsavedChanges, state.documentBuffer.undoLog.HasUnsavedChanges())
		})
	}
}

func TestReplaceChar(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		newText        string
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "empty",
			inputString:    "",
			newText:        "a",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:           "replace char",
			inputString:    "abcd",
			newText:        "x",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 1},
			expectedText:   "axcd",
		},
		{
			name:           "empty line",
			inputString:    "ab\n\ncd",
			newText:        "x",
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
			expectedText:   "ab\n\ncd",
		},
		{
			name:           "insert newline",
			inputString:    "abcd",
			newText:        "\n",
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 3},
			expectedText:   "ab\nd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			ReplaceChar(state, tc.newText)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestJoinLines(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		expectedText   string
		expectedCursor cursorState
	}{
		{
			name:           "empty",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			expectedText:   "",
			expectedCursor: cursorState{position: 0},
		},
		{
			name:           "two lines, no indentation, cursor at start",
			inputString:    "abc\ndef",
			initialCursor:  cursorState{position: 0},
			expectedText:   "abc def",
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "two lines, no indentation, cursor before newline",
			inputString:    "abc\ndef",
			initialCursor:  cursorState{position: 2},
			expectedText:   "abc def",
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "two lines, no indentation, cursor on newline",
			inputString:    "abc\ndef",
			initialCursor:  cursorState{position: 3},
			expectedText:   "abc def",
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "two lines, second line indented with spaces",
			inputString:    "abc\n    def",
			initialCursor:  cursorState{position: 2},
			expectedText:   "abc def",
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "two lines, second line indented with tabs",
			inputString:    "abc\n\t\tdef",
			initialCursor:  cursorState{position: 2},
			expectedText:   "abc def",
			expectedCursor: cursorState{position: 3},
		},
		{
			name:           "multiple lines, on last line",
			inputString:    "abc\ndef\nghijk",
			initialCursor:  cursorState{position: 10},
			expectedText:   "abc\ndef\nghijk",
			expectedCursor: cursorState{position: 10},
		},
		{
			name:           "second-to-last line, last line is whitespace",
			inputString:    "abc\n     ",
			initialCursor:  cursorState{position: 2},
			expectedText:   "abc",
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "before empty line",
			inputString:    "abc\n\ndef",
			initialCursor:  cursorState{position: 1},
			expectedText:   "abc\ndef",
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "before multiple empty lines",
			inputString:    "abc\n\n\n\ndef",
			initialCursor:  cursorState{position: 1},
			expectedText:   "abc\n\n\ndef",
			expectedCursor: cursorState{position: 2},
		},
		{
			name:           "on empty line before non-empty line",
			inputString:    "abc\n\ndef\nxyz",
			initialCursor:  cursorState{position: 4},
			expectedText:   "abc\ndef\nxyz",
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "on empty line before empty line",
			inputString:    "abc\n\n\n\ndef",
			initialCursor:  cursorState{position: 4},
			expectedText:   "abc\n\n\ndef",
			expectedCursor: cursorState{position: 4},
		},
		{
			name:           "before line all whitespace",
			inputString:    "abc\n       \ndef",
			initialCursor:  cursorState{position: 2},
			expectedText:   "abc\ndef",
			expectedCursor: cursorState{position: 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			JoinLines(state)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}
