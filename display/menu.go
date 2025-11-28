package display

import (
	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/state"
)

// DrawMenu draws the menu at the top of the screen.
func DrawMenu(screen tcell.Screen, palette *Palette, menu *state.MenuState) {
	screenWidth, screenHeight := screen.Size()
	if screenHeight == 0 || screenWidth == 0 {
		return
	}

	// Leave one line at the bottom for the status bar.
	height := screenHeight - 1

	// Search input
	row := 0
	searchInputRegion := NewScreenRegion(screen, 0, row, screenWidth, 1)
	drawSearchInput(searchInputRegion, palette, menu.Style(), menu.SearchQuery())
	row++

	// Filtered menu items (search results)
	items, selectedIdx := menu.SearchResults()
	items, selectedIdx = filterForVisibleItems(items, selectedIdx, height)
	for i := 0; i < len(items) && row < height; i++ {
		menuItemRegion := NewScreenRegion(screen, 0, row, screenWidth, 1)
		isSelected := i == selectedIdx
		drawMenuItem(menuItemRegion, palette, items[i], isSelected)
		row++
	}

	// Bottom border
	if row < height {
		borderRegion := NewScreenRegion(screen, 0, row, screenWidth, 1)
		borderRegion.Fill(tcell.RuneHLine, palette.StyleForMenuBorder())
		row++
	}
}

func filterForVisibleItems(items []menu.Item, selectedIdx int, height int) ([]menu.Item, int) {
	offset := 0
	limit := maxNumVisibleItems(len(items), height)
	if limit == 0 {
		return nil, 0
	}

	if selectedIdx >= limit {
		offset = selectedIdx - limit + 1
		selectedIdx = limit - 1
	}
	return items[offset : offset+limit], selectedIdx
}

func maxNumVisibleItems(numItems int, height int) int {
	maxVisible := height - 1 // Leave one line above for the search bar.
	if maxVisible < 0 {
		maxVisible = 0
	}
	if numItems > maxVisible {
		numItems = maxVisible
	}
	return numItems
}

func drawSearchInput(sr *ScreenRegion, palette *Palette, style state.MenuStyle, query string) {
	sr.Clear()
	col := sr.PutStrStyled(0, 0, menuIconForStyle(style), palette.StyleForMenuIcon())
	if len(query) == 0 {
		sr.ShowCursor(col, 0)
		sr.PutStrStyled(col, 0, menuPromptForStyle(style), palette.StyleForMenuPrompt())
		return
	}

	col = sr.PutStrStyled(col, 0, query, palette.StyleForMenuQuery())
	sr.ShowCursor(col, 0)
}

func menuIconForStyle(style state.MenuStyle) string {
	switch style {
	case state.MenuStyleCommand:
		return ":"
	case state.MenuStyleFilePath:
		return "./"
	case state.MenuStyleFileLocation:
		return "@"
	case state.MenuStyleInsertChoice:
		return "+ "
	case state.MenuStyleChildDir, state.MenuStyleParentDir, state.MenuStyleWorkingDir:
		return "§ "
	default:
		panic("Unrecognized menu style")
	}
}

func menuPromptForStyle(style state.MenuStyle) string {
	switch style {
	case state.MenuStyleCommand:
		return "command"
	case state.MenuStyleFilePath:
		return "file path"
	case state.MenuStyleFileLocation:
		return ""
	case state.MenuStyleInsertChoice:
		return ""
	case state.MenuStyleChildDir, state.MenuStyleParentDir, state.MenuStyleWorkingDir:
		return "working directory"
	default:
		panic("Unrecognized menu style")
	}
}

func drawMenuItem(sr *ScreenRegion, palette *Palette, item menu.Item, selected bool) {
	sr.Clear()

	col := 2
	if selected {
		sr.Put(col, 0, ">", palette.StyleForMenuCursor())
	}
	col += 2

	style := palette.StyleForMenuItem(selected)
	sr.PutStrStyled(col, 0, item.Name, style)
}
