package parser

// Edit represents a change to a document.
type Edit struct {
	Pos         uint64 // Position of the first character inserted/deleted.
	NumInserted uint64
	NumDeleted  uint64
}

func (edit Edit) applyToPosition(pos uint64) uint64 {
	if pos >= edit.Pos {
		if updatedPos := pos + edit.NumInserted; updatedPos >= pos {
			pos = updatedPos
		} else {
			pos = uint64(0xFFFFFFFFFFFFFFFF) // overflow
		}

		if updatedPos := pos - edit.NumDeleted; updatedPos <= pos {
			pos = updatedPos
		} else {
			pos = 0 // underflow
		}

		if pos < edit.Pos {
			pos = edit.Pos
		}
	}
	return pos
}
