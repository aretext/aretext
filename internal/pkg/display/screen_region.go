package display

import (
	"github.com/gdamore/tcell"
)

// ScreenRegion draws to a rectangular region in a screen.
type ScreenRegion struct {
	screen              tcell.Screen
	x, y, width, height int
}

// NewScreenRegion defines a new rectangular region within a screen.
// If the dimensions are invalid (for example, if the width of the region is greater than the width of the screen),
// then this will panic.
func NewScreenRegion(screen tcell.Screen, x, y, width, height int) *ScreenRegion {
	r := &ScreenRegion{screen, x, y, width, height}
	r.validateDimensions()
	return r
}

// Resize changes the size of the region.
// This will panic if the new width/height are invalid.
func (r *ScreenRegion) Resize(newWidth, newHeight int) {
	r.width = newWidth
	r.height = newHeight
	r.validateDimensions()
}

func (r *ScreenRegion) validateDimensions() {
	screenWidth, screenHeight := r.screen.Size()
	if r.width > screenWidth || r.height > screenHeight || r.x >= screenWidth+r.width || r.y >= screenHeight+r.height {
		panic("Invalid dimensions for screen region")
	}
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
// Attempts to set content outside the region are ignored.
func (r *ScreenRegion) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) {
	if x < 0 || y < 0 || x >= r.width || y >= r.height {
		return
	}

	r.screen.SetContent(x+r.x, y+r.y, mainc, combc, style)
}

// Size returns the width and height of the screen region.
func (r *ScreenRegion) Size() (width int, height int) {
	return r.width, r.height
}
