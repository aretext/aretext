package display

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withMockScreen(t *testing.T, size vt.MockOptSize, f func(tcell.Screen, vt.MockBackend)) {
	mt := vt.NewMockTerm(size)
	s, err := tcell.NewTerminfoScreenFromTty(mt)
	require.NoError(t, err)

	err = s.Init()
	require.NoError(t, err)
	defer s.Fini()

	s.Clear()
	s.Sync()

	f(s, mt.Backend())
}

func assertCellContents(t *testing.T, s vt.MockBackend, expectedContents [][]string) {
	size := s.GetSize()
	require.Equal(t, len(expectedContents), int(size.Y))
	require.Equal(t, len(expectedContents[0]), int(size.X))
	for y := 0; y < int(size.Y); y++ {
		for x := 0; x < int(size.X); x++ {
			cell := s.GetCell(vt.Coord{X: vt.Col(x), Y: vt.Row(y)})
			actual := cell.C
			expected := expectedContents[y][x]
			assert.Equal(t, expected, actual, "Wrong contents at (%d, %d), expected %q but got %q", x, y, expected, actual)
		}
	}
}

func assertCellStyles(t *testing.T, s vt.MockBackend, expectedStyles [][]tcell.Style) {
	size := s.GetSize()
	require.Equal(t, len(expectedStyles), int(size.Y))
	require.Equal(t, len(expectedStyles[0]), int(size.X))
	for y := 0; y < int(size.Y); y++ {
		for x := 0; x < int(size.X); x++ {
			cell := s.GetCell(vt.Coord{X: vt.Col(x), Y: vt.Row(y)})
			expectedStyle := expectedStyles[y][x]
			// Compare styles by converting both to attributes and colors
			// since cell.S is vt.Style interface and expectedStyle is tcell.Style
			assert.Equal(t, expectedStyle, cell.S, "Wrong style at (%d, %d)", x, y)
		}
	}
}

func TestScreenRegionPut(t *testing.T) {
	withMockScreen(t, vt.MockOptSize{X: 10, Y: 10}, func(s tcell.Screen, b vt.MockBackend) {
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
		assertCellContents(t, b, [][]string{
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
	withMockScreen(t, vt.MockOptSize{X: 10, Y: 10}, func(s tcell.Screen, b vt.MockBackend) {
		r := NewScreenRegion(s, 1, 2, 5, 5)

		// Top of region, clipped
		r.PutStrStyled(0, 0, "hello world", tcell.StyleDefault)

		// Bottom of region, clipped
		r.PutStrStyled(2, 2, "goodbye", tcell.StyleDefault)

		// Redraw
		s.Sync()

		// Check that only the contents in the region are displayed
		assertCellContents(t, b, [][]string{
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
	withMockScreen(t, vt.MockOptSize{X: 10, Y: 10}, func(s tcell.Screen, b vt.MockBackend) {
		s.Fill('~', tcell.StyleDefault.Bold(true))
		r := NewScreenRegion(s, 1, 2, 5, 5)
		r.Clear()
		s.Sync()

		assertCellContents(t, b, [][]string{
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
	withMockScreen(t, vt.MockOptSize{X: 10, Y: 10}, func(s tcell.Screen, b vt.MockBackend) {
		r := NewScreenRegion(s, 1, 2, 5, 5)
		r.Fill('^', tcell.StyleDefault.Bold(true))
		s.Sync()

		assertCellContents(t, b, [][]string{
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
