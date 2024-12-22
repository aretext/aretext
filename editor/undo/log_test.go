package undo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUndoToLastCommitted(t *testing.T) {
	log := NewLog()
	hasEntry, ops, cursor := log.UndoToLastCommitted()
	assert.False(t, hasEntry)
	assert.Equal(t, 0, len(ops))
	assert.Equal(t, uint64(0), cursor)

	log.BeginEntry(0)
	log.TrackOp(InsertOp(0, "a"))
	log.TrackOp(InsertOp(1, "bc"))
	log.CommitEntry(1)
	hasEntry, ops, cursor = log.UndoToLastCommitted()
	assert.True(t, hasEntry)
	expectedOps := []Op{
		DeleteOp(1, "bc"),
		DeleteOp(0, "a"),
	}
	assert.Equal(t, expectedOps, ops)
	assert.Equal(t, uint64(0), cursor)

	log.BeginEntry(2)
	log.TrackOp(DeleteOp(3, "x"))
	log.TrackOp(InsertOp(4, "yz"))
	log.CommitEntry(3)
	log.BeginEntry(4)
	log.TrackOp(DeleteOp(5, "12"))
	log.TrackOp(InsertOp(6, "345"))
	log.CommitEntry(5)

	hasEntry, ops, cursor = log.UndoToLastCommitted()
	assert.True(t, hasEntry)
	expectedOps = []Op{
		DeleteOp(6, "345"),
		InsertOp(5, "12"),
	}
	assert.Equal(t, expectedOps, ops)
	assert.Equal(t, uint64(4), cursor)

	hasEntry, ops, cursor = log.UndoToLastCommitted()
	assert.True(t, hasEntry)
	expectedOps = []Op{
		DeleteOp(4, "yz"),
		InsertOp(3, "x"),
	}
	assert.Equal(t, expectedOps, ops)
	assert.Equal(t, uint64(2), cursor)

	hasEntry, ops, cursor = log.UndoToLastCommitted()
	assert.False(t, hasEntry)
	assert.Equal(t, 0, len(ops))
	assert.Equal(t, uint64(0), cursor)
}

func TestRedoToNextCommitted(t *testing.T) {
	log := NewLog()
	log.BeginEntry(0)
	log.TrackOp(InsertOp(0, "a"))
	log.TrackOp(InsertOp(1, "bc"))
	log.CommitEntry(1)
	log.BeginEntry(1)
	log.TrackOp(InsertOp(2, "de"))
	log.TrackOp(InsertOp(3, "fg"))
	log.CommitEntry(2)
	log.BeginEntry(2)
	log.TrackOp(InsertOp(4, "h"))
	log.CommitEntry(3)

	hasEntry, ops, cursor := log.RedoToNextCommitted()
	assert.False(t, hasEntry)
	assert.Equal(t, 0, len(ops))
	assert.Equal(t, uint64(0), cursor)

	log.UndoToLastCommitted()
	log.UndoToLastCommitted()

	hasEntry, ops, cursor = log.RedoToNextCommitted()
	assert.True(t, hasEntry)
	expectedOps := []Op{
		InsertOp(2, "de"),
		InsertOp(3, "fg"),
	}
	assert.Equal(t, expectedOps, ops)
	assert.Equal(t, uint64(2), cursor)

	hasEntry, ops, cursor = log.RedoToNextCommitted()
	assert.True(t, hasEntry)
	expectedOps = []Op{
		InsertOp(4, "h"),
	}
	assert.Equal(t, expectedOps, ops)
	assert.Equal(t, uint64(3), cursor)

	hasEntry, ops, cursor = log.RedoToNextCommitted()
	assert.False(t, hasEntry)
	assert.Equal(t, 0, len(ops))
	assert.Equal(t, uint64(0), cursor)

	log.UndoToLastCommitted()
	log.UndoToLastCommitted()
	log.BeginEntry(3)
	log.TrackOp(DeleteOp(5, "x"))
	log.CommitEntry(4)

	hasEntry, ops, cursor = log.RedoToNextCommitted()
	assert.False(t, hasEntry)
	assert.Equal(t, 0, len(ops))
	assert.Equal(t, uint64(0), cursor)
}

func TestCommitEntryWithNoOps(t *testing.T) {
	log := NewLog()

	log.BeginEntry(0)
	log.TrackOp(InsertOp(1, "a"))
	log.TrackOp(InsertOp(2, "b"))
	log.CommitEntry(2)

	log.BeginEntry(2)
	log.TrackOp(InsertOp(1, "c"))
	log.TrackOp(InsertOp(2, "d"))
	log.CommitEntry(3)

	log.BeginEntry(4)
	// No operations in this action.
	log.CommitEntry(5)

	hasEntry, ops, cursor := log.UndoToLastCommitted()
	assert.True(t, hasEntry)
	expectedOps := []Op{
		DeleteOp(2, "d"),
		DeleteOp(1, "c"),
	}
	assert.Equal(t, expectedOps, ops)
	assert.Equal(t, uint64(2), cursor)
}

func TestHasUnsavedChanges(t *testing.T) {
	log := NewLog()
	assert.False(t, log.HasUnsavedChanges())

	log.BeginEntry(0)
	log.TrackOp(InsertOp(0, "a"))
	log.CommitEntry(1)
	assert.True(t, log.HasUnsavedChanges())

	log.TrackSave()
	assert.False(t, log.HasUnsavedChanges())

	log.UndoToLastCommitted()
	assert.True(t, log.HasUnsavedChanges())

	log.RedoToNextCommitted()
	assert.False(t, log.HasUnsavedChanges())

	log.UndoToLastCommitted()
	log.BeginEntry(1)
	log.TrackOp(DeleteOp(1, "b"))
	log.CommitEntry(2)
	assert.True(t, log.HasUnsavedChanges())

	log.UndoToLastCommitted()
	assert.True(t, log.HasUnsavedChanges())

	log.TrackSave()
	assert.False(t, log.HasUnsavedChanges())
}
