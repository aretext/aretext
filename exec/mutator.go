package exec

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/text/segment"
)

// Mutator modifies the state of the editor.
type Mutator interface {
	fmt.Stringer

	// Mutate modifies the editor state.
	// All changes to editor state should be performed by mutators.
	Mutate(state *EditorState)
}

// CompositeMutator executes a series of mutations in order.
type CompositeMutator struct {
	subMutators []Mutator
}

func NewCompositeMutator(subMutators []Mutator) Mutator {
	return &CompositeMutator{subMutators}
}

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

// abortIfUnsavedChangesMutator executes a sub-mutator only if the document does not have unsaved changes.
type abortIfUnsavedChangesMutator struct {
	subMutator Mutator
	showStatus bool
}

func NewAbortIfUnsavedChangesMutator(subMutator Mutator, showStatus bool) Mutator {
	return &abortIfUnsavedChangesMutator{subMutator, showStatus}
}

func (am *abortIfUnsavedChangesMutator) Mutate(state *EditorState) {
	if state.hasUnsavedChanges {
		log.Printf("Aborting mutator %s because document has unsaved changes", am.subMutator)

		if am.showStatus {
			NewSetStatusMsgMutator(StatusMsg{
				Style: StatusMsgStyleError,
				Text:  "Document has unsaved changes - either save the changes or force-quit",
			}).Mutate(state)
		}

		return
	}

	am.subMutator.Mutate(state)
}

func (am *abortIfUnsavedChangesMutator) String() string {
	return fmt.Sprintf("AbortIfUnsavedChanges(%s, showStatus=%t)", am.subMutator, am.showStatus)
}

// loadDocumentMutator loads the document into the editor.
type loadDocumentMutator struct {
	path          string
	requireExists bool
	showStatus    bool
}

func NewLoadDocumentMutator(path string, requireExists bool, showStatus bool) Mutator {
	return &loadDocumentMutator{path, requireExists, showStatus}
}

func (ldm *loadDocumentMutator) Mutate(state *EditorState) {
	var fileExists bool
	tree, watcher, err := file.Load(ldm.path, file.DefaultPollInterval)
	if os.IsNotExist(err) && !ldm.requireExists {
		tree = text.NewTree()
		watcher = file.NewWatcher(file.DefaultPollInterval, ldm.path, time.Time{}, 0, "")
	} else if err != nil {
		ldm.reportLoadError(err, state)
		return
	} else {
		fileExists = true
	}

	oldPath := state.fileWatcher.Path()
	state.documentBuffer.textTree = tree
	state.fileWatcher.Stop()
	state.fileWatcher = watcher
	state.hasUnsavedChanges = false

	if ldm.path == oldPath {
		ldm.updateAfterReload(state)
		ldm.reportSuccess(state, fileExists)
		return
	}

	config := state.configRuleSet.ConfigForPath(ldm.path)
	if err := config.Validate(); err != nil {
		ldm.reportConfigError(err, state)
		return
	}

	ldm.initializeAfterLoad(state, config)
	ldm.reportSuccess(state, fileExists)
}

func (ldm *loadDocumentMutator) updateAfterReload(state *EditorState) {
	// Make sure that the cursor is a valid position in the updated document
	// and that the cursor is visible.  If not, adjust the cursor and scroll.
	NewCompositeMutator([]Mutator{
		NewSetSyntaxMutator(state.documentBuffer.syntaxLanguage),
		NewCursorMutator(NewOntoDocumentLocator()),
		NewScrollToCursorMutator(),
	}).Mutate(state)
}

func (ldm *loadDocumentMutator) initializeAfterLoad(state *EditorState, config config.Config) {
	state.documentBuffer.cursor = cursorState{}
	state.documentBuffer.view.textOrigin = 0
	state.documentBuffer.SetSyntax(syntax.LanguageFromString(config.SyntaxLanguage))
	state.documentBuffer.tabSize = uint64(config.TabSize) // safe b/c we validated the config.
	state.documentBuffer.tabExpand = config.TabExpand
	state.documentBuffer.autoIndent = config.AutoIndent
	state.customMenuItems = ldm.customMenuItems(config)
}

func (ldm *loadDocumentMutator) customMenuItems(config config.Config) []MenuItem {
	items := make([]MenuItem, 0, len(config.MenuCommands))
	for _, cmd := range config.MenuCommands {
		items = append(items, MenuItem{
			Name:   cmd.Name,
			Action: NewExecuteShellCmdMutator(cmd.ShellCmd),
		})
	}
	return items
}

func (ldm *loadDocumentMutator) reportLoadError(err error, state *EditorState) {
	log.Printf("Error loading file at '%s': %v\n", ldm.path, err)
	if ldm.showStatus {
		NewSetStatusMsgMutator(StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("Could not open %s", file.RelativePathCwd(ldm.path)),
		}).Mutate(state)
	}
}

func (ldm *loadDocumentMutator) reportConfigError(err error, state *EditorState) {
	log.Printf("Invalid configuration for file at '%s': %v\n", ldm.path, err)
	if ldm.showStatus {
		NewSetStatusMsgMutator(StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("Invalid configuration for file at %s: %v", file.RelativePathCwd(ldm.path), err),
		}).Mutate(state)
	}
}

func (ldm *loadDocumentMutator) reportSuccess(state *EditorState, fileExists bool) {
	log.Printf("Successfully loaded file from '%s'", ldm.path)
	if ldm.showStatus {
		var msg string
		relPath := file.RelativePathCwd(ldm.path)
		if fileExists {
			msg = fmt.Sprintf("Opened %s", relPath)
		} else {
			msg = fmt.Sprintf("New file %s", relPath)
		}

		NewSetStatusMsgMutator(StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  msg,
		}).Mutate(state)
	}
}

func (ldm *loadDocumentMutator) String() string {
	return fmt.Sprintf("LoadDocument(%s, showStatus=%t)", ldm.path, ldm.showStatus)
}

// reloadDocumentMutator reloads the current document.
type reloadDocumentMutator struct {
	showStatus bool
}

func NewReloadDocumentMutator(showStatus bool) Mutator {
	return &reloadDocumentMutator{showStatus}
}

func (rdm *reloadDocumentMutator) Mutate(state *EditorState) {
	path := state.fileWatcher.Path()
	NewLoadDocumentMutator(path, true, rdm.showStatus).Mutate(state)
}

func (rdm *reloadDocumentMutator) String() string {
	return fmt.Sprintf("ReloadDocument(showStatus=%t)", rdm.showStatus)
}

// saveDocumentMutator saves the currently loaded document to disk.
type saveDocumentMutator struct {
	force bool
}

func NewSaveDocumentMutator(force bool) Mutator {
	return &saveDocumentMutator{force}
}

func (sdm *saveDocumentMutator) Mutate(state *EditorState) {
	path := state.fileWatcher.Path()
	if state.fileWatcher.ChangedFlag() && !sdm.force {
		NewSetStatusMsgMutator(StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("%s has changed since last save.  Use \"force save\" to overwrite.", path),
		}).Mutate(state)
		return
	}

	tree := state.documentBuffer.textTree
	newWatcher, err := file.Save(path, tree, file.DefaultPollInterval)
	if err != nil {
		sdm.reportError(state, path, err)
		return
	}

	state.fileWatcher.Stop()
	state.fileWatcher = newWatcher
	state.hasUnsavedChanges = false
	sdm.reportSuccess(state, path)
}

func (sdm *saveDocumentMutator) reportError(state *EditorState, path string, err error) {
	log.Printf("Error saving file to '%s': %v", path, err)
	NewSetStatusMsgMutator(StatusMsg{
		Style: StatusMsgStyleError,
		Text:  fmt.Sprintf("Could not save %s", path),
	}).Mutate(state)
}

func (sdm *saveDocumentMutator) reportSuccess(state *EditorState, path string) {
	log.Printf("Successfully wrote file to '%s'", path)
	NewSetStatusMsgMutator(StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  fmt.Sprintf("Saved %s", path),
	}).Mutate(state)
}

func (sdm *saveDocumentMutator) String() string {
	return fmt.Sprintf("SaveDocument(force=%t)", sdm.force)
}

// setInputModeMutator sets the editor input mode.
type setInputModeMutator struct {
	mode InputMode
}

func NewSetInputModeMutator(mode InputMode) Mutator {
	return &setInputModeMutator{mode}
}

func (sim *setInputModeMutator) Mutate(state *EditorState) {
	state.inputMode = sim.mode
}

func (sim *setInputModeMutator) String() string {
	return fmt.Sprintf("SetInputMode(%s)", sim.mode)
}

// cursorMutator returns a mutator that updates the cursor location.
type cursorMutator struct {
	loc CursorLocator
}

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

// scrollToCursorMutator updates the view origin so that the cursor is visible.
type scrollToCursorMutator struct{}

func NewScrollToCursorMutator() Mutator {
	return &scrollToCursorMutator{}
}

func (sm *scrollToCursorMutator) Mutate(state *EditorState) {
	bufferState := state.documentBuffer
	bufferState.view.textOrigin = ScrollToCursor(
		bufferState.cursor.position,
		bufferState.textTree,
		bufferState.view.textOrigin,
		bufferState.view.width,
		bufferState.view.height,
		bufferState.tabSize)
}

func (sm *scrollToCursorMutator) String() string {
	return "ScrollToCursor()"
}

// scrollLinesMutator moves the view origin up/down by the specified number of lines.
type scrollLinesMutator struct {
	direction text.ReadDirection
	numLines  uint64
}

func NewScrollLinesMutator(direction text.ReadDirection, numLines uint64) Mutator {
	return &scrollLinesMutator{direction, numLines}
}

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

// insertRuneMutator inserts a rune at the current cursor location.
type insertRuneMutator struct {
	r rune
}

func NewInsertRuneMutator(r rune) Mutator {
	return &insertRuneMutator{r}
}

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
	state.hasUnsavedChanges = true
}

func (irm *insertRuneMutator) String() string {
	return fmt.Sprintf("InsertRune(%q)", irm.r)
}

// insertNewlineMutator inserts a newline at the current cursor position.
type insertNewlineMutator struct{}

func NewInsertNewlineMutator() Mutator {
	return &insertNewlineMutator{}
}

func (inm *insertNewlineMutator) Mutate(state *EditorState) {
	NewInsertRuneMutator('\n').Mutate(state)

	buffer := state.documentBuffer
	if buffer.autoIndent {
		numCols := inm.numColsIndentedPrevLine(buffer)
		inm.indentFromCursor(state, numCols)
	}
}

func (inm *insertNewlineMutator) numColsIndentedPrevLine(buffer *BufferState) uint64 {
	tabSize := buffer.tabSize
	lineNum := buffer.textTree.LineNumForPosition(buffer.cursor.position)
	if lineNum == 0 {
		return 0
	}

	prevLineStartPos := buffer.textTree.LineStartPosition(lineNum - 1)
	iter := gcIterForTree(buffer.textTree, prevLineStartPos, text.ReadDirectionForward)
	seg := segment.NewSegment()
	numCols := uint64(0)
	for {
		eof := nextSegmentOrEof(iter, seg)
		if eof {
			break
		}

		gc := seg.Runes()
		if gc[0] != '\t' && gc[0] != ' ' {
			break
		}

		numCols += GraphemeClusterWidth(gc, numCols, tabSize)
	}

	return numCols
}

func (inm *insertNewlineMutator) indentFromCursor(state *EditorState, numCols uint64) {
	tabSize := state.documentBuffer.tabSize
	tabExpand := state.documentBuffer.tabExpand
	insertTabMutator := NewInsertRuneMutator('\t')
	insertSpaceMutator := NewInsertRuneMutator(' ')

	i := uint64(0)
	for i < numCols {
		if !tabExpand && numCols-i >= tabSize {
			insertTabMutator.Mutate(state)
			i += tabSize
		} else {
			insertSpaceMutator.Mutate(state)
			i++
		}
	}
}

func (inm *insertNewlineMutator) String() string {
	return "InsertNewline()"
}

// insertTabMutator inserts a tab at the current cursor position.
type insertTabMutator struct{}

func NewInsertTabMutator() Mutator {
	return &insertTabMutator{}
}

func (itm *insertTabMutator) Mutate(state *EditorState) {
	if state.documentBuffer.tabExpand {
		itm.insertSpaces(state)
		return
	}
	NewInsertRuneMutator('\t').Mutate(state)
}

func (itm *insertTabMutator) insertSpaces(state *EditorState) {
	buffer := state.documentBuffer
	tabSize := buffer.tabSize
	offset := itm.offsetInLine(state.documentBuffer)
	numSpaces := tabSize - (offset % tabSize)
	insertSpaceMut := NewInsertRuneMutator(' ')
	for i := uint64(0); i < numSpaces; i++ {
		insertSpaceMut.Mutate(state)
	}
}

func (itm *insertTabMutator) offsetInLine(buffer *BufferState) uint64 {
	var offset uint64
	pos := lineStartPos(buffer.textTree, buffer.cursor.position)
	iter := gcIterForTree(buffer.textTree, pos, text.ReadDirectionForward)
	seg := segment.NewSegment()
	for pos < buffer.cursor.position {
		eof := nextSegmentOrEof(iter, seg)
		if eof {
			break
		}
		offset += GraphemeClusterWidth(seg.Runes(), offset, buffer.tabSize)
		pos += seg.NumRunes()
	}
	return offset
}

func (itm *insertTabMutator) String() string {
	return "InsertTab()"
}

// deleteMutator deletes characters from the cursor position up to (but not including) the position returned by the locator.
// It can delete either forwards or backwards from the cursor.
// The cursor position will be set to the start of the deleted region,
// which could be on a newline character or past the end of the text.
type deleteMutator struct {
	loc CursorLocator
}

func NewDeleteMutator(loc CursorLocator) Mutator {
	return &deleteMutator{loc}
}

func (dm *deleteMutator) Mutate(state *EditorState) {
	bufferState := state.documentBuffer
	startPos := bufferState.cursor.position
	deleteToPos := dm.loc.Locate(bufferState).position

	if startPos < deleteToPos {
		dm.deleteCharacters(bufferState, startPos, deleteToPos-startPos)
	} else if startPos > deleteToPos {
		dm.deleteCharacters(bufferState, deleteToPos, startPos-deleteToPos)
	}
	state.hasUnsavedChanges = true
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

// deleteLinesMutator deletes lines from the cursor's current line to the line of a target cursor.
// It moves the cursor to the start of the line following the last deleted line.
type deleteLinesMutator struct {
	targetLineLocator          CursorLocator
	abortIfTargetIsCurrentLine bool
}

func NewDeleteLinesMutator(targetLineLocator CursorLocator, abortIfTargetIsCurrentLine bool) Mutator {
	return &deleteLinesMutator{targetLineLocator, abortIfTargetIsCurrentLine}
}

func (dlm *deleteLinesMutator) Mutate(state *EditorState) {
	buffer := state.documentBuffer
	currentLine := buffer.textTree.LineNumForPosition(buffer.cursor.position)
	targetCursor := dlm.targetLineLocator.Locate(buffer)
	targetLine := buffer.textTree.LineNumForPosition(targetCursor.position)

	if targetLine == currentLine && dlm.abortIfTargetIsCurrentLine {
		return
	}

	if targetLine < currentLine {
		currentLine, targetLine = targetLine, currentLine
	}

	numLinesToDelete := targetLine - currentLine + 1
	for i := uint64(0); i < numLinesToDelete; i++ {
		dlm.deleteLine(state, currentLine)
	}
}

func (dlm *deleteLinesMutator) deleteLine(state *EditorState, lineNum uint64) {
	buffer := state.documentBuffer
	startOfLinePos := buffer.textTree.LineStartPosition(lineNum)
	startOfNextLinePos := buffer.textTree.LineStartPosition(lineNum + 1)

	isLastLine := lineNum+1 >= buffer.textTree.NumLines()
	if isLastLine && startOfLinePos > 0 {
		// The last line does not have a newline at the end, so delete the newline from the end of the previous line instead.
		startOfLinePos--
	}

	numToDelete := startOfNextLinePos - startOfLinePos
	for i := uint64(0); i < numToDelete; i++ {
		buffer.textTree.DeleteAtPosition(startOfLinePos)
	}

	edit := parser.Edit{Pos: startOfLinePos, NumDeleted: numToDelete}
	if err := buffer.retokenizeAfterEdit(edit); err != nil {
		log.Printf("Error retokenizing doument: %v\n", err)
	}

	buffer.cursor = cursorState{position: startOfLinePos}
	if buffer.cursor.position >= buffer.textTree.NumChars() {
		buffer.cursor = NewLastLineLocator().Locate(buffer)
	}

	state.hasUnsavedChanges = state.hasUnsavedChanges || numToDelete > 0
}

func (dlm *deleteLinesMutator) String() string {
	return fmt.Sprintf("DeleteLines(%s, abortIfTargetIsCurrentLine=%t)", dlm.targetLineLocator, dlm.abortIfTargetIsCurrentLine)
}

// setSyntaxMutator sets the syntax language for the current document.
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

// resizeMutator resizes the view to the specified width and height.
type resizeMutator struct {
	width, height uint64
}

func NewResizeMutator(width, height uint64) Mutator {
	return &resizeMutator{width, height}
}

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

// showMenuMutator displays the menu with the specified prompt and items.
type showMenuMutator struct {
	prompt              string
	loadItems           func() []MenuItem
	emptyQueryShowAll   bool
	showCustomMenuItems bool
}

func NewShowMenuMutator(prompt string, loadItems func() []MenuItem, emptyQueryShowAll bool, showCustomMenuItems bool) Mutator {
	return &showMenuMutator{
		prompt:              prompt,
		loadItems:           loadItems,
		emptyQueryShowAll:   emptyQueryShowAll,
		showCustomMenuItems: showCustomMenuItems,
	}
}

func NewShowMenuMutatorWithItems(prompt string, items []MenuItem, emptyQueryShowAll bool, showCustomMenuItems bool) Mutator {
	loadItems := func() []MenuItem { return items }
	return NewShowMenuMutator(prompt, loadItems, emptyQueryShowAll, showCustomMenuItems)
}

func (smm *showMenuMutator) Mutate(state *EditorState) {
	search := &MenuSearch{emptyQueryShowAll: smm.emptyQueryShowAll}
	search.AddItems(smm.loadItems())

	if smm.showCustomMenuItems {
		search.AddItems(state.customMenuItems)
	}

	state.menu = &MenuState{
		visible:           true,
		prompt:            smm.prompt,
		search:            search,
		selectedResultIdx: 0,
	}
	state.inputMode = InputModeMenu
}

func (smm *showMenuMutator) String() string {
	return fmt.Sprintf("ShowMenu(%s, emptyQueryShowAll=%t, showCustomMenuItems=%t)", smm.prompt, smm.emptyQueryShowAll, smm.showCustomMenuItems)
}

// hideMenuMutator hides the menu.
type hideMenuMutator struct{}

func NewHideMenuMutator() Mutator {
	return &hideMenuMutator{}
}

func (hmm *hideMenuMutator) Mutate(state *EditorState) {
	state.menu = &MenuState{}
	state.inputMode = InputModeNormal
}

func (hmm *hideMenuMutator) String() string {
	return "HideMenu()"
}

// executeSelectedMenuItemMutator executes the action of the selected menu item and closes the menu.
type executeSelectedMenuItemMutator struct{}

func NewExecuteSelectedMenuItemMutator() Mutator {
	return &executeSelectedMenuItemMutator{}
}

func (esm *executeSelectedMenuItemMutator) Mutate(state *EditorState) {
	search := state.menu.search
	results := search.Results()
	if len(results) == 0 {
		// If there are no results, then there is no action to perform.
		NewCompositeMutator([]Mutator{
			NewSetStatusMsgMutator(StatusMsg{
				Style: StatusMsgStyleError,
				Text:  "No menu item selected",
			}),
			NewHideMenuMutator(),
		}).Mutate(state)
		return
	}

	idx := state.menu.selectedResultIdx
	selectedItem := results[idx]
	log.Printf("Executing menu item %s at result index %d\n", selectedItem, idx)
	NewCompositeMutator([]Mutator{
		// Perform the action after hiding the menu in case the action so the action can show another menu.
		NewHideMenuMutator(),
		selectedItem.Action,
	}).Mutate(state)
}

func (esm *executeSelectedMenuItemMutator) String() string {
	return "ExecuteSelectedMenuItem()"
}

// moveMenuSelectionMutator moves the menu selection up or down with wraparound.
type moveMenuSelectionMutator struct {
	// Number of items to move up/down.
	// Negative deltas move the selection up; positive deltas move the selection down.
	delta int
}

func NewMoveMenuSelectionMutator(delta int) Mutator {
	return &moveMenuSelectionMutator{delta}
}

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

// appendMenuSearchMutator appends a rune to the menu search query.
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

// deleteMenuSearchMutator deletes a rune from the menu search query.
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

// setStatusMsgMutator sets the message displayed in the status bar.
type setStatusMsgMutator struct {
	statusMsg StatusMsg
}

func NewSetStatusMsgMutator(statusMsg StatusMsg) Mutator {
	return &setStatusMsgMutator{statusMsg}
}

func (smm *setStatusMsgMutator) Mutate(state *EditorState) {
	state.statusMsg = smm.statusMsg
}

func (smm *setStatusMsgMutator) String() string {
	return fmt.Sprintf("SetStatusMsg(%s, %q)", smm.statusMsg.Style, smm.statusMsg.Text)
}

// executeShellCmdMutator runs a command in a shell and pipes the output to a pager (usually `less`).
// Once the command completes and the user exits the pager, control returns to the event loop.
type executeShellCmdMutator struct {
	shellCmd string
}

func NewExecuteShellCmdMutator(shellCmd string) Mutator {
	return &executeShellCmdMutator{shellCmd}
}

func (esm *executeShellCmdMutator) Mutate(state *EditorState) {
	// The command will overwrite the terminal, so we need to redraw after it completes.
	state.forceRedrawFlag = true

	// Run the shell command and pipe the output to a pager (usually `less`).
	// This will take control of the tty, effectively pausing the event loop until the command completes and the user exits the pager.
	err := RunShellCmd(esm.shellCmd)
	if err != nil {
		esm.reportError(state, err)
		return
	}
	esm.reportSuccess(state)
}

func (esm *executeShellCmdMutator) reportError(state *EditorState, err error) {
	log.Printf("Error running shell cmd '%s': %v\n", esm.shellCmd, err)
	NewSetStatusMsgMutator(StatusMsg{
		Style: StatusMsgStyleError,
		Text:  err.Error(),
	}).Mutate(state)
}

func (esm *executeShellCmdMutator) reportSuccess(state *EditorState) {
	log.Printf("Completed shell cmd '%s'\n", esm.shellCmd)
	NewSetStatusMsgMutator(StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  "Command completed successfully",
	}).Mutate(state)
}

func (esm *executeShellCmdMutator) String() string {
	return fmt.Sprintf("ExecuteShellCmd('%s')", esm.shellCmd)
}

// quitMutator sets a flag that terminates the program.
type quitMutator struct{}

func NewQuitMutator() Mutator {
	return &quitMutator{}
}

func (qm *quitMutator) Mutate(state *EditorState) {
	state.fileWatcher.Stop()
	state.quitFlag = true
}

func (qm *quitMutator) String() string {
	return fmt.Sprintf("Quit()")
}
