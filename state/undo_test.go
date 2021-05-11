package state

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/syntax"
)

func TestUndoAndRedo(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)

	// Make some edits with undo checkpoints.
	InsertRune(state, 'a')
	InsertRune(state, 'b')
	InsertRune(state, 'c')
	CheckpointUndoLog(state)
	InsertNewline(state)
	InsertRune(state, 'd')
	CheckpointUndoLog(state)
	DeleteRunes(state, func(params LocatorParams) uint64 {
		return locate.PrevCharInLine(params.TextTree, 1, false, params.CursorPos)
	})
	CheckpointUndoLog(state)

	// Verify state before undo.
	assert.Equal(t, uint64(4), state.documentBuffer.cursor.position)
	assert.Equal(t, "abc\n", state.documentBuffer.textTree.String())

	// Undo the deletion.
	Undo(state)
	assert.Equal(t, uint64(4), state.documentBuffer.cursor.position)
	assert.Equal(t, "abc\nd", state.documentBuffer.textTree.String())

	// Undo the newline and insertion of 'd'
	Undo(state)
	assert.Equal(t, uint64(2), state.documentBuffer.cursor.position)
	assert.Equal(t, "abc", state.documentBuffer.textTree.String())

	// Redo the newline and insertion of 'd'
	Redo(state)
	assert.Equal(t, uint64(2), state.documentBuffer.cursor.position)
	assert.Equal(t, "abc\nd", state.documentBuffer.textTree.String())

	// Undo again.
	Undo(state)
	assert.Equal(t, uint64(2), state.documentBuffer.cursor.position)
	assert.Equal(t, "abc", state.documentBuffer.textTree.String())

	// Undo last change.
	Undo(state)
	assert.Equal(t, uint64(0), state.documentBuffer.cursor.position)
	assert.Equal(t, "", state.documentBuffer.textTree.String())
}

func TestUndoDeleteLinesWithIndentation(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)

	// Insert some lines with indentation.
	InsertTab(state)
	InsertRune(state, 'a')
	InsertRune(state, 'b')
	InsertNewline(state)
	InsertTab(state)
	InsertRune(state, 'c')
	InsertNewline(state)
	InsertRune(state, 'd')
	CheckpointUndoLog(state)

	// Delete second-to-last line, which is indented.
	MoveCursor(state, func(p LocatorParams) uint64 { return locate.StartOfLineNum(p.TextTree, 1) })
	DeleteLines(state, func(p LocatorParams) uint64 { return p.CursorPos }, false, false)
	CheckpointUndoLog(state)

	// Verify state before undo.
	assert.Equal(t, uint64(4), state.documentBuffer.cursor.position)
	assert.Equal(t, "\tab\nd", state.documentBuffer.textTree.String())

	// Undo the deletion.
	// Expect the cursor to land on the start of the restored line AFTER indentation.
	Undo(state)
	assert.Equal(t, uint64(5), state.documentBuffer.cursor.position)
	assert.Equal(t, "\tab\n\tc\nd", state.documentBuffer.textTree.String())
}

func TestUndoMultiByteUnicodeWithSyntaxHighlighting(t *testing.T) {
	state := NewEditorState(100, 100, nil, nil)
	SetSyntax(state, syntax.LanguageGo)

	// Insert multi-byte UTF-8 runes.
	for _, r := range "丂丄丅丆丏 ¢ह€한" {
		InsertRune(state, r)
	}
	CheckpointUndoLog(state)

	// This used to trigger a panic when retokenizing because the
	// deleted rune count was incorrect.
	Undo(state)
	assert.Equal(t, "", state.documentBuffer.textTree.String())

	Redo(state)
	assert.Equal(t, "丂丄丅丆丏 ¢ह€한", state.documentBuffer.textTree.String())
}
