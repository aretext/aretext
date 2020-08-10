package display

import (
	"io"

	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/text"
	"github.com/wedaly/aretext/internal/pkg/text/segment"
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
func ScrollToCursor(cursorPos uint64, tree *text.Tree, viewOrigin uint64, viewWidth, viewHeight int) uint64 {
	wrapConfig := segment.NewLineWrapConfig(uint64(viewWidth), exec.GraphemeClusterWidth)
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

func maxLinesAboveCursorScrollBackward(viewHeight int) int {
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

func maxLinesAboveCursorScrollForward(viewHeight int) int {
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
func visibleRangeWithinMargin(tree *text.Tree, viewOrigin uint64, wrapConfig segment.LineWrapConfig, viewHeight int) posRange {
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
func visibleLineRanges(tree *text.Tree, viewOrigin uint64, wrapConfig segment.LineWrapConfig, viewHeight int) []posRange {
	reader := tree.ReaderAtPosition(viewOrigin, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	wrappedLineIter := segment.NewWrappedLineIter(runeIter, wrapConfig)
	pos := viewOrigin
	lineRanges := make([]posRange, 0, viewHeight)

	for row := 0; row < viewHeight; row++ {
		wrappedLine, err := wrappedLineIter.NextSegment()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		lineRanges = append(lineRanges, posRange{
			startPos: pos,
			endPos:   pos + wrappedLine.NumRunes(),
		})

		pos += wrappedLine.NumRunes()
	}

	return lineRanges
}

// scrollToCursor returns a view origin at the start of a line such that the cursor is visible.
// It attempts to display maxLinesAboveCursor before the cursor's line unless this would go past the start of the text.
// The complexity is worst-case O(n) for n runes in the text due to the scan backwards for the start of the cursor's line.
func scrollToCursor(cursorPos uint64, maxLinesAboveCursor int, tree *text.Tree, wrapConfig segment.LineWrapConfig) uint64 {
	lineStartPos := startOfLine(cursorPos, tree)
	wrappedLines := softWrapLineUntil(lineStartPos, tree, wrapConfig, func(rng posRange) bool {
		return cursorPos >= rng.startPos && cursorPos < rng.endPos
	})

	// If we've found enough lines before the cursor, we're done.
	if len(wrappedLines) > maxLinesAboveCursor {
		return wrappedLines[len(wrappedLines)-1-maxLinesAboveCursor].startPos
	}

	if lineStartPos == 0 {
		return 0
	}

	// We still need more lines before the cursor, so recurse.
	endOfPrevLine := lineStartPos - 1
	remainingLines := maxLinesAboveCursor - len(wrappedLines)
	return scrollToCursor(endOfPrevLine, remainingLines, tree, wrapConfig)
}

// startOfLine returns the position of the first character in the line containing cursorPos.
func startOfLine(cursorPos uint64, tree *text.Tree) uint64 {
	reader := tree.ReaderAtPosition(cursorPos, text.ReadDirectionBackward)
	runeIter := text.NewCloneableBackwardRuneIter(reader)
	pos := cursorPos
	for {
		r, err := runeIter.NextRune()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		} else if r == '\n' {
			break
		}
		pos--
	}
	return pos
}

// softWrapLineUntil returns ranges for soft-wrapped lines in a line until a given stop condition occurs.
func softWrapLineUntil(lineStartPos uint64, tree *text.Tree, wrapConfig segment.LineWrapConfig, stopFunc func(posRange) bool) []posRange {
	reader := tree.ReaderAtPosition(lineStartPos, text.ReadDirectionForward)
	runeIter := text.NewCloneableForwardRuneIter(reader)
	wrappedLineIter := segment.NewWrappedLineIter(runeIter, wrapConfig)
	pos := lineStartPos
	result := make([]posRange, 0, 1)
	for {
		wrappedLine, err := wrappedLineIter.NextSegment()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		lineRange := posRange{
			startPos: pos,
			endPos:   pos + wrappedLine.NumRunes(),
		}

		result = append(result, lineRange)

		if wrappedLine.HasNewline() || stopFunc(lineRange) {
			break
		}

		pos = lineRange.endPos
	}

	return result
}
