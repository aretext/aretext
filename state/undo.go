package state

import (
	"log"
	"math"

	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/undo"
)

// CheckpointUndoLog sets a checkpoint at the current position in the undo log.
func CheckpointUndoLog(state *EditorState) {
	state.documentBuffer.undoLog.Checkpoint()
}

// Undo returns the document to its state at the last undo checkpoint.
func Undo(state *EditorState) {
	ops := state.documentBuffer.undoLog.UndoToLastCheckpoint()
	if len(ops) == 0 {
		return
	}

	minPos := uint64(math.MaxUint64)
	for _, op := range ops {
		log.Printf("Undo operation: %#v\n", op)
		if err := applyOpFromUndoLog(state, op); err != nil {
			log.Printf("Could not apply undo op %v: %v\n", op, err)
			continue
		}
		if pos := op.Position(); pos < minPos {
			minPos = pos
		}
	}
	MoveCursor(state, locateCursorAfterUndoOrRedo(minPos))
}

// Redo reverses the last undo operation.
func Redo(state *EditorState) {
	ops := state.documentBuffer.undoLog.RedoToNextCheckpoint()
	if len(ops) == 0 {
		return
	}

	minPos := uint64(math.MaxUint64)
	for _, op := range ops {
		log.Printf("Redo operation: %#v\n", op)
		if err := applyOpFromUndoLog(state, op); err != nil {
			log.Printf("Could not apply redo op %v: %v\n", op, err)
			continue
		}
		if pos := op.Position(); pos < minPos {
			minPos = pos
		}
	}
	MoveCursor(state, locateCursorAfterUndoOrRedo(minPos))
}

func applyOpFromUndoLog(state *EditorState, op undo.Op) error {
	pos := op.Position()
	if s := op.TextToInsert(); len(s) > 0 {
		return insertTextAtPosition(state, s, pos, false)
	} else if n := op.NumRunesToDelete(); n > 0 {
		deleteRunes(state, pos, uint64(n), false)
	}
	return nil
}

func locateCursorAfterUndoOrRedo(minPos uint64) Locator {
	return func(params LocatorParams) uint64 {
		lineStartPos := locate.PrevLineBoundary(params.TextTree, minPos)
		indentedLineStartPos := locate.NextNonWhitespaceOrNewline(params.TextTree, lineStartPos)
		if indentedLineStartPos > minPos {
			return indentedLineStartPos
		}
		return locate.ClosestCharOnLine(params.TextTree, minPos)
	}
}
