package state

// CheckpointUndoLog sets a checkpoint at the current position in the undo log.
func CheckpointUndoLog(state *EditorState) {
	state.documentBuffer.undoLog.Checkpoint()
}
