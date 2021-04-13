package selection

import (
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/text"
)

// Mode controls the selection behavior (charwise or linewise).
type Mode int

const (
	ModeNone = Mode(iota)
	ModeChar
	ModeLine
)

// Selector tracks the selected region of the document.
type Selector struct {
	mode      Mode
	anchorPos uint64
}

// Start begins a new selection from the current cursor position.
func (s *Selector) Start(mode Mode, cursorPos uint64) {
	s.mode = mode
	s.anchorPos = cursorPos
}

// Clear clears the currently selected region, if any.
func (s *Selector) Clear() {
	s.mode = ModeNone
	s.anchorPos = 0
}

// Mode returns the current selection mode.
func (s *Selector) Mode() Mode {
	return s.mode
}

// SetMode sets the selection mode.
func (s *Selector) SetMode(mode Mode) {
	s.mode = mode
}

// Region returns the currently selected region.
// If nothing is currently selected, the returned region will be empty.
func (s *Selector) Region(tree *text.Tree, cursorPos uint64) Region {
	switch s.mode {
	case ModeNone:
		return EmptyRegion
	case ModeChar:
		return s.selectedCharsRegion(tree, cursorPos)
	case ModeLine:
		return s.selectedLinesRegion(tree, cursorPos)
	default:
		panic("Unrecognized mode")
	}
}

func (s *Selector) selectedCharsRegion(tree *text.Tree, cursorPos uint64) Region {
	var r Region
	if cursorPos < s.anchorPos {
		r.StartPos, r.EndPos = cursorPos, s.anchorPos+1
	} else {
		r.StartPos, r.EndPos = s.anchorPos, cursorPos+1
	}
	return r.Clip(tree.NumChars())
}

func (s *Selector) selectedLinesRegion(tree *text.Tree, cursorPos uint64) Region {
	minPos, maxPos := cursorPos, s.anchorPos
	if minPos > maxPos {
		minPos, maxPos = maxPos, minPos
	}
	r := Region{
		StartPos: locate.StartOfLineAtPos(tree, minPos),
		EndPos:   locate.NextLineBoundary(tree, true, maxPos),
	}
	return r.Clip(tree.NumChars())
}
