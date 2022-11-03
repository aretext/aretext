package undo

// logEntry represents an entry in the undo log.
// Some entries are "checkpoints" that partition entries into groups that can be undone/redone together.
type logEntry struct {
	checkpoint bool
	op         Op
}

// Log tracks changes to a document and generates undo/redo operations.
type Log struct {
	// entries[0:numUndoEntries] are changes made before the current document state.  These changes can be undone.
	// entries[numUndoEntries:len(entries)-1] are changes made after the current document state.  These changes can be redone.
	entries              []logEntry
	numUndoEntries       int
	numEntriesAtLastSave int
}

// NewLog constructs a new, empty undo log.
func NewLog() *Log {
	return &Log{
		entries:              make([]logEntry, 0, 256),
		numUndoEntries:       0,
		numEntriesAtLastSave: 0,
	}
}

// TrackOp tracks a change to the document.
// This reverts any changes in the redo log, then appends the new, uncommitted change.
func (l *Log) TrackOp(op Op) {
	// Revert all changes from the redo log.
	// This differs from vim, which discards all changes in the redo log.
	// Instead, we're following the approach from
	// "Resolving the Great Undo-Redo Quandary"
	// (https://github.com/zaboople/klonk/blob/master/TheGURQ.md)
	// to allow restoration of changes in the redo log.
	for i := len(l.entries) - 1; i >= l.numUndoEntries; i-- {
		revertOp := l.entries[i].op.Inverse()
		l.entries = append(l.entries, logEntry{op: revertOp})
	}
	l.numUndoEntries = len(l.entries)

	// Append a new undo entry.
	l.entries = append(l.entries, logEntry{op: op})
	l.numUndoEntries++
}

// TrackLoad removes all changes and resets the savepoint.
func (l *Log) TrackLoad() {
	l.entries = l.entries[:0]
	l.numUndoEntries = 0
	l.numEntriesAtLastSave = 0
}

// TrackSave moves the savepoint to the current entry.
func (l *Log) TrackSave() {
	l.numEntriesAtLastSave = l.numUndoEntries
}

// Checkpoint marks the current entry as a checkpoint.
func (l *Log) Checkpoint() {
	if l.numUndoEntries == 0 {
		return
	}
	l.entries[l.numUndoEntries-1].checkpoint = true
}

// UndoToLastCheckpoint returns operations to transform the document back to its state at the previous checkpoint.
// It also moves the current position backwards in the log.
func (l *Log) UndoToLastCheckpoint() []Op {
	var ops []Op
	for i := l.numUndoEntries - 1; i >= 0; i-- {
		if i < l.numUndoEntries-1 && l.entries[i].checkpoint {
			break
		}
		ops = append(ops, l.entries[i].op.Inverse())
	}
	l.numUndoEntries -= len(ops)
	return ops
}

// RedoToNextCheckpoint returns operations to to transform the document to its state at the next checkpoint.
// It also moves the current position forward in the log.
func (l *Log) RedoToNextCheckpoint() []Op {
	var ops []Op
	for i := l.numUndoEntries; i < len(l.entries); i++ {
		ops = append(ops, l.entries[i].op)
		if i > l.numUndoEntries && l.entries[i].checkpoint {
			break
		}
	}
	l.numUndoEntries += len(ops)
	return ops
}

// HasUnsavedChanges returns whether the log has unsaved changes.
func (l *Log) HasUnsavedChanges() bool {
	return l.numUndoEntries != l.numEntriesAtLastSave
}
