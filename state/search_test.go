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
			name:              "unicode normalization has different offsets",
			text:              "<p>  &amp; © Æ Ď\n¾ ℋ ⅆ\n∲ ≧̸</p>\nfoobar",
			cursorPos:         0,
			query:             "foobar",
			direction:         SearchDirectionForward,
			expectedCursorPos: 32,
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

func TestSearchWordUnderCursor(t *testing.T) {
	testCases := []struct {
		name          string
		inputText     string
		direction     SearchDirection
		count         uint64
		pos           uint64
		expectedQuery string
		expectedPos   uint64
	}{
		{
			name:          "empty",
			inputText:     "",
			direction:     SearchDirectionForward,
			count:         1,
			pos:           0,
			expectedQuery: "",
			expectedPos:   0,
		},
		{
			name:          "start of word under cursor, search forward",
			inputText:     "foo bar baz bar",
			direction:     SearchDirectionForward,
			count:         1,
			pos:           4,
			expectedQuery: "bar\\C",
			expectedPos:   12,
		},
		{
			name:          "word under cursor, search forward",
			inputText:     "foo bar baz bar",
			direction:     SearchDirectionForward,
			count:         1,
			pos:           5,
			expectedQuery: "bar\\C",
			expectedPos:   12,
		},
		{
			name:          "word under cursor, search backward",
			inputText:     "foo bar baz bar",
			direction:     SearchDirectionForward,
			count:         1,
			pos:           14,
			expectedQuery: "bar\\C",
			expectedPos:   4,
		},
		{
			name:          "whitespace before word",
			inputText:     "foo   bar baz bar",
			direction:     SearchDirectionForward,
			count:         1,
			pos:           3,
			expectedQuery: "bar\\C",
			expectedPos:   6, // differs from vim, which would advance to the next occurrence.
		},
		{
			name:          "whitespace before end of line",
			inputText:     "foo bar   \nbaz",
			direction:     SearchDirectionForward,
			count:         1,
			pos:           9,
			expectedQuery: "baz\\C", // differs from vim, which aborts.
			expectedPos:   11,
		},
		{
			name:          "search forward with count",
			inputText:     "foo bar baz\nxyz\nfoo bar bat",
			direction:     SearchDirectionForward,
			count:         2,
			pos:           1,
			expectedQuery: "foo bar\\C",
			expectedPos:   16,
		},
		{
			name:          "search case sensitive",
			inputText:     "foo bar FOO BAR bar",
			direction:     SearchDirectionForward,
			count:         1,
			pos:           5,
			expectedQuery: "bar\\C",
			expectedPos:   16,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputText)
			require.NoError(t, err)
			state := NewEditorState(100, 100, nil, nil)
			buffer := state.documentBuffer
			buffer.textTree = textTree
			buffer.cursor.position = tc.pos

			// Search for the word under the cursor.
			SearchWordUnderCursor(state, SearchDirectionForward, tc.count)
			assert.Equal(t, InputModeNormal, state.inputMode)
			assert.Equal(t, tc.expectedQuery, buffer.search.query)
			assert.Nil(t, buffer.search.match)
			assert.Equal(t, cursorState{position: tc.expectedPos}, buffer.cursor)
		})
	}
}

func TestSetSearchQueryToPrevInHistory(t *testing.T) {
	textTree, err := text.NewTreeFromString("abc def ghi")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
	buffer := state.documentBuffer
	buffer.textTree = textTree

	// First search query, aborted.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "abc" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, false)

	// Second search query, committed.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "def" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, true)

	// Start a search, go back in history.
	StartSearch(state, SearchDirectionForward)
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "def", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)

	// Go back in the history again.
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "abc", buffer.search.query)
	assert.Equal(t, uint64(0), buffer.search.match.StartPos)

	// Go back in the history, no previous entry so no change.
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "abc", buffer.search.query)
	assert.Equal(t, uint64(0), buffer.search.match.StartPos)
}

func TestSetSearchQueryToNextInHistory(t *testing.T) {
	textTree, err := text.NewTreeFromString("abc def ghi")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
	buffer := state.documentBuffer
	buffer.textTree = textTree

	// First search query, aborted.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "abc" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, false)

	// Second search query, committed.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "def" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, true)

	// Go back to beginning of history.
	SetSearchQueryToPrevInHistory(state)
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "abc", buffer.search.query)
	assert.Equal(t, uint64(0), buffer.search.match.StartPos)

	// Go to next in history.
	SetSearchQueryToNextInHistory(state)
	assert.Equal(t, "def", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)

	// Forward again. No future entry, so no change.
	SetSearchQueryToNextInHistory(state)
	assert.Equal(t, "def", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)
}

func TestSearchQueryToPrevInHistoryThenAppendRunes(t *testing.T) {
	textTree, err := text.NewTreeFromString("abc def ghi")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
	buffer := state.documentBuffer
	buffer.textTree = textTree

	// First search query, aborted.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "abc" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, false)

	// Second search query, committed.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "def" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, true)

	// Start a search, go back to beginning of history.
	StartSearch(state, SearchDirectionForward)
	SetSearchQueryToPrevInHistory(state)
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "abc", buffer.search.query)
	assert.Equal(t, uint64(0), buffer.search.match.StartPos)

	// Edit the query by appending runes.
	AppendRuneToSearchQuery(state, 'x')
	AppendRuneToSearchQuery(state, 'y')
	AppendRuneToSearchQuery(state, 'z')
	assert.Equal(t, "abcxyz", buffer.search.query)
	assert.Nil(t, buffer.search.match)

	// Go back in history, confirm that the edit reset to the last entry.
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "def", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)
}

func TestSearchQueryToPrevInHistoryThenDeleteRunes(t *testing.T) {
	textTree, err := text.NewTreeFromString("abc def ghi")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
	buffer := state.documentBuffer
	buffer.textTree = textTree

	// First search query, aborted.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "abc" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, false)

	// Second search query, committed.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "def" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, true)

	// Start a search, go back to beginning of history.
	StartSearch(state, SearchDirectionForward)
	SetSearchQueryToPrevInHistory(state)
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "abc", buffer.search.query)
	assert.Equal(t, uint64(0), buffer.search.match.StartPos)

	// Edit the query by deleting runes.
	DeleteRuneFromSearchQuery(state)
	DeleteRuneFromSearchQuery(state)
	assert.Equal(t, "a", buffer.search.query)
	assert.Equal(t, uint64(0), buffer.search.match.StartPos)

	// Go back in history, confirm that the edit reset to the last entry.
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "def", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)
}

func TestSearchQueryHistoryExcludesEmptyQueries(t *testing.T) {
	textTree, err := text.NewTreeFromString("abc def ghi")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
	buffer := state.documentBuffer
	buffer.textTree = textTree

	// First search query.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "abc" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, false)

	// Several empty search queries, should not be added to history.
	for i := 0; i < 3; i++ {
		StartSearch(state, SearchDirectionForward)
		CompleteSearch(state, false)
	}

	// Start a search, back to previous entry.
	StartSearch(state, SearchDirectionForward)
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "abc", buffer.search.query)
	assert.Equal(t, uint64(0), buffer.search.match.StartPos)
}

func TestSearchQueryHistoryExcludesDuplicateQueries(t *testing.T) {
	textTree, err := text.NewTreeFromString("abc def ghi")
	require.NoError(t, err)
	state := NewEditorState(100, 100, nil, nil)
	buffer := state.documentBuffer
	buffer.textTree = textTree

	// First search query.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "abc" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, false)

	// Second search query.
	StartSearch(state, SearchDirectionForward)
	for _, r := range "def" {
		AppendRuneToSearchQuery(state, r)
	}
	CompleteSearch(state, false)

	// Repeat the query several times.
	for i := 0; i < 3; i++ {
		StartSearch(state, SearchDirectionForward)
		for _, r := range "def" {
			AppendRuneToSearchQuery(state, r)
		}
		CompleteSearch(state, false)
	}

	// Start a search, back to previous entry.
	StartSearch(state, SearchDirectionForward)
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "def", buffer.search.query)
	assert.Equal(t, uint64(4), buffer.search.match.StartPos)

	// Back again, expect that we're at the first entry (duplicate entries were excluded from history).
	SetSearchQueryToPrevInHistory(state)
	assert.Equal(t, "abc", buffer.search.query)
	assert.Equal(t, uint64(0), buffer.search.match.StartPos)
}
