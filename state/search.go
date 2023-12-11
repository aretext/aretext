package state

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/transform"

	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/locate"
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

// SearchCompleteAction is the action to perform when a user completes a search.
type SearchCompleteAction func(*EditorState, string, SearchDirection, SearchMatch)

// searchState represents the state of a text search.
type searchState struct {
	query          string
	direction      SearchDirection
	completeAction SearchCompleteAction
	prevQuery      string
	prevDirection  SearchDirection
	history        []string
	historyIdx     int
	match          *SearchMatch
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
func StartSearch(state *EditorState, direction SearchDirection, completeAction SearchCompleteAction) {
	search := &state.documentBuffer.search
	prevQuery, prevDirection := search.query, search.direction
	*search = searchState{
		direction:      direction,
		completeAction: completeAction,
		prevQuery:      prevQuery,
		prevDirection:  prevDirection,
		history:        search.history,
		historyIdx:     len(search.history),
	}
	SetInputMode(state, InputModeSearch)
}

// CompleteSearch terminates a text search and returns to normal mode.
// If commit is true, execute the complete search action.
// Otherwise, return to the original cursor position.
func CompleteSearch(state *EditorState, commit bool) {
	search := &state.documentBuffer.search

	if search.query != "" {
		if len(search.history) == 0 || search.history[len(search.history)-1] != search.query {
			search.history = append(search.history, search.query)
		}
	}

	// Return to normal mode.
	// This must run BEFORE executing the complete action, because some actions
	// change the input mode again to insert mode (specifically "c/" and "c?")
	SetInputMode(state, InputModeNormal)

	if commit {
		if search.match != nil {
			search.completeAction(state, search.query, search.direction, *search.match)
		}
	} else {
		prevQuery, prevDirection := search.prevQuery, search.prevDirection
		*search = searchState{
			query:     prevQuery,
			direction: prevDirection,
			history:   search.history,
		}
	}

	search.match = nil

	ScrollViewToCursor(state)
}

// AppendRuneToSearchQuery appends a rune to the text search query.
func AppendRuneToSearchQuery(state *EditorState, r rune) {
	search := &state.documentBuffer.search
	q := search.query + string(r)
	runTextSearchQuery(state, q)
	search.historyIdx = len(search.history)
}

// DeleteRuneFromSearchQuery deletes the last rune from the text search query.
// A deletion in an empty query aborts the search and returns the editor to normal mode.
func DeleteRuneFromSearchQuery(state *EditorState) {
	search := &state.documentBuffer.search
	if len(search.query) == 0 {
		CompleteSearch(state, false)
		return
	}

	q := search.query[0 : len(search.query)-1]
	runTextSearchQuery(state, q)
	search.historyIdx = len(search.history)
}

// SetSearchQueryToPrevInHistory sets the search query to a previous search query in the history.
func SetSearchQueryToPrevInHistory(state *EditorState) {
	search := &state.documentBuffer.search
	if search.historyIdx == 0 {
		return
	}
	search.historyIdx--
	q := search.history[search.historyIdx]
	runTextSearchQuery(state, q)
}

// SetSearchQueryToNextInHistory sets the search query to the next search query in the history.
func SetSearchQueryToNextInHistory(state *EditorState) {
	search := &state.documentBuffer.search
	if search.historyIdx >= len(search.history)-1 {
		return
	}

	search.historyIdx++
	q := search.history[search.historyIdx]
	runTextSearchQuery(state, q)
}

// SearchWordUnderCursor starts a search for the word under the cursor.
func SearchWordUnderCursor(state *EditorState, direction SearchDirection, completeAction SearchCompleteAction, targetCount uint64) {
	// Retrieve the current word under the cursor.
	// If the cursor is on leading whitespace, this will retrieve the word after the whitespace.
	buffer := state.documentBuffer
	wordStartPos, wordEndPos := locate.WordObject(buffer.textTree, buffer.cursor.position, targetCount)
	word := strings.TrimSpace(copyText(buffer.textTree, wordStartPos, wordEndPos-wordStartPos))
	if word == "" {
		return
	}

	query := fmt.Sprintf("%s\\C", word) // Force case-sensitive search.

	// Search for the word.
	StartSearch(state, direction, completeAction)
	runTextSearchQuery(state, query)
	CompleteSearch(state, true)
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
			buffer.cursor.position,
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

// searchTextForward finds the position of the next occurrence of a query string after the start position.
func searchTextForward(startPos uint64, tree *text.Tree, parsedQuery parsedQuery) (bool, uint64) {
	// Start the search one after the provided start position so we skip a match on the current position.
	startPos++

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
	limit := startPos + uint64(utf8.RuneCountInString(transformedQuery))
	if limit > 0 {
		limit--
	}
	foundMatch, matchOffset, err = searcher.Limit(limit).NextInReader(transformedReader)
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

// SearchCompleteMoveCursorToMatch is a SearchCompleteAction that moves the cursor to the start of the search match.
func SearchCompleteMoveCursorToMatch(state *EditorState, query string, direction SearchDirection, match SearchMatch) {
	state.documentBuffer.cursor = cursorState{position: match.StartPos}
}

// SearchCompleteDeleteToMatch is a SearchCompleteAction that deletes from the cursor position to the search match.
func SearchCompleteDeleteToMatch(clipboardPage clipboard.PageId) SearchCompleteAction {
	return func(state *EditorState, query string, direction SearchDirection, match SearchMatch) {
		completeAction := func(state *EditorState, query string, direction SearchDirection, match SearchMatch) {
			deleteToSearchMatch(state, direction, match, clipboardPage)
		}
		completeAction(state, query, direction, match)
		replaySearchInLastActionMacro(state, query, direction, completeAction)
	}
}

// SearchCompleteChangeToMatch is a SearchCompleteAction that deletes to the search match, then enters insert mode.
func SearchCompleteChangeToMatch(clipboardPage clipboard.PageId) SearchCompleteAction {
	return func(state *EditorState, query string, direction SearchDirection, match SearchMatch) {
		completeAction := func(state *EditorState, query string, direction SearchDirection, match SearchMatch) {
			// Delete to the match (exactly the same as the "search and delete" commands).
			// Then go to insert mode (override default transition back to normal mode).
			deleteToSearchMatch(state, direction, match, clipboardPage)
			SetInputMode(state, InputModeInsert)
		}
		completeAction(state, query, direction, match)
		replaySearchInLastActionMacro(state, query, direction, completeAction)
	}
}

// SearchCompleteCopyToMatch is a SearchCompleteAction that copies text from the cursor position to the search match.
func SearchCompleteCopyToMatch(clipboardPage clipboard.PageId) SearchCompleteAction {
	return func(state *EditorState, query string, direction SearchDirection, match SearchMatch) {
		// If the search wraps around, then the range start will be >= range end,
		// so nothing will be copied.
		CopyRange(state, clipboardPage, func(params LocatorParams) (uint64, uint64) {
			if direction == SearchDirectionForward {
				return params.CursorPos, match.StartPos
			} else {
				return match.EndPos, params.CursorPos
			}
		})
	}
}

func deleteToSearchMatch(state *EditorState, direction SearchDirection, match SearchMatch, clipboardPage clipboard.PageId) {
	DeleteToPos(state, func(params LocatorParams) uint64 {
		if direction == SearchDirectionForward {
			return match.StartPos
		} else {
			if params.CursorPos > match.EndPos {
				return match.EndPos
			} else {
				// Match vim's behavior for backward search with wraparound.
				return match.StartPos
			}
		}
	}, clipboardPage)
}

func replaySearchInLastActionMacro(state *EditorState, query string, direction SearchDirection, completeAction SearchCompleteAction) {
	// The "last action" must use the original search query (even if the user subsequently searched for something else).
	// Construct a macro action that performs the original search again, then performs the specified complete action.
	ClearLastActionMacro(state)
	AddToLastActionMacro(state, func(state *EditorState) {
		StartSearch(state, direction, completeAction)
		for _, r := range query {
			AppendRuneToSearchQuery(state, r)
		}
		CompleteSearch(state, true)
	})
}
