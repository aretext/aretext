package state

import (
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/text"
)

// LocatorParams are inputs to a function that locates a position in the document.
type LocatorParams struct {
	TextTree          *text.Tree
	CursorPos         uint64
	AutoIndentEnabled bool
	TabSize           uint64
}

func locatorParamsForBuffer(buffer *BufferState) LocatorParams {
	return LocatorParams{
		TextTree:          buffer.textTree,
		CursorPos:         buffer.cursor.position,
		AutoIndentEnabled: buffer.autoIndent,
		TabSize:           buffer.tabSize,
	}
}

// Locator is a function that locates a position in the document.
type Locator func(LocatorParams) uint64

// RangeLocator is a function that locates the start and end positions
// of a range in the document (for example, a word or selection).
// The (start, end] interval does NOT include the end position.
type RangeLocator func(LocatorParams) (uint64, uint64)

// SelectionEndLocator returns a locator for the end of a selection.
// For example, suppose a user has selected the first two lines of a document.
// When the user repeats an action for this selection (by using the "." command),
// the command repeats for the two lines starting at the *new* cursor location.
// The selection end locator determines the end of the selection at the new cursor.
//
// For linewise selections, it attempts to select the same number of lines.
// Example where cursor moves from line two to line three:
//
//    abcd        abcd
//    [efg   -->  efg
//    hij]        [hij
//    klm         klm]
//
// For charwise selections it attempts to select down the same number of lines
// and over the same number of columns (grapheme clusters) on the final line
// of the selection.
// Example where the cursor moves from the second col in the first line
// to the third col in the second line:
//
//   ab[cd        abcd
//   ef]g    -->  ef[g
//   hij          hi]j
//   klm          klm
//
func SelectionEndLocator(textTree *text.Tree, cursorPos uint64, selector *selection.Selector) Locator {
	r := selector.Region(textTree, cursorPos)
	switch selector.Mode() {
	case selection.ModeNone:
		return nil
	case selection.ModeChar:
		return charwiseSelectionEndLocator(textTree, r)
	case selection.ModeLine:
		return linewiseSelectionEndLocator(textTree, r)
	default:
		panic("Unrecognized selection mode")
	}
}

func charwiseSelectionEndLocator(textTree *text.Tree, r selection.Region) Locator {
	startLineNum := textTree.LineNumForPosition(r.StartPos)
	endLineNum := textTree.LineNumForPosition(r.EndPos)
	if startLineNum == endLineNum {
		numGcFromStartToEnd := locate.NumGraphemeClustersInRange(textTree, r.StartPos, r.EndPos)
		return func(p LocatorParams) uint64 {
			return locate.NextCharInLine(p.TextTree, numGcFromStartToEnd, true, p.CursorPos)
		}
	}

	numLinesDown := endLineNum - startLineNum
	startOfLinePos := locate.StartOfLineAtPos(textTree, r.EndPos)
	numGcPastStartOfLine := locate.NumGraphemeClustersInRange(textTree, startOfLinePos, r.EndPos)
	return func(p LocatorParams) uint64 {
		startOfLineBelowPos := locate.StartOfLineBelow(p.TextTree, numLinesDown, p.CursorPos)
		if startOfLineBelowPos > p.CursorPos {
			// Moved down a line, now move over.
			return locate.NextCharInLine(p.TextTree, numGcPastStartOfLine, true, startOfLineBelowPos)
		} else {
			// On last line, so select to the end of the document.
			return locate.NextLineBoundary(p.TextTree, true, p.CursorPos)
		}
	}
}

func linewiseSelectionEndLocator(textTree *text.Tree, r selection.Region) Locator {
	startLineNum := textTree.LineNumForPosition(r.StartPos)
	endLineNum := textTree.LineNumForPosition(r.EndPos)
	numLinesDown := endLineNum - startLineNum
	return func(p LocatorParams) uint64 {
		pos := locate.StartOfLineBelow(p.TextTree, numLinesDown, p.CursorPos)
		return locate.NextLineBoundary(p.TextTree, true, pos)
	}
}
