package breaks

import (
	"io"

	"github.com/wedaly/aretext/internal/pkg/text"
)

//go:generate go run gen_props.go --prefix gb --dataPath data/GraphemeBreakProperty.txt --dataPath data/emoji-data.txt --propertyName Prepend --propertyName Control --propertyName Extend --propertyName Regional_Indicator --propertyName SpacingMark --propertyName L --propertyName V --propertyName T --propertyName LV --propertyName LVT --propertyName ZWJ --propertyName CR --propertyName LF --propertyName Extended_Pictographic --outputPath grapheme_clusters_props.go

// GraphemeClusterBreakIter identifies valid breakpoints between grapheme clusters.
// A grapheme cluster is a user-perceived character.  These may be composed of multiple unicode codepoints.
// For full details see https://www.unicode.org/reports/tr29/ version 13.0.0, revision 37.
type GraphemeClusterBreakIter struct {
	runeIter                         text.RuneIter
	runeCount                        uint64
	endOfText                        bool
	lastProp                         gbProp
	inExtendedPictographic           bool
	afterExtendedPictographicPlusZWJ bool
	lastPropsWereRIOdd               bool
	lastPropsWereRIEven              bool
}

// NewGraphemeClusterBreakIter initializes a new iterator.
// The iterator assumes that the first character it receives is at a break point
// (either the start of the text or the beginning of a new grapheme cluster).
// The input reader MUST produce valid UTF-8 codepoints.
func NewGraphemeClusterBreakIter(runeIter text.RuneIter) BreakIter {
	return &GraphemeClusterBreakIter{runeIter: runeIter}
}

// NextBreak implements BreakIter#NextBreak()
func (g *GraphemeClusterBreakIter) NextBreak() (uint64, error) {
	for {
		r, err := g.runeIter.NextRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}

		canBreakBefore := g.processRune(r)
		g.runeCount++
		if canBreakBefore {
			return g.runeCount - 1, nil
		}
	}

	if !g.endOfText {
		g.endOfText = true
		return g.runeCount, nil
	}

	return 0, io.EOF
}

// processRune determines whether the position before the rune is a valid breakpoint (starts a new grapheme cluster).
func (g *GraphemeClusterBreakIter) processRune(r rune) (canBreakBefore bool) {
	prop := gbPropForRune(r)

	defer func() {
		g.lastPropsWereRIEven = bool(prop == gbPropRegional_Indicator && g.lastPropsWereRIOdd)
		g.lastPropsWereRIOdd = bool(prop == gbPropRegional_Indicator && !g.lastPropsWereRIOdd)
		g.afterExtendedPictographicPlusZWJ = bool(g.inExtendedPictographic && prop == gbPropZWJ)
		g.inExtendedPictographic = bool(
			prop == gbPropExtended_Pictographic ||
				g.inExtendedPictographic && prop == gbPropExtend)
		g.lastProp = prop
	}()

	// GB1: sot ÷ Any
	if g.runeCount == 0 {
		return true
	}

	// GB2: Any ÷ eot
	// This rule is implemented by the caller, which is able to detect EOF.

	// GB3: CR × LF
	if prop == gbPropLF && g.lastProp == gbPropCR {
		return false
	}

	// GB4: (Control | CR | LF) ÷
	if g.lastProp == gbPropControl || g.lastProp == gbPropCR || g.lastProp == gbPropLF ||
		// GB5: ÷ (Control | CR | LF)
		prop == gbPropControl || prop == gbPropCR || prop == gbPropLF {
		return true
	}

	// GB6: L × (L | V | LV | LVT)
	if (g.lastProp == gbPropL && (prop == gbPropL || prop == gbPropV || prop == gbPropLV || prop == gbPropLVT)) ||

		// GB7: (LV | V) × (V | T)
		((g.lastProp == gbPropLV || g.lastProp == gbPropV) && (prop == gbPropV || prop == gbPropT)) ||
		// GB8: (LVT | T) × T
		((g.lastProp == gbPropLVT || g.lastProp == gbPropT) && prop == gbPropT) ||

		// GB9: × (Extend | ZWJ)
		prop == gbPropExtend || prop == gbPropZWJ ||

		// GB9a: × SpacingMark
		prop == gbPropSpacingMark ||

		// GB9b: Prepend ×
		g.lastProp == gbPropPrepend ||

		// GB11: \p{Extended_Pictographic} Extend* ZWJ × \p{Extended_Pictographic}
		(g.afterExtendedPictographicPlusZWJ && prop == gbPropExtended_Pictographic) ||

		// GB12: sot (RI RI)* RI × RI
		// GB13: [^RI] (RI RI)* RI × RI
		(g.lastPropsWereRIOdd && prop == gbPropRegional_Indicator) {
		return false
	}

	// GB999: Any ÷ Any
	return true
}

// ReverseGraphemeClusterBreakIter identifies valid breakpoints between grapheme clusters in a reversed-order sequence of runes.
type ReverseGraphemeClusterBreakIter struct {
	runeIter    text.CloneableRuneIter
	runeCount   uint64
	startOfText bool
	lastProp    gbProp
}

// NewReverseGraphemeClusterBreakIter constructs a new BreakIter from a runeIter that yields runes in reverse order.
func NewReverseGraphemeClusterBreakIter(runeIter text.CloneableRuneIter) BreakIter {
	return &ReverseGraphemeClusterBreakIter{runeIter: runeIter}
}

// NextBreak implements BreakIter#NextBreak()
// The returned locations are relative to the end of the text.
func (g *ReverseGraphemeClusterBreakIter) NextBreak() (uint64, error) {
	for {
		r, err := g.runeIter.NextRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}

		// "After" is relative to the original (non-reversed) rune order.
		// So if the original string was "abcd" and we're iterating through it backwards,
		// then the break between "b" and "c" would be *after* "b".
		canBreakAfter := g.processRune(r)
		g.runeCount++
		if canBreakAfter {
			return g.runeCount - 1, nil
		}
	}

	if !g.startOfText {
		g.startOfText = true
		return g.runeCount, nil
	}

	return 0, io.EOF
}

func (g *ReverseGraphemeClusterBreakIter) processRune(r rune) (canBreakAfter bool) {
	prop := gbPropForRune(r)
	defer func() { g.lastProp = prop }()

	// GB1: sot ÷ Any
	// This rule is implemented by the caller, which is able to detect EOF in the reversed input.

	// GB2: Any ÷ eot
	if g.runeCount == 0 {
		return true
	}

	// GB3: CR × LF
	if prop == gbPropCR && g.lastProp == gbPropLF {
		return false
	}

	// GB4: (Control | CR | LF) ÷
	if prop == gbPropControl || prop == gbPropCR || prop == gbPropLF {
		return true
	}

	// GB5: ÷ (Control | CR | LF)
	if g.lastProp == gbPropControl || g.lastProp == gbPropCR || g.lastProp == gbPropLF {
		return true
	}

	// GB6: L × (L | V | LV | LVT)
	if prop == gbPropL && (g.lastProp == gbPropL || g.lastProp == gbPropV || g.lastProp == gbPropLV || g.lastProp == gbPropLVT) {
		return false
	}

	// GB7: (LV | V) × (V | T)
	if (prop == gbPropLV || prop == gbPropV) && (g.lastProp == gbPropV || g.lastProp == gbPropT) {
		return false
	}

	// GB8: (LVT | T) × T
	if (prop == gbPropLVT || prop == gbPropT) && g.lastProp == gbPropT {
		return false
	}

	// GB9: × (Extend | ZWJ)
	if g.lastProp == gbPropExtend || g.lastProp == gbPropZWJ {
		return false
	}

	// GB9a: × SpacingMark
	if g.lastProp == gbPropSpacingMark {
		return false
	}

	// GB9b: Prepend ×
	if prop == gbPropPrepend {
		return false
	}

	// GB11: \p{Extended_Pictographic} Extend* ZWJ × \p{Extended_Pictographic}
	if prop == gbPropZWJ && g.lastProp == gbPropExtended_Pictographic && g.lookaheadExtendedPictographicSequence() {
		return false
	}

	// GB12: sot (RI RI)* RI × RI
	// GB13: [^RI] (RI RI)* RI × RI
	if g.lastProp == gbPropRegional_Indicator && prop == gbPropRegional_Indicator && g.lookaheadEvenRI() {
		return false
	}

	// GB999: Any ÷ Any
	return true
}

func (g *ReverseGraphemeClusterBreakIter) lookaheadExtendedPictographicSequence() bool {
	iterClone := g.runeIter.Clone()

	// Check for Extend* followed by \p{Extended_Pictographic}
	for {
		r, err := iterClone.NextRune()
		if err != nil {
			return false
		}

		prop := gbPropForRune(r)
		if prop == gbPropExtend {
			continue
		} else if prop == gbPropExtended_Pictographic {
			return true
		} else {
			return false
		}
	}
}

func (g *ReverseGraphemeClusterBreakIter) lookaheadEvenRI() bool {
	riCount := 0
	iterClone := g.runeIter.Clone()
	for {
		r, err := iterClone.NextRune()
		if err != nil {
			break
		}

		prop := gbPropForRune(r)
		if prop == gbPropRegional_Indicator {
			riCount++
		} else {
			break
		}
	}

	return riCount%2 == 0
}
