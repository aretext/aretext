package parser

// Edit represents a change to a document.
type Edit struct {
	pos         uint64 // Position of the first character inserted/deleted.
	numInserted uint64
	numDeleted  uint64
}

func NewInsertEdit(pos, numInserted uint64) Edit {
	return Edit{pos: pos, numInserted: numInserted}
}

func NewDeleteEdit(pos, numDeleted uint64) Edit {
	return Edit{pos: pos, numDeleted: numDeleted}
}
