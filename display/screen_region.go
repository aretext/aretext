package display

import (
	"github.com/gdamore/tcell/v2"
)

// ScreenRegion draws to a rectangular region in a screen.
type ScreenRegion struct {
	screen              tcell.Screen
	x, y, width, height int
}

// NewScreenRegion defines a new rectangular region within a screen.
func NewScreenRegion(screen tcell.Screen, x, y, width, height int) *ScreenRegion {
	return &ScreenRegion{screen, x, y, width, height}
}

// Resize changes the size of the region.
func (r *ScreenRegion) Resize(newWidth, newHeight int) {
	r.width = newWidth
	r.height = newHeight
}

// Clear resets a rectangular region of the screen to its initial state.
func (r *ScreenRegion) Clear() {
	r.Fill(' ', tcell.StyleDefault)
}

// Fill fills a rectangular region of the screen with a character.
func (r *ScreenRegion) Fill(c rune, style tcell.Style) {
	x, y := 0, 0
	for {
		r.SetContent(x, y, c, nil, style)
		x++
		if x >= r.width {
			x = 0
			y++
		}

		if y >= r.height {
			break
		}
	}
}

// SetContent sets the content of a cell in the screen region.
// The x and y coordinates are relative to the origin of the region.
// Attempts to set content outside the region or screen are ignored.
func (r *ScreenRegion) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) {
	if x < 0 || y < 0 || x >= r.width || y >= r.height {
		return
	}

	r.screen.SetContent(x+r.x, y+r.y, mainc, combc, style)
}

// GetContent returns the content of a cell in the screen region.
// The x and y coordinates are relative to the origin of the region.
// If the coordinates are out of range, zero values will be returned.
func (r *ScreenRegion) GetContent(x, y int) (mainc rune, combc []rune, style tcell.Style) {
	if x < 0 || y < 0 || x >= r.width || y >= r.height {
		return 0, nil, tcell.StyleDefault
	}

	mainc, combc, style, _ = r.screen.GetContent(x+r.x, y+r.y)
	return mainc, combc, style
}

// HideCursor prevents the cursor from being displayed.
func (r *ScreenRegion) HideCursor() {
	r.screen.ShowCursor(-1, -1)
}

// ShowCursor sets the location of the cursor on the screen.
// The x and y coordinates are relative to the origin of the region.
// If the coodinates are outside of the screen region, the cursor will be hidden.
func (r *ScreenRegion) ShowCursor(x, y int) {
	if x < 0 || y < 0 || x >= r.width || y >= r.height {
		r.HideCursor()
		return
	}

	r.screen.ShowCursor(x+r.x, y+r.y)
}

// Size returns the width and height of the screen region.
func (r *ScreenRegion) Size() (width int, height int) {
	return r.width, r.height
}
