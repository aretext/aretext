package undo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUndoToLastCheckpoint(t *testing.T) {
	log := NewLog()
	ops := log.UndoToLastCheckpoint()
	assert.Equal(t, 0, len(ops))

	log.TrackOp(InsertOp(0, "a"))
	log.TrackOp(InsertOp(1, "bc"))
	log.Checkpoint()
	ops = log.UndoToLastCheckpoint()
	expectedOps := []Op{
		DeleteOp(1, "bc"),
		DeleteOp(0, "a"),
	}
	assert.Equal(t, expectedOps, ops)

	log.TrackOp(DeleteOp(3, "x"))
	log.TrackOp(InsertOp(4, "yz"))
	log.Checkpoint()
	log.TrackOp(DeleteOp(5, "12"))
	log.TrackOp(InsertOp(6, "345"))

	ops = log.UndoToLastCheckpoint()
	expectedOps = []Op{
		DeleteOp(6, "345"),
		InsertOp(5, "12"),
	}
	assert.Equal(t, expectedOps, ops)

	ops = log.UndoToLastCheckpoint()
	expectedOps = []Op{
		DeleteOp(4, "yz"),
		InsertOp(3, "x"),
		InsertOp(0, "a"),
		InsertOp(1, "bc"),
	}
	assert.Equal(t, expectedOps, ops)

	ops = log.UndoToLastCheckpoint()
	expectedOps = []Op{
		DeleteOp(1, "bc"),
		DeleteOp(0, "a"),
	}
	assert.Equal(t, expectedOps, ops)

	ops = log.UndoToLastCheckpoint()
	assert.Equal(t, 0, len(ops))
}

func TestRedoToNextCheckpoint(t *testing.T) {
	log := NewLog()
	log.TrackOp(InsertOp(0, "a"))
	log.TrackOp(InsertOp(1, "bc"))
	log.Checkpoint()
	log.TrackOp(InsertOp(2, "de"))
	log.TrackOp(InsertOp(3, "fg"))
	log.Checkpoint()
	log.TrackOp(InsertOp(4, "h"))
	log.Checkpoint()

	ops := log.RedoToNextCheckpoint()
	assert.Equal(t, 0, len(ops))

	log.UndoToLastCheckpoint()
	log.UndoToLastCheckpoint()

	ops = log.RedoToNextCheckpoint()
	expectedOps := []Op{
		InsertOp(2, "de"),
		InsertOp(3, "fg"),
	}
	assert.Equal(t, expectedOps, ops)

	ops = log.RedoToNextCheckpoint()
	expectedOps = []Op{
		InsertOp(4, "h"),
	}
	assert.Equal(t, expectedOps, ops)

	ops = log.RedoToNextCheckpoint()
	assert.Equal(t, 0, len(ops))

	log.UndoToLastCheckpoint()
	log.UndoToLastCheckpoint()
	log.TrackOp(DeleteOp(5, "x"))

	ops = log.RedoToNextCheckpoint()
	assert.Equal(t, 0, len(ops))
}

func TestHasUnsavedChanges(t *testing.T) {
	log := NewLog()
	assert.False(t, log.HasUnsavedChanges())

	log.TrackOp(InsertOp(0, "a"))
	assert.True(t, log.HasUnsavedChanges())

	log.TrackSave()
	assert.False(t, log.HasUnsavedChanges())

	log.UndoToLastCheckpoint()
	assert.True(t, log.HasUnsavedChanges())

	log.RedoToNextCheckpoint()
	assert.False(t, log.HasUnsavedChanges())

	log.UndoToLastCheckpoint()
	log.TrackOp(DeleteOp(1, "b"))
	assert.True(t, log.HasUnsavedChanges())

	log.UndoToLastCheckpoint()
	assert.True(t, log.HasUnsavedChanges())

	log.TrackSave()
	assert.False(t, log.HasUnsavedChanges())
}

func TestTrackLoad(t *testing.T) {
	log := NewLog()
	log.TrackOp(InsertOp(0, "a"))
	log.TrackOp(InsertOp(1, "b"))
	log.Checkpoint()
	log.TrackSave()

	log.TrackOp(InsertOp(2, "c"))
	assert.True(t, log.HasUnsavedChanges())

	log.TrackLoad()
	assert.False(t, log.HasUnsavedChanges())
	assert.Equal(t, 0, len(log.UndoToLastCheckpoint()))
	assert.Equal(t, 0, len(log.RedoToNextCheckpoint()))
}
