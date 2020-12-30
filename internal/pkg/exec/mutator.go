package exec

import (
	"fmt"
	"log"
	"strings"

	"github.com/wedaly/aretext/internal/pkg/file"
	"github.com/wedaly/aretext/internal/pkg/syntax"
	"github.com/wedaly/aretext/internal/pkg/syntax/parser"
	"github.com/wedaly/aretext/internal/pkg/text"
)

// Mutator modifies the state of the editor.
type Mutator interface {
	fmt.Stringer

	// Mutate modifies the editor state.
	// All changes to editor state should be performed by mutators.
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

type loadDocumentMutator struct {
	textTree    *text.Tree
	fileWatcher *file.Watcher
}

func NewLoadDocumentMutator(textTree *text.Tree, fileWatcher *file.Watcher) Mutator {
	return &loadDocumentMutator{textTree, fileWatcher}
}

// Mutate loads the document into the editor.
func (ldm *loadDocumentMutator) Mutate(state *EditorState) {
	state.documentBuffer.textTree = ldm.textTree

	state.fileWatcher.Stop()
	state.fileWatcher = ldm.fileWatcher

	// Make sure that the cursor is a valid position in the new document
	// and that the cursor is visible.  If not, adjust the cursor and scroll.
	NewCompositeMutator([]Mutator{
		NewSetSyntaxMutator(state.documentBuffer.syntaxLanguage),
		NewCursorMutator(NewOntoDocumentLocator()),
		NewScrollToCursorMutator(),
	}).Mutate(state)
}

func (ldm *loadDocumentMutator) String() string {
	return "LoadDocument()"
}

type cursorMutator struct {
	loc CursorLocator
}

// NewCursorMutator returns a mutator that updates the cursor location.
func NewCursorMutator(loc CursorLocator) Mutator {
	return &cursorMutator{loc}
}

func (cpm *cursorMutator) Mutate(state *EditorState) {
	bufferState := state.documentBuffer
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
	bufferState := state.documentBuffer
	bufferState.view.textOrigin = ScrollToCursor(
		bufferState.cursor.position,
		bufferState.textTree,
		bufferState.view.textOrigin,
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
	bufferState := state.documentBuffer
	lineNum := bufferState.textTree.LineNumForPosition(bufferState.view.textOrigin)
	if sm.direction == text.ReadDirectionForward {
		lineNum += sm.numLines
	} else if lineNum >= sm.numLines {
		lineNum -= sm.numLines
	} else {
		lineNum = 0
	}

	lineNum = closestValidLineNum(bufferState.textTree, lineNum)

	// When scrolling to the end of the file, we want most of the last lines to remain visible.
	// To achieve this, set the view origin (viewHeight - scrollMargin) lines above
	// the last line.  This will leave a few blank lines past the end of the document
	// (the scroll margin) for consistency with ScrollToCursor.
	lastLineNum := closestValidLineNum(bufferState.textTree, bufferState.textTree.NumLines())
	if lastLineNum-lineNum < bufferState.view.height {
		if lastLineNum+scrollMargin+1 > bufferState.view.height {
			lineNum = lastLineNum + scrollMargin + 1 - bufferState.view.height
		} else {
			lineNum = 0
		}
	}

	bufferState.view.textOrigin = bufferState.textTree.LineStartPosition(lineNum)
}

func (sm *scrollLinesMutator) String() string {
	return fmt.Sprintf("ScrollLines(%s, %d)", sm.direction, sm.numLines)
}

type insertRuneMutator struct {
	r rune
}

func NewInsertRuneMutator(r rune) Mutator {
	return &insertRuneMutator{r}
}

// Mutate inserts a rune at the current cursor location.
func (irm *insertRuneMutator) Mutate(state *EditorState) {
	bufferState := state.documentBuffer
	startPos := bufferState.cursor.position

	if err := bufferState.textTree.InsertAtPosition(startPos, irm.r); err != nil {
		// Invalid UTF-8 character; ignore it.
		return
	}

	edit := parser.Edit{Pos: startPos, NumInserted: 1}
	if err := bufferState.retokenizeAfterEdit(edit); err != nil {
		log.Printf("Error retokenizing document: %v\n", err)
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
	bufferState := state.documentBuffer
	startPos := bufferState.cursor.position
	deleteToPos := dm.loc.Locate(bufferState).position

	if startPos < deleteToPos {
		dm.deleteCharacters(bufferState, startPos, deleteToPos-startPos)
	} else if startPos > deleteToPos {
		dm.deleteCharacters(bufferState, deleteToPos, startPos-deleteToPos)
	}
}

func (dm *deleteMutator) deleteCharacters(bufferState *BufferState, pos uint64, count uint64) {
	for i := uint64(0); i < count; i++ {
		bufferState.textTree.DeleteAtPosition(pos)
	}

	edit := parser.Edit{Pos: pos, NumDeleted: count}
	if err := bufferState.retokenizeAfterEdit(edit); err != nil {
		log.Printf("Error retokenizing document: %v\n", err)
	}

	bufferState.cursor = cursorState{position: pos}
}

func (dm *deleteMutator) String() string {
	return fmt.Sprintf("Delete(%s)", dm.loc)
}

type setSyntaxMutator struct {
	language syntax.Language
}

func NewSetSyntaxMutator(language syntax.Language) Mutator {
	return &setSyntaxMutator{language}
}

func (ssm *setSyntaxMutator) Mutate(state *EditorState) {
	buffer := state.documentBuffer
	if err := buffer.SetSyntax(ssm.language); err != nil {
		log.Printf("Error setting syntax: %v\n", err)
	}
}

func (ssm *setSyntaxMutator) String() string {
	return fmt.Sprintf("SetSyntax(%s)", ssm.language)
}

type resizeMutator struct {
	width, height uint64
}

func NewResizeMutator(width, height uint64) Mutator {
	return &resizeMutator{width, height}
}

// Mutate resizes the view to the specified width and height.
func (rm *resizeMutator) Mutate(state *EditorState) {
	state.screenWidth = rm.width
	state.screenHeight = rm.height
	state.documentBuffer.view.x = 0
	state.documentBuffer.view.y = 0
	state.documentBuffer.view.width = state.screenWidth
	state.documentBuffer.view.height = 0
	if rm.height > 0 {
		// Leave one line for the status bar at the bottom.
		state.documentBuffer.view.height = rm.height - 1
	}
}

func (rm *resizeMutator) String() string {
	return fmt.Sprintf("Resize(%d,%d)", rm.width, rm.height)
}

type showMenuMutator struct {
	prompt string
	items  []MenuItem
}

func NewShowMenuMutator(prompt string, items []MenuItem) Mutator {
	return &showMenuMutator{prompt, items}
}

// Mutate displays the menu with the specified prompt and items.
func (smm *showMenuMutator) Mutate(state *EditorState) {
	search := &MenuSearch{}
	search.AddItems(smm.items)
	state.menu = &MenuState{
		visible:           true,
		prompt:            smm.prompt,
		search:            search,
		selectedResultIdx: 0,
	}
}

func (smm *showMenuMutator) String() string {
	return fmt.Sprintf("ShowMenu(%s)", smm.prompt)
}

type hideMenuMutator struct{}

func NewHideMenuMutator() Mutator {
	return &hideMenuMutator{}
}

// Mutate hides the menu.
func (hmm *hideMenuMutator) Mutate(state *EditorState) {
	state.menu = &MenuState{}
}

func (hmm *hideMenuMutator) String() string {
	return "HideMenu()"
}

type executeSelectedMenuItemMutator struct{}

func NewExecuteSelectedMenuItemMutator() Mutator {
	return &executeSelectedMenuItemMutator{}
}

// Mutate executes the action of the selected menu item and closes the menu.
func (esm *executeSelectedMenuItemMutator) Mutate(state *EditorState) {
	search := state.menu.search
	results := search.Results()
	if len(results) == 0 {
		// If there are no results, then there is no action to perform.
		NewHideMenuMutator().Mutate(state)
		return
	}

	idx := state.menu.selectedResultIdx
	selectedItem := results[idx]
	log.Printf("Executing menu item %s at result index %d\n", selectedItem, idx)
	NewCompositeMutator([]Mutator{
		selectedItem.Action,
		NewHideMenuMutator(),
	}).Mutate(state)
}

func (esm *executeSelectedMenuItemMutator) String() string {
	return "ExecuteSelectedMenuItem()"
}

type moveMenuSelectionMutator struct {
	// Number of items to move up/down.
	// Negative deltas move the selection up; positive deltas move the selection down.
	delta int
}

func NewMoveMenuSelectionMutator(delta int) Mutator {
	return &moveMenuSelectionMutator{delta}
}

// Mutate moves the menu selection up or down with wraparound.
func (mms *moveMenuSelectionMutator) Mutate(state *EditorState) {
	numResults := len(state.menu.search.Results())
	if numResults == 0 {
		return
	}

	newIdx := (state.menu.selectedResultIdx + mms.delta) % numResults
	if newIdx < 0 {
		newIdx = numResults + newIdx
	}

	state.menu.selectedResultIdx = newIdx
}

func (mms *moveMenuSelectionMutator) String() string {
	return fmt.Sprintf("MoveMenuSelection(%d)", mms.delta)
}

type appendMenuSearchMutator struct {
	r rune
}

func NewAppendMenuSearchMutator(r rune) Mutator {
	return &appendMenuSearchMutator{r}
}

func (ams *appendMenuSearchMutator) Mutate(state *EditorState) {
	menu := state.menu
	newQuery := menu.search.Query() + string(ams.r)
	menu.search.SetQuery(newQuery)
	menu.selectedResultIdx = 0
}

func (ams *appendMenuSearchMutator) String() string {
	return fmt.Sprintf("AppendMenuSearch(%q)", ams.r)
}

type deleteMenuSearchMutator struct{}

func NewDeleteMenuSearchMutator() Mutator {
	return &deleteMenuSearchMutator{}
}

func (dms *deleteMenuSearchMutator) Mutate(state *EditorState) {
	menu := state.menu
	query := menu.search.Query()
	if len(query) > 0 {
		queryRunes := []rune(query)
		newQueryRunes := queryRunes[0 : len(queryRunes)-1]
		menu.search.SetQuery(string(newQueryRunes))
		menu.selectedResultIdx = 0
	}
}

func (dms *deleteMenuSearchMutator) String() string {
	return "DeleteMenuSearch()"
}

type setStatusMsgMutator struct {
	statusMsg StatusMsg
}

func NewSetStatusMsgMutator(statusMsg StatusMsg) Mutator {
	return &setStatusMsgMutator{statusMsg}
}

// Mutate sets the message displayed in the status bar.
func (smm *setStatusMsgMutator) Mutate(state *EditorState) {
	state.statusMsg = smm.statusMsg
}

func (smm *setStatusMsgMutator) String() string {
	return fmt.Sprintf("SetStatusMsg(%s, %q)", smm.statusMsg.Style, smm.statusMsg.Text)
}

type quitMutator struct{}

func NewQuitMutator() Mutator {
	return &quitMutator{}
}

func (qm *quitMutator) Mutate(state *EditorState) {
	state.quitFlag = true
}

func (qm *quitMutator) String() string {
	return "Quit()"
}
