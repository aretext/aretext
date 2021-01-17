package exec

import (
	"io"
	"log"

	"github.com/aretext/aretext/internal/pkg/text"
	"github.com/aretext/aretext/internal/pkg/text/segment"
)

// posRange is a contiguous range of positions within a text.
type posRange struct {
	startPos uint64 // inclusive
	endPos   uint64 // exclusive
}

// scrollMargin is the number of lines at the beginning and end of the displayed text
// where a cursor movement would trigger a scroll.
const scrollMargin = 3

// ScrollToCursor returns a new view origin such that the cursor is visible.
// It attempts to display a few lines before/after the cursor to help the user navigate.
func ScrollToCursor(cursorPos uint64, tree *text.Tree, viewOrigin, viewWidth, viewHeight uint64) uint64 {
	wrapConfig := segment.NewLineWrapConfig(uint64(viewWidth), GraphemeClusterWidth)
	rng := visibleRangeWithinMargin(tree, viewOrigin, wrapConfig, viewHeight)
	if cursorPos < rng.startPos {
		// scroll backward
		return scrollToCursor(cursorPos, maxLinesAboveCursorScrollBackward(viewHeight), tree, wrapConfig)
	} else if cursorPos >= rng.endPos {
		// scroll forward
		return scrollToCursor(cursorPos, maxLinesAboveCursorScrollForward(viewHeight), tree, wrapConfig)
	} else {
		// cursor is already visible and within the margins, so don't move the view origin
		return viewOrigin
	}
}

func maxLinesAboveCursorScrollBackward(viewHeight uint64) uint64 {
	// ===================
	// |  scroll margin  | <- return this height
	// -------------------
	// |                 |
	// |                 |
	// ===================
	if scrollMargin < viewHeight {
		return scrollMargin
	} else if viewHeight > 0 {
		return viewHeight - 1
	} else {
		return 0
	}
}

func maxLinesAboveCursorScrollForward(viewHeight uint64) uint64 {
	// ===================
	// |                 |
	// |                 | <- return this height
	// |                 |
	// -------------------
	// |  scroll margin  |
	// ===================
	if viewHeight > scrollMargin {
		return viewHeight - scrollMargin - 1
	} else if viewHeight > 0 {
		return viewHeight - 1
	} else {
		return 0
	}
}

// visibleRangeWithinMargin returns a range of visible characters, excluding the scroll margin at the top and bottom.
// Cursor movements within this range will NOT trigger scrolling.
// This is an important performance optimization because scrolling is computationally expensive.
func visibleRangeWithinMargin(tree *text.Tree, viewOrigin uint64, wrapConfig segment.LineWrapConfig, viewHeight uint64) posRange {
	lines := visibleLineRanges(tree, viewOrigin, wrapConfig, viewHeight)

	if len(lines) == 0 {
		return posRange{}
	}

	margin := 0
	if len(lines) > scrollMargin*2 {
		margin = scrollMargin
	} else if len(lines) >= 3 {
		margin = 1
	}

	rng := posRange{
		startPos: lines[margin].startPos,
		endPos:   lines[len(lines)-1-margin].endPos,
	}

	if lines[0].startPos == 0 {
		rng.startPos = lines[0].startPos
	}

	if lines[len(lines)-1].endPos == tree.NumChars() {
		rng.endPos = lines[len(lines)-1].endPos + 1
	}

	return rng
}

// visibleLineRanges returns the range for each soft- or hard-wrapped line visible in the current view.
// For hard-wrapped lines, the newline character position is included in the line it terminates.
func visibleLineRanges(tree *text.Tree, viewOrigin uint64, wrapConfig segment.LineWrapConfig, viewHeight uint64) []posRange {
	reader := tree.ReaderAtPosition(viewOrigin, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	wrappedLineIter := segment.NewWrappedLineIter(runeIter, wrapConfig)
	wrappedLine := segment.NewSegment()
	pos := viewOrigin
	lineRanges := make([]posRange, 0, viewHeight)
	var prevHadNewline bool

	for row := uint64(0); row < viewHeight; row++ {
		err := wrappedLineIter.NextSegment(wrappedLine)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("%s", err)
		}

		lineRanges = append(lineRanges, posRange{
			startPos: pos,
			endPos:   pos + wrappedLine.NumRunes(),
		})

		pos += wrappedLine.NumRunes()
		prevHadNewline = wrappedLine.HasNewline()
	}

	// If the last visible line ends with a newline, then a cursor positioned at the end of the line
	// would be displayed on the next line, which is not visible.  For this reason, we exclude
	// the newline character from the visible range.
	if prevHadNewline && uint64(len(lineRanges)) == viewHeight {
		lineRanges[len(lineRanges)-1].endPos--
	}

	return lineRanges
}

// scrollToCursor returns a view origin at the start of a line such that the cursor is visible.
// It attempts to display maxLinesAboveCursor before the cursor's line unless this would go past the start of the text.
// The complexity is worst-case O(n) for n runes in the text due to the scan backwards for the start of the cursor's line.
func scrollToCursor(cursorPos uint64, maxLinesAboveCursor uint64, tree *text.Tree, wrapConfig segment.LineWrapConfig) uint64 {
	lineStartPos := tree.LineStartPosition(tree.LineNumForPosition(cursorPos))
	wrappedLines := softWrapLineUntil(lineStartPos, tree, wrapConfig, func(rng posRange) bool {
		return cursorPos >= rng.startPos && cursorPos < rng.endPos
	})

	// If we've found enough lines before the cursor, we're done.
	numWrappedLines := uint64(len(wrappedLines))
	if numWrappedLines > maxLinesAboveCursor {
		return wrappedLines[numWrappedLines-1-maxLinesAboveCursor].startPos
	}

	if lineStartPos == 0 {
		return 0
	}

	// We still need more lines before the cursor, so recurse.
	endOfPrevLine := lineStartPos - 1
	remainingLines := maxLinesAboveCursor - numWrappedLines
	return scrollToCursor(endOfPrevLine, remainingLines, tree, wrapConfig)
}

// softWrapLineUntil returns ranges for soft-wrapped lines in a line until a given stop condition occurs.
func softWrapLineUntil(lineStartPos uint64, tree *text.Tree, wrapConfig segment.LineWrapConfig, stopFunc func(posRange) bool) []posRange {
	reader := tree.ReaderAtPosition(lineStartPos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	wrappedLineIter := segment.NewWrappedLineIter(runeIter, wrapConfig)
	wrappedLine := segment.NewSegment()
	pos := lineStartPos
	result := make([]posRange, 0, 1)
	prevHadNewline := true // Assume we're at the start of a hard-wrapped line.

	for {
		err := wrappedLineIter.NextSegment(wrappedLine)
		if err == io.EOF {
			if prevHadNewline {
				// If the text ends with a newline, then there is one empty line at the end.
				result = append(result, posRange{
					startPos: pos,
					endPos:   pos,
				})
			}
			break
		} else if err != nil {
			log.Fatalf("%s", err)
		}

		lineRange := posRange{
			startPos: pos,
			endPos:   pos + wrappedLine.NumRunes(),
		}

		result = append(result, lineRange)

		hasNewline := wrappedLine.HasNewline()
		if hasNewline || stopFunc(lineRange) {
			break
		}

		pos = lineRange.endPos
		prevHadNewline = hasNewline
	}

	return result
}
