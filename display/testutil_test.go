package display

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockScreen struct {
	tcell.Screen

	mt vt.MockTerm
}

func WithMockScreen(t *testing.T, f func(*MockScreen)) {
	mt := vt.NewMockTerm()
	s, err := tcell.NewTerminfoScreenFromTty(mt)
	require.NoError(err)

	err = s.Init()
	require.NoError(err)
	defer s.Fini()

	ms := MockScreen{s, mt}

	f(&ms)
}

func (ms *MockScreen) AssertCellContents(t *testing.T, expectedContents [][]string) {
	ws, err := ms.mt.WindowSize()
	require.NoError(t, err)

	// Assert expectedContents has the same dimensions as the window.
	require.Equal(t, len(expectedContents), ws.Height)
	for _, row := range expectedContents {
		require.Equal(t, len(row), ws.Width)
	}

	// Assert every cell has the expected contents.
	for y := 0; y < ws.Height; y++ {
		for x := 0; x < ws.Width; x++ {
			actual, _, _ := ms.Get(x, y)
			expected := expectedContents[y][x]
			assert.Equal(t, expected, actual, "Wrong contents at (%d, %d), expected %q but got %q", x, y, expected, actual)
		}
	}
}

func (ms *MockScreen) AssertCellStyles(t *testing.T, expectedStyles [][]tcell.Style) {
	ws, err := ms.mt.WindowSize()
	require.NoError(t, err)

	// Assert expectedStyles has the same dimensions as the window.
	require.Equal(t, len(expectedStyles), ws.Height)
	for _, row := range expectedStyles {
		require.Equal(t, len(row), ws.Width)
	}

	// Assert every cell has the expected contents.
	for y := 0; y < ws.Height; y++ {
		for x := 0; x < ws.Width; x++ {
			_, actual, _ := ms.Get(x, y)
			expected := expectedStyles[y][x]
			assert.Equal(t, expectedStyle, actualStyle, "Wrong style at (%d, %d)", x, y)
		}
	}
}

func (ms *MockScreen) AssertCursor(t *testing.T, expectVisible bool, expectCol int, expectRow int) {
	visible := ms.mt.Backend().GetCursor().IsVisible()
	assert.Equal(t, visible, expectVisible)

	if expectVisible {
		cursorPos := ms.mt.Pos()
		assert.Equal(t, cursorPos.X, vt.Col(expectCol))
		assert.Equal(t, cursorPos.Y, vt.Row(expectRow))
	}
}
