package state

import (
	"unicode/utf8"

	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/text"
)

// searchState represents the state of a text search.
type searchState struct {
	query         string
	direction     text.ReadDirection
	prevQuery     string
	prevDirection text.ReadDirection
	match         *SearchMatch
}

// SearchMatch represents the successful result of a text search.
type SearchMatch struct {
	StartPos uint64
	EndPos   uint64
}

func (sm *SearchMatch) ContainsPosition(pos uint64) bool {
	return sm != nil && pos >= sm.StartPos && pos < sm.EndPos
}

// StartSearch initiates a new text search.
func StartSearch(state *EditorState, direction text.ReadDirection) {
	buffer := state.documentBuffer
	prevQuery, prevDirection := buffer.search.query, buffer.search.direction
	buffer.search = searchState{
		direction:     direction,
		prevQuery:     prevQuery,
		prevDirection: prevDirection,
	}
	state.inputMode = InputModeSearch
}

// CompleteSearch terminates a text search and returns to normal mode.
// If commit is true, jump to the matching search result.
// Otherwise, return to the original cursor position.
func CompleteSearch(state *EditorState, commit bool) {
	buffer := state.documentBuffer
	if commit {
		if buffer.search.match != nil {
			buffer.cursor = cursorState{position: buffer.search.match.StartPos}
		}
	} else {
		prevQuery, prevDirection := buffer.search.prevQuery, buffer.search.prevDirection
		buffer.search = searchState{
			query:     prevQuery,
			direction: prevDirection,
		}
	}
	buffer.search.match = nil
	state.inputMode = InputModeNormal
	ScrollViewToCursor(state)
}

// AppendRuneToSearchQuery appends a rune to the text search query.
func AppendRuneToSearchQuery(state *EditorState, r rune) {
	q := state.documentBuffer.search.query
	q = q + string(r)
	runTextSearchQuery(state, q)
}

// DeleteRuneFromSearchQuery
// A deletion in an empty query aborts the search and returns the editor to normal mode.
func DeleteRuneFromSearchQuery(state *EditorState) {
	q := state.documentBuffer.search.query
	if len(q) == 0 {
		CompleteSearch(state, false)
		return
	}

	q = q[0 : len(q)-1]
	runTextSearchQuery(state, q)
}

func runTextSearchQuery(state *EditorState, q string) {
	buffer := state.documentBuffer
	buffer.search.query = q
	foundMatch, matchStartPos := false, uint64(0)
	if buffer.search.direction == text.ReadDirectionForward {
		foundMatch, matchStartPos = searchTextForward(
			buffer.cursor.position,
			buffer.textTree,
			buffer.search.query)
	} else {
		foundMatch, matchStartPos = searchTextBackward(
			buffer.cursor.position,
			buffer.textTree,
			buffer.search.query)
	}

	if !foundMatch {
		buffer.search.match = nil
		ScrollViewToCursor(state)
		return
	}

	buffer.search.match = &SearchMatch{
		StartPos: matchStartPos,
		EndPos:   matchStartPos + uint64(utf8.RuneCountInString(q)),
	}
	buffer.view.textOrigin = locate.ViewOriginAfterScroll(
		matchStartPos,
		buffer.textTree,
		buffer.view.textOrigin,
		buffer.view.width,
		buffer.view.height,
		buffer.tabSize)
}

// FindNextMatch moves the cursor to the next position matching the search query.
func FindNextMatch(state *EditorState, reverse bool) {
	buffer := state.documentBuffer

	direction := buffer.search.direction
	if reverse {
		direction = direction.Reverse()
	}

	foundMatch, newCursorPos := false, uint64(0)
	if direction == text.ReadDirectionForward {
		foundMatch, newCursorPos = searchTextForward(
			buffer.cursor.position+1,
			buffer.textTree,
			buffer.search.query)
	} else {
		foundMatch, newCursorPos = searchTextBackward(
			buffer.cursor.position,
			buffer.textTree,
			buffer.search.query)
	}

	if foundMatch {
		buffer.cursor = cursorState{position: newCursorPos}
	}
}

// searchTextForward finds the position of the next occurrence of a query string on or after the start position.
func searchTextForward(startPos uint64, tree *text.Tree, query string) (bool, uint64) {
	r := tree.ReaderAtPosition(startPos, text.ReadDirectionForward)
	foundMatch, matchOffset, err := text.Search(query, r)
	if err != nil {
		panic(err) // should never happen because the tree reader shouldn't return an error.
	}

	if !foundMatch {
		return false, 0
	}

	return true, startPos + matchOffset
}

// searchTextBackward finds the beginning of the previous match before the start position.
func searchTextBackward(startPos uint64, tree *text.Tree, query string) (bool, uint64) {
	if len(query) == 0 {
		return false, 0
	}

	// Since we're searching backwards through the text, we need to find
	// the mirror image of the query string.  Note that we are reversing the bytes
	// of the query string, not runes or grapheme clusters.
	reversedQuery := make([]byte, len(query))
	for i := 0; i < len(query); i++ {
		reversedQuery[i] = query[len(query)-1-i]
	}

	// It is possible for the cursor to be in the middle of a search query,
	// in which case we want to match the beginning of the query.
	// Example: if the text is "...ab[c]d..." (where [] shows the cursor position)
	// and we're searching backwards for "abcd", the cursor should end up on "a".
	// To ensure that we find these matches, we need to start searching from the current
	// position plus one less than the length of the query (or the end of text if that comes sooner).
	numRunesInQuery := uint64(utf8.RuneCountInString(query))
	pos := startPos + numRunesInQuery - 1
	if n := tree.NumChars(); pos >= n {
		if n > 0 {
			pos = n - 1
		} else {
			pos = 0
		}
	}

	r := tree.ReaderAtPosition(pos, text.ReadDirectionBackward)
	foundMatch, matchOffset, err := text.Search(string(reversedQuery), r)
	if err != nil {
		panic(err) // should never happen because the tree reader shouldn't return an error.
	}

	if !foundMatch {
		return false, 0
	}

	matchStartPos := pos - matchOffset - numRunesInQuery
	return true, matchStartPos
}
