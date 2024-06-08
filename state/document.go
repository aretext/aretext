package state

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/text"
	"github.com/aretext/aretext/undo"
)

// NewDocument opens a new document at the given path.
// Returns an error if the file already exists or the directory doesn't exist.
// This won't create a new file on disk until the user saves it.
func NewDocument(state *EditorState, path string) error {
	err := file.ValidateCreate(path)
	if err != nil {
		return err
	}
	// Initialize the editor with the file path.
	// This won't create the file on disk until the user saves it.
	// It's possible that some other process created a file at the path
	// before or after it's loaded, but this is handled gracefully.
	LoadDocument(state, path, false, func(_ LocatorParams) uint64 { return 0 })
	return nil
}

// RenameDocument moves a document to a different file path.
// Returns an error if the file already exists or the directory doesn't exist.
func RenameDocument(state *EditorState, newPath string) error {
	// Validate that we can create a file at the new path.
	// This isn't 100% reliable, since some other process could create a file
	// at the target path between this check and the rename below, but it at least
	// reduces the risk of overwriting another file.
	err := file.ValidateCreate(newPath)
	if err != nil {
		return err
	}

	// Move the file on disk. Ignore fs.ErrNotExist which can happen if
	// the file was never saved to the old path.
	//
	// The rename won't trigger a reload of the old document because:
	// 1. file.Watcher's check loop ignores fs.ErrNotExist.
	// 2. LoadDocument below starts a new file watcher, so the main event loop
	//    won't check the old file.Watcher's changed channel anyway.
	oldPath := state.fileWatcher.Path()
	err = os.Rename(oldPath, newPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	// Load the document at the new path, retaining the original cursor position.
	cursorPos := state.documentBuffer.cursor.position
	LoadDocument(state, newPath, false, func(_ LocatorParams) uint64 { return cursorPos })
	return nil
}

// LoadDocument loads a file into the editor.
func LoadDocument(state *EditorState, path string, requireExists bool, cursorLoc Locator) {
	timelineState := currentTimelineState(state)
	fileExists, err := loadDocumentAndResetState(state, path, requireExists)
	if err != nil {
		// If this is the first document loaded into the editor, set a watcher
		// even if the load failed.  This retains the attempted path so the user
		// can try saving or reloading the document later.
		if state.fileWatcher.Path() == "" {
			state.fileWatcher = file.NewWatcherForNewFile(file.DefaultPollInterval, path)
		}

		reportLoadError(state, err, path)
		return
	}

	if !timelineState.Empty() {
		state.fileTimeline.TransitionFrom(timelineState)
	}

	setCursorAfterLoad(state, cursorLoc)

	if fileExists {
		reportOpenSuccess(state, path)
	} else {
		reportCreateSuccess(state, path)
	}
}

// ReloadDocument reloads the current document.
func ReloadDocument(state *EditorState) {
	path := state.fileWatcher.Path()

	// Store the configuration we want to preserve.
	oldTextTree := state.documentBuffer.textTree
	oldText := oldTextTree.String()
	oldTextOriginLineNum := oldTextTree.LineNumForPosition(state.documentBuffer.view.textOrigin)
	oldCursorLineNum, oldCursorCol := locate.PosToLineNumAndCol(oldTextTree, state.documentBuffer.cursor.position)
	oldSearch := state.documentBuffer.search
	oldAutoIndent := state.documentBuffer.autoIndent
	oldShowTabs := state.documentBuffer.showTabs
	oldShowSpaces := state.documentBuffer.showSpaces
	oldShowLineNum := state.documentBuffer.showLineNum
	oldLineNumberMode := state.documentBuffer.lineNumberMode

	// Reload the document.
	_, err := loadDocumentAndResetState(state, path, true)
	if err != nil {
		reportLoadError(state, err, path)
		return
	}

	// Attempt to restore the original cursor and scroll positions, aligned to the new document.
	newTextTree := state.documentBuffer.textTree
	newTreeReader := newTextTree.ReaderAtPosition(0)
	oldReader := strings.NewReader(oldText)
	lineMatches, err := text.Align(oldReader, &newTreeReader)
	if err != nil {
		panic(err) // Should never happen since we're reading from in-memory strings.
	}
	state.documentBuffer.cursor.position = locate.LineNumAndColToPos(
		newTextTree,
		translateLineNum(lineMatches, oldCursorLineNum),
		oldCursorCol,
	)
	state.documentBuffer.view.textOrigin = newTextTree.LineStartPosition(
		translateLineNum(lineMatches, oldTextOriginLineNum),
	)
	ScrollViewToCursor(state)

	// Restore search query, direction, and history.
	state.documentBuffer.search = searchState{
		query:     oldSearch.query,
		direction: oldSearch.direction,
		history:   oldSearch.history,
	}

	// Restore other configuration that might have been toggled with menu commands.
	state.documentBuffer.autoIndent = oldAutoIndent
	state.documentBuffer.showTabs = oldShowTabs
	state.documentBuffer.showSpaces = oldShowSpaces
	state.documentBuffer.showLineNum = oldShowLineNum
	state.documentBuffer.lineNumberMode = oldLineNumberMode

	reportReloadSuccess(state, path)
}

func translateLineNum(lineMatches []text.LineMatch, lineNum uint64) uint64 {
	matchIdx := sort.Search(len(lineMatches), func(i int) bool {
		return lineMatches[i].LeftLineNum >= lineNum
	})

	if matchIdx < len(lineMatches) && lineMatches[matchIdx].LeftLineNum == lineNum {
		alignedLineNum := lineMatches[matchIdx].RightLineNum
		log.Printf("Aligned line %d in old document with line %d in new document\n", lineNum, alignedLineNum)
		return lineMatches[matchIdx].RightLineNum
	} else {
		log.Printf("Could not find alignment for line number %d\n", lineNum)
		return lineNum
	}
}

// LoadPrevDocument loads the previous document from the timeline in the editor.
// The cursor is moved to the start of the line from when the document was last open.
func LoadPrevDocument(state *EditorState) {
	prev := state.fileTimeline.PeekBackward()
	if prev.Empty() {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  "No previous document to open",
		})
		return
	}

	timelineState := currentTimelineState(state)
	path := prev.Path
	_, err := loadDocumentAndResetState(state, path, false)
	if err != nil {
		reportLoadError(state, err, path)
		return
	}

	state.fileTimeline.TransitionBackwardFrom(timelineState)
	setCursorAfterLoad(state, func(p LocatorParams) uint64 {
		return locate.LineNumAndColToPos(p.TextTree, prev.LineNum, prev.Col)
	})
	reportOpenSuccess(state, path)
}

// LoadNextDocument loads the next document from the timeline in the editor.
// The cursor is moved to the start of the line from when the document was last open.
func LoadNextDocument(state *EditorState) {
	next := state.fileTimeline.PeekForward()
	if next.Empty() {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  "No next document to open",
		})
		return
	}

	timelineState := currentTimelineState(state)
	path := next.Path
	_, err := loadDocumentAndResetState(state, path, false)
	if err != nil {
		reportLoadError(state, err, path)
		return
	}

	state.fileTimeline.TransitionForwardFrom(timelineState)
	setCursorAfterLoad(state, func(p LocatorParams) uint64 {
		return locate.LineNumAndColToPos(p.TextTree, next.LineNum, next.Col)
	})
	reportOpenSuccess(state, path)
}

func currentTimelineState(state *EditorState) file.TimelineState {
	buffer := state.documentBuffer
	lineNum, col := locate.PosToLineNumAndCol(buffer.textTree, buffer.cursor.position)
	return file.TimelineState{
		Path:    state.fileWatcher.Path(),
		LineNum: lineNum,
		Col:     col,
	}
}

func loadDocumentAndResetState(state *EditorState, path string, requireExists bool) (fileExists bool, err error) {
	cfg := state.configRuleSet.ConfigForPath(path)
	loadedFile, err := file.Load(path, file.DefaultPollInterval)
	if errors.Is(err, fs.ErrNotExist) && !requireExists {
		loadedFile = file.LoadedFile{
			TextTree:    text.NewTree(),
			FileWatcher: file.NewWatcherForNewFile(file.DefaultPollInterval, path),
			PosixEof:    true,
		}
	} else if err != nil {
		return false, err
	} else {
		fileExists = true
	}

	CancelTaskIfRunning(state)
	state.documentLoadCount++
	state.documentBuffer.textTree = loadedFile.TextTree
	state.documentBuffer.posixEof = loadedFile.PosixEof
	state.fileWatcher.Stop()
	state.fileWatcher = loadedFile.FileWatcher
	state.inputMode = InputModeNormal
	state.documentBuffer.cursor = cursorState{}
	state.documentBuffer.view.textOrigin = 0
	state.documentBuffer.selector.Clear()
	state.documentBuffer.search = searchState{}
	state.documentBuffer.tabSize = uint64(cfg.TabSize) // safe b/c we validated the config.
	state.documentBuffer.tabExpand = cfg.TabExpand
	state.documentBuffer.showTabs = cfg.ShowTabs
	state.documentBuffer.showSpaces = cfg.ShowSpaces
	state.documentBuffer.autoIndent = cfg.AutoIndent
	state.documentBuffer.showLineNum = cfg.ShowLineNumbers
	state.documentBuffer.lineNumberMode = config.LineNumberMode(cfg.LineNumberMode)
	state.documentBuffer.lineWrapAllowCharBreaks = bool(cfg.LineWrap == config.LineWrapCharacter)
	state.documentBuffer.undoLog = undo.NewLog()
	state.menu = &MenuState{}
	state.customMenuItems = customMenuItems(cfg)
	state.dirPatternsToHide = cfg.HideDirectories
	state.styles = cfg.Styles
	setSyntaxAndRetokenize(state.documentBuffer, syntax.Language(cfg.SyntaxLanguage))

	return fileExists, nil
}

func setCursorAfterLoad(state *EditorState, cursorLoc Locator) {
	// First, scroll to the last line.
	MoveCursor(state, func(p LocatorParams) uint64 {
		return locate.StartOfLastLine(p.TextTree)
	})
	ScrollViewToCursor(state)

	// Then, scroll up to the target location.
	// This ensures that the target line appears near the top
	// of the view instead of near the bottom.
	MoveCursor(state, cursorLoc)
	ScrollViewToCursor(state)
}

func customMenuItems(cfg config.Config) []menu.Item {
	// Deduplicate commands with the same name.
	// Later commands take priority.
	uniqueItemMap := make(map[string]menu.Item, len(cfg.MenuCommands))
	for _, cmd := range cfg.MenuCommands {
		uniqueItemMap[cmd.Name] = menu.Item{
			Name:   cmd.Name,
			Action: actionForCustomMenuItem(cmd),
		}
	}

	// Convert the map to a slice.
	items := make([]menu.Item, 0, len(uniqueItemMap))
	for _, item := range uniqueItemMap {
		items = append(items, item)
	}

	// Sort the slice ascending by name.
	// This isn't strictly necessary since menu search will reorder
	// the commands based on the user's search query, but do it anyway
	// so the output is deterministic.
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items
}

func actionForCustomMenuItem(cmd config.MenuCommandConfig) func(*EditorState) {
	if cmd.Save {
		return func(state *EditorState) {
			AbortIfFileChanged(state, func(state *EditorState) {
				SaveDocumentIfUnsavedChanges(state)
				RunShellCmd(state, cmd.ShellCmd, cmd.Mode)
			})
		}
	} else {
		return func(state *EditorState) {
			RunShellCmd(state, cmd.ShellCmd, cmd.Mode)
		}
	}
}

func reportOpenSuccess(state *EditorState, path string) {
	log.Printf("Successfully opened file from %q", path)
	msg := fmt.Sprintf("Opened %s", file.RelativePathCwd(path))
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  msg,
	})
}

func reportCreateSuccess(state *EditorState, path string) {
	log.Printf("Successfully created file at %q", path)
	msg := fmt.Sprintf("New file %s", file.RelativePathCwd(path))
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  msg,
	})
}

func reportReloadSuccess(state *EditorState, path string) {
	log.Printf("Successfully reloaded file from %q", path)
	msg := fmt.Sprintf("Reloaded %s", file.RelativePathCwd(path))
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  msg,
	})
}

func reportLoadError(state *EditorState, err error, path string) {
	log.Printf("Error loading file at %q: %v\n", path, err)
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleError,
		Text:  fmt.Sprintf("Could not open %q: %s", file.RelativePathCwd(path), err),
	})
}

// SaveDocument saves the currently loaded document to disk.
func SaveDocument(state *EditorState) {
	path := state.fileWatcher.Path()
	tree := state.documentBuffer.textTree
	appendPosixEof := state.documentBuffer.posixEof
	newWatcher, err := file.Save(path, tree, appendPosixEof, file.DefaultPollInterval)
	if err != nil {
		reportSaveError(state, err, path)
		return
	}

	state.fileWatcher.Stop()
	state.fileWatcher = newWatcher
	state.documentBuffer.undoLog.TrackSave()
	reportSaveSuccess(state, path)
}

// SaveDocumentIfUnsavedChanges saves the document only if it has been edited
// or the file does not exist on disk.
func SaveDocumentIfUnsavedChanges(state *EditorState) {
	path := state.fileWatcher.Path()
	_, err := os.Stat(path)
	undoLog := state.documentBuffer.undoLog
	if undoLog.HasUnsavedChanges() || errors.Is(err, os.ErrNotExist) {
		SaveDocument(state)
	}
}

func reportSaveError(state *EditorState, err error, path string) {
	log.Printf("Error saving file to %q: %v", path, err)
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleError,
		Text:  fmt.Sprintf("Could not save %q: %s", file.RelativePathCwd(path), err),
	})
}

func reportSaveSuccess(state *EditorState, path string) {
	log.Printf("Successfully wrote file to %q", path)
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  fmt.Sprintf("Saved %s", path),
	})
}

const DefaultUnsavedChangesAbortMsg = `Document has unsaved changes. Either save them ("force save") or discard them ("force reload") and try again`

// AbortIfUnsavedChanges executes a function only if the document does not have unsaved changes and shows an error status msg otherwise.
func AbortIfUnsavedChanges(state *EditorState, abortMsg string, f func(*EditorState)) {
	if state.documentBuffer.undoLog.HasUnsavedChanges() {
		log.Printf("Aborting operation because document has unsaved changes\n")
		if abortMsg != "" {
			SetStatusMsg(state, StatusMsg{
				Style: StatusMsgStyleError,
				Text:  abortMsg,
			})
		}
		return
	}

	// Document has no unsaved changes, so execute the operation.
	f(state)
}

// AbortIfFileChanged aborts with an error message if the file has changed on disk.
// Specifically, abort if the file was moved/deleted or its content checksum has changed.
func AbortIfFileChanged(state *EditorState, f func(*EditorState)) {
	path := state.fileWatcher.Path()
	filename := filepath.Base(path)

	movedOrDeleted, err := state.fileWatcher.CheckFileMovedOrDeleted()
	if err != nil {
		log.Printf("Aborting operation because error occurred checking if file was moved or deleted: %s\n", err)
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("Could not check file: %s", err),
		})
		return
	}

	if movedOrDeleted {
		log.Printf("Aborting operation because file was moved or deleted\n")
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("File %s was moved or deleted. Use \"force save\" to save the file at the current path", filename),
		})
		return
	}

	changed, err := state.fileWatcher.CheckFileContentsChanged()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Printf("Aborting operation because error occurred checking the file contents: %s\n", err)
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("Could not checksum file: %s", err),
		})
		return
	}

	if changed {
		log.Printf("Aborting operation because file changed on disk\n")
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("File %s has changed since last save.  Use \"force save\" to overwrite.", filename),
		})
		return
	}

	// All checks passed, so execute the action.
	f(state)
}
