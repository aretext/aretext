package exec

import (
	"fmt"
	"io"

	"github.com/wedaly/aretext/internal/pkg/text"
	"github.com/wedaly/aretext/internal/pkg/text/breaks"
)

// Locator finds the position of the cursor according to some criteria.
type Locator interface {
	fmt.Stringer

	// Locate finds the next position of the cursor based on the current state and criteria for this locator.
	Locate(state *State) cursorState
}

// charInLineLocator locates a character (grapheme cluster) in the current line.
type charInLineLocator struct {
	direction              text.ReadDirection
	count                  uint64
	includeEndOfLineOrFile bool
}

// NewCharInLineLocator builds a new locator for a character on the same line as the cursor.
// The direction arg indicates whether to read forward or backwards from the cursor.
// The count arg is the maximum number of characters to move the cursor.
func NewCharInLineLocator(direction text.ReadDirection, count uint64, includeEndOfLineOrFile bool) Locator {
	if count == 0 {
		panic("Count must be greater than zero")
	}
	return &charInLineLocator{direction, count, includeEndOfLineOrFile}
}

func (loc *charInLineLocator) String() string {
	return fmt.Sprintf("CharInLineLocator(%s, %d, %t)", directionString(loc.direction), loc.count, loc.includeEndOfLineOrFile)
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

	// Ignore the breakpoint at the end of the text
	if err := breaks.SkipBreak(gcIter); err != nil {
		panic(err)
	}

	var prevBreak uint64
	for i := uint64(0); i < loc.count; i++ {
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
		if gcHasNewline(runeIter, nextBreak-prevBreak) {
			if loc.includeEndOfLineOrFile {
				prevBreak = nextBreak
			}
			break
		}

		prevBreak = nextBreak
	}

	return startPos - prevBreak
}

func (loc *charInLineLocator) findPositionAfterCursor(state *State) uint64 {
	startPos := state.cursor.position
	reader := state.tree.ReaderAtPosition(startPos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	gcIter := breaks.NewGraphemeClusterBreakIter(runeIter.Clone())

	// Ignore the breakpoint at the start of the text
	if err := breaks.SkipBreak(gcIter); err != nil {
		panic(err)
	}

	var endOfLineOrFile bool
	var prevBreak, prevPrevBreak uint64
	for i := uint64(0); i <= loc.count; i++ {
		nextBreak, err := gcIter.NextBreak()
		if err == io.EOF {
			endOfLineOrFile = true
			break
		} else if err != nil {
			panic(err) // We assume the input is valid UTF-8, so this should never happen.
		}

		// Check if the next grapheme cluster contains a newline.
		// This will consume runes from `runeIter`, which is a clone of the rune iter
		// used by gcIter.  This keeps both gcIter and runeIter synchronized.
		if gcHasNewline(runeIter, nextBreak-prevBreak) {
			endOfLineOrFile = true
			break
		}

		prevPrevBreak = prevBreak
		prevBreak = nextBreak
	}

	if endOfLineOrFile && loc.includeEndOfLineOrFile {
		return startPos + prevBreak
	}
	return startPos + prevPrevBreak
}

// ontoLineLocator finds the closest grapheme cluster on a line (not newline or past end of text).
// This is useful for "resetting" the cursor onto a line
// (for example, after deleting the last character on the line or exiting insert mode).
type ontoLineLocator struct {
}

func NewOntoLineLocator() Locator {
	return &ontoLineLocator{}
}

// Locate finds the closest grapheme cluster on a line (not newline or past end of text).
func (loc *ontoLineLocator) Locate(state *State) cursorState {
	// If past the end of the text, return the start of the last grapheme cluster.
	numChars := state.tree.NumChars()
	if state.cursor.position >= numChars {
		newPos := loc.findPrevGraphemeCluster(state.tree, numChars, 1)
		return cursorState{position: newPos}
	}

	// If on a grapheme cluster with a newline (either "\n" or "\r\n"), return the start
	// of the last grapheme cluster before the current grapheme cluster.
	if hasNewline, afterNewlinePos := loc.findNewlineAtPos(state.tree, state.cursor.position); hasNewline {
		newPos := loc.findPrevGraphemeCluster(state.tree, afterNewlinePos, 2)
		return cursorState{position: newPos}
	}

	// The cursor is already on a line, so do nothing.
	return cursorState{position: state.cursor.position}
}

func (loc *ontoLineLocator) findNewlineAtPos(tree *text.Tree, pos uint64) (bool, uint64) {
	reader := tree.ReaderAtPosition(pos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	gcIter := breaks.NewGraphemeClusterBreakIter(runeIter.Clone())

	// Skip break at start of text.
	if err := breaks.SkipBreak(gcIter); err != nil {
		panic(err)
	}

	// Check if the next grapheme cluster has a newline.
	nextBreak, err := gcIter.NextBreak()
	if err == io.EOF {
		return false, 0
	} else if err != nil {
		panic(err)
	}

	if gcHasNewline(runeIter, nextBreak) {
		return true, pos + nextBreak
	}

	return false, 0
}

func (loc *ontoLineLocator) findPrevGraphemeCluster(tree *text.Tree, pos uint64, count int) uint64 {
	reader := tree.ReaderAtPosition(pos, text.ReadDirectionBackward)
	runeIter := text.NewCloneableBackwardRuneIter(reader)
	gcIter := breaks.NewReverseGraphemeClusterBreakIter(runeIter)

	for i := 0; i < count; i++ {
		if err := breaks.SkipBreak(gcIter); err != nil {
			panic(err)
		}
	}

	nextBreak, err := gcIter.NextBreak()
	if err == io.EOF {
		return 0
	} else if err != nil {
		panic(err)
	}
	return pos - nextBreak
}

func (loc *ontoLineLocator) String() string {
	return "OntoLineLocator()"
}

// Check if the current grapheme cluster contains a newline.
// runeIter must be positioned at the beginning of the grapheme cluster.
// This will consume from runeIter, so callers that want to retain the current position
// should provide a clone of the iterator.
func gcHasNewline(runeIter text.RuneIter, gcLen uint64) bool {
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

func directionString(direction text.ReadDirection) string {
	switch direction {
	case text.ReadDirectionForward:
		return "forward"
	case text.ReadDirectionBackward:
		return "backward"
	default:
		panic("Unrecognized direction")
	}
}
