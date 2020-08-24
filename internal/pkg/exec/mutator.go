package exec

import (
	"fmt"
	"strings"

	"github.com/wedaly/aretext/internal/pkg/text"
)

// Mutator modifies the state of the cursor or text.
type Mutator interface {
	fmt.Stringer
	Mutate(state *EditorState)
}

// CompositeMutator executes a series of mutations.
type CompositeMutator struct {
	subMutators []Mutator
}

func NewCompositeMutator(subMutators []Mutator) Mutator {
	return &CompositeMutator{subMutators}
}

// Mutate executes a series of mutations in order.
func (cm *CompositeMutator) Mutate(state *EditorState) {
	for _, mut := range cm.subMutators {
		mut.Mutate(state)
	}
}

func (cm *CompositeMutator) String() string {
	args := make([]string, 0, len(cm.subMutators))
	for _, mut := range cm.subMutators {
		args = append(args, mut.String())
	}
	return fmt.Sprintf("Composite(%s)", strings.Join(args, ","))
}

type cursorMutator struct {
	loc CursorLocator
}

// NewCursorMutator returns a mutator that updates the cursor location.
func NewCursorMutator(loc CursorLocator) Mutator {
	return &cursorMutator{loc}
}

func (cpm *cursorMutator) Mutate(state *EditorState) {
	bufferState := state.FocusedBuffer()
	bufferState.cursor = cpm.loc.Locate(bufferState)
}

func (cpm *cursorMutator) String() string {
	return fmt.Sprintf("MutateCursor(%s)", cpm.loc)
}

type scrollToCursorMutator struct{}

func NewScrollToCursorMutator() Mutator {
	return &scrollToCursorMutator{}
}

// Mutate updates the view origin so that the cursor is visible.
func (sm *scrollToCursorMutator) Mutate(state *EditorState) {
	bufferState := state.FocusedBuffer()
	bufferState.view.origin = ScrollToCursor(
		bufferState.cursor.position,
		bufferState.tree,
		bufferState.view.origin,
		bufferState.view.width,
		bufferState.view.height)
}

func (sm *scrollToCursorMutator) String() string {
	return "ScrollToCursor()"
}

type scrollLinesMutator struct {
	direction text.ReadDirection
	numLines  uint64
}

func NewScrollLinesMutator(direction text.ReadDirection, numLines uint64) Mutator {
	return &scrollLinesMutator{direction, numLines}
}

// Mutate moves the view origin up/down by the specified number of lines.
func (sm *scrollLinesMutator) Mutate(state *EditorState) {
	bufferState := state.FocusedBuffer()
	lineNum := bufferState.tree.LineNumForPosition(bufferState.view.origin)
	if sm.direction == text.ReadDirectionForward {
		lineNum += sm.numLines
	} else if lineNum >= sm.numLines {
		lineNum -= sm.numLines
	} else {
		lineNum = 0
	}

	lineNum = closestValidLineNum(bufferState.tree, lineNum)

	// When scrolling to the end of the file, we want most of the last lines to remain visible.
	// To achieve this, set the view origin (viewHeight - scrollMargin) lines above
	// the last line.  This will leave a few blank lines past the end of the document
	// (the scroll margin) for consistency with ScrollToCursor.
	lastLineNum := closestValidLineNum(bufferState.tree, bufferState.tree.NumLines())
	if lastLineNum-lineNum < bufferState.view.height {
		if lastLineNum+scrollMargin+1 > bufferState.view.height {
			lineNum = lastLineNum + scrollMargin + 1 - bufferState.view.height
		} else {
			lineNum = 0
		}
	}

	bufferState.view.origin = bufferState.tree.LineStartPosition(lineNum)
}

func (sm *scrollLinesMutator) String() string {
	return fmt.Sprintf("ScrollLines(%s, %d)", directionString(sm.direction), sm.numLines)
}

type insertRuneMutator struct {
	r rune
}

func NewInsertRuneMutator(r rune) Mutator {
	return &insertRuneMutator{r}
}

// Mutate inserts a rune at the current cursor location.
func (irm *insertRuneMutator) Mutate(state *EditorState) {
	bufferState := state.FocusedBuffer()
	startPos := bufferState.cursor.position
	if err := bufferState.tree.InsertAtPosition(startPos, irm.r); err != nil {
		// Invalid UTF-8 character; ignore it.
		return
	}

	bufferState.cursor.position = startPos + 1
}

func (irm *insertRuneMutator) String() string {
	return fmt.Sprintf("InsertRune(%q)", irm.r)
}

type deleteMutator struct {
	loc CursorLocator
}

func NewDeleteMutator(loc CursorLocator) Mutator {
	return &deleteMutator{loc}
}

// Mutate deletes characters from the cursor position up to (but not including) the position returned by the locator.
// It can delete either forwards or backwards from the cursor.
// The cursor position will be set to the start of the deleted region,
// which could be on a newline character or past the end of the text.
func (dm *deleteMutator) Mutate(state *EditorState) {
	bufferState := state.FocusedBuffer()
	startPos := bufferState.cursor.position
	deleteToPos := dm.loc.Locate(bufferState).position

	if startPos < deleteToPos {
		dm.deleteCharacters(bufferState.tree, startPos, deleteToPos-startPos)
		bufferState.cursor = cursorState{position: startPos}
	} else if startPos > deleteToPos {
		dm.deleteCharacters(bufferState.tree, deleteToPos, startPos-deleteToPos)
		bufferState.cursor = cursorState{position: deleteToPos}
	}
}

func (dm *deleteMutator) deleteCharacters(tree *text.Tree, pos uint64, count uint64) {
	for i := uint64(0); i < count; i++ {
		tree.DeleteAtPosition(pos)
	}
}

func (dm *deleteMutator) String() string {
	return fmt.Sprintf("Delete(%s)", dm.loc)
}
