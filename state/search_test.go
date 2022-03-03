package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretext/aretext/text"
)

func TestSearchAndCommit(t *testing.T) {
	textTree, err := text.NewTreeFromString("foo bar baz")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
	buffer := state.documentBuffer
	buffer.textTree = textTree

	// Start a search.
	StartSearch(state, SearchDirectionForward)
	assert.Equal(t, state.inputMode, InputModeSearch)
	assert.Equal(t, buffer.search.query, "")

	// Enter a search query.
	AppendRuneToSearchQuery(state, 'b')
	assert.Equal(t, "b", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)
	assert.Equal(t, uint64(5), buffer.search.match.EndPos)

	AppendRuneToSearchQuery(state, 'a')
	assert.Equal(t, "ba", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)
	assert.Equal(t, uint64(6), buffer.search.match.EndPos)

	AppendRuneToSearchQuery(state, 'r')
	assert.Equal(t, "bar", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)
	assert.Equal(t, uint64(7), buffer.search.match.EndPos)

	DeleteRuneFromSearchQuery(state)
	assert.Equal(t, "ba", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)
	assert.Equal(t, uint64(6), buffer.search.match.EndPos)

	// Commit the search.
	CompleteSearch(state, true)
	assert.Equal(t, state.inputMode, InputModeNormal)
	assert.Equal(t, "ba", buffer.search.query)
	assert.Nil(t, buffer.search.match)
	assert.Equal(t, cursorState{position: 4}, buffer.cursor)
}

func TestSearchAndAbort(t *testing.T) {
	textTree, err := text.NewTreeFromString("foo bar baz")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
	buffer := state.documentBuffer
	buffer.textTree = textTree
	buffer.search.query = "xyz"

	// Start a search.
	StartSearch(state, SearchDirectionForward)
	assert.Equal(t, state.inputMode, InputModeSearch)
	assert.Equal(t, buffer.search.query, "")
	assert.Equal(t, buffer.search.prevQuery, "xyz")

	// Enter a search query.
	AppendRuneToSearchQuery(state, 'b')
	assert.Equal(t, "b", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)
	assert.Equal(t, uint64(5), buffer.search.match.EndPos)

	// Abort the search.
	CompleteSearch(state, false)
	assert.Equal(t, state.inputMode, InputModeNormal)
	assert.Equal(t, "xyz", buffer.search.query)
	assert.Nil(t, buffer.search.match)
	assert.Equal(t, cursorState{position: 0}, buffer.cursor)
}

func TestSearchAndBackspaceEmptyQuery(t *testing.T) {
	textTree, err := text.NewTreeFromString("foo bar baz")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
	buffer := state.documentBuffer
	buffer.textTree = textTree

	// Start a search.
	StartSearch(state, SearchDirectionForward)
	assert.Equal(t, state.inputMode, InputModeSearch)
	assert.Equal(t, buffer.search.query, "")

	// Delete from the empty query, equivalent to aborting the search.
	DeleteRuneFromSearchQuery(state)
	assert.Equal(t, state.inputMode, InputModeNormal)
	assert.Equal(t, "", buffer.search.query)
	assert.Nil(t, buffer.search.match)
	assert.Equal(t, cursorState{position: 0}, buffer.cursor)
}

func TestSearchCaseSensitivity(t *testing.T) {
	testCases := []struct {
		name             string
		text             string
		query            string
		expectedMatchPos uint64
	}{
		{
			name:             "lowercase query, case-insensitive search",
			text:             "abc Foo foo xyz",
			query:            "foo",
			expectedMatchPos: 4,
		},
		{
			name:             "mixed-case query, case-sensitive search",
			text:             "abc foo Foo xyz",
			query:            "Foo",
			expectedMatchPos: 8,
		},
		{
			name:             "lowercase query, force case-sensitive search",
			text:             "abc Foo foo xyz",
			query:            "foo\\C",
			expectedMatchPos: 8,
		},
		{
			name:             "mixed-case query, force case-insensitive search",
			text:             "abc Foo foo xyz",
			query:            "FOO\\c",
			expectedMatchPos: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.text)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			buffer := state.documentBuffer
			buffer.textTree = textTree

			StartSearch(state, SearchDirectionForward)
			for _, r := range tc.query {
				AppendRuneToSearchQuery(state, r)
			}
			CompleteSearch(state, true)

			assert.Equal(t, cursorState{position: tc.expectedMatchPos}, buffer.cursor)
		})
	}
}

func TestFindNextMatch(t *testing.T) {
	testCases := []struct {
		name              string
		text              string
		cursorPos         uint64
		query             string
		direction         SearchDirection
		reverse           bool
		expectedCursorPos uint64
	}{
		{
			name:              "empty text",
			text:              "",
			cursorPos:         0,
			query:             "abc",
			direction:         SearchDirectionForward,
			expectedCursorPos: 0,
		},
		{
			name:              "find next after cursor",
			text:              "foo bar baz",
			cursorPos:         1,
			query:             "ba",
			direction:         SearchDirectionForward,
			expectedCursorPos: 4,
		},
		{
			name:              "find next after cursor already on match",
			text:              "foo bar baz",
			cursorPos:         4,
			query:             "ba",
			direction:         SearchDirectionForward,
			expectedCursorPos: 8,
		},
		{
			name:              "find next at end of text, not found in wraparound",
			text:              "foo bar baz",
			cursorPos:         10,
			query:             "xa",
			direction:         SearchDirectionForward,
			expectedCursorPos: 10,
		},
		{
			name:              "find next at end of text, found in wraparound",
			text:              "foo bar baz",
			cursorPos:         10,
			query:             "ba",
			direction:         SearchDirectionForward,
			expectedCursorPos: 4,
		},
		{
			name:              "find next with multi-byte unicode",
			text:              "丂丄丅丆丏 ¢ह€한",
			cursorPos:         0,
			query:             "丅丆",
			direction:         SearchDirectionForward,
			expectedCursorPos: 2,
		},
		{
			name:              "empty text, reverse search",
			text:              "",
			cursorPos:         0,
			query:             "abc",
			expectedCursorPos: 0,
			direction:         SearchDirectionForward,
			reverse:           true,
		},
		{
			name:              "find prev",
			text:              "foo bar baz xyz",
			cursorPos:         14,
			query:             "ba",
			expectedCursorPos: 8,
			direction:         SearchDirectionForward,
			reverse:           true,
		},
		{
			name:              "find prev from current match",
			text:              "foo bar baz xyz",
			cursorPos:         8,
			query:             "ba",
			direction:         SearchDirectionForward,
			expectedCursorPos: 4,
			reverse:           true,
		},
		{
			name:              "find prev from middle of current match",
			text:              "foo bar baz xyz",
			cursorPos:         9,
			query:             "ba",
			direction:         SearchDirectionForward,
			expectedCursorPos: 8,
			reverse:           true,
		},
		{
			name:              "find prev from start of text, not found in wraparound",
			text:              "foo bar baz xyz",
			cursorPos:         0,
			query:             "lm",
			direction:         SearchDirectionForward,
			expectedCursorPos: 0,
			reverse:           true,
		},
		{
			name:              "find prev from start of text, found in wraparound",
			text:              "foo bar baz xyz",
			cursorPos:         0,
			query:             "ba",
			direction:         SearchDirectionForward,
			expectedCursorPos: 8,
			reverse:           true,
		},
		{
			name:              "find prev with multi-byte unicode",
			text:              "丂丄丅丆丏 ¢ह€한",
			cursorPos:         9,
			query:             "丅丆",
			direction:         SearchDirectionForward,
			expectedCursorPos: 2,
			reverse:           true,
		},
		{
			name:              "backward search equivalent to reverse forward search",
			text:              "foo bar baz xyz",
			cursorPos:         14,
			query:             "ba",
			direction:         SearchDirectionBackward,
			expectedCursorPos: 8,
			reverse:           false,
		},
		{
			name:              "reverse backward search equivalent to forward search",
			text:              "foo bar baz xyz",
			cursorPos:         0,
			query:             "ba",
			direction:         SearchDirectionBackward,
			expectedCursorPos: 4,
			reverse:           true,
		},
		{
			name:              "unicode normalization, forward search",
			text:              "ohm \u03a9 xyz",
			cursorPos:         0,
			query:             "\u2126",
			direction:         SearchDirectionForward,
			expectedCursorPos: 4,
		},
		{
			name:              "unicode normalization, backward search",
			text:              "ohm \u03a9 xyz",
			cursorPos:         14,
			query:             "\u2126",
			direction:         SearchDirectionBackward,
			expectedCursorPos: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.text)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			buffer := state.documentBuffer
			buffer.textTree = textTree
			buffer.cursor = cursorState{position: tc.cursorPos}
			buffer.search.query = tc.query
			buffer.search.direction = tc.direction
			FindNextMatch(state, tc.reverse)
			assert.Equal(t, tc.expectedCursorPos, buffer.cursor.position)
		})
	}
}
