package segment

import (
	"io"

	"github.com/aretext/aretext/text"
)

//go:generate go run gen_props.go --prefix gb --dataPath data/GraphemeBreakProperty.txt --dataPath data/emoji-data.txt --propertyName Prepend --propertyName Control --propertyName Extend --propertyName Regional_Indicator --propertyName SpacingMark --propertyName L --propertyName V --propertyName T --propertyName LV --propertyName LVT --propertyName ZWJ --propertyName CR --propertyName LF --propertyName Extended_Pictographic --outputPath grapheme_cluster_props.go

// graphemeClusterIter segments text into grapheme clusters.
// A grapheme cluster is a user-perceived character, which can be composed of multiple unicode codepoints.
// For full details see https://www.unicode.org/reports/tr29/ version 13.0.0, revision 37.
type graphemeClusterIter struct {
	runeIter                         text.CloneableRuneIter
	buffer                           []rune
	lastProp                         gbProp
	inExtendedPictographic           bool
	afterExtendedPictographicPlusZWJ bool
	lastPropsWereRIOdd               bool
	lastPropsWereRIEven              bool
}

// NewGraphemeClusterIter initializes a new iterator.
// The iterator assumes that the first character it receives is at a break point
// (either the start of the text or the beginning of a new grapheme cluster).
// The input reader MUST produce valid UTF-8 codepoints.
func NewGraphemeClusterIter(runeIter text.CloneableRuneIter) CloneableSegmentIter {
	return &graphemeClusterIter{
		runeIter: runeIter,
		buffer:   make([]rune, 0, 1),
	}
}

// Clone implements CloneableSegmentIter#Clone()
func (g *graphemeClusterIter) Clone() CloneableSegmentIter {
	var clone graphemeClusterIter
	clone = *g
	clone.runeIter = g.runeIter.Clone()
	clone.buffer = make([]rune, len(g.buffer))
	copy(clone.buffer, g.buffer)
	return &clone
}

// NextSegment implements SegmentIter#NextSegment()
func (g *graphemeClusterIter) NextSegment(segment *Segment) error {
	segment.Clear()
	for {
		r, err := g.runeIter.NextRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if canBreakBefore := g.processRune(r); canBreakBefore && len(g.buffer) > 0 {
			segment.Extend(g.buffer)
			g.buffer = append(g.buffer[:0], r)
			return nil
		}

		g.buffer = append(g.buffer, r)
	}

	if len(g.buffer) > 0 {
		segment.Extend(g.buffer)
		g.buffer = g.buffer[:0]
		return nil
	}

	return io.EOF
}

// processRune determines whether the position before the rune is a valid breakpoint (starts a new grapheme cluster).
func (g *graphemeClusterIter) processRune(r rune) (canBreakBefore bool) {
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
	// GB2: Any ÷ eot
	// We don't need to implement these because we're only interested in non-empty segments.

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

// reverseGraphemeClusterIter identifies valid breakpoints between grapheme clusters in a reversed-order sequence of runes.
type reverseGraphemeClusterIter struct {
	runeIter text.CloneableRuneIter
	buffer   []rune
	lastProp gbProp
}

// NewReverseGraphemeClusterIter constructs a new BreakIter from a runeIter that yields runes in reverse order.
func NewReverseGraphemeClusterIter(runeIter text.CloneableRuneIter) CloneableSegmentIter {
	return &reverseGraphemeClusterIter{
		runeIter: runeIter,
		buffer:   make([]rune, 0, 1),
	}
}

// Clone implements CloneableSegmentIter#Clone()
func (g *reverseGraphemeClusterIter) Clone() CloneableSegmentIter {
	var clone reverseGraphemeClusterIter
	clone = *g
	clone.runeIter = g.runeIter.Clone()
	clone.buffer = make([]rune, len(g.buffer))
	copy(clone.buffer, g.buffer)
	return &clone
}

// NextBreak implements SegmentIter#NextSegment()
// The returned locations are relative to the end of the text.
func (g *reverseGraphemeClusterIter) NextSegment(segment *Segment) error {
	segment.Clear()
	for {
		r, err := g.runeIter.NextRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// "After" is relative to the original (non-reversed) rune order.
		// So if the original string was "abcd" and we're iterating through it backwards,
		// then the break between "b" and "c" would be *after* "b".
		if canBreakAfter := g.processRune(r); canBreakAfter && len(g.buffer) > 0 {
			segment.Extend(g.buffer).ReverseRunes()
			g.buffer = append(g.buffer[:0], r)
			return nil
		}

		g.buffer = append(g.buffer, r)
	}

	if len(g.buffer) > 0 {
		segment.Extend(g.buffer).ReverseRunes()
		g.buffer = g.buffer[:0]
		return nil
	}

	return io.EOF
}

func (g *reverseGraphemeClusterIter) processRune(r rune) (canBreakAfter bool) {
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

func (g *reverseGraphemeClusterIter) lookaheadExtendedPictographicSequence() bool {
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

func (g *reverseGraphemeClusterIter) lookaheadEvenRI() bool {
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

// GraphemeClusterIsEmoji checks whether the grapheme cluster represents an emoji.
func GraphemeClusterIsEmoji(gc []rune) bool {
	return len(gc) > 0 && gbPropForRune(gc[0]) == gbPropExtended_Pictographic
}

// GraphemeClusterIsRegionalIndicator checks whether the grapheme cluster represents a regional indicator (usually rendered as a flag).
func GraphemeClusterIsRegionalIndicator(gc []rune) bool {
	return len(gc) > 0 && gbPropForRune(gc[0]) == gbPropRegional_Indicator
}
