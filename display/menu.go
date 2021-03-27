package display

import (
	"github.com/aretext/aretext/exec"
	"github.com/aretext/aretext/menu"
	"github.com/gdamore/tcell/v2"
)

// DrawMenu draws the menu at the top of the screen.
func DrawMenu(screen tcell.Screen, menu *exec.MenuState) {
	if !menu.Visible() {
		return
	}

	screenWidth, screenHeight := screen.Size()
	if screenHeight == 0 || screenWidth == 0 {
		return
	}

	// Leave one line at the bottom for the status bar.
	height := screenHeight - 1

	// Search input
	row := 0
	searchInputRegion := NewScreenRegion(screen, 0, row, screenWidth, 1)
	drawSearchInput(searchInputRegion, menu.Prompt(), menu.SearchQuery())
	row++

	// Filtered menu items (search results)
	items, selectedIdx := menu.SearchResults()
	items, selectedIdx = filterForVisibleItems(items, selectedIdx, height)
	for i := 0; i < len(items) && row < height; i++ {
		menuItemRegion := NewScreenRegion(screen, 0, row, screenWidth, 1)
		isSelected := i == selectedIdx
		drawMenuItem(menuItemRegion, items[i], isSelected)
		row++
	}

	// Bottom border
	if row < height {
		borderRegion := NewScreenRegion(screen, 0, row, screenWidth, 1)
		borderRegion.Fill(tcell.RuneHLine, tcell.StyleDefault.Dim(true))
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

func drawSearchInput(sr *ScreenRegion, prompt string, query string) {
	sr.Clear()
	sr.SetContent(0, 0, ':', nil, tcell.StyleDefault)
	col := 1
	if len(query) == 0 {
		sr.ShowCursor(col, 0)
		drawStringNoWrap(sr, prompt, col, 0, tcell.StyleDefault.Dim(true))
		return
	}

	col = drawStringNoWrap(sr, query, col, 0, tcell.StyleDefault)
	sr.ShowCursor(col, 0)
}

func drawMenuItem(sr *ScreenRegion, item menu.Item, selected bool) {
	sr.Clear()

	col := 2
	if selected {
		sr.SetContent(col, 0, '>', nil, tcell.StyleDefault.Bold(true))
	}
	col += 2

	style := tcell.StyleDefault
	if selected {
		style = style.Underline(true)
	}
	drawStringNoWrap(sr, item.Name, col, 0, style)
}
