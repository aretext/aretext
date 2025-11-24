package segment

import (
	"io"

	"github.com/aretext/aretext/text"
)

//go:generate go run gen_props.go --prefix gb --dataPath data/GraphemeBreakProperty.txt --dataPath data/emoji-data.txt --propertyName Prepend --propertyName Control --propertyName Extend --propertyName Regional_Indicator --propertyName SpacingMark --propertyName L --propertyName V --propertyName T --propertyName LV --propertyName LVT --propertyName ZWJ --propertyName CR --propertyName LF --propertyName Extended_Pictographic --outputPath grapheme_cluster_props.go

// GraphemeClusterBreaker finds breakpoints between grapheme clusters in a sequence of runes.
type GraphemeClusterBreaker struct {
	lastProp                         gbProp
	inExtendedPictographic           bool
	afterExtendedPictographicPlusZWJ bool
	lastPropsWereRIOdd               bool
}

// ProcessRune determines whether the position before the rune is a valid breakpoint (starts a new grapheme cluster).
func (gb *GraphemeClusterBreaker) ProcessRune(r rune) (canBreakBefore bool) {
	prop := gbPropForRune(r)

	// GB1: sot ÷ Any
	// GB2: Any ÷ eot
	// We don't need to implement these because we're only interested in non-empty segments.

	// GB3: CR × LF
	if prop == gbPropLF && gb.lastProp == gbPropCR {
		canBreakBefore = false
		goto done
	}

	// GB4: (Control | CR | LF) ÷
	if gb.lastProp == gbPropControl || gb.lastProp == gbPropCR || gb.lastProp == gbPropLF ||
		// GB5: ÷ (Control | CR | LF)
		prop == gbPropControl || prop == gbPropCR || prop == gbPropLF {
		canBreakBefore = true
		goto done
	}

	// GB6: L × (L | V | LV | LVT)
	if (gb.lastProp == gbPropL && (prop == gbPropL || prop == gbPropV || prop == gbPropLV || prop == gbPropLVT)) ||

		// GB7: (LV | V) × (V | T)
		((gb.lastProp == gbPropLV || gb.lastProp == gbPropV) && (prop == gbPropV || prop == gbPropT)) ||
		// GB8: (LVT | T) × T
		((gb.lastProp == gbPropLVT || gb.lastProp == gbPropT) && prop == gbPropT) ||

		// GB9: × (Extend | ZWJ)
		prop == gbPropExtend || prop == gbPropZWJ ||

		// GB9a: × SpacingMark
		prop == gbPropSpacingMark ||

		// GB9b: Prepend ×
		gb.lastProp == gbPropPrepend ||

		// GB11: \p{Extended_Pictographic} Extend* ZWJ × \p{Extended_Pictographic}
		(gb.afterExtendedPictographicPlusZWJ && prop == gbPropExtended_Pictographic) ||

		// GB12: sot (RI RI)* RI × RI
		// GB13: [^RI] (RI RI)* RI × RI
		(gb.lastPropsWereRIOdd && prop == gbPropRegional_Indicator) {
		canBreakBefore = false
		goto done
	}

	// GB999: Any ÷ Any
	canBreakBefore = true

done:
	gb.lastPropsWereRIOdd = bool(prop == gbPropRegional_Indicator && !gb.lastPropsWereRIOdd)
	gb.afterExtendedPictographicPlusZWJ = bool(gb.inExtendedPictographic && prop == gbPropZWJ)
	gb.inExtendedPictographic = bool(
		prop == gbPropExtended_Pictographic ||
			gb.inExtendedPictographic && prop == gbPropExtend)
	gb.lastProp = prop

	return canBreakBefore
}

// GraphemeClusterIter segments text into grapheme clusters.
// A grapheme cluster is a user-perceived character, which can be composed of multiple unicode codepoints.
// For full details see https://www.unicode.org/reports/tr29/ version 13.0.0, revision 37.
// Copying the struct produces a new, independent iterator.
type GraphemeClusterIter struct {
	reader           text.Reader
	breaker          GraphemeClusterBreaker
	hasCarryoverRune bool
	carryoverRune    rune
}

// NewGraphemeClusterIter initializes a new iterator.
// The iterator assumes that the first character it receives is at a break point
// (either the start of the text or the beginning of a new grapheme cluster).
// The input reader MUST produce valid UTF-8 codepoints.
func NewGraphemeClusterIter(reader text.Reader) GraphemeClusterIter {
	return GraphemeClusterIter{reader: reader}
}

// NextSegment retrieves the next grapheme cluster.
func (g *GraphemeClusterIter) NextSegment(segment *Segment) error {
	segment.Clear()

	if g.hasCarryoverRune {
		segment.Append(g.carryoverRune)
		g.hasCarryoverRune = false
	}

	for {
		r, _, err := g.reader.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if canBreakBefore := g.breaker.ProcessRune(r); canBreakBefore && segment.NumRunes() > 0 {
			g.hasCarryoverRune = true
			g.carryoverRune = r
			return nil
		}

		segment.Append(r)
	}

	if segment.NumRunes() > 0 {
		return nil
	}

	return io.EOF
}

// ReverseGraphemeClusterIter identifies valid breakpoints between grapheme clusters in a reversed-order sequence of runes.
// Copying the struct produces a new, independent iterator.
type ReverseGraphemeClusterIter struct {
	reader           text.ReverseReader
	hasCarryoverRune bool
	carryoverRune    rune
	lastProp         gbProp
}

// NewReverseGraphemeClusterIter constructs a new BreakIter from a runeIter that yields runes in reverse order.
func NewReverseGraphemeClusterIter(reader text.ReverseReader) ReverseGraphemeClusterIter {
	return ReverseGraphemeClusterIter{reader: reader}
}

// NextSegment retrives the next grapheme cluster reading backwards.
// The returned locations are relative to the end of the text.
func (g *ReverseGraphemeClusterIter) NextSegment(segment *Segment) error {
	segment.Clear()

	if g.hasCarryoverRune {
		segment.Append(g.carryoverRune)
		g.hasCarryoverRune = false
	}

	for {
		r, _, err := g.reader.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// "After" is relative to the original (non-reversed) rune order.
		// So if the original string was "abcd" and we're iterating through it backwards,
		// then the break between "b" and "c" would be *after* "b".
		if canBreakAfter := g.processRune(r); canBreakAfter && segment.NumRunes() > 0 {
			segment.ReverseRunes()
			g.hasCarryoverRune = true
			g.carryoverRune = r
			return nil
		}

		segment.Append(r)
	}

	if segment.NumRunes() > 0 {
		segment.ReverseRunes()
		return nil
	}

	return io.EOF
}

func (g *ReverseGraphemeClusterIter) processRune(r rune) (canBreakAfter bool) {
	prop := gbPropForRune(r)
	defer func() { g.lastProp = prop }()

	// GB1: sot ÷ Any
	// GB2: Any ÷ eot
	// We don't need to implement these because we're only interested in non-empty segments.

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

func (g *ReverseGraphemeClusterIter) lookaheadExtendedPictographicSequence() bool {
	rclone := g.reader

	// Check for Extend* followed by \p{Extended_Pictographic}
	for {
		r, _, err := rclone.ReadRune()
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

func (g *ReverseGraphemeClusterIter) lookaheadEvenRI() bool {
	riCount := 0
	rclone := g.reader
	for {
		r, _, err := rclone.ReadRune()
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
