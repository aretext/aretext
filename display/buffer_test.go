package display

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/selection"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/syntax"
)

func drawBuffer(t *testing.T, screen tcell.Screen, setupState func(*state.EditorState)) {
	screenWidth, screenHeight := screen.Size()
	editorState := state.NewEditorState(uint64(screenWidth), uint64(screenHeight+1), nil, nil)
	setupState(editorState)
	palette := NewPalette()
	buffer := editorState.DocumentBuffer()
	DrawBuffer(screen, palette, buffer)
	screen.Sync()
}

func TestDrawBuffer(t *testing.T) {
	testCases := []struct {
		name             string
		inputString      string
		expectedContents [][]rune
	}{
		{
			name:        "empty",
			inputString: "",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "short line",
			inputString: "abc",
			expectedContents: [][]rune{
				{'a', 'b', 'c', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "wrapping line",
			inputString: "abcdefghijklmnopqrstuvwxyz",
			expectedContents: [][]rune{
				{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
				{'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't'},
				{'u', 'v', 'w', 'x', 'y', 'z', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "newline",
			inputString: "abc\ndefghi\njkl",
			expectedContents: [][]rune{
				{'a', 'b', 'c', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'d', 'e', 'f', 'g', 'h', 'i', ' ', ' ', ' ', ' '},
				{'j', 'k', 'l', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "carriage return and line feed",
			inputString: "abc\r\ndef",
			expectedContents: [][]rune{
				{'a', 'b', 'c', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{'d', 'e', 'f', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "full-width characters, no wrapping",
			inputString: "abc界xyz",
			expectedContents: [][]rune{
				{'a', 'b', 'c', '界', 'X', 'x', 'y', 'z', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "full-width characters wrapped at end of line",
			inputString: "abcdefghi界jklmn",
			expectedContents: [][]rune{
				{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', ' '},
				{'界', 'X', 'j', 'k', 'l', 'm', 'n', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "trademark character occupies one cell",
			inputString: "™,",
			expectedContents: [][]rune{
				{'™', ',', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "angle brackets are displayed",
			inputString: "⟦A⟧ ⇔ ⟪B⟫",
			expectedContents: [][]rune{
				{'⟦', 'A', '⟧', ' ', '⇔', ' ', '⟪', 'B', '⟫', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(10, 10)
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
				})
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}

func TestDrawBufferCarriageReturnAndLineFeedNotRendered(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(5, 2)
		drawBuffer(t, s, func(editorState *state.EditorState) {
			state.InsertRune(editorState, '\r')
			state.InsertRune(editorState, '\n')
		})

		// Every rune, including combining runes, should be empty.
		// This detects a bug where tcell would write the combining
		// char ('\n') to the terminal, which caused
		// strange display artifacts.
		cells, width, height := s.GetContents()
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				for _, r := range cells[x+y*width].Runes {
					assert.Equal(t, ' ', r)
				}
			}
		}
	})
}

func TestGraphemeClustersWithMultipleRunes(t *testing.T) {
	testCases := []struct {
		name              string
		inputString       string
		expectedCellRunes [][]rune
	}{
		{
			name:        "ascii",
			inputString: "abcd1234",
			expectedCellRunes: [][]rune{
				{'a'}, {'b'}, {'c'}, {'d'}, {'1'}, {'2'}, {'3'}, {'4'},
			},
		},
		{
			name:        "thai",
			inputString: "\u0E04\u0E49\u0E33",
			expectedCellRunes: [][]rune{
				{'\u0E04', '\u0E49'}, {'\u0E33'},
			},
		},
		{
			name:        "emoji with zero-width joiner",
			inputString: "\U0001f9db\u200d\u2640\U0001f469\u200d\U0001f467\u200d\U0001f467",
			expectedCellRunes: [][]rune{
				{'\U0001f9db', '\u200d', '\u2640'},
				{'X'},
				{'\U0001f469', '\u200d', '\U0001f467', '\u200d', '\U0001f467'},
				{'X'},
			},
		},
		{
			name:        "regional indicator",
			inputString: "\U0001f1fa\U0001f1f8 (usa!)",
			expectedCellRunes: [][]rune{
				{'\U0001f1fa', '\U0001f1f8'},
				{' '}, {'('}, {'u'}, {'s'}, {'a'}, {'!'}, {')'},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(100, 1)
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
				})
				contents, _, _ := s.GetContents()
				for i, expectedRunes := range tc.expectedCellRunes {
					assert.Equal(t, expectedRunes, contents[i].Runes)
				}
			})
		})
	}
}

func TestDrawBufferSizeTooSmall(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 4)
		drawBuffer(t, s, func(editorState *state.EditorState) {
			for _, r := range "ab界cd" {
				state.InsertRune(editorState, r)
			}
		})
		assertCellContents(t, s, [][]rune{
			{'a'},
			{'b'},
			{' '},
			{'c'},
		})
	})
}

func TestDrawBufferCursor(t *testing.T) {
	testCases := []struct {
		name                  string
		inputString           string
		cursorPosition        uint64
		expectedCursorVisible bool
		expectedCursorCol     int
		expectedCursorRow     int
	}{
		{
			name:                  "empty",
			inputString:           "",
			cursorPosition:        0,
			expectedCursorVisible: true,
			expectedCursorCol:     0,
			expectedCursorRow:     0,
		},
		{
			name:                  "single character",
			inputString:           "a",
			cursorPosition:        0,
			expectedCursorVisible: true,
			expectedCursorCol:     0,
			expectedCursorRow:     0,
		},
		{
			name:                  "single character, past end of line",
			inputString:           "a",
			cursorPosition:        1,
			expectedCursorVisible: true,
			expectedCursorCol:     1,
			expectedCursorRow:     0,
		},
		{
			name:                  "multiple characters, within line",
			inputString:           "abcde",
			cursorPosition:        3,
			expectedCursorVisible: true,
			expectedCursorCol:     3,
			expectedCursorRow:     0,
		},
		{
			name:                  "multiple characters, at end of line",
			inputString:           "abcde",
			cursorPosition:        4,
			expectedCursorVisible: true,
			expectedCursorCol:     4,
			expectedCursorRow:     0,
		},
		{
			name:                  "multiple characters, past end of wrapped line",
			inputString:           "abcdefghijkl",
			cursorPosition:        5,
			expectedCursorVisible: true,
			expectedCursorCol:     0,
			expectedCursorRow:     1,
		},
		{
			name:                  "multiple characters, on newline",
			inputString:           "ab\ncdefghijkl",
			cursorPosition:        2,
			expectedCursorVisible: true,
			expectedCursorCol:     2,
			expectedCursorRow:     0,
		},
		{
			name:                  "multiple characters, past newline",
			inputString:           "ab\ncdefghijkl",
			cursorPosition:        3,
			expectedCursorVisible: true,
			expectedCursorCol:     0,
			expectedCursorRow:     1,
		},
		{
			name:                  "single newline, on newline",
			inputString:           "\n",
			cursorPosition:        0,
			expectedCursorVisible: true,
			expectedCursorCol:     0,
			expectedCursorRow:     0,
		},
		{
			name:                  "single newline, past newline",
			inputString:           "\n",
			cursorPosition:        1,
			expectedCursorVisible: true,
			expectedCursorCol:     0,
			expectedCursorRow:     1,
		},
		{
			name:                  "multiple lines, at end of file",
			inputString:           "ab\ncdefghi\njkl",
			cursorPosition:        13,
			expectedCursorVisible: true,
			expectedCursorCol:     2,
			expectedCursorRow:     3,
		},
		{
			name:                  "multiple lines, past end of file",
			inputString:           "ab\ncdefghi\njkl",
			cursorPosition:        14,
			expectedCursorVisible: true,
			expectedCursorCol:     3,
			expectedCursorRow:     3,
		},
		{
			name:                  "cursor past end of screen",
			inputString:           "abcdefghijklmnopqrstuvwxyz",
			cursorPosition:        26,
			expectedCursorVisible: false,
		},
		{
			name:                  "cursor past end of screen width",
			inputString:           "abcde",
			cursorPosition:        5,
			expectedCursorVisible: true,
			expectedCursorCol:     0,
			expectedCursorRow:     1,
		},
		{
			name:                  "cursor position equal to max line width",
			inputString:           "ab\ncd",
			cursorPosition:        5,
			expectedCursorVisible: true,
			expectedCursorCol:     2,
			expectedCursorRow:     1,
		},
		{
			name:                  "cursor position at end of line ending on newline",
			inputString:           "abcde\nfg",
			cursorPosition:        5,
			expectedCursorVisible: true,
			expectedCursorCol:     0,
			expectedCursorRow:     1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(5, 5)
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
					state.MoveCursor(editorState, func(state.LocatorParams) uint64 {
						return tc.cursorPosition
					})
				})
				cursorCol, cursorRow, cursorVisible := s.GetCursor()
				assert.Equal(t, tc.expectedCursorVisible, cursorVisible)
				if tc.expectedCursorVisible {
					assert.Equal(t, tc.expectedCursorCol, cursorCol)
					assert.Equal(t, tc.expectedCursorRow, cursorRow)
				}
			})
		})
	}
}

func TestSyntaxHighlighting(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(18, 1)
		drawBuffer(t, s, func(editorState *state.EditorState) {
			state.SetSyntax(editorState, syntax.LanguageGo)
			for _, r := range `const foo = "test"` {
				state.InsertRune(editorState, r)
			}
		})
		assertCellStyles(t, s, [][]tcell.Style{
			{
				// `const` highlighted as keyword.
				tcell.StyleDefault.Foreground(tcell.ColorOlive),
				tcell.StyleDefault.Foreground(tcell.ColorOlive),
				tcell.StyleDefault.Foreground(tcell.ColorOlive),
				tcell.StyleDefault.Foreground(tcell.ColorOlive),
				tcell.StyleDefault.Foreground(tcell.ColorOlive),

				// ` foo ` has no highlighting.
				tcell.StyleDefault,
				tcell.StyleDefault,
				tcell.StyleDefault,
				tcell.StyleDefault,
				tcell.StyleDefault,

				// `=` highlighted as an operator.
				tcell.StyleDefault.Foreground(tcell.ColorPurple),

				// ` ` has no highlighting.
				tcell.StyleDefault,

				// `"test"` highlighted as a string.
				tcell.StyleDefault.Foreground(tcell.ColorMaroon),
				tcell.StyleDefault.Foreground(tcell.ColorMaroon),
				tcell.StyleDefault.Foreground(tcell.ColorMaroon),
				tcell.StyleDefault.Foreground(tcell.ColorMaroon),
				tcell.StyleDefault.Foreground(tcell.ColorMaroon),
				tcell.StyleDefault.Foreground(tcell.ColorMaroon),
			},
		})
	})
}

func TestSearchMatch(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(12, 1)
		query := "d1"
		drawBuffer(t, s, func(editorState *state.EditorState) {
			for _, r := range `abcd1234` {
				state.InsertRune(editorState, r)
			}
			state.MoveCursor(editorState, func(state.LocatorParams) uint64 { return 0 })
			state.StartSearch(editorState, state.SearchDirectionForward)
			for _, r := range query {
				state.AppendRuneToSearchQuery(editorState, r)
			}
		})
		assertCellStyles(t, s, [][]tcell.Style{
			{
				tcell.StyleDefault,
				tcell.StyleDefault,
				tcell.StyleDefault,
				tcell.StyleDefault.Reverse(true),
				tcell.StyleDefault.Reverse(true),
				tcell.StyleDefault,
				tcell.StyleDefault,
				tcell.StyleDefault,
				tcell.StyleDefault,
				tcell.StyleDefault,
				tcell.StyleDefault,
				tcell.StyleDefault,
			},
		})
	})
}

func TestSelection(t *testing.T) {
	testCases := []struct {
		name              string
		width, height     int
		inputString       string
		selectionStartPos uint64
		selectionEndPos   uint64
		expectedStyles    [][]tcell.Style
	}{
		{
			name:              "selection within line",
			width:             5,
			height:            1,
			inputString:       "abcd",
			selectionStartPos: 1,
			selectionEndPos:   2,
			expectedStyles: [][]tcell.Style{
				{
					tcell.StyleDefault,
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault,
					tcell.StyleDefault,
				},
			},
		},
		{
			name:              "selection with line ending",
			width:             5,
			height:            2,
			inputString:       "abc\nefg",
			selectionStartPos: 1,
			selectionEndPos:   5,
			expectedStyles: [][]tcell.Style{
				{
					tcell.StyleDefault,
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault,
				},
				{
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault,
					tcell.StyleDefault,
					tcell.StyleDefault,
				},
			},
		},
		{
			name:              "selection on empty line",
			width:             3,
			height:            3,
			inputString:       "a\n\nb",
			selectionStartPos: 2,
			selectionEndPos:   2,
			expectedStyles: [][]tcell.Style{
				{
					tcell.StyleDefault,
					tcell.StyleDefault,
					tcell.StyleDefault,
				},
				{
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault,
					tcell.StyleDefault,
				},
				{
					tcell.StyleDefault,
					tcell.StyleDefault,
					tcell.StyleDefault,
				},
			},
		},
		{
			name:              "selection with tab",
			width:             8,
			height:            1,
			inputString:       "ab\tbc",
			selectionStartPos: 1,
			selectionEndPos:   3,
			expectedStyles: [][]tcell.Style{
				{
					tcell.StyleDefault,
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault.Reverse(true).Dim(true),
					tcell.StyleDefault,
					tcell.StyleDefault,
					tcell.StyleDefault,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(tc.width, tc.height)
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
					state.MoveCursor(editorState, func(state.LocatorParams) uint64 {
						return tc.selectionStartPos
					})
					state.ToggleVisualMode(editorState, selection.ModeChar)
					state.MoveCursor(editorState, func(state.LocatorParams) uint64 {
						return tc.selectionEndPos
					})
				})
				assertCellStyles(t, s, tc.expectedStyles)
			})
		})
	}
}

func TestShowLineNumbers(t *testing.T) {
	testCases := []struct {
		name             string
		width, height    int
		inputString      string
		expectedContents [][]rune
	}{
		{
			name:        "empty",
			width:       5,
			height:      5,
			inputString: "",
			expectedContents: [][]rune{
				{' ', '1', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "single line",
			width:       5,
			height:      5,
			inputString: "ab",
			expectedContents: [][]rune{
				{' ', '1', ' ', 'a', 'b'},
				{' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "single line, soft-wrapped",
			width:       5,
			height:      5,
			inputString: "abcde",
			expectedContents: [][]rune{
				{' ', '1', ' ', 'a', 'b'},
				{' ', ' ', ' ', 'c', 'd'},
				{' ', ' ', ' ', 'e', ' '},
				{' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "multiple lines",
			width:       5,
			height:      5,
			inputString: "ab\nc\nde",
			expectedContents: [][]rune{
				{' ', '1', ' ', 'a', 'b'},
				{' ', '2', ' ', 'c', ' '},
				{' ', '3', ' ', 'd', 'e'},
				{' ', ' ', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "multiple lines, last line empty",
			width:       5,
			height:      5,
			inputString: "ab\nc\nd\n",
			expectedContents: [][]rune{
				{' ', '1', ' ', 'a', 'b'},
				{' ', '2', ' ', 'c', ' '},
				{' ', '3', ' ', 'd', ' '},
				{' ', '4', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "multiple lines, last line empty with newline at view width",
			width:       5,
			height:      5,
			inputString: "ab\nc\nde\n",
			expectedContents: [][]rune{
				{' ', '1', ' ', 'a', 'b'},
				{' ', '2', ' ', 'c', ' '},
				{' ', '3', ' ', 'd', 'e'},
				{' ', '4', ' ', ' ', ' '},
				{' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "collapsed margin",
			width:       3,
			height:      5,
			inputString: "ab\ncd",
			expectedContents: [][]rune{
				{'a', 'b', ' '},
				{'c', 'd', ' '},
				{' ', ' ', ' '},
				{' ', ' ', ' '},
				{' ', ' ', ' '},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(tc.width, tc.height)
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
					state.ToggleShowLineNumbers(editorState)
				})
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}

func TestShowTabs(t *testing.T) {
	testCases := []struct {
		name             string
		width, height    int
		showTabs         bool
		inputString      string
		expectedContents [][]rune
	}{
		{
			name:        "hide tabs",
			width:       8,
			height:      2,
			showTabs:    false,
			inputString: "\ta\t\nb\t",
			expectedContents: [][]rune{
				{' ', ' ', ' ', ' ', 'a', ' ', ' ', ' '},
				{'b', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
		{
			name:        "show tabs",
			width:       8,
			height:      2,
			showTabs:    true,
			inputString: "\ta\t\nb\t",
			expectedContents: [][]rune{
				{tcell.RuneRArrow, ' ', ' ', ' ', 'a', tcell.RuneRArrow, ' ', ' '},
				{'b', tcell.RuneRArrow, ' ', ' ', ' ', ' ', ' ', ' '},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withSimScreen(t, func(s tcell.SimulationScreen) {
				s.SetSize(tc.width, tc.height)
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
					if tc.showTabs {
						state.ToggleShowTabs(editorState)
					}
				})
				assertCellContents(t, s, tc.expectedContents)
			})
		})
	}
}

func TestShowTabsWithSelectionStyle(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(8, 1)
		drawBuffer(t, s, func(editorState *state.EditorState) {
			state.InsertRune(editorState, '\t')
			state.InsertRune(editorState, 'a')
			state.InsertRune(editorState, 'b')
			state.InsertRune(editorState, 'c')
			state.ToggleShowTabs(editorState)
			state.MoveCursor(editorState, func(state.LocatorParams) uint64 {
				return 0
			})
			state.ToggleVisualMode(editorState, selection.ModeChar)
			state.MoveCursor(editorState, func(state.LocatorParams) uint64 {
				return 2
			})
		})
		assertCellContents(t, s, [][]rune{
			{tcell.RuneRArrow, ' ', ' ', ' ', 'a', 'b', 'c', ' '},
		})
		assertCellStyles(t, s, [][]tcell.Style{
			{
				tcell.StyleDefault.Reverse(true).Dim(true),
				tcell.StyleDefault.Reverse(true).Dim(true),
				tcell.StyleDefault.Reverse(true).Dim(true),
				tcell.StyleDefault.Reverse(true).Dim(true),
				tcell.StyleDefault.Reverse(true).Dim(true),
				tcell.StyleDefault.Reverse(true).Dim(true),
				tcell.StyleDefault,
				tcell.StyleDefault,
			},
		})
	})
}
