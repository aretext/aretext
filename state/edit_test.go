package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/text"
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
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			InsertRune(state, tc.insertRune)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestInsertText(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		insertText     string
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "empty",
			initialCursor:  cursorState{position: 0},
			inputString:    "",
			insertText:     "",
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:           "ascii",
			initialCursor:  cursorState{position: 1},
			inputString:    "abc",
			insertText:     "xyz",
			expectedCursor: cursorState{position: 4},
			expectedText:   "axyzbc",
		},
		{
			name:           "non-ascii unicode with multi-byte runes",
			initialCursor:  cursorState{position: 1},
			inputString:    "abc",
			insertText:     "丂丄丅丆丏 ¢ह€한",
			expectedCursor: cursorState{position: 11},
			expectedText:   "a丂丄丅丆丏 ¢ह€한bc",
		},
		{
			name:          "invalid UTF-8",
			initialCursor: cursorState{position: 0},
			inputString:   "",
			// 0xff is invalid UTF-8
			insertText:     string([]byte{'H', 'i', 0xff, '!'}),
			expectedCursor: cursorState{position: 4},
			expectedText:   "Hi�!",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			InsertText(state, tc.insertText)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestDeleteToPos(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		initialCursor     cursorState
		locator           func(LocatorParams) uint64
		expectedCursor    cursorState
		expectedText      string
		expectedClipboard clipboard.PageContent
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
			expectedCursor:    cursorState{position: 0},
			expectedText:      "bcd",
			expectedClipboard: clipboard.PageContent{Text: "a"},
		},
		{
			name:          "delete from end of text",
			inputString:   "abcd",
			initialCursor: cursorState{position: 3},
			locator: func(params LocatorParams) uint64 {
				return locate.NextCharInLine(params.TextTree, 1, true, params.CursorPos)
			},
			expectedCursor:    cursorState{position: 3},
			expectedText:      "abc",
			expectedClipboard: clipboard.PageContent{Text: "d"},
		},
		{
			name:          "delete multiple characters",
			inputString:   "abcd",
			initialCursor: cursorState{position: 1},
			locator: func(params LocatorParams) uint64 {
				return locate.NextCharInLine(params.TextTree, 10, true, params.CursorPos)
			},
			expectedCursor:    cursorState{position: 1},
			expectedText:      "a",
			expectedClipboard: clipboard.PageContent{Text: "bcd"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			DeleteToPos(state, tc.locator, clipboard.PageDefault)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
			assert.Equal(t, tc.expectedClipboard, state.clipboard.Get(clipboard.PageDefault))
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
			state := NewEditorState(100, 100, nil, nil)
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

func TestClearAutoIndentWhitespaceLine(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		cursorPos         uint64
		targetLinePos     uint64
		expectedText      string
		expectedCursorPos uint64
	}{
		{
			name:              "empty",
			inputString:       "",
			cursorPos:         0,
			targetLinePos:     0,
			expectedText:      "",
			expectedCursorPos: 0,
		},
		{
			name:              "line with non-whitespace chars",
			inputString:       "    abc",
			cursorPos:         0,
			targetLinePos:     0,
			expectedText:      "    abc",
			expectedCursorPos: 0,
		},
		{
			name:              "line with only spaces",
			inputString:       "    ",
			cursorPos:         1,
			targetLinePos:     0,
			expectedText:      "",
			expectedCursorPos: 0,
		},
		{
			name:              "line with only tabs",
			inputString:       "\t\t",
			cursorPos:         1,
			targetLinePos:     0,
			expectedText:      "",
			expectedCursorPos: 0,
		},
		{
			name:              "cursor after target line",
			inputString:       "    ab\n    \n    cd",
			cursorPos:         17,
			targetLinePos:     7,
			expectedText:      "    ab\n\n    cd",
			expectedCursorPos: 13,
		},
		{
			name:              "cursor before target line",
			inputString:       "    ab\n    \n    cd",
			cursorPos:         5,
			targetLinePos:     7,
			expectedText:      "    ab\n\n    cd",
			expectedCursorPos: 5,
		},
		{
			name:              "cursor on target line",
			inputString:       "    ab\n    \n    cd",
			cursorPos:         9,
			targetLinePos:     7,
			expectedText:      "    ab\n\n    cd",
			expectedCursorPos: 7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = cursorState{position: tc.cursorPos}
			state.documentBuffer.autoIndent = true
			ClearAutoIndentWhitespaceLine(state, func(p LocatorParams) uint64 {
				return locate.StartOfLineAtPos(p.TextTree, tc.targetLinePos)
			})
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
			state := NewEditorState(100, 100, nil, nil)
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
		replaceWithEmptyLine       bool
		expectedCursor             cursorState
		expectedText               string
		expectedClipboard          clipboard.PageContent
	}{
		{
			name:          "empty",
			inputString:   "",
			initialCursor: cursorState{position: 0},
			targetLineLocator: func(params LocatorParams) uint64 {
				return locate.StartOfLineBelow(params.TextTree, 1, params.CursorPos)
			},
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:          "delete single line",
			inputString:   "abcd",
			initialCursor: cursorState{position: 2},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
			expectedClipboard: clipboard.PageContent{
				Text:     "abcd",
				Linewise: true,
			},
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
		},
		{
			name:          "delete single line, first line",
			inputString:   "abcd\nefgh\nijk",
			initialCursor: cursorState{position: 2},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			expectedCursor: cursorState{position: 0},
			expectedText:   "efgh\nijk",
			expectedClipboard: clipboard.PageContent{
				Text:     "abcd",
				Linewise: true,
			},
		},
		{
			name:          "delete single line, interior line",
			inputString:   "abcd\nefgh\nijk",
			initialCursor: cursorState{position: 6},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			expectedCursor: cursorState{position: 5},
			expectedText:   "abcd\nijk",
			expectedClipboard: clipboard.PageContent{
				Text:     "efgh",
				Linewise: true,
			},
		},
		{
			name:          "delete single line, last line",
			inputString:   "abcd\nefgh\nijk",
			initialCursor: cursorState{position: 12},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			expectedCursor: cursorState{position: 5},
			expectedText:   "abcd\nefgh",
			expectedClipboard: clipboard.PageContent{
				Text:     "ijk",
				Linewise: true,
			},
		},
		{
			name:          "delete empty line",
			inputString:   "abcd\n\nefgh",
			initialCursor: cursorState{position: 5},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			expectedCursor: cursorState{position: 5},
			expectedText:   "abcd\nefgh",
			expectedClipboard: clipboard.PageContent{
				Text:     "",
				Linewise: true,
			},
		},
		{
			name:          "delete multiple lines down",
			inputString:   "abcd\nefgh\nijk\nlmnop",
			initialCursor: cursorState{position: 0},
			targetLineLocator: func(params LocatorParams) uint64 {
				return locate.StartOfLineBelow(params.TextTree, 2, params.CursorPos)
			},
			expectedCursor: cursorState{position: 0},
			expectedText:   "lmnop",
			expectedClipboard: clipboard.PageContent{
				Text:     "abcd\nefgh\nijk",
				Linewise: true,
			},
		},
		{
			name:          "delete multiple lines up",
			inputString:   "abcd\nefgh\nijk\nlmnop",
			initialCursor: cursorState{position: 16},
			targetLineLocator: func(params LocatorParams) uint64 {
				return locate.StartOfLineAbove(params.TextTree, 2, params.CursorPos)
			},
			expectedCursor: cursorState{position: 0},
			expectedText:   "abcd",
			expectedClipboard: clipboard.PageContent{
				Text:     "efgh\nijk\nlmnop",
				Linewise: true,
			},
		},
		{
			name:          "replace with empty line, empty document",
			inputString:   "",
			initialCursor: cursorState{position: 0},
			targetLineLocator: func(params LocatorParams) uint64 {
				return locate.StartOfLineBelow(params.TextTree, 1, params.CursorPos)
			},
			replaceWithEmptyLine: true,
			expectedCursor:       cursorState{position: 0},
			expectedText:         "",
		},
		{
			name:          "replace with empty line, on first line",
			inputString:   "abc\nefgh",
			initialCursor: cursorState{position: 0},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			replaceWithEmptyLine: true,
			expectedCursor:       cursorState{position: 0},
			expectedText:         "\nefgh",
			expectedClipboard: clipboard.PageContent{
				Text:     "abc",
				Linewise: true,
			},
		},
		{
			name:          "replace with empty line, on middle line",
			inputString:   "abc\nefg\nhij",
			initialCursor: cursorState{position: 5},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			replaceWithEmptyLine: true,
			expectedCursor:       cursorState{position: 4},
			expectedText:         "abc\n\nhij",
			expectedClipboard: clipboard.PageContent{
				Text:     "efg",
				Linewise: true,
			},
		},
		{
			name:          "replace with empty line, on empty line",
			inputString:   "abc\n\n\nhij",
			initialCursor: cursorState{position: 4},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			replaceWithEmptyLine: true,
			expectedCursor:       cursorState{position: 4},
			expectedText:         "abc\n\n\nhij",
			expectedClipboard: clipboard.PageContent{
				Text:     "",
				Linewise: true,
			},
		},
		{
			name:          "replace with empty line, on last line",
			inputString:   "abc\nefg\nhij",
			initialCursor: cursorState{position: 8},
			targetLineLocator: func(params LocatorParams) uint64 {
				return params.CursorPos
			},
			replaceWithEmptyLine: true,
			expectedCursor:       cursorState{position: 8},
			expectedText:         "abc\nefg\n",
			expectedClipboard: clipboard.PageContent{
				Text:     "hij",
				Linewise: true,
			},
		},
		{
			name:                 "replace with empty line, multiple lines selected",
			inputString:          "abc\nefg\nhij\nlmnop",
			initialCursor:        cursorState{position: 5},
			targetLineLocator:    func(params LocatorParams) uint64 { return 9 },
			replaceWithEmptyLine: true,
			expectedCursor:       cursorState{position: 4},
			expectedText:         "abc\n\nlmnop",
			expectedClipboard: clipboard.PageContent{
				Text:     "efg\nhij",
				Linewise: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			DeleteLines(state, tc.targetLineLocator, tc.abortIfTargetIsCurrentLine, tc.replaceWithEmptyLine, clipboard.PageDefault)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
			assert.Equal(t, tc.expectedClipboard, state.clipboard.Get(clipboard.PageDefault))
		})
	}
}

func TestReplaceChar(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		newChar        rune
		autoIndent     bool
		tabExpand      bool
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "empty",
			inputString:    "",
			newChar:        'a',
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:           "replace char",
			inputString:    "abcd",
			newChar:        'x',
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 1},
			expectedText:   "axcd",
		},
		{
			name:           "empty line",
			inputString:    "ab\n\ncd",
			newChar:        'x',
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
			expectedText:   "ab\n\ncd",
		},
		{
			name:           "insert newline",
			inputString:    "abcd",
			newChar:        '\n',
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 3},
			expectedText:   "ab\nd",
		},
		{
			name:           "insert newline with autoindent",
			inputString:    "\tabcd",
			newChar:        '\n',
			initialCursor:  cursorState{position: 2},
			autoIndent:     true,
			expectedCursor: cursorState{position: 4},
			expectedText:   "\ta\n\tcd",
		},
		{
			name:           "insert tab, no expand",
			inputString:    "abcd",
			newChar:        '\t',
			initialCursor:  cursorState{position: 2},
			expectedCursor: cursorState{position: 2},
			expectedText:   "ab\td",
		},
		{
			name:           "insert tab, expand",
			inputString:    "abcd",
			newChar:        '\t',
			initialCursor:  cursorState{position: 2},
			tabExpand:      true,
			expectedCursor: cursorState{position: 3},
			expectedText:   "ab  d",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			state.documentBuffer.autoIndent = tc.autoIndent
			state.documentBuffer.tabExpand = tc.tabExpand
			state.documentBuffer.tabSize = 4
			ReplaceChar(state, tc.newChar)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestToggleCaseAtCursor(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "empty",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:           "toggle lowercase to uppercase",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 2},
			expectedText:   "aBcd",
		},
		{
			name:           "toggle uppercase to lowercase",
			inputString:    "ABCD",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 2},
			expectedText:   "AbCD",
		},
		{
			name:           "toggle number",
			inputString:    "1234",
			initialCursor:  cursorState{position: 1},
			expectedCursor: cursorState{position: 2},
			expectedText:   "1234",
		},
		{
			name:           "empty line",
			inputString:    "ab\n\ncd",
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
			expectedText:   "ab\n\ncd",
		},
		{
			name:           "toggle at end of line",
			inputString:    "abcd\nefgh",
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
			expectedText:   "abcD\nefgh",
		},
		{
			name:           "toggle at end of document",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 3},
			expectedCursor: cursorState{position: 3},
			expectedText:   "abcD",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			ToggleCaseAtCursor(state)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestToggleCaseInSelection(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		selectionMode     selection.Mode
		selectionStartPos uint64
		selectionEndPos   uint64
		expectedCursor    cursorState
		expectedText      string
	}{
		{
			name:              "empty",
			inputString:       "",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 0,
			selectionEndPos:   0,
			expectedCursor:    cursorState{position: 0},
			expectedText:      "",
		},
		{
			name:              "select single character",
			inputString:       "abcdefgh",
			selectionMode:     selection.ModeChar,
			selectionStartPos: 2,
			selectionEndPos:   3,
			expectedCursor:    cursorState{position: 2},
			expectedText:      "abCdefgh",
		},
		{
			name:              "select multiple characters",
			inputString:       "abcdefgh",
			selectionMode:     selection.ModeLine,
			selectionStartPos: 2,
			selectionEndPos:   6,
			expectedCursor:    cursorState{position: 2},
			expectedText:      "abCDEFgh",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = cursorState{position: tc.selectionStartPos}
			selectionEndLoc := func(p LocatorParams) uint64 { return tc.selectionEndPos }
			ToggleCaseInSelection(state, selectionEndLoc)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestIndentLines(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		cursorPos      uint64
		targetLinePos  uint64
		count          uint64
		tabExpand      bool
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "empty",
			inputString:    "",
			cursorPos:      0,
			targetLinePos:  0,
			count:          1,
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:           "empty line",
			inputString:    "abc\n\ndef",
			cursorPos:      4,
			targetLinePos:  4,
			count:          1,
			expectedCursor: cursorState{position: 4},
			expectedText:   "abc\n\ndef",
		},
		{
			name:           "empty line with carriage return",
			inputString:    "abc\r\n\r\ndef",
			cursorPos:      5,
			targetLinePos:  5,
			count:          1,
			expectedCursor: cursorState{position: 5},
			expectedText:   "abc\r\n\r\ndef",
		},
		{
			name:           "line with single character",
			inputString:    "a",
			cursorPos:      0,
			targetLinePos:  0,
			count:          1,
			expectedCursor: cursorState{position: 1},
			expectedText:   "\ta",
		},
		{
			name:           "line with single character, tab expand",
			tabExpand:      true,
			inputString:    "a",
			cursorPos:      0,
			targetLinePos:  0,
			count:          1,
			expectedCursor: cursorState{position: 4},
			expectedText:   "    a",
		},
		{
			name:           "first line, cursor at start",
			inputString:    "abc\ndef\nghi",
			cursorPos:      0,
			targetLinePos:  0,
			count:          1,
			expectedCursor: cursorState{position: 1},
			expectedText:   "\tabc\ndef\nghi",
		},
		{
			name:           "first line, cursor past start",
			inputString:    "abc\ndef\nghi",
			cursorPos:      1,
			targetLinePos:  1,
			count:          1,
			expectedCursor: cursorState{position: 1},
			expectedText:   "\tabc\ndef\nghi",
		},
		{
			name:           "second line, cursor at start",
			inputString:    "abc\ndef\nghi",
			cursorPos:      4,
			targetLinePos:  4,
			count:          1,
			expectedCursor: cursorState{position: 5},
			expectedText:   "abc\n\tdef\nghi",
		},
		{
			name:           "second line, cursor past start",
			inputString:    "abc\ndef\nghi",
			cursorPos:      6,
			targetLinePos:  6,
			count:          1,
			expectedCursor: cursorState{position: 5},
			expectedText:   "abc\n\tdef\nghi",
		},
		{
			name:           "last line, cursor at end",
			inputString:    "abc\ndef\nghi",
			cursorPos:      11,
			targetLinePos:  11,
			count:          1,
			expectedCursor: cursorState{position: 9},
			expectedText:   "abc\ndef\n\tghi",
		},
		{
			name:           "tab expand, aligned",
			inputString:    "abc\ndef\nghi",
			cursorPos:      6,
			targetLinePos:  6,
			count:          1,
			tabExpand:      true,
			expectedCursor: cursorState{position: 8},
			expectedText:   "abc\n    def\nghi",
		},
		{
			name:           "tab expand, line with whitespace at start",
			inputString:    "abc\n  def\nghi",
			cursorPos:      7,
			targetLinePos:  7,
			count:          1,
			tabExpand:      true,
			expectedCursor: cursorState{position: 10},
			expectedText:   "abc\n      def\nghi",
		},
		{
			name:           "multiple lines",
			inputString:    "ab\ncd\nef\ngh",
			cursorPos:      4,
			targetLinePos:  7,
			count:          1,
			expectedCursor: cursorState{position: 4},
			expectedText:   "ab\n\tcd\n\tef\ngh",
		},
		{
			name:           "repeat count times",
			inputString:    "ab\ncd\nef\ngh",
			cursorPos:      4,
			targetLinePos:  7,
			count:          3,
			expectedCursor: cursorState{position: 6},
			expectedText:   "ab\n\t\t\tcd\n\t\t\tef\ngh",
		},
		{
			name:           "tab expand, repeat count times",
			inputString:    "abc\n  def\nghi",
			cursorPos:      7,
			targetLinePos:  7,
			count:          3,
			tabExpand:      true,
			expectedCursor: cursorState{position: 18},
			expectedText:   "abc\n              def\nghi",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = cursorState{position: tc.cursorPos}
			state.documentBuffer.tabExpand = tc.tabExpand
			targetLineLoc := func(p LocatorParams) uint64 { return tc.targetLinePos }
			IndentLines(state, targetLineLoc, tc.count)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestOutdentLines(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		cursorPos      uint64
		targetLinePos  uint64
		count          uint64
		tabSize        uint64
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "empty",
			inputString:    "",
			cursorPos:      0,
			targetLinePos:  0,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:           "outdent first line starting with a single tab, on tab",
			inputString:    "\tabc",
			cursorPos:      0,
			targetLinePos:  0,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 0},
			expectedText:   "abc",
		},
		{
			name:           "outdent first line starting with a single tab, on start of text",
			inputString:    "\tabc",
			cursorPos:      1,
			targetLinePos:  1,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 0},
			expectedText:   "abc",
		},
		{
			name:           "outdent first line starting with a single tab, on end of text",
			inputString:    "\tabc",
			cursorPos:      3,
			targetLinePos:  3,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 0},
			expectedText:   "abc",
		},
		{
			name:           "outdent first line starting with multiple tabs",
			inputString:    "\t\t\tabc",
			cursorPos:      4,
			targetLinePos:  4,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 2},
			expectedText:   "\t\tabc",
		},
		{
			name:           "outdent first line starting with spaces less than tabsize",
			inputString:    "  abc",
			cursorPos:      2,
			targetLinePos:  2,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 0},
			expectedText:   "abc",
		},
		{
			name:           "outdent first line starting with spaces equal to tabsize",
			inputString:    "    abc",
			cursorPos:      2,
			targetLinePos:  2,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 0},
			expectedText:   "abc",
		},
		{
			name:           "outdent first line starting with spaces greater than tabsize",
			inputString:    "    abc",
			cursorPos:      2,
			targetLinePos:  2,
			count:          1,
			tabSize:        2,
			expectedCursor: cursorState{position: 2},
			expectedText:   "  abc",
		},
		{
			name:           "outdent empty line",
			inputString:    "abc\n\ndef",
			cursorPos:      5,
			targetLinePos:  5,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 5},
			expectedText:   "abc\n\ndef",
		},
		{
			name:           "outdent line with only space",
			inputString:    "abc\n      \ndef",
			cursorPos:      5,
			targetLinePos:  5,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 6},
			expectedText:   "abc\n  \ndef",
		},
		{
			name:           "outdent middle line",
			inputString:    "abc\n\t\tdef\nghi",
			cursorPos:      7,
			targetLinePos:  7,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 5},
			expectedText:   "abc\n\tdef\nghi",
		},
		{
			name:           "outdent mix of tabs and spaces",
			inputString:    "  \t abc",
			cursorPos:      5,
			targetLinePos:  5,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 1},
			expectedText:   " abc",
		},
		{
			name:           "multiple lines",
			inputString:    "ab\n\tcd\n\t\tef\ngh",
			cursorPos:      5,
			targetLinePos:  8,
			count:          1,
			tabSize:        4,
			expectedCursor: cursorState{position: 3},
			expectedText:   "ab\ncd\n\tef\ngh",
		},
		{
			name:           "repeat count times",
			inputString:    "ab\n\t\t\tcd\n\t\t\t\tef\ngh",
			cursorPos:      5,
			targetLinePos:  14,
			count:          3,
			tabSize:        4,
			expectedCursor: cursorState{position: 3},
			expectedText:   "ab\ncd\n\tef\ngh",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = cursorState{position: tc.cursorPos}
			state.documentBuffer.tabSize = tc.tabSize
			targetLineLoc := func(p LocatorParams) uint64 { return tc.targetLinePos }
			OutdentLines(state, targetLineLoc, tc.count)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestBeginNewLineAbove(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		cursorPos      uint64
		autoIndent     bool
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "empty, no autoindent",
			inputString:    "",
			cursorPos:      0,
			autoIndent:     false,
			expectedCursor: cursorState{position: 0},
			expectedText:   "\n",
		},
		{
			name:           "empty, autoindent",
			inputString:    "",
			cursorPos:      0,
			autoIndent:     true,
			expectedCursor: cursorState{position: 0},
			expectedText:   "\n",
		},
		{
			name:           "multiple lines, no indentation, no autoindent",
			inputString:    "abc\ndef\nhij",
			cursorPos:      5,
			autoIndent:     false,
			expectedCursor: cursorState{position: 4},
			expectedText:   "abc\n\ndef\nhij",
		},
		{
			name:           "multiple lines, indentation, no autoindent",
			inputString:    "abc\n\t\tdef\nhij",
			cursorPos:      5,
			autoIndent:     false,
			expectedCursor: cursorState{position: 4},
			expectedText:   "abc\n\n\t\tdef\nhij",
		},
		{
			name:           "multiple lines, no indentation, autoindent",
			inputString:    "abc\ndef\nhij",
			cursorPos:      5,
			autoIndent:     true,
			expectedCursor: cursorState{position: 4},
			expectedText:   "abc\n\ndef\nhij",
		},
		{
			name:           "multiple lines, indentation, autoindent",
			inputString:    "abc\n\t\tdef\nhij",
			cursorPos:      5,
			autoIndent:     true,
			expectedCursor: cursorState{position: 6},
			expectedText:   "abc\n\t\t\n\t\tdef\nhij",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = cursorState{position: tc.cursorPos}
			state.documentBuffer.autoIndent = tc.autoIndent
			BeginNewLineAbove(state)
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
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			JoinLines(state)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestCopyRange(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		loc               RangeLocator
		expectedClipboard clipboard.PageContent
	}{
		{
			name:              "empty",
			inputString:       "",
			loc:               func(p LocatorParams) (uint64, uint64) { return 0, 0 },
			expectedClipboard: clipboard.PageContent{},
		},
		{
			name:              "start pos equal to end pos",
			inputString:       "abcd",
			loc:               func(p LocatorParams) (uint64, uint64) { return 2, 2 },
			expectedClipboard: clipboard.PageContent{},
		},
		{
			name:              "start pos after  end pos",
			inputString:       "abcd",
			loc:               func(p LocatorParams) (uint64, uint64) { return 3, 2 },
			expectedClipboard: clipboard.PageContent{},
		},
		{
			name:              "start pos before end pos",
			inputString:       "abcd",
			loc:               func(p LocatorParams) (uint64, uint64) { return 1, 3 },
			expectedClipboard: clipboard.PageContent{Text: "bc"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			CopyRange(state, clipboard.PageDefault, tc.loc)
			assert.Equal(t, tc.expectedClipboard, state.clipboard.Get(clipboard.PageDefault))
		})
	}
}

func TestCopyLine(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		initialCursor     cursorState
		expectedClipboard clipboard.PageContent
	}{
		{
			name:          "empty",
			inputString:   "",
			initialCursor: cursorState{position: 0},
			expectedClipboard: clipboard.PageContent{
				Linewise: true,
			},
		},
		{
			name:          "single line, cursor at start",
			inputString:   "abcd",
			initialCursor: cursorState{position: 0},
			expectedClipboard: clipboard.PageContent{
				Text:     "abcd",
				Linewise: true,
			},
		},
		{
			name:          "single line, cursor in middle",
			inputString:   "abcd",
			initialCursor: cursorState{position: 2},
			expectedClipboard: clipboard.PageContent{
				Text:     "abcd",
				Linewise: true,
			},
		},
		{
			name:          "single line, cursor at end",
			inputString:   "abcd",
			initialCursor: cursorState{position: 4},
			expectedClipboard: clipboard.PageContent{
				Text:     "abcd",
				Linewise: true,
			},
		},
		{
			name:          "multiple lines, cursor on first line",
			inputString:   "abcd\nefgh\nijkl",
			initialCursor: cursorState{position: 2},
			expectedClipboard: clipboard.PageContent{
				Text:     "abcd",
				Linewise: true,
			},
		},
		{
			name:          "multiple lines, cursor on middle line",
			inputString:   "abcd\nefgh\nijkl",
			initialCursor: cursorState{position: 5},
			expectedClipboard: clipboard.PageContent{
				Text:     "efgh",
				Linewise: true,
			},
		},
		{
			name:          "multiple lines, cursor on last line",
			inputString:   "abcd\nefgh\nijkl",
			initialCursor: cursorState{position: 10},
			expectedClipboard: clipboard.PageContent{
				Text:     "ijkl",
				Linewise: true,
			},
		},
		{
			name:          "cursor on empty line",
			inputString:   "abcd\n\n\nefgh",
			initialCursor: cursorState{position: 5},
			expectedClipboard: clipboard.PageContent{
				Text:     "",
				Linewise: true,
			},
		},
		{
			name:          "multi-byte unicode",
			inputString:   "丂丄丅丆丏 ¢ह€한",
			initialCursor: cursorState{position: 2},
			expectedClipboard: clipboard.PageContent{
				Text:     "丂丄丅丆丏 ¢ह€한",
				Linewise: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			CopyLine(state, clipboard.PageDefault)
			assert.Equal(t, tc.initialCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedClipboard, state.clipboard.Get(clipboard.PageDefault))
		})
	}
}

func TestCopySelection(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		selectionMode     selection.Mode
		cursorStartPos    uint64
		cursorEndPos      uint64
		expectedCursor    cursorState
		expectedText      string
		expectedClipboard clipboard.PageContent
	}{
		{
			name:              "empty document, select charwise",
			inputString:       "",
			selectionMode:     selection.ModeChar,
			cursorStartPos:    0,
			cursorEndPos:      0,
			expectedCursor:    cursorState{position: 0},
			expectedText:      "",
			expectedClipboard: clipboard.PageContent{Text: ""},
		},
		{
			name:           "empty document, select linewise",
			inputString:    "",
			selectionMode:  selection.ModeLine,
			cursorStartPos: 0,
			cursorEndPos:   0,
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
			expectedClipboard: clipboard.PageContent{
				Text:     "",
				Linewise: true,
			},
		},
		{
			name:              "nonempty charwise selection",
			inputString:       "abcd1234",
			selectionMode:     selection.ModeChar,
			cursorStartPos:    1,
			cursorEndPos:      3,
			expectedCursor:    cursorState{position: 1},
			expectedText:      "abcd1234",
			expectedClipboard: clipboard.PageContent{Text: "bcd"},
		},
		{
			name:           "nonempty linewise selection",
			inputString:    "ab\ncde\nfgh\n12\n34",
			selectionMode:  selection.ModeLine,
			cursorStartPos: 4,
			cursorEndPos:   8,
			expectedCursor: cursorState{position: 3},
			expectedText:   "ab\ncde\nfgh\n12\n34",
			expectedClipboard: clipboard.PageContent{
				Text:     "cde\nfgh",
				Linewise: true,
			},
		},
		{
			name:              "empty line, select charwise",
			inputString:       "abc\n\ndef",
			selectionMode:     selection.ModeChar,
			cursorStartPos:    4,
			cursorEndPos:      4,
			expectedCursor:    cursorState{position: 4},
			expectedText:      "abc\n\ndef",
			expectedClipboard: clipboard.PageContent{Text: "\n"},
		},
		{
			name:           "empty line, select linewise",
			inputString:    "abc\n\ndef",
			selectionMode:  selection.ModeLine,
			cursorStartPos: 4,
			cursorEndPos:   4,
			expectedCursor: cursorState{position: 4},
			expectedText:   "abc\n\ndef",
			expectedClipboard: clipboard.PageContent{
				Text:     "",
				Linewise: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.selector.Start(tc.selectionMode, tc.cursorStartPos)
			state.documentBuffer.cursor = cursorState{position: tc.cursorEndPos}
			CopySelection(state, clipboard.PageDefault)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
			assert.Equal(t, tc.expectedClipboard, state.clipboard.Get(clipboard.PageDefault))
			assert.Equal(t, false, state.documentBuffer.undoLog.HasUnsavedChanges())
		})
	}
}

func TestPasteAfterCursor(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		clipboard      clipboard.PageContent
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "empty document, empty clipboard",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			clipboard:      clipboard.PageContent{},
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:          "empty document, empty clipboard insert on next line",
			inputString:   "",
			initialCursor: cursorState{position: 0},
			clipboard: clipboard.PageContent{
				Linewise: true,
			},
			expectedCursor: cursorState{position: 1},
			expectedText:   "\n",
		},
		{
			name:          "paste after cursor",
			inputString:   "abcd",
			initialCursor: cursorState{position: 2},
			clipboard: clipboard.PageContent{
				Text:     "xyz",
				Linewise: false,
			},
			expectedCursor: cursorState{position: 5},
			expectedText:   "abcxyzd",
		},
		{
			name:          "paste after cursor insert on next line",
			inputString:   "abcd",
			initialCursor: cursorState{position: 2},
			clipboard: clipboard.PageContent{
				Text:     "xyz",
				Linewise: true,
			},
			expectedCursor: cursorState{position: 5},
			expectedText:   "abcd\nxyz",
		},
		{
			name:          "paste newline after cursor",
			inputString:   "abcd",
			initialCursor: cursorState{position: 1},
			clipboard: clipboard.PageContent{
				Text:     "\n",
				Linewise: false,
			},
			expectedCursor: cursorState{position: 3},
			expectedText:   "ab\ncd",
		},
		{
			name:          "multi-byte unicode",
			inputString:   "abc",
			initialCursor: cursorState{position: 1},
			clipboard: clipboard.PageContent{
				Text:     "丂丄丅丆丏 ¢ह€한",
				Linewise: false,
			},
			expectedCursor: cursorState{position: 11},
			expectedText:   "ab丂丄丅丆丏 ¢ह€한c",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			state.clipboard.Set(clipboard.PageDefault, tc.clipboard)
			PasteAfterCursor(state, clipboard.PageDefault)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestPasteBeforeCursor(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		clipboard      clipboard.PageContent
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "empty document, empty clipboard",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			clipboard:      clipboard.PageContent{},
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:          "empty document, empty clipboard insert on next line",
			inputString:   "",
			initialCursor: cursorState{position: 0},
			clipboard: clipboard.PageContent{
				Linewise: true,
			},
			expectedCursor: cursorState{position: 0},
			expectedText:   "\n",
		},
		{
			name:          "paste before cursor",
			inputString:   "abcd",
			initialCursor: cursorState{position: 2},
			clipboard: clipboard.PageContent{
				Text:     "xyz",
				Linewise: false,
			},
			expectedCursor: cursorState{position: 4},
			expectedText:   "abxyzcd",
		},
		{
			name:          "paste before cursor insert on next line",
			inputString:   "abcd",
			initialCursor: cursorState{position: 2},
			clipboard: clipboard.PageContent{
				Text:     "xyz",
				Linewise: true,
			},
			expectedCursor: cursorState{position: 0},
			expectedText:   "xyz\nabcd",
		},
		{
			name:          "paste newline before cursor",
			inputString:   "abcd",
			initialCursor: cursorState{position: 2},
			clipboard: clipboard.PageContent{
				Text:     "\n",
				Linewise: false,
			},
			expectedCursor: cursorState{position: 1},
			expectedText:   "ab\ncd",
		},
		{
			name:          "multi-byte unicode",
			inputString:   "abc",
			initialCursor: cursorState{position: 1},
			clipboard: clipboard.PageContent{
				Text:     "丂丄丅丆丏 ¢ह€한",
				Linewise: false,
			},
			expectedCursor: cursorState{position: 10},
			expectedText:   "a丂丄丅丆丏 ¢ह€한bc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			state.documentBuffer.textTree = textTree
			state.documentBuffer.cursor = tc.initialCursor
			state.clipboard.Set(clipboard.PageDefault, tc.clipboard)
			PasteBeforeCursor(state, clipboard.PageDefault)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}
