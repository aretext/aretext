package display

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScreenRegionPut(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 10)
		r := NewScreenRegion(s, 1, 2, 5, 5)

		// Inside the region, at each corner
		r.Put(0, 0, "a", tcell.StyleDefault)
		r.Put(4, 0, "b", tcell.StyleDefault)
		r.Put(4, 4, "c", tcell.StyleDefault)
		r.Put(0, 4, "d", tcell.StyleDefault)

		// Outside of the region
		r.Put(-1, -1, "x", tcell.StyleDefault)
		r.Put(5, 5, "y", tcell.StyleDefault)

		// Redraw
		s.Sync()

		// Check that only the contents in the region are displayed
		assertCellContents(t, s, [][]string{
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", "a", " ", " ", " ", "b", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", "d", " ", " ", " ", "c", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
		})
	})
}

func TestScreenRegionPutStrStyled(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 10)
		r := NewScreenRegion(s, 1, 2, 5, 5)

		// Top of region, clipped
		r.PutStrStyled(0, 0, "hello world", tcell.StyleDefault)

		// Bottom of region, clipped
		r.PutStrStyled(2, 2, "goodbye", tcell.StyleDefault)

		// Redraw
		s.Sync()

		// Check that only the contents in the region are displayed
		assertCellContents(t, s, [][]string{
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", "h", "e", "l", "l", "o", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", "g", "o", "o", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
		})
	})
}

func TestScreenRegionClear(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 10)
		s.Fill('~', tcell.StyleDefault.Bold(true))
		r := NewScreenRegion(s, 1, 2, 5, 5)
		r.Clear()
		s.Sync()

		assertCellContents(t, s, [][]string{
			{"~", "~", "~", "~", "~", "~", "~", "~", "~", "~"},
			{"~", "~", "~", "~", "~", "~", "~", "~", "~", "~"},
			{"~", " ", " ", " ", " ", " ", "~", "~", "~", "~"},
			{"~", " ", " ", " ", " ", " ", "~", "~", "~", "~"},
			{"~", " ", " ", " ", " ", " ", "~", "~", "~", "~"},
			{"~", " ", " ", " ", " ", " ", "~", "~", "~", "~"},
			{"~", " ", " ", " ", " ", " ", "~", "~", "~", "~"},
			{"~", "~", "~", "~", "~", "~", "~", "~", "~", "~"},
			{"~", "~", "~", "~", "~", "~", "~", "~", "~", "~"},
			{"~", "~", "~", "~", "~", "~", "~", "~", "~", "~"},
		})
	})
}

func TestScreenRegionFill(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 10)
		r := NewScreenRegion(s, 1, 2, 5, 5)
		r.Fill('^', tcell.StyleDefault.Bold(true))
		s.Sync()

		assertCellContents(t, s, [][]string{
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", "^", "^", "^", "^", "^", " ", " ", " ", " "},
			{" ", "^", "^", "^", "^", "^", " ", " ", " ", " "},
			{" ", "^", "^", "^", "^", "^", " ", " ", " ", " "},
			{" ", "^", "^", "^", "^", "^", " ", " ", " ", " "},
			{" ", "^", "^", "^", "^", "^", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
			{" ", " ", " ", " ", " ", " ", " ", " ", " ", " "},
		})
	})
}
