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
	runeIter                         text.CloneableRuneIter
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
func NewGraphemeClusterBreakIter(in text.CloneableReader) *GraphemeClusterBreakIter {
	runeIter := text.NewForwardRuneIter(in)
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
