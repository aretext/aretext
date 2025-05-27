package state

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/shellcmd"
	"github.com/aretext/aretext/text"
)

// SuspendScreenFunc suspends the screen, executes a function, then resumes the screen.
// This allows the shell command to take control of the terminal.
type SuspendScreenFunc func(func() error) error

// RunShellCmd executes the command in a shell.
// Mode must be a valid command mode, as defined in config.
// All modes run as an asynchronous task that the user can cancel,
// except for CmdModeTerminal which takes over stdin/stdout.
func RunShellCmd(state *EditorState, shellCmd string, mode string) {
	log.Printf("Running shell command: %q\n", shellCmd)

	env := envVars(state) // Read-only copy of env vars is safe to pass to other goroutines.

	switch mode {
	case config.CmdModeTerminal:
		// Run synchronously because the command takes over stdin/stdout.
		ctx := context.Background()
		err := state.suspendScreenFunc(func() error {
			return shellcmd.RunInTerminal(ctx, shellCmd, env)
		})
		setStatusForShellCmdResult(state, err)

	case config.CmdModeSilent:
		StartTask(state, func(ctx context.Context) func(*EditorState) {
			err := shellcmd.RunSilent(ctx, shellCmd, env)
			return func(state *EditorState) {
				setStatusForShellCmdResult(state, err)
			}
		})

	case config.CmdModeInsert:
		StartTask(state, func(ctx context.Context) func(*EditorState) {
			output, err := shellcmd.RunAndCaptureOutput(ctx, shellCmd, env)
			return func(state *EditorState) {
				if err == nil {
					insertShellCmdOutput(state, output)
				}
				setStatusForShellCmdResult(state, err)
			}
		})

	case config.CmdModeInsertChoice:
		StartTask(state, func(ctx context.Context) func(*EditorState) {
			output, err := shellcmd.RunAndCaptureOutput(ctx, shellCmd, env)
			return func(state *EditorState) {
				if err == nil {
					err = showInsertChoiceMenuForShellCmdOutput(state, output)
				}
				setStatusForShellCmdResult(state, err)
			}
		})

	case config.CmdModeFileLocations:
		StartTask(state, func(ctx context.Context) func(*EditorState) {
			output, err := shellcmd.RunAndCaptureOutput(ctx, shellCmd, env)
			return func(state *EditorState) {
				if err == nil {
					err = showFileLocationsMenuForShellCmdOutput(state, output)
				}
				setStatusForShellCmdResult(state, err)
			}
		})

	case config.CmdModeWorkingDir:
		StartTask(state, func(ctx context.Context) func(*EditorState) {
			output, err := shellcmd.RunAndCaptureOutput(ctx, shellCmd, env)
			return func(state *EditorState) {
				if err == nil {
					err = showWorkingDirMenuForShellCmdOutput(state, output)
				}
				setStatusForShellCmdResult(state, err)
			}
		})

	default:
		// This should never happen because the config validates the mode.
		panic("Unrecognized shell cmd mode")
	}
}

func setStatusForShellCmdResult(state *EditorState, err error) {
	if err != nil {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  fmt.Sprintf("Shell command failed: %s", err),
		})
		return
	}

	SetStatusMsg(state, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  "Shell command completed successfully",
	})
}

func envVars(state *EditorState) []string {
	env := os.Environ()

	// $FILEPATH is the path to the current file.
	filePath := state.fileWatcher.Path()
	env = append(env, fmt.Sprintf("FILEPATH=%s", filePath))

	// $WORD is the current word under the cursor (excluding whitespace).
	currentWord := currentWordEnvVar(state)
	env = append(env, fmt.Sprintf("WORD=%s", currentWord))

	// $LINE is the line number of the cursor, starting from one.
	// $COLUMN is the column position of the cursor in bytes, starting from one.
	lineNum, columnNum := lineAndColumnEnvVars(state)
	env = append(env,
		fmt.Sprintf("LINE=%d", lineNum),
		fmt.Sprintf("COLUMN=%d", columnNum))

	// $SELECTION is the current visual mode selection, if any.
	selection, _ := copySelectionText(state.documentBuffer)
	if len(selection) > 0 {
		env = append(env, fmt.Sprintf("SELECTION=%s", selection))
	}

	return env
}

func currentWordEnvVar(state *EditorState) string {
	buffer := state.documentBuffer
	textTree := buffer.textTree
	cursorPos := buffer.cursor.position
	wordStartPos, wordEndPos := locate.InnerWordObject(textTree, cursorPos, 1)
	word := copyText(textTree, wordStartPos, wordEndPos-wordStartPos)
	return strings.TrimSpace(word)
}

func lineAndColumnEnvVars(state *EditorState) (uint64, uint64) {
	buffer := state.documentBuffer
	textTree := buffer.textTree
	cursorPos := buffer.cursor.position
	lineNum := textTree.LineNumForPosition(cursorPos)
	startOfLinePos := textTree.LineStartPosition(lineNum)
	columnNum := countBytesBetweenPositions(textTree, startOfLinePos, cursorPos)
	// convert 0-indexed to 1-indexed
	return lineNum + 1, columnNum + 1
}

func countBytesBetweenPositions(textTree *text.Tree, startPos, endPos uint64) uint64 {
	var byteCount uint64
	reader := textTree.ReaderAtPosition(startPos)
	for i := startPos; i < endPos; i++ {
		_, numBytes, err := reader.ReadRune()
		if err != nil {
			break
		}
		byteCount += uint64(numBytes)
	}
	return byteCount
}

func insertShellCmdOutput(state *EditorState, shellCmdOutput string) {
	BeginUndoEntry(state)

	tree := state.documentBuffer.textTree
	pos := state.documentBuffer.cursor.position
	if state.documentBuffer.selector.Mode() == selection.ModeNone {
		// If there isn't any selection, then the text should be inserted
		// immediately *after* the cursor position.
		//
		// This is more intuitive when inserting text at the end of a line,
		// Escape to normal mode will place the cursor on the last
		// character of the line. Executing a shell cmd to insert text
		// should insert after the last character in the line, not before.
		pos = locate.NextCharInLine(tree, 1, true, pos)
	} else {
		// If there is a selection, delete it before inserting.
		// Deleting the selection moves the cursor, so insert text at the updated
		// cursor position.
		deleteCurrentSelection(state)
		pos = state.documentBuffer.cursor.position
	}

	insertTextAtPosition(state, shellCmdOutput, pos, true)

	posAfterInsert := pos + uint64(utf8.RuneCountInString(shellCmdOutput))
	MoveCursor(state, func(params LocatorParams) uint64 {
		newPos := locate.PrevChar(params.TextTree, 1, posAfterInsert)
		return locate.ClosestCharOnLine(params.TextTree, newPos)
	})

	CommitUndoEntry(state)

	setInputMode(state, InputModeNormal)
}

func deleteCurrentSelection(state *EditorState) {
	selectionMode := state.documentBuffer.selector.Mode()
	selectedRegion := state.documentBuffer.SelectedRegion()
	MoveCursor(state, func(p LocatorParams) uint64 { return selectedRegion.StartPos })
	selectionEndLoc := func(p LocatorParams) uint64 { return selectedRegion.EndPos }
	if selectionMode == selection.ModeChar {
		DeleteToPos(state, selectionEndLoc, clipboard.PageDefault)
	} else if selectionMode == selection.ModeLine {
		DeleteLines(state, selectionEndLoc, false, true, clipboard.PageDefault)
	}
}

func showInsertChoiceMenuForShellCmdOutput(state *EditorState, shellCmdOutput string) error {
	var menuItems []menu.Item
	for _, line := range strings.Split(shellCmdOutput, "\n") {
		name := strings.TrimRight(line, "\r") // If output is CRLF, strip the CR as well.
		if len(name) == 0 {
			continue
		}
		menuItems = append(menuItems, menu.Item{
			Name: name,
			Action: func(s *EditorState) {
				insertShellCmdOutput(state, name)
			},
		})
	}

	if len(menuItems) == 0 {
		return fmt.Errorf("No lines in command output")
	}

	ShowMenu(state, MenuStyleInsertChoice, menuItems)
	return nil
}

func showWorkingDirMenuForShellCmdOutput(state *EditorState, shellCmdOutput string) error {
	var menuItems []menu.Item
	for _, line := range strings.Split(shellCmdOutput, "\n") {
		dirPath := strings.TrimRight(line, "\r") // If output is CRLF, strip the CR as well.
		if len(dirPath) == 0 {
			continue
		}

		menuItems = append(menuItems, menu.Item{
			Name: dirPath,
			Action: func(s *EditorState) {
				SetWorkingDirectory(s, dirPath)
			},
		})
	}

	if len(menuItems) == 0 {
		return fmt.Errorf("No lines in command output")
	}

	ShowMenu(state, MenuStyleWorkingDir, menuItems)
	return nil
}

func showFileLocationsMenuForShellCmdOutput(state *EditorState, shellCmdOutput string) error {
	locations, err := shellcmd.FileLocationsFromLines(strings.NewReader(shellCmdOutput))
	if err != nil {
		return err
	}

	if len(locations) == 0 {
		return fmt.Errorf("No file locations in cmd output")
	}

	menuItems, err := menuItemsFromFileLocations(locations)
	if err != nil {
		return err
	}

	ShowMenu(state, MenuStyleFileLocation, menuItems)
	return nil
}

func menuItemsFromFileLocations(locations []shellcmd.FileLocation) ([]menu.Item, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("os.Getwd: %w", err)
	}

	menuItems := make([]menu.Item, 0, len(locations))
	for _, loc := range locations {
		name := formatFileLocationName(loc)
		path := absPath(loc.Path, cwd)
		lineNum := translateFileLocationLineNum(loc.LineNum)
		menuItems = append(menuItems, menu.Item{
			Name: name,
			Action: func(s *EditorState) {
				abortMsg := "Document has unsaved changes"
				AbortIfUnsavedChanges(s, abortMsg, func(s *EditorState) {
					LoadDocument(s, path, true, func(p LocatorParams) uint64 {
						return locate.StartOfLineNum(p.TextTree, lineNum)
					})
				})
			},
		})
	}
	return menuItems, nil
}

func formatFileLocationName(loc shellcmd.FileLocation) string {
	return fmt.Sprintf("%s:%d  %s", loc.Path, loc.LineNum, loc.Snippet)
}

func absPath(p, wd string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(wd, p)
}

func translateFileLocationLineNum(lineNum uint64) uint64 {
	if lineNum > 0 {
		return lineNum - 1
	} else {
		return lineNum
	}
}
