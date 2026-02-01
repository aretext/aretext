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

	f(s, mt.Backend())
}

func assertCellContents(t *testing.T, s vt.MockBackend, expectedContents [][]string) {
	size := s.GetSize()
	require.Equal(t, len(expectedContents), size.Y)
	require.Equal(t, len(expectedContents[0]), size.X)
	for y := vt.Row(0); y <  size.Y; y++ {
		for x := vt.Col(0); x <  size.X; x++ {
			actual := s.GetCell(vt.Coord{X:x, Y:y}).C
			expected := expectedContents[y][x]
			assert.Equal(t, expected, actual, "Wrong contents at (%d, %d), expected %q but got %q", x, y, expected, actual)
		}
	}
}

func assertCellStyles(t *testing.T, s vt.MockBackend, expectedStyles [][]tcell.Style) {
	size := s.GetSize()
	require.Equal(t, len(expectedStyles), size.Y)
	require.Equal(t, len(expectedStyles[0]), size.X)
	for y := vt.Row(0); y < size.Y; y++ {
		for x := vt.Col(0); x < size.X; x++ {
			actualStyle := s.GetCell(vt.Coord{X:x, Y:y}).S
			expectedStyle := expectedStyles[y][x]
			assert.Equal(t, expectedStyle, actualStyle, "Wrong style at (%d, %d)", x, y)
		}
	}
}

func TestScreenRegionPut(t *testing.T) {
	withMockScreen(t, vt.MockOptSize{X:10, Y:10}, func(s tcell.Screen, b vt.MockBackend) {
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
	withMockScreen(t, vt.MockOptSize{X:10, Y: 10}, func(s tcell.Screen, b vt.MockBackend) {
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
	withMockScreen(t, vt.MockOptSize{X:10, Y: 10}, func(s tcell.Screen, b vt.MockBackend) {
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
	withMockScreen(t, vt.MockOptSize{X:10, Y:10}, func(s tcell.Screen, b vt.MockBackend) {
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
