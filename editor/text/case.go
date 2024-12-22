package text

import (
	"unicode"
)

// ToggleRuneCase changes the case of the rune from lower-to-upper or vice-versa.
func ToggleRuneCase(r rune) rune {
	if unicode.IsUpper(r) {
		return unicode.ToLower(r)
	} else if unicode.IsLower(r) {
		return unicode.ToUpper(r)
	} else {
		return r
	}
}
