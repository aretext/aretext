package display

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/config"
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
	inputMode := editorState.InputMode()
	DrawBuffer(screen, palette, buffer, inputMode)
	screen.Sync()
}

func TestDrawBuffer(t *testing.T) {
	testCases := []struct {
		name             string
		inputString      string
		expectedContents [][]string
	}{
		{
			name:        "empty",
			inputString: "",
			expectedContents: [][]string{
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "short line",
			inputString: "abc",
			expectedContents: [][]string{
				{"a", "b", "c", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "wrapping line",
			inputString: "abcdefghijklmnopqrstuvwxyz",
			expectedContents: [][]string{
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
				{"k", "l", "m", "n", "o", "p", "q", "r", "s", "t"},
				{"u", "v", "w", "x", "y", "z", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "newline",
			inputString: "abc\ndefghi\njkl",
			expectedContents: [][]string{
				{"a", "b", "c", " ", "", "", "", "", "", ""},
				{"d", "e", "f", "g", "h", "i", " ", "", "", ""},
				{"j", "k", "l", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "carriage return and line feed",
			inputString: "abc\r\ndef",
			expectedContents: [][]string{
				{"a", "b", "c", " ", "", "", "", "", "", ""},
				{"d", "e", "f", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "full-width characters, no wrapping",
			inputString: "abc界xyz",
			expectedContents: [][]string{
				{"a", "b", "c", "界", "", "x", "y", "z", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "full-width characters wrapped at end of line",
			inputString: "abcdefghi界jklmn",
			expectedContents: [][]string{
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", ""},
				{"界", "", "j", "k", "l", "m", "n", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "trademark character occupies one cell",
			inputString: "™,",
			expectedContents: [][]string{
				{"™", ",", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "angle brackets are displayed",
			inputString: "⟦A⟧ ⇔ ⟪B⟫",
			expectedContents: [][]string{
				{"⟦", "A", "⟧", " ", "⇔", " ", "⟪", "B", "⟫", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "emoji presentation selector",
			inputString: "\u2139\ufe0f abc",
			expectedContents: [][]string{
				// tcell v3 splits the emoji presentation into separate cells
				{"\u2139", "\ufe0f", " ", "a", "b", "c", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name: "space then country-flag",
			// This reproduces a bug where the grapheme cluster breaker
			// wasn't reset correctly in the display loop, causing an incorrect
			// gc break between the two codepoints of the country-flag emoji.
			inputString: " \U0001F1FA\U0001F1F8",
			expectedContents: [][]string{
				// tcell v3 splits the flag emoji into separate cells
				{" ", "\U0001F1FA", "", "\U0001F1F8", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name: "space with combining mark modifier",
			// This is a space with a combining mark macron.
			// tcell v3 splits the space and combining mark into separate cells.
			inputString: " \u0304",
			expectedContents: [][]string{
				{" ", "\u0304", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name: "tab with combining mark modifier",
			// tcell v3 renders tabs as spaces
			inputString: "\t\u0304abc",
			expectedContents: [][]string{
				{" ", " ", " ", " ", "a", "b", "c", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withMockScreen(t, vt.MockOptSize{X: 10, Y: 10}, func(s tcell.Screen, b vt.MockBackend) {
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
				})
				assertCellContents(t, b, tc.expectedContents)
			})
		})
	}
}

func TestDrawBufferCarriageReturnAndLineFeedNotRendered(t *testing.T) {
	withMockScreen(t, vt.MockOptSize{X: 5, Y: 2}, func(s tcell.Screen, b vt.MockBackend) {
		drawBuffer(t, s, func(editorState *state.EditorState) {
			state.InsertRune(editorState, '\r')
			state.InsertRune(editorState, '\n')
		})

		// tcell v3 writes space for cells after newlines
		size := b.GetSize()
		for y := vt.Row(0); y < size.Y; y++ {
			for x := vt.Col(0); x < size.X; x++ {
				cell := b.GetCell(vt.Coord{X: x, Y: y})
				// First cell on first line gets a space after \r\n
				if y == 0 && x == 0 {
					assert.Equal(t, " ", cell.C)
				} else {
					assert.Equal(t, "", cell.C)
				}
			}
		}
	})
}

func TestGraphemeClustersWithMultipleRunes(t *testing.T) {
	pad := func(arr []string, width int) []string {
		result := make([]string, width)
		copy(result, arr)
		for i := len(arr); i < width; i++ {
			result[i] = ""
		}
		return result
	}

	testCases := []struct {
		name             string
		inputString      string
		expectedContents [][]string
	}{
		{
			name:        "ascii",
			inputString: "abcd1234",
			expectedContents: [][]string{
				pad([]string{"a", "b", "c", "d", "1", "2", "3", "4"}, 100),
			},
		},
		{
			name:        "thai",
			inputString: "\u0E04\u0E49\u0E33",
			expectedContents: [][]string{
				pad([]string{"\u0E04\u0E49\u0E33"}, 100),
			},
		},
		{
			name:        "emoji with zero-width joiner",
			inputString: "\U0001f9db\u200d\u2640\U0001f469\u200d\U0001f467\u200d\U0001f467",
			expectedContents: [][]string{
				pad([]string{"\U0001f9db\u200d\u2640", "", "\U0001f469\u200d\U0001f467\u200d\U0001f467", ""}, 100),
			},
		},
		{
			name:        "regional indicator",
			inputString: "\U0001f1fa\U0001f1f8 (usa!)",
			expectedContents: [][]string{
				pad([]string{"\U0001f1fa\U0001f1f8", "", " ", "(", "u", "s", "a", "!", ")"}, 100),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withMockScreen(t, vt.MockOptSize{X: 100, Y: 1}, func(s tcell.Screen, b vt.MockBackend) {
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
				})
				assertCellContents(t, b, tc.expectedContents)
			})
		})
	}
}

func TestDrawBufferSizeTooSmall(t *testing.T) {
	withMockScreen(t, vt.MockOptSize{X: 1, Y: 4}, func(s tcell.Screen, b vt.MockBackend) {
		drawBuffer(t, s, func(editorState *state.EditorState) {
			for _, r := range "ab界cd" {
				state.InsertRune(editorState, r)
			}
		})
		assertCellContents(t, b, [][]string{
			{"a"},
			{"b"},
			{""},
			{"c"},
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
			withMockScreen(t, vt.MockOptSize{X: 5, Y: 5}, func(s tcell.Screen, b vt.MockBackend) {
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
					state.MoveCursor(editorState, func(state.LocatorParams) uint64 {
						return tc.cursorPosition
					})
				})
				cursorStyle := b.GetCursor()
				assert.Equal(t, tc.expectedCursorVisible, cursorStyle.IsVisible())
				if tc.expectedCursorVisible {
					cursorPos := b.GetPosition()
					assert.Equal(t, tc.expectedCursorCol, cursorPos.X)
					assert.Equal(t, tc.expectedCursorRow, cursorPos.Y)
				}
			})
		})
	}
}

func TestSyntaxHighlighting(t *testing.T) {
	withMockScreen(t, vt.MockOptSize{X:18,Y:1}, func(s tcell.Screen, b vt.MockBackend) {
		drawBuffer(t, s, func(editorState *state.EditorState) {
			state.SetSyntax(editorState, syntax.LanguageGo)
			for _, r := range `const foo = "test"` {
				state.InsertRune(editorState, r)
			}
		})
		assertCellStyles(t, b, [][]tcell.Style{
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
	withMockScreen(t, vt.MockOptSize{X:12,Y:1}, func(s tcell.Screen, b vt.MockBackend) {
		query := "d1"
		drawBuffer(t, s, func(editorState *state.EditorState) {
			for _, r := range `abcd1234` {
				state.InsertRune(editorState, r)
			}
			state.MoveCursor(editorState, func(state.LocatorParams) uint64 { return 0 })
			state.StartSearch(editorState, state.SearchDirectionForward, state.SearchCompleteMoveCursorToMatch)
			for _, r := range query {
				state.AppendRuneToSearchQuery(editorState, r)
			}
		})
		assertCellStyles(t, b, [][]tcell.Style{
			{
				tcell.StyleDefault.Reverse(true).Dim(true),
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
			withMockScreen(t, vt.MockOptSize{X:vt.Col(tc.width),Y:vt.Row(tc.height)}, func(s tcell.Screen, b vt.MockBackend) {
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
				assertCellStyles(t, b, tc.expectedStyles)
			})
		})
	}
}

func TestShowLineNumbers(t *testing.T) {
	testCases := []struct {
		name             string
		width, height    int
		inputString      string
		expectedContents [][]string
	}{
		{
			name:        "empty",
			width:       5,
			height:      5,
			inputString: "",
			expectedContents: [][]string{
				{"", "1", "", "", ""},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
			},
		},
		{
			name:        "single line",
			width:       5,
			height:      5,
			inputString: "ab",
			expectedContents: [][]string{
				{"", "1", "", "a", "b"},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
			},
		},
		{
			name:        "single line, soft-wrapped",
			width:       5,
			height:      5,
			inputString: "abcde",
			expectedContents: [][]string{
				{"", "1", "", "a", "b"},
				{"", "", "", "c", "d"},
				{"", "", "", "e", ""},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
			},
		},
		{
			name:        "multiple lines",
			width:       5,
			height:      5,
			inputString: "ab\nc\nde",
			expectedContents: [][]string{
				{"", "1", "", "a", "b"},
				{"", "2", "", "c", ""},
				{"", "3", "", "d", "e"},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
			},
		},
		{
			name:        "multiple lines, last line empty",
			width:       5,
			height:      5,
			inputString: "ab\nc\nd\n",
			expectedContents: [][]string{
				{"", "1", "", "a", "b"},
				{"", "2", "", "c", ""},
				{"", "3", "", "d", ""},
				{"", "4", "", "", ""},
				{"", "", "", "", ""},
			},
		},
		{
			name:        "multiple lines, last line empty with newline at view width",
			width:       5,
			height:      5,
			inputString: "ab\nc\nde\n",
			expectedContents: [][]string{
				{"", "1", "", "a", "b"},
				{"", "2", "", "c", ""},
				{"", "3", "", "d", "e"},
				{"", "4", "", "", ""},
				{"", "", "", "", ""},
			},
		},
		{
			name:        "collapsed margin",
			width:       3,
			height:      5,
			inputString: "ab\ncd",
			expectedContents: [][]string{
				{"a", "b", ""},
				{"c", "d", ""},
				{"", "", ""},
				{"", "", ""},
				{"", "", ""},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withMockScreen(t, vt.MockOptSize{X:vt.Col(tc.width), Y:vt.Row(tc.height)}, func(s tcell.Screen, b vt.MockBackend) {
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
					state.ToggleShowLineNumbers(editorState)
				})
				assertCellContents(t, b, tc.expectedContents)
			})
		})
	}
}

func TestLineNumberMode(t *testing.T) {
	testCases := []struct {
		name             string
		width, height    int
		inputString      string
		cursorPosition   uint64
		lineNumMode      config.LineNumberMode
		showLineNum      bool
		expectedContents [][]string
	}{
		{
			name:  "absolute with cursor on first line",
			width: 5, height: 5,
			inputString:    "ab\nc\nde",
			cursorPosition: 0,
			lineNumMode:    config.LineNumberModeAbsolute,
			showLineNum:    true,
			expectedContents: [][]string{
				{"", "1", "", "a", "b"},
				{"", "2", "", "c", ""},
				{"", "3", "", "d", "e"},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
			},
		},
		{
			name:  "absolute with cursor on third line",
			width: 5, height: 5,
			inputString:    "ab\nc\nde",
			cursorPosition: 5, // cursor is on the 'd'
			lineNumMode:    config.LineNumberModeAbsolute,
			showLineNum:    true,
			expectedContents: [][]string{
				{"", "1", "", "a", "b"},
				{"", "2", "", "c", ""},
				{"", "3", "", "d", "e"},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
			},
		},
		{
			name:  "with line numbers disabled",
			width: 5, height: 5,
			inputString:    "ab\nc\nde\nfg\nhi",
			cursorPosition: 5, // cursor is on 'd'
			lineNumMode:    config.LineNumberModeAbsolute,
			showLineNum:    false,
			expectedContents: [][]string{
				{"a", "b", "", "", ""},
				{"c", "", "", "", ""},
				{"d", "e", "", "", ""},
				{"f", "g", "", "", ""},
				{"h", "i", "", "", ""},
			},
		},
		{
			name:  "relative in empty document",
			width: 5, height: 5,
			inputString:    "",
			cursorPosition: 0,
			lineNumMode:    config.LineNumberModeRelative,
			showLineNum:    true,
			expectedContents: [][]string{
				{"", "0", "", "", ""},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
			},
		},
		{
			name:  "relative with cursor on first line",
			width: 5, height: 5,
			inputString:    "ab\nc\nde",
			cursorPosition: 0,
			lineNumMode:    config.LineNumberModeRelative,
			showLineNum:    true,
			expectedContents: [][]string{
				{"", "0", "", "a", "b"},
				{"", "1", "", "c", ""},
				{"", "2", "", "d", "e"},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
			},
		},
		{
			name:  "relative with cursor on second line",
			width: 5, height: 5,
			inputString:    "ab\nc\nde\nfg\nhi",
			cursorPosition: 3, // cursor is on 'c'
			lineNumMode:    config.LineNumberModeRelative,
			showLineNum:    true,
			expectedContents: [][]string{
				{"", "1", "", "a", "b"},
				{"", "0", "", "c", ""},
				{"", "1", "", "d", "e"},
				{"", "2", "", "f", "g"},
				{"", "3", "", "h", "i"},
			},
		},
		{
			name:  "relative with cursor on third line",
			width: 5, height: 5,
			inputString:    "ab\nc\nde\nfg\nhi",
			cursorPosition: 5, // cursor is on 'd'
			lineNumMode:    config.LineNumberModeRelative,
			showLineNum:    true,
			expectedContents: [][]string{
				{"", "2", "", "a", "b"},
				{"", "1", "", "c", ""},
				{"", "0", "", "d", "e"},
				{"", "1", "", "f", "g"},
				{"", "2", "", "h", "i"},
			},
		},
		{
			name:  "relative with cursor on third line without line numbers",
			width: 5, height: 5,
			inputString:    "ab\nc\nde\nfg\nhi",
			cursorPosition: 5, // cursor is on 'd'
			lineNumMode:    config.LineNumberModeRelative,
			showLineNum:    false,
			expectedContents: [][]string{
				{"a", "b", "", "", ""},
				{"c", "", "", "", ""},
				{"d", "e", "", "", ""},
				{"f", "g", "", "", ""},
				{"h", "i", "", "", ""},
			},
		},
		{
			name:  "relative with cursor on 12th line",
			width: 5, height: 25,
			cursorPosition: 25,
			lineNumMode:    config.LineNumberModeRelative,
			showLineNum:    true,
			inputString:    "ab\nc\nde\nfg\nhi\n\nj\nk\nl\nm\nn\no\np\nq\nr\ns\nt\nu\nv\nw\nx\ny\nz\n",
			expectedContents: [][]string{
				{"1", "1", "", "a", "b"},
				{"1", "0", "", "c", ""},
				{"", "9", "", "d", "e"},
				{"", "8", "", "f", "g"},
				{"", "7", "", "h", "i"},
				{"", "6", "", "", ""},
				{"", "5", "", "j", ""},
				{"", "4", "", "k", ""},
				{"", "3", "", "l", ""},
				{"", "2", "", "m", ""},
				{"", "1", "", "n", ""},
				{"", "0", "", "o", ""},
				{"", "1", "", "p", ""},
				{"", "2", "", "q", ""},
				{"", "3", "", "r", ""},
				{"", "4", "", "s", ""},
				{"", "5", "", "t", ""},
				{"", "6", "", "u", ""},
				{"", "7", "", "v", ""},
				{"", "8", "", "w", ""},
				{"", "9", "", "x", ""},
				{"1", "0", "", "y", ""},
				{"1", "1", "", "z", ""},
				{"1", "2", "", "", ""},
				{"", "", "", "", ""},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withMockScreen(t, vt.MockOptSize{X:vt.Col(tc.width), Y: vt.Row(tc.height)}, func(s tcell.Screen, b vt.MockBackend) {
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
					state.MoveCursor(editorState, func(state.LocatorParams) uint64 {
						return tc.cursorPosition
					})
					if tc.showLineNum {
						state.ToggleShowLineNumbers(editorState)
					}

					state.SetLineNumberMode(editorState, tc.lineNumMode)
				})

				assertCellContents(t, b, tc.expectedContents)
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
		expectedContents [][]string
	}{
		{
			name:        "hide tabs",
			width:       8,
			height:      2,
			showTabs:    false,
			inputString: "\ta\t\nb\t",
			expectedContents: [][]string{
				{"", "", "", "", "a", "", "", ""},
				{"b", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "show tabs",
			width:       8,
			height:      2,
			showTabs:    true,
			inputString: "\ta\t\nb\t",
			expectedContents: [][]string{
				{string([]rune{tcell.RuneRArrow}), "", "", "", "a", string([]rune{tcell.RuneRArrow}), "", ""},
				{"b", string([]rune{tcell.RuneRArrow}), "", "", "", "", "", ""},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withMockScreen(t, vt.MockOptSize{X:vt.Col(tc.width), Y: vt.Row(tc.height)}, func(s tcell.Screen, b vt.MockBackend) {
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
					if tc.showTabs {
						state.ToggleShowTabs(editorState)
					}
				})
				assertCellContents(t, b, tc.expectedContents)
			})
		})
	}
}

func TestShowTabsWithSelectionStyle(t *testing.T) {
	withMockScreen(t, vt.MockOptSize{X:8, Y: 1}, func(s tcell.Screen, b vt.MockBackend) {
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
		assertCellContents(t, b, [][]string{
			{string([]rune{tcell.RuneRArrow}), "", "", "", "a", "b", "c", ""},
		})
		assertCellStyles(t, b, [][]tcell.Style{
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

func TestShowSpaces(t *testing.T) {
	testCases := []struct {
		name             string
		width, height    int
		showSpaces       bool
		inputString      string
		expectedContents [][]string
	}{
		{
			name:        "hide spaces",
			width:       8,
			height:      2,
			showSpaces:  false,
			inputString: " a \nb ",
			expectedContents: [][]string{
				{"", "a", "", "", "", "", "", ""},
				{"b", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "show spaces",
			width:       8,
			height:      2,
			showSpaces:  true,
			inputString: " a \nb ",
			expectedContents: [][]string{
				{string([]rune{tcell.RuneBullet}), "a", string([]rune{tcell.RuneBullet}), "", "", "", "", ""},
				{"b", string([]rune{tcell.RuneBullet}), "", "", "", "", "", ""},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withMockScreen(t, vt.MockOptSize{X:vt.Col(tc.width), Y: vt.Row(tc.height)}, func(s tcell.Screen, b vt.MockBackend) {
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
					if tc.showSpaces {
						state.ToggleShowSpaces(editorState)
					}
				})
				assertCellContents(t, b, tc.expectedContents)
			})
		})
	}
}

func TestShowUnicode(t *testing.T) {
	testCases := []struct {
		name             string
		width, height    int
		showUnicode      bool
		inputString      string
		expectedContents [][]string
	}{
		{
			name:        "render unicode",
			width:       10,
			height:      2,
			showUnicode: false,
			// emoji "frowning face"
			inputString: "\u2639",
			expectedContents: [][]string{
				{"\u2639", "", "", "", "", "", "", "", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			name:        "escape unicode",
			width:       10,
			height:      2,
			showUnicode: true,
			// emoji "frowning face"
			inputString: "\u2639",
			expectedContents: [][]string{
				{"<", "U", "+", "2", "6", "3", "9", ">", "", ""},
				{"", "", "", "", "", "", "", "", "", ""},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withMockScreen(t, vt.MockOptSize{X:vt.Col(tc.width), Y: vt.Row(tc.height)}, func(s tcell.Screen, b vt.MockBackend) {
				drawBuffer(t, s, func(editorState *state.EditorState) {
					for _, r := range tc.inputString {
						state.InsertRune(editorState, r)
					}
					if tc.showUnicode {
						state.ToggleShowUnicode(editorState)
					}
				})
				assertCellContents(t, b, tc.expectedContents)
			})
		})
	}
}

