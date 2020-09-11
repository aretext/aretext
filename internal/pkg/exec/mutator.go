package exec

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/wedaly/aretext/internal/pkg/repl"
	"github.com/wedaly/aretext/internal/pkg/text"
)

// Mutator modifies the state of the editor.
type Mutator interface {
	fmt.Stringer

	// Mutate modifies the editor state.
	// All changes to editor state should be performed by mutators.
	Mutate(state *EditorState)

	// RestrictToReplInput prevents the mutator from navigating to or modifying text before the REPL input start position in the REPL buffer.
	// This is called by input modes that allow the user to edit the current REPL command but not prior commands or output from the REPL.
	RestrictToReplInput()
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

func (cm *CompositeMutator) RestrictToReplInput() {
	for _, mut := range cm.subMutators {
		mut.RestrictToReplInput()
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
	loc                 CursorLocator
	restrictToReplInput bool
}

// NewCursorMutator returns a mutator that updates the cursor location.
func NewCursorMutator(loc CursorLocator) Mutator {
	return &cursorMutator{loc, false}
}

func (cpm *cursorMutator) Mutate(state *EditorState) {
	bufferState := state.FocusedBuffer()
	bufferState.cursor = cpm.loc.Locate(bufferState)

	if cpm.restrictToReplInput && bufferState.cursor.position < state.replInputStartPos {
		bufferState.cursor = cursorState{position: state.replInputStartPos}
	}
}

func (cpm *cursorMutator) RestrictToReplInput() {
	cpm.restrictToReplInput = true
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
	bufferState.view.textOrigin = ScrollToCursor(
		bufferState.cursor.position,
		bufferState.tree,
		bufferState.view.textOrigin,
		bufferState.view.width,
		bufferState.view.height)
}

func (sm *scrollToCursorMutator) RestrictToReplInput() {}

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
	lineNum := bufferState.tree.LineNumForPosition(bufferState.view.textOrigin)
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

	bufferState.view.textOrigin = bufferState.tree.LineStartPosition(lineNum)
}

func (sm *scrollLinesMutator) RestrictToReplInput() {}

func (sm *scrollLinesMutator) String() string {
	return fmt.Sprintf("ScrollLines(%s, %d)", directionString(sm.direction), sm.numLines)
}

type insertRuneMutator struct {
	r                   rune
	restrictToReplInput bool
}

func NewInsertRuneMutator(r rune) Mutator {
	return &insertRuneMutator{r, false}
}

// Mutate inserts a rune at the current cursor location.
func (irm *insertRuneMutator) Mutate(state *EditorState) {
	bufferState := state.FocusedBuffer()
	startPos := bufferState.cursor.position

	if irm.restrictToReplInput && startPos < state.replInputStartPos {
		startPos = state.replInputStartPos
	}

	if err := bufferState.tree.InsertAtPosition(startPos, irm.r); err != nil {
		// Invalid UTF-8 character; ignore it.
		return
	}

	bufferState.cursor.position = startPos + 1
}

func (irm *insertRuneMutator) RestrictToReplInput() {
	irm.restrictToReplInput = true
}

func (irm *insertRuneMutator) String() string {
	return fmt.Sprintf("InsertRune(%q)", irm.r)
}

type deleteMutator struct {
	loc                 CursorLocator
	restrictToReplInput bool
}

func NewDeleteMutator(loc CursorLocator) Mutator {
	return &deleteMutator{loc, false}
}

// Mutate deletes characters from the cursor position up to (but not including) the position returned by the locator.
// It can delete either forwards or backwards from the cursor.
// The cursor position will be set to the start of the deleted region,
// which could be on a newline character or past the end of the text.
func (dm *deleteMutator) Mutate(state *EditorState) {
	bufferState := state.FocusedBuffer()
	startPos := bufferState.cursor.position
	deleteToPos := dm.loc.Locate(bufferState).position

	if dm.restrictToReplInput {
		if startPos < state.replInputStartPos {
			startPos = state.replInputStartPos
		}
		if deleteToPos < state.replInputStartPos {
			deleteToPos = state.replInputStartPos
		}
	}

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

func (dm *deleteMutator) RestrictToReplInput() {
	dm.restrictToReplInput = true
}

func (dm *deleteMutator) String() string {
	return fmt.Sprintf("Delete(%s)", dm.loc)
}

type resizeMutator struct {
	width, height uint64
}

func NewResizeMutator(width, height uint64) Mutator {
	return &resizeMutator{width, height}
}

// Mutate resizes the view to the specified width and height.
func (rm *resizeMutator) Mutate(state *EditorState) {
	state.SetScreenSize(rm.width, rm.height)

	// The layout mutator will update the dimensions of each buffer according to the current layout.
	NewLayoutMutator(state.layout).Mutate(state)
}

func (rm *resizeMutator) RestrictToReplInput() {}

func (rm *resizeMutator) String() string {
	return fmt.Sprintf("Resize(%d,%d)", rm.width, rm.height)
}

type layoutMutator struct {
	layout Layout
}

func NewLayoutMutator(layout Layout) Mutator {
	return &layoutMutator{layout}
}

func (lm *layoutMutator) Mutate(state *EditorState) {
	if lm.layout == LayoutDocumentOnly {
		lm.setLayoutDocumentOnly(state)
	} else if lm.layout == LayoutDocumentAndRepl {
		lm.setLayoutDocumentAndRepl(state)
	} else {
		log.Fatalf("Unrecognized layout: %d", lm.layout)
	}

	state.layout = lm.layout
}

func (lm *layoutMutator) setLayoutDocumentOnly(state *EditorState) {
	state.documentBuffer.focus = true
	state.replBuffer.focus = false

	state.documentBuffer.view.x = 0
	state.documentBuffer.view.y = 0
	state.documentBuffer.view.width = state.screenWidth
	state.documentBuffer.view.height = state.screenHeight

	state.replBuffer.view.x = 0
	state.replBuffer.view.y = 0
	state.replBuffer.view.width = 0
	state.replBuffer.view.height = 0
}

func (lm *layoutMutator) setLayoutDocumentAndRepl(state *EditorState) {
	state.documentBuffer.focus = false
	state.documentBuffer.view.x = 0
	state.documentBuffer.view.y = 0
	state.documentBuffer.view.width = state.screenWidth
	state.replBuffer.view.height = 0

	state.replBuffer.focus = true
	state.replBuffer.view.x = 0
	state.replBuffer.view.y = 0
	state.replBuffer.view.width = state.screenWidth
	state.replBuffer.view.height = 0

	if state.screenHeight > 2 {
		// Shrink the document to leave space for the REPL.
		state.documentBuffer.view.height = state.screenHeight / 2

		// Make the REPL visible, leaving one line at the top for a border.
		state.replBuffer.view.y = state.documentBuffer.view.height + 1
		state.replBuffer.view.width = state.screenWidth
		state.replBuffer.view.height = state.screenHeight/2 - 1
	} else if state.screenHeight > 0 {
		// Display only the REPL
		state.documentBuffer.view.height = 0
		state.replBuffer.view.height = state.screenHeight
	}
}

func (lm *layoutMutator) RestrictToReplInput() {}

func (lm *layoutMutator) String() string {
	var layout string
	if lm.layout == LayoutDocumentOnly {
		layout = "DocumentOnly"
	} else if lm.layout == LayoutDocumentAndRepl {
		layout = "DocumentAndRepl"
	} else {
		log.Fatalf("Unrecognized layout: %d", lm.layout)
	}
	return fmt.Sprintf("SetLayout(%s)", layout)
}

type outputReplMutator struct {
	s string
}

func NewOutputReplMutator(s string) Mutator {
	return &outputReplMutator{s}
}

// Mutate outputs a string to the REPL.
func (orm *outputReplMutator) Mutate(state *EditorState) {
	tree := state.replBuffer.tree
	for _, r := range orm.s {
		tree.InsertAtPosition(state.replInputStartPos, r)
		state.replInputStartPos++
		state.replBuffer.cursor.position++
	}
}

func (orm *outputReplMutator) RestrictToReplInput() {}

func (orm *outputReplMutator) String() string {
	return fmt.Sprintf("OutputReplMutator('%s')", orm.s)
}

type submitReplMutator struct {
	repl repl.Repl
}

func NewSubmitReplMutator(repl repl.Repl) Mutator {
	return &submitReplMutator{repl}
}

// Mutate executes the current repl input and updates the cursor.
func (srm *submitReplMutator) Mutate(state *EditorState) {
	tree := state.replBuffer.tree
	reader := tree.ReaderAtPosition(state.replInputStartPos, text.ReadDirectionForward)
	inputBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatalf("%s", err)
	}

	tree.InsertAtPosition(tree.NumChars(), '\n')
	srm.repl.SubmitInput(string(inputBytes))
	state.replInputStartPos = tree.NumChars()
	state.replBuffer.cursor = cursorState{position: state.replInputStartPos}
}

func (srm *submitReplMutator) RestrictToReplInput() {}

func (srm *submitReplMutator) String() string {
	return "SubmitRepl()"
}

type quitMutator struct{}

func NewQuitMutator() Mutator {
	return &quitMutator{}
}

func (qm *quitMutator) Mutate(state *EditorState) {
	state.quitFlag = true
}

func (qm *quitMutator) RestrictToReplInput() {}

func (qm *quitMutator) String() string {
	return "Quit()"
}
