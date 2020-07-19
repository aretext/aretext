package exec

import (
	"io"

	"github.com/wedaly/aretext/internal/pkg/text"
	"github.com/wedaly/aretext/internal/pkg/text/breaks"
)

// Locator finds the position of the cursor according to some criteria.
type Locator interface {
	// Locate finds the next position of the cursor based on the current state and criteria for this locator.
	Locate(state *State) cursorState
}

// charInLineLocator locates a character (grapheme cluster) in the current line.
type charInLineLocator struct {
	direction text.ReadDirection
	count     uint64
}

// NewCharInLineLocator builds a new locator for a character on the same line as the cursor.
// The direction arg indicates whether to read forward or backwards from the cursor.
// The count arg is the maximum number of characters to move the cursor.
func NewCharInLineLocator(direction text.ReadDirection, count uint64) Locator {
	return &charInLineLocator{direction, count}
}

// Locate finds a character to the right of the cursor on the current line.
func (loc *charInLineLocator) Locate(state *State) cursorState {
	newPosition := loc.findPosition(state)

	logicalOffset := uint64(0)
	if newPosition == state.cursor.position {
		// This handles the case where the user is moving the cursor up to a shorter line,
		// then tries to move the cursor to the right at the end of the line.
		// The cursor doesn't actually move, so when the user moves up another line,
		// it should use the offset from the longest line.
		logicalOffset = state.cursor.logicalOffset
	}

	return cursorState{
		position:      newPosition,
		logicalOffset: logicalOffset,
	}
}

func (loc *charInLineLocator) findPosition(state *State) uint64 {
	if loc.direction == text.ReadDirectionBackward {
		return loc.findPositionBeforeCursor(state)
	}
	return loc.findPositionAfterCursor(state)
}

func (loc *charInLineLocator) findPositionBeforeCursor(state *State) uint64 {
	startPos := state.cursor.position
	reader := state.tree.ReaderAtPosition(startPos, text.ReadDirectionBackward)
	runeIter := text.NewCloneableBackwardRuneIter(reader)
	gcIter := breaks.NewReverseGraphemeClusterBreakIter(runeIter.Clone())
	var moveCount, prevBreak uint64
	for moveCount < loc.count {
		nextBreak, err := gcIter.NextBreak()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err) // We assume the input is valid UTF-8, so this should never happen.
		}

		if nextBreak >= startPos {
			return 0
		}

		// Check if the next grapheme cluster contains a newline.
		// This will consume runes from `runeIter`, which is a clone of the rune iter
		// used by gcIter.  This keeps both gcIter and runeIter synchronized.
		if loc.gcHasNewline(runeIter, nextBreak-prevBreak) {
			break
		}

		prevBreak = nextBreak

		// The first break is always at the start of the text, which we can ignore
		// because it won't move the cursor.
		if nextBreak > 0 {
			moveCount++
		}
	}

	return startPos - prevBreak
}

func (loc *charInLineLocator) findPositionAfterCursor(state *State) uint64 {
	startPos := state.cursor.position
	reader := state.tree.ReaderAtPosition(startPos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	eofBreak := state.tree.NumChars() - startPos
	gcIter := breaks.NewGraphemeClusterBreakIter(runeIter.Clone())
	var moveCount, prevBreak, prevPrevBreak uint64
	for moveCount <= loc.count {
		nextBreak, err := gcIter.NextBreak()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err) // We assume the input is valid UTF-8, so this should never happen.
		}

		if nextBreak > eofBreak {
			break
		}

		// Check if the next grapheme cluster contains a newline.
		// This will consume runes from `runeIter`, which is a clone of the rune iter
		// used by gcIter.  This keeps both gcIter and runeIter synchronized.
		if loc.gcHasNewline(runeIter, nextBreak-prevBreak) {
			break
		}

		prevPrevBreak = prevBreak
		prevBreak = nextBreak

		// The first break is always at the start of the text, which we can ignore
		// because it won't move the cursor.
		if nextBreak > 0 {
			moveCount++
		}
	}

	return startPos + prevPrevBreak
}

func (loc *charInLineLocator) gcHasNewline(runeIter text.RuneIter, gcLen uint64) bool {
	for i := uint64(0); i < gcLen; i++ {
		r, err := runeIter.NextRune()
		if err != nil {
			// We assume that the input text is valid UTF-8 and that we're given
			// a grapheme cluster before the end of the text, so this should never happen.
			panic(err)
		}

		if r == '\n' {
			return true
		}
	}

	return false
}
