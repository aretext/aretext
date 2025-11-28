package display

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withSimScreen(t *testing.T, f func(tcell.SimulationScreen)) {
	s := tcell.NewSimulationScreen("")
	require.NotNil(t, s)
	err := s.Init()

	// Sometime between tcell v2.9 and 2.12 tcell simulation screen went from
	// " " to "X" as the default value of each cell. Restore the old behavior
	// by explicitly clearning the screen before each test.
	s.Clear()

	require.NoError(t, err)
	defer s.Fini()
	f(s)
}

func assertCellContents(t *testing.T, s tcell.SimulationScreen, expectedChars [][]rune) {
	cells, width, height := s.GetContents()
	require.Equal(t, len(expectedChars), height)
	require.Equal(t, len(expectedChars[0]), width)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			actualChar := cells[x+y*width].Runes[0]
			expectedChar := expectedChars[y][x]
			assert.Equal(t, expectedChar, actualChar, "Wrong character at (%d, %d), expected '%c' but got '%c'", x, y, expectedChar, actualChar)
		}
	}
}

func assertCellStyles(t *testing.T, s tcell.SimulationScreen, expectedStyles [][]tcell.Style) {
	cells, width, height := s.GetContents()
	require.Equal(t, height, len(expectedStyles))
	require.Equal(t, width, len(expectedStyles[0]))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			actualStyle := cells[x+y*width].Style
			expectedStyle := expectedStyles[y][x]
			assert.Equal(t, expectedStyle, actualStyle, "Wrong style at (%d, %d)", x, y)
		}
	}
}

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
		assertCellContents(t, s, [][]rune{
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', 'a', ' ', ' ', ' ', 'b', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', 'd', ' ', ' ', ' ', 'c', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
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
		assertCellContents(t, s, [][]rune{
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', 'h', 'e', 'l', 'l', 'o', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', 'g', 'o', 'o', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
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

		assertCellContents(t, s, [][]rune{
			{'~', '~', '~', '~', '~', '~', '~', '~', '~', '~'},
			{'~', '~', '~', '~', '~', '~', '~', '~', '~', '~'},
			{'~', ' ', ' ', ' ', ' ', ' ', '~', '~', '~', '~'},
			{'~', ' ', ' ', ' ', ' ', ' ', '~', '~', '~', '~'},
			{'~', ' ', ' ', ' ', ' ', ' ', '~', '~', '~', '~'},
			{'~', ' ', ' ', ' ', ' ', ' ', '~', '~', '~', '~'},
			{'~', ' ', ' ', ' ', ' ', ' ', '~', '~', '~', '~'},
			{'~', '~', '~', '~', '~', '~', '~', '~', '~', '~'},
			{'~', '~', '~', '~', '~', '~', '~', '~', '~', '~'},
			{'~', '~', '~', '~', '~', '~', '~', '~', '~', '~'},
		})
	})
}

func TestScreenRegionFill(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 10)
		r := NewScreenRegion(s, 1, 2, 5, 5)
		r.Fill('^', tcell.StyleDefault.Bold(true))
		s.Sync()

		assertCellContents(t, s, [][]rune{
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', '^', '^', '^', '^', '^', ' ', ' ', ' ', ' '},
			{' ', '^', '^', '^', '^', '^', ' ', ' ', ' ', ' '},
			{' ', '^', '^', '^', '^', '^', ' ', ' ', ' ', ' '},
			{' ', '^', '^', '^', '^', '^', ' ', ' ', ' ', ' '},
			{' ', '^', '^', '^', '^', '^', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
		})
	})
}
