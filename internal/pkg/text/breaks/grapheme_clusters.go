package breaks

//go:generate go run gen_props.go --prefix gb --dataPath data/GraphemeBreakProperty.txt --dataPath data/emoji-data.txt --propertyName Prepend --propertyName Control --propertyName Extend --propertyName Regional_Indicator --propertyName SpacingMark --propertyName L --propertyName V --propertyName T --propertyName LV --propertyName LVT --propertyName ZWJ --propertyName CR --propertyName LF --propertyName Extended_Pictographic --outputPath grapheme_clusters_props.go
//go:generate go run gen_tests.go --dataPath data/GraphemeBreakTest.txt --outputPath grapheme_clusters_generated_test.go

// GraphemeClusterBreakFinder identifies valid breakpoints between grapheme clusters.
// A grapheme cluster is a user-perceived character.  These may be composed of multiple unicode codepoints.
// For full details see https://www.unicode.org/reports/tr29/ version 13.0.0, revision 37.
type GraphemeClusterBreakFinder struct {
	startOfText                      bool
	lastProp                         gbProp
	inExtendedPictographic           bool
	afterExtendedPictographicPlusZWJ bool
	lastPropsWereRIOdd               bool
	lastPropsWereRIEven              bool
}

// NewGraphemeClusterBreakFinder initializes a new finder.
// The finder assumes that the first character it receives is at a break point
// (either the start of the text or the beginning of a new grapheme cluster).
func NewGraphemeClusterBreakFinder() *GraphemeClusterBreakFinder {
	return &GraphemeClusterBreakFinder{startOfText: true}
}

// Reset resets the finder to its initial state.
func (g *GraphemeClusterBreakFinder) Reset() {
	*g = GraphemeClusterBreakFinder{startOfText: true}
}

// ProcessCharacter determines whether the position before the character is a valid breakpoint (starts a new grapheme cluster).
func (g *GraphemeClusterBreakFinder) ProcessCharacter(c rune) (canBreakBefore bool) {
	prop := gbPropForRune(c)

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
	if g.startOfText {
		g.startOfText = false
		return true
	}

	// GB2: Any ÷ eot
	// We don't implement this rule, because we can't detect the end of the file.

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
