package undo

// Op represents an insert or delete operation on a document.
type Op struct {
	pos        uint64
	insertText string
	deleteText string
}

// InsertOp constructs a new operation to insert text at a position.
func InsertOp(pos uint64, text string) Op {
	return Op{
		pos:        pos,
		insertText: text,
	}
}

// DeleteOp constructs a new operation to delete text at a position.
func DeleteOp(pos uint64, text string) Op {
	return Op{
		pos:        pos,
		deleteText: text,
	}
}

// Inverse returns an op that reverses the effect of the op.
func (op Op) Inverse() Op {
	return Op{
		pos:        op.pos,
		insertText: op.deleteText,
		deleteText: op.insertText,
	}
}

// Position returns the position at which the operation occurred.
func (op Op) Position() uint64 {
	return op.pos
}

// TextToInsert returns the text inserted by the op.
// This will be an empty string if NumRunesToDelete is greater than zero.
func (op Op) TextToInsert() string {
	return op.insertText
}

// NumRunesToDelete returns the number of runes deleted at the position.
// This will be zero if TextToInsert is a non-empty string.
func (op Op) NumRunesToDelete() int {
	return len(op.deleteText)
}
