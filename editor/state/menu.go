package state

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/aretext/aretext/editor/file"
	"github.com/aretext/aretext/editor/menu"
	"github.com/aretext/aretext/editor/text"
)

type MenuStyle int

const (
	MenuStyleCommand = MenuStyle(iota)
	MenuStyleFilePath
	MenuStyleFileLocation
	MenuStyleChildDir
	MenuStyleParentDir
	MenuStyleInsertChoice
	MenuStyleWorkingDir
)

// EmptyQueryShowAll returns whether an empty query should show all items.
func (s MenuStyle) EmptyQueryShowAll() bool {
	switch s {
	case MenuStyleFilePath, MenuStyleFileLocation, MenuStyleChildDir, MenuStyleParentDir, MenuStyleInsertChoice, MenuStyleWorkingDir:
		return true
	default:
		return false
	}
}

// MenuState represents the menu for searching and selecting items.
type MenuState struct {
	// style controls how the menu is displayed.
	style MenuStyle

	// query is the text input by the user to search for a menu item.
	query text.RuneStack

	// search controls which items are visible based on the user's current search query.
	search *menu.Search

	// selectedResultIdx is the index of the currently selected search result.
	// If there are no results, this is set to zero.
	// If there are results, this must be less than the number of results.
	selectedResultIdx int

	// prevInputMode is the input mode to set after exiting menu mode.
	prevInputMode InputMode
}

func (m *MenuState) Style() MenuStyle {
	return m.style
}

func (m *MenuState) SearchQuery() string {
	return m.query.String()
}

func (m *MenuState) SearchResults() (results []menu.Item, selectedResultIdx int) {
	if m.search == nil {
		return nil, 0
	}
	return m.search.Results(), m.selectedResultIdx
}

// ShowMenu displays the menu with the specified style and items.
func ShowMenu(state *EditorState, style MenuStyle, items []menu.Item) {
	if style == MenuStyleCommand {
		items = append(items, state.customMenuItems...)
	}

	switch style {
	case MenuStyleParentDir:
		// Sort lexicographic order descending.
		// This ensures that longer paths appear first when listing parent directory paths.
		sort.SliceStable(items, func(i, j int) bool { return items[i].Name > items[j].Name })

	case MenuStyleCommand, MenuStyleFilePath, MenuStyleChildDir:
		// Sort lexicographic order ascending.
		sort.SliceStable(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	}

	search := menu.NewSearch(items, style.EmptyQueryShowAll())
	state.menu = &MenuState{
		style:             style,
		search:            search,
		selectedResultIdx: 0,
		prevInputMode:     state.inputMode,
	}
	setInputMode(state, InputModeMenu)
}

// ShowFileMenu displays a menu for finding and loading files in the current working directory.
// The files are loaded asynchronously as a task that the user can cancel.
func ShowFileMenu(s *EditorState, hidePatterns []string) {
	log.Printf("Scheduling task to load file menu items...\n")
	StartTask(s, func(ctx context.Context) func(*EditorState) {
		log.Printf("Starting to load file menu items...\n")
		items := loadFileMenuItems(ctx, hidePatterns)
		log.Printf("Successfully loaded %d file menu items\n", len(items))
		return func(s *EditorState) {
			ShowMenu(s, MenuStyleFilePath, items)
		}
	})
}

func loadFileMenuItems(ctx context.Context, hidePatterns []string) []menu.Item {
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Error loading menu items: %v\n", fmt.Errorf("os.GetCwd: %w", err))
		return nil
	}

	paths := file.ListDir(ctx, dir, file.ListDirOptions{
		HidePatterns: hidePatterns,
	})
	log.Printf("Listed %d paths for dir %q\n", len(paths), dir)

	items := make([]menu.Item, 0, len(paths))
	for _, p := range paths {
		menuPath := p // reference path in this iteration of the loop
		items = append(items, menu.Item{
			Name: file.RelativePath(menuPath, dir),
			Action: func(s *EditorState) {
				LoadDocument(s, menuPath, true, func(LocatorParams) uint64 {
					return 0
				})
			},
		})
	}
	return items
}

// ShowChildDirsMenu displays a menu for changing the working directory to a child directory.
func ShowChildDirsMenu(s *EditorState, hidePatterns []string) {
	log.Printf("Scheduling task to load child dir menu items...\n")
	StartTask(s, func(ctx context.Context) func(*EditorState) {
		log.Printf("Starting to load child dir menu items...\n")
		items := loadChildDirMenuItems(ctx, hidePatterns)
		log.Printf("Successfully loaded %d child dir menu items\n", len(items))
		return func(s *EditorState) {
			ShowMenu(s, MenuStyleChildDir, items)
		}
	})
}

func loadChildDirMenuItems(ctx context.Context, hidePatterns []string) []menu.Item {
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Error loading menu items: %v\n", fmt.Errorf("os.GetCwd: %w", err))
		return nil
	}

	paths := file.ListDir(ctx, dir, file.ListDirOptions{
		DirectoriesOnly: true,
		HidePatterns:    hidePatterns,
	})
	log.Printf("Listed %d subdirectory paths for dir %q\n", len(paths), dir)

	items := make([]menu.Item, 0, len(paths))
	for _, p := range paths {
		menuPath := p // reference path in this iteration of the loop
		items = append(items, menu.Item{
			Name: fmt.Sprintf("./%s", file.RelativePath(menuPath, dir)),
			Action: func(s *EditorState) {
				SetWorkingDirectory(s, menuPath)
			},
		})
	}
	return items
}

// ShowParentDirsMenu displays a menu for changing the working directory to a parent directory.
func ShowParentDirsMenu(s *EditorState) {
	ShowMenu(s, MenuStyleParentDir, parentDirMenuItems())
}

func parentDirMenuItems() []menu.Item {
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Error loading menu items: %v\n", fmt.Errorf("os.GetCwd: %w", err))
		return nil
	}

	dir = filepath.Clean(dir)

	// Create an item for each parent directory.
	// We can detect when we've reached the root directory by checking the last character
	// of the path because both filepath.Clean and filepath.Dir
	// guarantee that only the root directory ends in a separator.
	var items []menu.Item
	for len(dir) > 0 && dir[len(dir)-1] != os.PathSeparator {
		dir = filepath.Dir(dir)
		menuDir := dir // reference path in this iteration of the loop
		items = append(items, menu.Item{
			Name: dir,
			Action: func(s *EditorState) {
				SetWorkingDirectory(s, menuDir)
			},
		})
	}
	return items
}

// HideMenu hides the menu.
func HideMenu(state *EditorState) {
	prevInputMode := state.menu.prevInputMode
	state.menu = &MenuState{}
	setInputMode(state, prevInputMode)
}

// ExecuteSelectedMenuItem executes the action of the selected menu item and closes the menu.
func ExecuteSelectedMenuItem(state *EditorState) {
	search := state.menu.search
	results := search.Results()
	if len(results) == 0 {
		// If there are no results, then there is no action to perform.
		SetStatusMsg(state, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  "No menu item selected",
		})
		HideMenu(state)
		return
	}

	idx := state.menu.selectedResultIdx
	selectedItem := results[idx]

	// Some menu commands enter a different input mode (like task mode for shell commands),
	// then return to whatever the input mode was at the start of the action.
	// Hide the menu first so that these actions return to normal/visual mode, not menu mode.
	HideMenu(state)

	executeMenuItemAction(state, selectedItem)
	ScrollViewToCursor(state)
}

func executeMenuItemAction(state *EditorState, item menu.Item) {
	log.Printf("Executing menu item %q\n", item.Name)
	actionFunc, ok := item.Action.(func(*EditorState))
	if !ok {
		log.Printf("Invalid action for menu item %q\n", item.Name)
		return
	}
	actionFunc(state)
}

// MoveMenuSelection moves the menu selection up or down with wraparound.
func MoveMenuSelection(state *EditorState, delta int) {
	numResults := len(state.menu.search.Results())
	if numResults == 0 {
		return
	}

	newIdx := (state.menu.selectedResultIdx + delta) % numResults
	if newIdx < 0 {
		newIdx = numResults + newIdx
	}

	state.menu.selectedResultIdx = newIdx
}

// AppendMenuSearch appends a rune to the menu search query.
func AppendRuneToMenuSearch(state *EditorState, r rune) {
	menu := state.menu
	menu.query.Push(r)
	menu.search.Execute(menu.query.String())
	menu.selectedResultIdx = 0
}

// DeleteMenuSearch deletes a rune from the menu search query.
func DeleteRuneFromMenuSearch(state *EditorState) {
	menu := state.menu
	if menu.query.Len() > 0 {
		menu.query.Pop()
		menu.search.Execute(menu.query.String())
		menu.selectedResultIdx = 0
	}
}
