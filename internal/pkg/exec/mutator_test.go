package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wedaly/aretext/internal/pkg/file"
	"github.com/wedaly/aretext/internal/pkg/text"
)

func TestLoadDocumentMutator(t *testing.T) {
	state := NewEditorState(100, 100)

	// Load a new document.
	textTree, err := text.NewTreeFromString("abcd")
	require.NoError(t, err)
	watcher := file.NewEmptyWatcher()
	NewLoadDocumentMutator(textTree, watcher).Mutate(state)

	// Expect that the text and watcher are installed.
	assert.Equal(t, "abcd", state.documentBuffer.textTree.String())
	assert.Equal(t, watcher, state.FileWatcher())
}

func TestLoadDocumentMutatorMoveCursorOntoDocument(t *testing.T) {
	textTree, err := text.NewTreeFromString("abcd\nefghi\njklmnop\nqrst")
	require.NoError(t, err)
	state := NewEditorState(5, 2)
	state.documentBuffer.textTree = textTree
	state.documentBuffer.cursor.position = 22

	// Scroll to cursor at end of document.
	NewScrollToCursorMutator().Mutate(state)
	assert.Equal(t, uint64(16), state.documentBuffer.view.textOrigin)

	// Load a new document with a shorter text.
	textTree, err = text.NewTreeFromString("ab")
	require.NoError(t, err)
	watcher := file.NewEmptyWatcher()
	NewLoadDocumentMutator(textTree, watcher).Mutate(state)

	// Expect that the cursor moved back to the end of the text,
	// and the view scrolled to make the cursor visible.
	assert.Equal(t, "ab", state.documentBuffer.textTree.String())
	assert.Equal(t, uint64(2), state.documentBuffer.cursor.position)
	assert.Equal(t, uint64(0), state.documentBuffer.view.textOrigin)
}

func TestLoadDocumentMutatorWithReplOpen(t *testing.T) {
	state := NewEditorState(100, 100)

	// Insert text in both buffers, focus on REPL buffer.
	setupMutator := NewCompositeMutator([]Mutator{
		NewInsertRuneMutator('x'),
		NewInsertRuneMutator('y'),
		NewInsertRuneMutator('z'),
		NewOutputReplMutator("hello\n>>> "),
		NewLayoutMutator(LayoutDocumentAndRepl),
		NewFocusBufferMutator(BufferIdRepl),
		NewInsertRuneMutator('a'),
		NewInsertRuneMutator('b'),
		NewInsertRuneMutator('c'),
	})
	setupMutator.Mutate(state)

	// Load a new document.
	textTree, err := text.NewTreeFromString("updated")
	require.NoError(t, err)
	watcher := file.NewEmptyWatcher()
	NewLoadDocumentMutator(textTree, watcher).Mutate(state)

	// Expect that only the document buffer changed.
	assert.Equal(t, "updated", state.documentBuffer.textTree.String())
	assert.Equal(t, "hello\n>>> abc", state.replBuffer.textTree.String())
}

func TestCursorMutator(t *testing.T) {
	textTree, err := text.NewTreeFromString("abcd")
	require.NoError(t, err)
	state := NewEditorState(100, 100)
	state.documentBuffer = &BufferState{
		textTree: textTree,
		cursor:   cursorState{position: 2},
	}
	mutator := NewCursorMutator(NewCharInLineLocator(text.ReadDirectionForward, 1, false))
	mutator.Mutate(state)
	assert.Equal(t, uint64(3), state.documentBuffer.cursor.position)
}

func TestCursorMutatorRestrictToReplInput(t *testing.T) {
	state := NewEditorState(100, 100)
	textTree, err := text.NewTreeFromString(">>> abcd")
	require.NoError(t, err)
	state.replBuffer.textTree = textTree

	setupMutator := NewCompositeMutator([]Mutator{
		NewLayoutMutator(LayoutDocumentAndRepl),
		NewFocusBufferMutator(BufferIdRepl),
	})
	setupMutator.Mutate(state)
	state.replBuffer.cursor = cursorState{position: 6}
	state.SetReplInputStartPos(4)

	startOfLineLoc := NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	mutator := NewCursorMutator(startOfLineLoc)
	mutator.RestrictToReplInput()
	mutator.Mutate(state)
	assert.Equal(t, uint64(4), state.replBuffer.cursor.position)
}

func TestInsertRuneMutator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		insertRune     rune
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "insert into empty string",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			insertRune:     'x',
			expectedCursor: cursorState{position: 1},
			expectedText:   "x",
		},
		{
			name:           "insert in middle of string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 1},
			insertRune:     'x',
			expectedCursor: cursorState{position: 2},
			expectedText:   "axbcd",
		},
		{
			name:           "insert at end of string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 4},
			insertRune:     'x',
			expectedCursor: cursorState{position: 5},
			expectedText:   "abcdx",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100)
			state.documentBuffer = &BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			mutator := NewInsertRuneMutator(tc.insertRune)
			mutator.Mutate(state)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestInsertRuneMutatorRestrictToReplInput(t *testing.T) {
	state := NewEditorState(100, 100)
	textTree, err := text.NewTreeFromString(">>> abcd")
	require.NoError(t, err)
	state.replBuffer.textTree = textTree

	setupMutator := NewCompositeMutator([]Mutator{
		NewLayoutMutator(LayoutDocumentAndRepl),
		NewFocusBufferMutator(BufferIdRepl),
	})
	setupMutator.Mutate(state)
	state.replBuffer.cursor = cursorState{position: 2}
	state.SetReplInputStartPos(4)

	mutator := NewInsertRuneMutator('x')
	mutator.RestrictToReplInput()
	mutator.Mutate(state)
	assert.Equal(t, uint64(5), state.replBuffer.cursor.position)
	assert.Equal(t, ">>> xabcd", textTree.String())
}

func TestDeleteMutator(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		initialCursor  cursorState
		locator        CursorLocator
		expectedCursor cursorState
		expectedText   string
	}{
		{
			name:           "delete from empty string",
			inputString:    "",
			initialCursor:  cursorState{position: 0},
			locator:        NewCharInLineLocator(text.ReadDirectionForward, 1, true),
			expectedCursor: cursorState{position: 0},
			expectedText:   "",
		},
		{
			name:           "delete next character at start of string",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 0},
			locator:        NewCharInLineLocator(text.ReadDirectionForward, 1, true),
			expectedCursor: cursorState{position: 0},
			expectedText:   "bcd",
		},
		{
			name:           "delete from end of text",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 3},
			locator:        NewCharInLineLocator(text.ReadDirectionForward, 1, true),
			expectedCursor: cursorState{position: 3},
			expectedText:   "abc",
		},
		{
			name:           "delete multiple characters",
			inputString:    "abcd",
			initialCursor:  cursorState{position: 1},
			locator:        NewCharInLineLocator(text.ReadDirectionForward, 10, true),
			expectedCursor: cursorState{position: 1},
			expectedText:   "a",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100)
			state.documentBuffer = &BufferState{
				textTree: textTree,
				cursor:   tc.initialCursor,
			}
			mutator := NewDeleteMutator(tc.locator)
			mutator.Mutate(state)
			assert.Equal(t, tc.expectedCursor, state.documentBuffer.cursor)
			assert.Equal(t, tc.expectedText, textTree.String())
		})
	}
}

func TestDeleteMutatorRestrictToReplInput(t *testing.T) {
	state := NewEditorState(100, 100)
	textTree, err := text.NewTreeFromString(">>> abcd")
	require.NoError(t, err)
	state.replBuffer.textTree = textTree

	setupMutator := NewCompositeMutator([]Mutator{
		NewLayoutMutator(LayoutDocumentAndRepl),
		NewFocusBufferMutator(BufferIdRepl),
	})
	setupMutator.Mutate(state)
	state.replBuffer.cursor = cursorState{position: 6}
	state.SetReplInputStartPos(4)

	startOfLineLoc := NewLineBoundaryLocator(text.ReadDirectionBackward, false)
	mutator := NewDeleteMutator(startOfLineLoc)
	mutator.RestrictToReplInput()
	mutator.Mutate(state)
	assert.Equal(t, uint64(4), state.replBuffer.cursor.position)
	assert.Equal(t, ">>> cd", textTree.String())
}

func TestScrollLinesMutator(t *testing.T) {
	testCases := []struct {
		name               string
		inputString        string
		initialView        viewState
		direction          text.ReadDirection
		numLines           uint64
		expectedtextOrigin uint64
	}{
		{
			name:               "empty, scroll up",
			inputString:        "",
			initialView:        viewState{textOrigin: 0, height: 100, width: 100},
			direction:          text.ReadDirectionBackward,
			numLines:           1,
			expectedtextOrigin: 0,
		},
		{
			name:               "empty, scroll down",
			inputString:        "",
			initialView:        viewState{textOrigin: 0, height: 100, width: 100},
			direction:          text.ReadDirectionForward,
			numLines:           1,
			expectedtextOrigin: 0,
		},
		{
			name:               "scroll up",
			inputString:        "ab\ncd\nef\ngh\nij\nkl\nmn",
			initialView:        viewState{textOrigin: 12, height: 2, width: 100},
			direction:          text.ReadDirectionBackward,
			numLines:           3,
			expectedtextOrigin: 3,
		},
		{
			name:               "scroll down",
			inputString:        "ab\ncd\nef\ngh\nij\nkl\nmn",
			initialView:        viewState{textOrigin: 3, height: 2, width: 100},
			direction:          text.ReadDirectionForward,
			numLines:           3,
			expectedtextOrigin: 12,
		},
		{
			name:               "scroll down to last line",
			inputString:        "ab\ncd\nef\ngh\nij\nkl\nmn",
			initialView:        viewState{textOrigin: 0, height: 6, width: 100},
			numLines:           10,
			expectedtextOrigin: 12,
		},
		{
			name:               "scroll down view taller than document",
			inputString:        "ab\ncd\nef\ngh",
			initialView:        viewState{textOrigin: 0, height: 100, width: 100},
			numLines:           1,
			expectedtextOrigin: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			state := NewEditorState(100, 100)
			state.documentBuffer = &BufferState{
				textTree: textTree,
				view:     tc.initialView,
			}
			mutator := NewScrollLinesMutator(tc.direction, tc.numLines)
			mutator.Mutate(state)
			assert.Equal(t, tc.expectedtextOrigin, state.documentBuffer.view.textOrigin)
		})
	}
}

type stubRepl struct {
	inputs      []string
	interrupted bool
}

func newStubRepl() *stubRepl {
	return &stubRepl{inputs: make([]string, 0, 1)}
}

func (r *stubRepl) Start() error {
	return nil
}

func (r *stubRepl) Terminate() error {
	return nil
}

func (r *stubRepl) Interrupt() error {
	r.interrupted = true
	return nil
}

func (r *stubRepl) SubmitInput(s string) error {
	r.inputs = append(r.inputs, s)
	return nil
}

func (r *stubRepl) OutputChan() chan string {
	return nil
}

func TestSubmitReplMutator(t *testing.T) {
	state := NewEditorState(100, 100)
	textTree, err := text.NewTreeFromString(">>> abcd")
	require.NoError(t, err)
	state.replBuffer.textTree = textTree

	NewLayoutMutator(LayoutDocumentAndRepl).Mutate(state)
	state.replBuffer.cursor = cursorState{position: 6}
	state.SetReplInputStartPos(4)

	repl := newStubRepl()
	mutator := NewSubmitReplMutator(repl)
	mutator.Mutate(state)
	assert.Equal(t, uint64(9), state.replBuffer.cursor.position)
	assert.Equal(t, uint64(9), state.replInputStartPos)
	assert.Equal(t, ">>> abcd\n", textTree.String())
	assert.Equal(t, []string{"abcd"}, repl.inputs)
}

func TestInterruptReplMutator(t *testing.T) {
	textTree, err := text.NewTreeFromString("")
	require.NoError(t, err)

	state := NewEditorState(100, 100)
	state.documentBuffer = &BufferState{
		textTree: textTree,
		view:     viewState{height: 100, width: 100},
	}

	setupMutator := NewCompositeMutator([]Mutator{
		NewOutputReplMutator("hello\n>>> "),
		NewLayoutMutator(LayoutDocumentAndRepl),
		NewFocusBufferMutator(BufferIdRepl),
		NewInsertRuneMutator('a'),
		NewInsertRuneMutator('b'),
		NewInsertRuneMutator('c'),
	})
	setupMutator.Mutate(state)

	replBuffer := state.Buffer(BufferIdRepl)
	assert.Equal(t, "hello\n>>> abc", replBuffer.TextTree().String())
	assert.Equal(t, uint64(10), state.ReplInputStartPos())
	assert.Equal(t, uint64(13), replBuffer.CursorPosition())

	repl := newStubRepl()
	NewInterruptReplMutator(repl).Mutate(state)
	assert.Equal(t, "hello\n>>> ", replBuffer.TextTree().String())
	assert.Equal(t, uint64(10), state.ReplInputStartPos())
	assert.Equal(t, uint64(10), replBuffer.CursorPosition())
	assert.True(t, repl.interrupted)
}
