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

// Clear resets a rectangular region of the screen to its initial state.
func (r *ScreenRegion) Clear() {
	r.Fill(' ', tcell.StyleDefault)
}

// Fill fills a rectangular region of the screen with a character.
func (r *ScreenRegion) Fill(c rune, style tcell.Style) {
	str := string(c)
	x, y := 0, 0
	for {
		_, width := r.Put(x, y, str, style)
		x += width
		if x >= r.width {
			x = 0
			y++
		}

		if y >= r.height {
			break
		}
	}
}

// Put outputs a single grapheme cluster to the screen.
// The x and y coordinates are relative to the origin of the region.
// Attempts to set content outside the region or screen are ignored.
func (r *ScreenRegion) Put(x int, y int, str string, style tcell.Style) (remain string, width int) {
	if x < 0 || y < 0 || x >= r.width || y >= r.height {
		return
	}

	return r.screen.Put(x+r.x, y+r.y, str, style)
}

// PutStrStyled prints the string clipped to the screen region without wrapping.
func (r *ScreenRegion) PutStrStyled(x int, y int, str string, style tcell.Style) int {
	width := 0
	for str != "" && x < r.width && y < r.height {
		str, width = r.Put(x, y, str, style)
		if width == 0 {
			break
		}
		x += width
	}

	return x
}

// SetStyle sets the style of a cell without changing its content.
// Attempts to set content outside the region or screen are ignored.
func (r *ScreenRegion) SetStyleInCell(x, y int, style tcell.Style) {
	if x < 0 || y < 0 || x >= r.width || y >= r.height {
		return
	}

	str, _, _ := r.screen.Get(r.x+x, r.y+y)
	r.screen.Put(r.x+x, r.y+y, str, style)
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
