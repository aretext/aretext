package display

import (
	"bufio"
	"log"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/wedaly/aretext/internal/pkg/text"
)

// TextView displays text in a terminal, clipping and scrolling as necessary.
type TextView struct {
	tree         *text.Tree
	screenRegion *ScreenRegion
}

// NewTextView initializes a text view for a text tree and screen.
func NewTextView(tree *text.Tree, screenRegion *ScreenRegion) *TextView {
	return &TextView{tree, screenRegion}
}

// Resize notifies the text view that the terminal size has changed.
func (v *TextView) Resize(width, height int) {
	v.screenRegion.Resize(width, height)
}

// Draw draws text to the screen.
func (v *TextView) Draw() {
	width, height := v.screenRegion.Size()

	if width < 2 {
		// If the view is too narrow to display full-width characters (occupying 2 cells), just fill it.
		// Everything after this point assumes that there's space on each line for at least one full-width char.
		v.screenRegion.Fill('~', tcell.StyleDefault.Dim(true))
	} else if height > 0 {
		v.drawText(width, height)
	}
}

func (v *TextView) drawText(width, height int) {
	reader := v.tree.ReaderAtPosition(0, text.ReadDirectionForward)
	scanner := bufio.NewScanner(reader)
	scanner.Split(splitUtf8Cells)

	x, y := 0, 0
	for scanner.Scan() {
		s := scanner.Text()

		// If newline, skip to the next line.
		if s[0] == '\n' {
			x = 0
			y++
			continue
		}

		r, n := utf8.DecodeRune([]byte(s))
		if n == 1 && r == utf8.RuneError {
			// This should never happen because text.Tree validates UTF-8 characters.
			log.Fatal("invalid rune")
		}

		rw := runewidth.RuneWidth(r)

		// If a full-width character would be clipped, push it to the next line.
		if x+rw > width {
			x = 0
			y++
			if y >= height {
				break
			}
		}

		if rw > 0 {
			// Display the characters in this cell.  The first rune is a non-zero-width character,
			// and subsequent runes are combining characters (like accent marks).
			v.screenRegion.SetContent(x, y, r, []rune(s[n:]), tcell.StyleDefault)
		} else {
			// Replace non-displayable character.  This won't happen unless a zero-width character is in the wrong location
			// (for example, a combining accent mark at the very beginning of the string).
			v.screenRegion.SetContent(x, y, utf8.RuneError, nil, tcell.StyleDefault)
		}

		// Move the cursor forward, wrapping to the next line if necessary.
		x += rw
		if x >= width {
			x = 0
			y++
			if y >= height {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		// This should never happen because splitUtf8Cells never returns errors.
		log.Fatalf("error scanning text: %v", err)
	}
}
