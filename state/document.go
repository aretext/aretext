package state

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/text"
)

// LoadDocument loads a file into the editor.
func LoadDocument(state *EditorState, path string, requireExists bool, showStatus bool) {
	var fileExists bool
	tree, watcher, err := file.Load(path, file.DefaultPollInterval)
	if os.IsNotExist(err) && !requireExists {
		tree = text.NewTree()
		watcher = file.NewWatcher(file.DefaultPollInterval, path, time.Time{}, 0, "")
	} else if err != nil {
		reportLoadError(state, err, path, showStatus)
		return
	} else {
		fileExists = true
	}

	oldPath := state.fileWatcher.Path()
	state.documentBuffer.textTree = tree
	state.fileWatcher.Stop()
	state.fileWatcher = watcher
	state.hasUnsavedChanges = false

	if path == oldPath {
		updateAfterReload(state)
		reportLoadSuccess(state, fileExists, path, showStatus)
		return
	}

	config := state.configRuleSet.ConfigForPath(path)
	if err := config.Validate(); err != nil {
		reportConfigError(state, err, path, showStatus)
		return
	}

	initializeAfterLoad(state, config)
	reportLoadSuccess(state, fileExists, path, showStatus)
}

func updateAfterReload(state *EditorState) {
	// Tokenize the document using the current syntax language.
	SetSyntax(state, state.documentBuffer.syntaxLanguage)

	// Ensure that the cursor position is within the document and within the current view.
	MoveCursorOntoDocument(state)
	ScrollViewToCursor(state)
}

func initializeAfterLoad(state *EditorState, config config.Config) {
	state.documentBuffer.cursor = cursorState{}
	state.documentBuffer.view.textOrigin = 0
	state.documentBuffer.tabSize = uint64(config.TabSize) // safe b/c we validated the config.
	state.documentBuffer.tabExpand = config.TabExpand
	state.documentBuffer.autoIndent = config.AutoIndent
	state.customMenuItems = customMenuItems(config)
	state.dirNamesToHide = stringSliceToMap(config.HideDirectories)
	setSyntaxAndRetokenize(state.documentBuffer, syntax.LanguageFromString(config.SyntaxLanguage))
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
		items = append(items, menu.Item{
			Name: cmd.Name,
			Action: func(state *EditorState) {
				ScheduleShellCmd(state, cmd.ShellCmd)
			},
		})
	}
	return items
}

func reportLoadError(state *EditorState, err error, path string, showStatus bool) {
	log.Printf("Error loading file at '%s': %v\n", path, err)
	if showStatus {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("Could not open %s", file.RelativePathCwd(path)),
		})
	}
}

func reportLoadSuccess(state *EditorState, fileExists bool, path string, showStatus bool) {
	log.Printf("Successfully loaded file from '%s'", path)
	if showStatus {
		var msg string
		relPath := file.RelativePathCwd(path)
		if fileExists {
			msg = fmt.Sprintf("Opened %s", relPath)
		} else {
			msg = fmt.Sprintf("New file %s", relPath)
		}

		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  msg,
		})
	}
}

func reportConfigError(state *EditorState, err error, path string, showStatus bool) {
	log.Printf("Invalid configuration for file at '%s': %v\n", path, err)
	if showStatus {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("Invalid configuration for file at %s: %v", file.RelativePathCwd(path), err),
		})
	}
}

// ReloadDocument reloads the current document.
func ReloadDocument(state *EditorState, showStatus bool) {
	path := state.fileWatcher.Path()
	LoadDocument(state, path, false, showStatus)
}

// SaveDocument saves the currently loaded document to disk.
func SaveDocument(state *EditorState, force bool) {
	path := state.fileWatcher.Path()
	if state.fileWatcher.ChangedFlag() && !force {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("%s has changed since last save.  Use \"force save\" to overwrite.", path),
		})
		return
	}

	tree := state.documentBuffer.textTree
	newWatcher, err := file.Save(path, tree, file.DefaultPollInterval)
	if err != nil {
		reportSaveError(state, err, path)
		return
	}

	state.fileWatcher.Stop()
	state.fileWatcher = newWatcher
	state.hasUnsavedChanges = false
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
	if state.hasUnsavedChanges {
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
