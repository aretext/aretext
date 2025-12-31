package text

import (
	"fmt"
	"strings"
)

// Escaper outputs an ASCII representation of unicode codepoints.
type Escaper struct {
	sb strings.Builder
}

func (e *Escaper) RunesToStr(runes []rune) string {
	e.escapeRunes(runes)
	return e.sb.String()
}

func (e *Escaper) RunesToStrLen(runes []rune) int {
	e.escapeRunes(runes)
	return e.sb.Len()
}

func (e *Escaper) escapeRunes(runes []rune) {
	e.sb.Reset()
	e.sb.WriteRune('<')
	for i, r := range runes {
		fmt.Fprintf(&e.sb, "U+%04X", r)
		if i < len(runes)-1 {
			e.sb.WriteRune(',')
		}
	}
	e.sb.WriteRune('>')
}
