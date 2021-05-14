package state

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/aretext/aretext/clipboard"
	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/locate"
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/shellcmd"
)

// SuspendScreenFunc suspends the screen, executes a function, then resumes the screen.
// This is allows the shell command to take control of the terminal.
type SuspendScreenFunc func(func() error) error

// RunShellCmd executes the command in a shell.
// Mode must be a valid command mode, as defined in config.
func RunShellCmd(state *EditorState, shellCmd string, mode string) {
	log.Printf("Running shell command: '%s'\n", shellCmd)
	env := envVars(state)

	var err error
	switch mode {
	case config.CmdModeSilent:
		err = shellcmd.RunSilent(shellCmd, env)
	case config.CmdModeTerminal:
		err = state.suspendScreenFunc(func() error {
			return shellcmd.RunInTerminal(shellCmd, env)
		})
	case config.CmdModeInsert:
		err = runInShellAndInsertOutput(state, shellCmd, env)
	case config.CmdModeFileLocations:
		err = runInShellAndShowFileLocationsMenu(state, shellCmd, env)
	default:
		// This should never happen because the config validates the mode.
		panic("Unrecognized shell cmd mode")
	}

	if err != nil {
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  err.Error(),
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
	tokenTree := buffer.tokenTree
	cursorPos := buffer.cursor.position
	wordStartPos := locate.CurrentWordStart(textTree, tokenTree, cursorPos)
	wordEndPos := locate.CurrentWordEnd(textTree, tokenTree, cursorPos)
	word := copyText(textTree, wordStartPos, wordEndPos-wordStartPos)
	return strings.TrimSpace(word)
}

func runInShellAndInsertOutput(state *EditorState, shellCmd string, env []string) error {
	text, err := shellcmd.RunAndCaptureOutput(shellCmd, env)
	if err != nil {
		return err
	}
	page := clipboard.PageContent{Text: text}
	state.clipboard.Set(clipboard.PageShellCmdOutput, page)
	DeleteSelection(state, false)
	PasteBeforeCursor(state, clipboard.PageShellCmdOutput)
	SetInputMode(state, InputModeNormal)
	return nil
}

func runInShellAndShowFileLocationsMenu(state *EditorState, shellCmd string, env []string) error {
	text, err := shellcmd.RunAndCaptureOutput(shellCmd, env)
	if err != nil {
		return err
	}

	locations, err := shellcmd.FileLocationsFromLines(strings.NewReader(text))
	if err != nil {
		return err
	}

	if len(locations) == 0 {
		return errors.New("No file locations in cmd output")
	}

	menuItems, err := menuItemsFromFileLocations(locations)
	if err != nil {
		return err
	}

	ShowMenu(state, MenuStyleFileLocation, func() []menu.Item {
		return menuItems
	})

	return nil
}

func menuItemsFromFileLocations(locations []shellcmd.FileLocation) ([]menu.Item, error) {
	menuItems := make([]menu.Item, 0, len(locations))
	for _, loc := range locations {
		name := formatFileLocationName(loc)
		path, err := filepath.Abs(loc.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "filepath.Abs")
		}
		lineNum := translateFileLocationLineNum(loc.LineNum)
		menuItems = append(menuItems, menu.Item{
			Name: name,
			Action: func(s *EditorState) {
				LoadDocument(s, path, true)
				MoveCursor(s, func(p LocatorParams) uint64 {
					return locate.StartOfLineNum(p.TextTree, lineNum)
				})
			},
		})
	}
	return menuItems, nil
}

func formatFileLocationName(loc shellcmd.FileLocation) string {
	return fmt.Sprintf("%s:%d  %s", loc.Path, loc.LineNum, loc.Snippet)
}

func translateFileLocationLineNum(lineNum uint64) uint64 {
	if lineNum > 0 {
		return lineNum - 1
	} else {
		return lineNum
	}
}
