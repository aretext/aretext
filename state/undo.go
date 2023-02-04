package state

import (
	"log"

	"github.com/aretext/aretext/undo"
)

// BeginUndoEntry begins a new undo entry.
// This should be called before tracking any undo operations.
func BeginUndoEntry(state *EditorState) {
	if state.macroState.isReplayingUserMacro {
		log.Printf("Skip begin undo entry because we're replaying a user macro\n")
		return
	}

	log.Printf("Begin undo entry\n")
	buffer := state.documentBuffer
	buffer.undoLog.BeginEntry(buffer.cursor.position)
}

// CommitUndoEntry commits the current undo entry.
// This should be called after completing an action that can be undone.
func CommitUndoEntry(state *EditorState) {
	if state.macroState.isReplayingUserMacro {
		log.Printf("Skip commit undo entry because we're replaying a user macro\n")
		return
	}

	log.Printf("Commit undo entry\n")
	buffer := state.documentBuffer
	buffer.undoLog.CommitEntry(buffer.cursor.position)
}

// Undo returns the document to its state at the last undo entry.
func Undo(state *EditorState) {
	hasEntry, undoOps, cursor := state.documentBuffer.undoLog.UndoToLastCommitted()
	if !hasEntry {
		return
	}

	for _, op := range undoOps {
		log.Printf("Undo operation: %#v\n", op)
		if err := applyOpFromUndoLog(state, op); err != nil {
			log.Printf("Could not apply undo op %v: %v\n", op, err)
			continue
		}
	}

	MoveCursor(state, func(LocatorParams) uint64 {
		return cursor
	})
}

// Redo reverses the last undo operation.
func Redo(state *EditorState) {
	hasEntry, redoOps, cursor := state.documentBuffer.undoLog.RedoToNextCommitted()
	if !hasEntry {
		return
	}

	for _, op := range redoOps {
		log.Printf("Redo operation: %#v\n", op)
		if err := applyOpFromUndoLog(state, op); err != nil {
			log.Printf("Could not apply redo op %v: %v\n", op, err)
			continue
		}
	}

	MoveCursor(state, func(LocatorParams) uint64 {
		return cursor
	})
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
