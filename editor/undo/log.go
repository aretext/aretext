package undo

// LogEntry represents an entry in the undo log.
type LogEntry struct {
	Ops         []Op
	CursorBegin uint64
	CursorEnd   uint64
}

// Log tracks changes to a document and generates undo/redo operations.
type Log struct {
	stagedEntry          LogEntry
	committedEntries     []LogEntry
	numUndoEntries       int
	numEntriesAtLastSave int
}

// NewLog constructs a new, empty undo log.
func NewLog() *Log {
	return &Log{
		stagedEntry:          LogEntry{},
		committedEntries:     nil,
		numUndoEntries:       0,
		numEntriesAtLastSave: 0,
	}
}

// BeginEntry starts a new undo entry.
// This should be called before tracking any operations.
func (l *Log) BeginEntry(cursorPos uint64) {
	l.stagedEntry = LogEntry{CursorBegin: cursorPos}
}

// CommitEntry completes an undo entry.
// This should be called after BeginEntry.
// If no operations were tracked, this does nothing.
func (l *Log) CommitEntry(cursorPos uint64) {
	if len(l.stagedEntry.Ops) == 0 {
		return
	}

	if len(l.committedEntries) > l.numUndoEntries {
		// Invalidate future changes.
		l.committedEntries = l.committedEntries[0:l.numUndoEntries]
	}

	if l.numEntriesAtLastSave > l.numUndoEntries {
		// Invalidate a save point in the future.
		l.numEntriesAtLastSave = -1
	}

	l.stagedEntry.CursorEnd = cursorPos
	l.committedEntries = append(l.committedEntries, l.stagedEntry)
	l.stagedEntry = LogEntry{}
	l.numUndoEntries++
}

// TrackOp tracks a change to the document.
// This appends a new, uncommitted change and invalidates any future changes.
func (l *Log) TrackOp(op Op) {
	// Stage a new undo entry.
	l.stagedEntry.Ops = append(l.stagedEntry.Ops, op)
}

// TrackSave moves the savepoint to the current entry.
func (l *Log) TrackSave() {
	l.numEntriesAtLastSave = l.numUndoEntries
}

// UndoToLastCommitted returns operations to transform the document back to its state before the last entry.
// It also moves the current position backwards in the log.
func (l *Log) UndoToLastCommitted() (hasEntry bool, ops []Op, cursor uint64) {
	if l.numUndoEntries == 0 {
		return false, nil, 0
	}

	entry := l.committedEntries[l.numUndoEntries-1]
	ops = make([]Op, 0, len(entry.Ops))
	for i := len(entry.Ops) - 1; i >= 0; i-- {
		ops = append(ops, entry.Ops[i].Inverse())
	}

	l.numUndoEntries--

	return true, ops, entry.CursorBegin
}

// RedoToNextCommitted returns operations to to transform the document to its state after the next entry.
// It also moves the current position forward in the log.
func (l *Log) RedoToNextCommitted() (hasEntry bool, ops []Op, cursor uint64) {
	if l.numUndoEntries == len(l.committedEntries) {
		return false, nil, 0
	}

	entry := l.committedEntries[l.numUndoEntries]
	ops = make([]Op, 0, len(entry.Ops))
	ops = append(ops, entry.Ops...)

	l.numUndoEntries++

	return true, ops, entry.CursorEnd
}

// HasUnsavedChanges returns whether the log has unsaved changes.
func (l *Log) HasUnsavedChanges() bool {
	return l.numUndoEntries != l.numEntriesAtLastSave
}
