package state

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/text"
)

// LoadDocument loads a file into the editor.
func LoadDocument(state *EditorState, path string, requireExists bool, cursorLoc Locator) {
	timelineState := currentTimelineState(state)
	fileExists, err := loadDocumentAndResetState(state, path, requireExists)
	if err != nil {
		// If this is the first document loaded into the editor, set a watcher
		// even if the load failed.  This retains the attempted path so the user
		// can try saving or reloading the document later.
		if state.fileWatcher.Path() == "" {
			state.fileWatcher = file.NewWatcher(file.DefaultPollInterval, path, time.Time{}, 0, "")
		}

		reportLoadError(state, err, path)
		return
	}

	if !timelineState.Empty() && timelineState.Path != path {
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
	oldSyntaxLanguage := state.documentBuffer.syntaxLanguage
	oldCursorPos := state.documentBuffer.cursor.position

	// Reload the document.
	_, err := loadDocumentAndResetState(state, path, true)
	if err != nil {
		reportLoadError(state, err, path)
		return
	}

	// Retokenize if the configured language doesn't match the previous language.
	// This can happen only when the language was changed through a menu command.
	if oldSyntaxLanguage != state.documentBuffer.syntaxLanguage {
		setSyntaxAndRetokenize(state.documentBuffer, oldSyntaxLanguage)
	}

	// Attempt to restore the original cursor position.
	setCursorAfterLoad(state, func(LocatorParams) uint64 {
		return oldCursorPos
	})

	reportReloadSuccess(state, path)
}

// LoadPrevDocument loads the previous document from the timeline in the editor.
// The cursor is moved to the start of the line from the when the document was last open.
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
		return locate.StartOfLineNum(p.TextTree, prev.LineNum)
	})
	reportOpenSuccess(state, path)
}

// LoadNextDocument loads the next document from the timeline in the editor.
// The cursor is moved to the start of the line from the when the document was last open.
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
		return locate.StartOfLineNum(p.TextTree, next.LineNum)
	})
	reportOpenSuccess(state, path)
}

func currentTimelineState(state *EditorState) file.TimelineState {
	buffer := state.documentBuffer
	cursorPos := buffer.cursor.position
	lineNum := buffer.textTree.LineNumForPosition(cursorPos)
	return file.TimelineState{
		Path:    state.fileWatcher.Path(),
		LineNum: lineNum,
	}
}

func loadDocumentAndResetState(state *EditorState, path string, requireExists bool) (fileExists bool, err error) {
	config := state.configRuleSet.ConfigForPath(path)
	tree, watcher, err := file.Load(path, file.DefaultPollInterval)
	if err := errors.Cause(err); errors.Is(err, os.ErrNotExist) && !requireExists {
		tree = text.NewTree()
		watcher = file.NewWatcher(file.DefaultPollInterval, path, time.Time{}, 0, "")
	} else if err != nil {
		return false, err
	} else {
		fileExists = true
	}

	state.documentBuffer.textTree = tree
	state.fileWatcher.Stop()
	state.fileWatcher = watcher
	state.inputMode = InputModeNormal
	state.prevInputMode = InputModeNormal
	state.documentBuffer.cursor = cursorState{}
	state.documentBuffer.view.textOrigin = 0
	state.documentBuffer.selector.Clear()
	state.documentBuffer.search = searchState{}
	state.documentBuffer.tabSize = uint64(config.TabSize) // safe b/c we validated the config.
	state.documentBuffer.tabExpand = config.TabExpand
	state.documentBuffer.showTabs = config.ShowTabs
	state.documentBuffer.autoIndent = config.AutoIndent
	state.documentBuffer.showLineNum = config.ShowLineNumbers
	state.documentBuffer.undoLog.TrackLoad()
	state.customMenuItems = customMenuItems(config)
	state.dirNamesToHide = stringSliceToMap(config.HideDirectories)
	setSyntaxAndRetokenize(state.documentBuffer, syntax.LanguageFromString(config.SyntaxLanguage))

	return fileExists, nil
}

func setCursorAfterLoad(state *EditorState, cursorLoc Locator) {
	MoveCursor(state, cursorLoc)
	ScrollViewToCursor(state)
}

func stringSliceToMap(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func customMenuItems(config config.Config) []menu.Item {
	items := make([]menu.Item, 0, len(config.MenuCommands))
	for _, cmd := range config.MenuCommands {
		shellCmd := cmd.ShellCmd
		shellCmdMode := cmd.Mode
		items = append(items, menu.Item{
			Name: cmd.Name,
			Action: func(state *EditorState) {
				RunShellCmd(state, shellCmd, shellCmdMode)
			},
		})
	}
	return items
}

func reportOpenSuccess(state *EditorState, path string) {
	log.Printf("Successfully opened file from '%s'", path)
	msg := fmt.Sprintf("Opened %s", file.RelativePathCwd(path))
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  msg,
	})
}

func reportCreateSuccess(state *EditorState, path string) {
	log.Printf("Successfully created file at '%s'", path)
	msg := fmt.Sprintf("New file %s", file.RelativePathCwd(path))
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  msg,
	})
}

func reportReloadSuccess(state *EditorState, path string) {
	log.Printf("Successfully reloaded file from '%s'", path)
	msg := fmt.Sprintf("Reloaded %s", file.RelativePathCwd(path))
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  msg,
	})
}

func reportLoadError(state *EditorState, err error, path string) {
	log.Printf("Error loading file at '%s': %v\n", path, err)
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleError,
		Text:  fmt.Sprintf("Could not open '%s': %s", file.RelativePathCwd(path), errors.Cause(err)),
	})
}

// SaveDocument saves the currently loaded document to disk.
func SaveDocument(state *EditorState) {
	path := state.fileWatcher.Path()
	tree := state.documentBuffer.textTree
	newWatcher, err := file.Save(path, tree, file.DefaultPollInterval)
	if err != nil {
		reportSaveError(state, err, path)
		return
	}

	state.fileWatcher.Stop()
	state.fileWatcher = newWatcher
	state.documentBuffer.undoLog.TrackSave()
	reportSaveSuccess(state, path)
}

func reportSaveError(state *EditorState, err error, path string) {
	log.Printf("Error saving file to '%s': %v", path, err)
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleError,
		Text:  fmt.Sprintf("Could not save %s", path),
	})
}

func reportSaveSuccess(state *EditorState, path string) {
	log.Printf("Successfully wrote file to '%s'", path)
	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  fmt.Sprintf("Saved %s", path),
	})
}

// AbortIfUnsavedChanges executes a function only if the document does not have unsaved changes and shows an error status msg otherwise.
func AbortIfUnsavedChanges(state *EditorState, f func(*EditorState), showStatus bool) {
	if state.documentBuffer.undoLog.HasUnsavedChanges() {
		log.Printf("Aborting operation because document has unsaved changes\n")
		if showStatus {
			SetStatusMsg(state, StatusMsg{
				Style: StatusMsgStyleError,
				Text:  "Document has unsaved changes - either save the changes or force-quit",
			})
		}
	} else {
		f(state)
	}
}

// AbortIfFileChanged executes a function only if the file has not changed on disk; otherwise, it aborts and shows an error message.
func AbortIfFileChanged(state *EditorState, f func(*EditorState)) {
	path := state.fileWatcher.Path()
	if state.fileWatcher.ChangedFlag() {
		log.Printf("Aborting operation because file changed on disk\n")
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("%s has changed since last save.  Use \"force save\" to overwrite.", path),
		})
	} else {
		f(state)
	}
}
