package segment

import (
	"io"
	"unicode"

	"github.com/aretext/aretext/cellwidth"
	"github.com/aretext/aretext/text"
)

//go:generate go run gen_props.go --prefix lb --dataPath data/LineBreak.txt --propertyName BK --propertyName CM --propertyName CR --propertyName GL --propertyName LF --propertyName NL --propertyName SP --propertyName WJ --propertyName ZW --propertyName ZW --propertyName AI --propertyName AL --propertyName B2 --propertyName BA --propertyName BB --propertyName CB --propertyName CJ --propertyName CL --propertyName CP --propertyName EB --propertyName EM --propertyName EX --propertyName H2 --propertyName H3 --propertyName HL --propertyName HY --propertyName ID --propertyName IN --propertyName IS --propertyName JL --propertyName JT --propertyName JV --propertyName NS --propertyName NU --propertyName OP --propertyName PO --propertyName PR --propertyName QU --propertyName RI --propertyName SA --propertyName SG --propertyName SY --propertyName XX --propertyName ZWJ --outputPath line_break_props.go

//go:generate go run gen_props.go --prefix ea --dataPath data/EastAsianWidth.txt --propertyName F --propertyName W --propertyName H --outputPath east_asian_width_props.go

//go:generate go run gen_props.go --prefix gc --dataPath data/DerivedGeneralCategory.txt --propertyName Cn --outputPath general_category_props.go

//go:generate go run gen_props.go --prefix em --dataPath data/emoji-data.txt --propertyName Extended_Pictographic --outputPath emoji_props.go

type LineBreakDecision byte

const (
	NoLineBreak = LineBreakDecision(iota)
	AllowLineBreakBefore
	RequireLineBreakBefore
	RequireLineBreakAfter
)

// LineBreaker finds possible breakpoints between lines.
// This uses the Unicode line breaking algorithm from
// https://www.unicode.org/reports/tr14/
type LineBreaker struct {
	lastProp                         lbProp
	lastLastProp                     lbProp
	inZeroWidthSpaceSeq              bool
	inLeftBraceSpaceSeq              bool
	inQuotationSpaceSeq              bool
	inClosePunctSpaceSeq             bool
	inDashSpaceSeq                   bool
	lastPropsWereRIOdd               bool
	lastWasExtendedPictographicAndCn bool
}

// ProcessRune finds valid breakpoints between lines.
func (lb *LineBreaker) ProcessRune(r rune) (decision LineBreakDecision) {
	// LB1: Assign a line breaking class to each code point of the input.
	prop := lbPropForRune(r)
	if prop == lbPropNone {
		// LineBreak.txt says "All code points, assigned and unassigned, that are not listed explicitly are given the value XX"
		prop = lbPropXX
	}

	if prop == lbPropAI || prop == lbPropSG || prop == lbPropXX {
		prop = lbPropAL
	} else if prop == lbPropSA {
		if unicode.In(r, unicode.Mn, unicode.Mc) {
			prop = lbPropCM
		} else {
			prop = lbPropAL
		}
	} else if prop == lbPropCJ {
		prop = lbPropNS
	}

	// LB2: Never break at the start of text.
	// LB3: Always break at the end of text.
	// We don't care about empty segments, so we don't need to do anything here.

	// LB4: Always break after hard line breaks.
	if prop == lbPropBK && lb.lastProp != lbPropCR {
		decision = RequireLineBreakAfter
		goto done
	}

	// LB5: Treat CR followed by LF, as well as CR, LF, and NL as hard line breaks.
	if lb.lastProp == lbPropCR && prop == lbPropLF {
		decision = RequireLineBreakAfter
		goto done
	} else if lb.lastProp == lbPropCR {
		decision = RequireLineBreakBefore
		goto done
	} else if prop == lbPropLF || prop == lbPropNL {
		decision = RequireLineBreakAfter
		goto done
	}

	// LB6: Do not break before hard line breaks.
	if prop == lbPropBK || prop == lbPropCR || prop == lbPropLF || prop == lbPropNL {
		goto done
	}

	// LB7: Do not break before spaces or zero width space.
	if prop == lbPropSP || prop == lbPropZW {
		goto done
	}

	// LB8: Break before any character following a zero-width space, even if one or more spaces intervene.
	if lb.inZeroWidthSpaceSeq && prop != lbPropSP {
		decision = AllowLineBreakBefore
		goto done
	}

	// LB8a: Do not break after a zero width joiner.
	if lb.lastProp == lbPropZWJ {
		goto done
	}

	// LB9: Do not break a combining character sequence; treat it as if it has the line breaking class of the base character in all of the following rules. Treat ZWJ as if it were CM.
	if (prop == lbPropCM || prop == lbPropZWJ) &&
		lb.lastProp != lbPropBK &&
		lb.lastProp != lbPropCR &&
		lb.lastProp != lbPropLF &&
		lb.lastProp != lbPropNL &&
		lb.lastProp != lbPropSP &&
		lb.lastProp != lbPropZW &&
		lb.lastProp != lbPropNone {

		prop = lb.lastProp

		// We don't want to *count* the combining mark as a regional indicator,
		// even if we treat it like a regional indicator in other respects.
		// Compensate by flipping "odd number of regional indicators" flag,
		// which will get flipped back to its original value b/c
		// the CM is treated like an RI below.
		if prop == lbPropRI {
			lb.lastPropsWereRIOdd = !lb.lastPropsWereRIOdd
		}

		goto done
	}

	// LB10 Treat any remaining combining mark or ZWJ as AL.
	// This runs at the end so it applies even if other rules short-circuit above.

	// LB11: Do not break before or after Word joiner and related characters.
	if lb.lastProp == lbPropWJ || prop == lbPropWJ {
		goto done
	}

	// LB12: Do not break after NBSP and related characters.
	if lb.lastProp == lbPropGL {
		goto done
	}

	// B12a: Do not break before NBSP and related characters, except after spaces and hyphens.
	if lb.lastProp != lbPropSP && lb.lastProp != lbPropBA && lb.lastProp != lbPropHY && prop == lbPropGL {
		goto done
	}

	// LB13: Do not break before ‘]’ or ‘!’ or ‘;’ or ‘/’, even after spaces.
	if prop == lbPropCL || prop == lbPropCP || prop == lbPropEX || prop == lbPropIS || prop == lbPropSY {
		goto done
	}

	// LB14: Do not break after ‘[’, even after spaces.
	if lb.inLeftBraceSpaceSeq && prop != lbPropSP {
		goto done
	}

	// LB15: Do not break within ‘”[’, even with intervening spaces.
	if lb.inQuotationSpaceSeq && prop == lbPropOP {
		goto done
	}

	// LB16: Do not break between closing punctuation and a nonstarter (lb=NS), even with intervening spaces.
	if lb.inClosePunctSpaceSeq && prop == lbPropNS {
		goto done
	}

	// LB17: Do not break within ‘——’, even with intervening spaces.
	if lb.inDashSpaceSeq && prop == lbPropB2 {
		goto done
	}

	// LB18: Break after spaces.
	if lb.lastProp == lbPropSP {
		decision = AllowLineBreakBefore
		goto done
	}

	// LB19: Do not break before or after quotation marks, such as ‘ ” ’.
	if lb.lastProp == lbPropQU || prop == lbPropQU {
		goto done
	}

	// LB20: Break before and after unresolved CB.
	if lb.lastProp == lbPropCB || prop == lbPropCB {
		decision = AllowLineBreakBefore
		goto done
	}

	// LB21: Do not break before hyphen-minus, other hyphens, fixed-width spaces, small kana, and other non-starters, or after acute accents.
	if prop == lbPropBA || prop == lbPropHY || prop == lbPropNS || lb.lastProp == lbPropBB {
		goto done
	}

	// LB21a: Don't break after Hebrew + Hyphen.
	if lb.lastLastProp == lbPropHL && (lb.lastProp == lbPropHY || lb.lastProp == lbPropBA) {
		goto done
	}

	// LB21b: Don’t break between Solidus and Hebrew letters.
	if lb.lastProp == lbPropSY && prop == lbPropHL {
		goto done
	}

	// LB22: Do not break before ellipses.
	if prop == lbPropIN {
		goto done
	}

	// LB23: Do not break between digits and letters.
	if ((lb.lastProp == lbPropAL || lb.lastProp == lbPropHL) && prop == lbPropNU) ||
		(lb.lastProp == lbPropNU && (prop == lbPropAL || prop == lbPropHL)) {
		goto done
	}

	// LB23a: Do not break between numeric prefixes and ideographs, or between ideographs and numeric postfixes.
	if (lb.lastProp == lbPropPR && (prop == lbPropID || prop == lbPropEB || prop == lbPropEM)) ||
		((lb.lastProp == lbPropID || lb.lastProp == lbPropEB || lb.lastProp == lbPropEM) && prop == lbPropPO) {
		goto done
	}

	// LB24: Do not break between numeric prefix/postfix and letters, or between letters and prefix/postfix.
	if ((lb.lastProp == lbPropPR || lb.lastProp == lbPropPO) && (prop == lbPropAL || prop == lbPropHL)) ||
		((lb.lastProp == lbPropAL || lb.lastProp == lbPropHL) && (prop == lbPropPR || prop == lbPropPO)) {
		goto done
	}

	// LB25: Do not break between the following pairs of classes relevant to numbers.
	// TR14 says this can be tailored for better results (see the "Regex-Number" rule in
	// Section 8.1, example 7), but for now we're using the simple version.
	if (lb.lastProp == lbPropCL && prop == lbPropPO) ||
		(lb.lastProp == lbPropCP && prop == lbPropPO) ||
		(lb.lastProp == lbPropCL && prop == lbPropPR) ||
		(lb.lastProp == lbPropCP && prop == lbPropPR) ||
		(lb.lastProp == lbPropNU && prop == lbPropPO) ||
		(lb.lastProp == lbPropNU && prop == lbPropPR) ||
		(lb.lastProp == lbPropPO && prop == lbPropOP) ||
		(lb.lastProp == lbPropPO && prop == lbPropNU) ||
		(lb.lastProp == lbPropPR && prop == lbPropOP) ||
		(lb.lastProp == lbPropPR && prop == lbPropNU) ||
		(lb.lastProp == lbPropHY && prop == lbPropNU) ||
		(lb.lastProp == lbPropIS && prop == lbPropNU) ||
		(lb.lastProp == lbPropNU && prop == lbPropNU) ||
		(lb.lastProp == lbPropSY && prop == lbPropNU) {
		goto done
	}

	// LB26: Do not break a Korean syllable.
	if (lb.lastProp == lbPropJL && (prop == lbPropJL || prop == lbPropJV || prop == lbPropH2 || prop == lbPropH3)) ||
		((lb.lastProp == lbPropJV || lb.lastProp == lbPropH2) && (prop == lbPropJV || prop == lbPropJT)) ||
		((lb.lastProp == lbPropJT || lb.lastProp == lbPropH3) && prop == lbPropJT) {
		goto done
	}

	// LB27: Treat a Korean Syllable Block the same as ID.
	if ((lb.lastProp == lbPropJL || lb.lastProp == lbPropJV || lb.lastProp == lbPropJT || lb.lastProp == lbPropH2 || lb.lastProp == lbPropH3) && prop == lbPropPO) ||
		(lb.lastProp == lbPropPR && (prop == lbPropJL || prop == lbPropJV || prop == lbPropJT || prop == lbPropH2 || prop == lbPropH3)) {
		goto done
	}

	// LB28 Do not break between alphabetics (“at”).
	if (lb.lastProp == lbPropAL || lb.lastProp == lbPropHL) && (prop == lbPropAL || prop == lbPropHL) {
		goto done
	}

	// LB29: Do not break between numeric punctuation and alphabetics (“e.g.”).
	if lb.lastProp == lbPropIS && (prop == lbPropAL || prop == lbPropHL) {
		goto done
	}

	// LB30: Do not break between letters, numbers, or ordinary symbols and opening or closing parentheses.
	if ((lb.lastProp == lbPropAL || lb.lastProp == lbPropHL || lb.lastProp == lbPropNU) && prop == lbPropOP) ||
		(lb.lastProp == lbPropCP && (prop == lbPropAL || prop == lbPropHL || prop == lbPropNU)) {
		eaProp := eaPropForRune(r)
		if eaProp != eaPropF && eaProp != eaPropW && eaProp != eaPropH {
			goto done
		}
	}

	// LB30a: Break between two regional indicator symbols if and only if there are an even number of regional indicators preceding the position of the break.
	if lb.lastPropsWereRIOdd && prop == lbPropRI {
		goto done
	}

	// LB30b: Do not break between an emoji base (or potential emoji) and an emoji modifier.
	if (lb.lastProp == lbPropEB && prop == lbPropEM) ||
		(lb.lastWasExtendedPictographicAndCn && prop == lbPropEM) {
		goto done
	}

	// LB31: Break everywhere else.
	decision = AllowLineBreakBefore

done:
	// This is LB10, which we run at the end so it applies even if other rules short-circuit.
	if prop == lbPropCM || (prop == lbPropZWJ && lb.lastProp != lbPropNone) {
		prop = lbPropAL
	}

	lb.lastLastProp = lb.lastProp
	lb.lastProp = prop
	lb.inZeroWidthSpaceSeq = bool(prop == lbPropZW || (lb.inZeroWidthSpaceSeq && prop == lbPropSP))
	lb.inLeftBraceSpaceSeq = bool(prop == lbPropOP || (lb.inLeftBraceSpaceSeq && prop == lbPropSP))
	lb.inQuotationSpaceSeq = bool(prop == lbPropQU || (lb.inQuotationSpaceSeq && prop == lbPropSP))
	lb.inClosePunctSpaceSeq = bool((prop == lbPropCL || prop == lbPropCP) || (lb.inClosePunctSpaceSeq && prop == lbPropSP))
	lb.inDashSpaceSeq = bool(prop == lbPropB2 || (lb.inDashSpaceSeq && prop == lbPropSP))
	lb.lastPropsWereRIOdd = bool(prop == lbPropRI && !lb.lastPropsWereRIOdd)
	lb.lastWasExtendedPictographicAndCn = bool(emPropForRune(r) == emPropExtended_Pictographic && gcPropForRune(r) == gcPropCn)
	return decision
}

// LineWrapConfig controls how lines should be soft-wrapped.
type LineWrapConfig struct {
	MaxLineWidth    uint64 // Maximum number of cells per line, which must be at least one.
	AllowCharBreaks bool   // Allow breaks at grapheme cluster boundaries.
	CellWidthSizer  cellwidth.Sizer
}

// WrappedLineIter iterates through soft- and hard-wrapped lines.
type WrappedLineIter struct {
	wrapConfig LineWrapConfig
	textTree   *text.Tree
	pos        uint64
	gc         []rune
}

// NewWrappedLineIter constructs a segment iterator for soft- and hard-wrapped lines.
func NewWrappedLineIter(wrapConfig LineWrapConfig, textTree *text.Tree, startPos uint64) WrappedLineIter {
	if wrapConfig.MaxLineWidth == 0 {
		panic("MaxLineWidth must be greater than zero")
	}

	return WrappedLineIter{
		wrapConfig: wrapConfig,
		textTree:   textTree,
		pos:        startPos,
		gc:         make([]rune, 0, 4), // allocate once and reuse for all lines.
	}
}

// NextSegment retrieves the next soft- or hard-wrapped line.
// For hard-wrapped lines, the grapheme cluster containing the newline will be included at the end of the line.
// If a segment is too long to fit on any line, put it in its own line.
func (iter *WrappedLineIter) NextSegment(segment *Segment) error {
	segment.Clear()

	lineBreakPos, err := iter.lookaheadLineBreakPos()
	if err != nil {
		return err
	}

	reader := iter.textTree.ReaderAtPosition(iter.pos)
	for iter.pos < lineBreakPos {
		r, _, err := reader.ReadRune()
		if err != nil {
			return err
		}
		segment.Append(r)
		iter.pos++
	}

	return nil
}

func (iter *WrappedLineIter) lookaheadLineBreakPos() (uint64, error) {
	var (
		gcBreaker   GraphemeClusterBreaker
		lineWidth   uint64
		lineBreaker LineBreaker
	)
	iter.gc = iter.gc[:0]
	pos := iter.pos
	lineBreakPos := iter.pos
	reader := iter.textTree.ReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err == io.EOF && pos > iter.pos {
			lineBreakPos = pos
			break
		} else if err != nil {
			return 0, err
		}

		canBreakBeforeGc := gcBreaker.ProcessRune(r)
		if canBreakBeforeGc {
			lineWidth += iter.wrapConfig.CellWidthSizer.GraphemeClusterWidth(iter.gc, lineWidth)
			// Check if we've exceeded the max line width. If so, exit the loop and return
			// the last breakpoint.
			// If the rune is '\r' or '\n', continue so LineBreaker can hard-wrap on the next loop iteration.
			if lineWidth >= iter.wrapConfig.MaxLineWidth && r != '\r' && r != '\n' {
				if lineBreakPos == iter.pos || iter.wrapConfig.AllowCharBreaks {
					// Break at the last grapheme cluster boundary if:
					// 1) The user has configured lineWrap="character", or
					// 2) There is no other break opportunity within maxLineWidth.
					lineBreakPos = pos
				}
				break
			}
			iter.gc = append(iter.gc[:0], r)
		} else {
			iter.gc = append(iter.gc, r)
		}

		decision := lineBreaker.ProcessRune(r)
		if decision == AllowLineBreakBefore {
			lineBreakPos = pos
		} else if decision == RequireLineBreakBefore {
			lineBreakPos = pos
			break
		} else if decision == RequireLineBreakAfter {
			lineBreakPos = pos + 1
			break
		}

		pos++
	}

	return lineBreakPos, nil
}
