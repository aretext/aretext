package state

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/transform"

	"github.com/aretext/aretext/text"
)

// SearchDirection represents the direction of the search (forward or backward).
type SearchDirection int

const (
	SearchDirectionForward = SearchDirection(iota)
	SearchDirectionBackward
)

// Reverse returns the opposite direction.
func (d SearchDirection) Reverse() SearchDirection {
	switch d {
	case SearchDirectionForward:
		return SearchDirectionBackward
	case SearchDirectionBackward:
		return SearchDirectionForward
	default:
		panic("Unrecognized direction")
	}
}

// searchState represents the state of a text search.
type searchState struct {
	query         string
	direction     SearchDirection
	prevQuery     string
	prevDirection SearchDirection
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
func StartSearch(state *EditorState, direction SearchDirection) {
	buffer := state.documentBuffer
	prevQuery, prevDirection := buffer.search.query, buffer.search.direction
	buffer.search = searchState{
		direction:     direction,
		prevQuery:     prevQuery,
		prevDirection: prevDirection,
	}
	SetInputMode(state, InputModeSearch)
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
	SetInputMode(state, InputModeNormal)
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
	parsedQuery := parseQuery(q)
	if buffer.search.direction == SearchDirectionForward {
		foundMatch, matchStartPos = searchTextForward(
			buffer.cursor.position,
			buffer.textTree,
			parsedQuery)
	} else {
		foundMatch, matchStartPos = searchTextBackward(
			buffer.cursor.position,
			buffer.textTree,
			parsedQuery)
	}

	if !foundMatch {
		buffer.search.match = nil
		ScrollViewToCursor(state)
		return
	}

	buffer.search.match = &SearchMatch{
		StartPos: matchStartPos,
		EndPos:   matchStartPos + uint64(utf8.RuneCountInString(parsedQuery.queryText)),
	}
	scrollViewToPosition(buffer, matchStartPos)
}

// FindNextMatch moves the cursor to the next position matching the search query.
func FindNextMatch(state *EditorState, reverse bool) {
	buffer := state.documentBuffer
	parsedQuery := parseQuery(buffer.search.query)

	direction := buffer.search.direction
	if reverse {
		direction = direction.Reverse()
	}

	foundMatch, newCursorPos := false, uint64(0)
	if direction == SearchDirectionForward {
		foundMatch, newCursorPos = searchTextForward(
			buffer.cursor.position+1,
			buffer.textTree,
			parsedQuery)
	} else {
		foundMatch, newCursorPos = searchTextBackward(
			buffer.cursor.position,
			buffer.textTree,
			parsedQuery)
	}

	if foundMatch {
		buffer.cursor = cursorState{position: newCursorPos}
	}
}

type parsedQuery struct {
	queryText     string
	caseSensitive bool
}

// parseQuery interprets the user's search query.
// By default, if the query is all lowercase, it's case-insensitive;
// otherwise, it's case-sensitive (equivalent to vim's smartcase option).
// Users can override this by setting the suffix to "\c" for case-insensitive
// and "\C" for case-sensitive.
func parseQuery(rawQuery string) parsedQuery {
	if strings.HasSuffix(rawQuery, `\c`) {
		return parsedQuery{
			queryText:     rawQuery[0 : len(rawQuery)-2],
			caseSensitive: false,
		}
	}

	if strings.HasSuffix(rawQuery, `\C`) {
		return parsedQuery{
			queryText:     rawQuery[0 : len(rawQuery)-2],
			caseSensitive: true,
		}
	}

	var caseSensitive bool
	for _, r := range rawQuery {
		if unicode.IsUpper(r) {
			caseSensitive = true
			break
		}
	}

	return parsedQuery{
		queryText:     rawQuery,
		caseSensitive: caseSensitive,
	}

}

func transformerForSearch(caseSensitive bool) transform.Transformer {
	if caseSensitive {
		// No transformation for case-sensitive search.
		return transform.Nop
	} else {
		// Make the search case-insensitive by lowercasing the query and document.
		return cases.Lower(language.Und)
	}
}

// searchTextForward finds the position of the next occurrence of a query string on or after the start position.
func searchTextForward(startPos uint64, tree *text.Tree, parsedQuery parsedQuery) (bool, uint64) {
	transformer := transformerForSearch(parsedQuery.caseSensitive)
	transformedQuery, _, err := transform.String(transformer, parsedQuery.queryText)
	if err != nil {
		panic(err)
	}

	// Search forward from the start position to the end of the text, looking for the first match.
	searcher := text.NewSearcher(transformedQuery)
	treeReader := tree.ReaderAtPosition(startPos)
	transformedReader := transform.NewReader(&treeReader, transformer)
	foundMatch, matchOffset, err := searcher.NextInReader(transformedReader)
	if err != nil {
		panic(err) // should never happen for text.Reader.
	}

	if foundMatch {
		return true, startPos + matchOffset
	}

	// Wraparound search from the beginning of the text to the start position.
	treeReader = tree.ReaderAtPosition(0)
	transformedReader = transform.NewReader(&treeReader, transformer)
	foundMatch, matchOffset, err = searcher.Limit(startPos).NextInReader(transformedReader)
	if err != nil {
		panic(err)
	}
	return foundMatch, matchOffset
}

// searchTextBackward finds the beginning of the previous match before the start position.
func searchTextBackward(startPos uint64, tree *text.Tree, parsedQuery parsedQuery) (bool, uint64) {
	transformer := transformerForSearch(parsedQuery.caseSensitive)
	transformedQuery, _, err := transform.String(transformer, parsedQuery.queryText)
	if err != nil {
		panic(err)
	}

	// Search from the beginning of the text just past the start position, looking for the last match.
	// Set the limit to startPos + queryLen - 1 to include matches overlapping startPos, but not startPos itself.
	searcher := text.NewSearcher(transformedQuery)
	treeReader := tree.ReaderAtPosition(0)
	transformedReader := transform.NewReader(&treeReader, transformer)
	limit := startPos + uint64(utf8.RuneCountInString(transformedQuery))
	if limit > 0 {
		limit--
	}
	foundMatch, matchOffset, err := searcher.Limit(limit).LastInReader(transformedReader)
	if err != nil {
		panic(err) // should never happen for text.Reader.
	}

	if foundMatch {
		return true, matchOffset
	}

	// Wraparound search from the start position to the end of the text, looking for the last match.
	// Begin the search at startPos + 1 to exclude a potential match at startPos.
	readerStartPos := startPos + 1
	treeReader = tree.ReaderAtPosition(readerStartPos)
	transformedReader = transform.NewReader(&treeReader, transformer)
	foundMatch, matchOffset, err = searcher.NoLimit().LastInReader(transformedReader)
	if err != nil {
		panic(err)
	}
	return foundMatch, readerStartPos + matchOffset
}
