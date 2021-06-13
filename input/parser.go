package input

import (
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ParseResult is the result of parsing an input key event.
type ParseResult struct {
	// Accepted indicates whether the input sequence was accepted.
	// The following fields are set only if Accepted is true.
	Accepted bool

	// Rule is the rule triggered by the sequence of input key events.
	Rule Rule

	// Input is the sequence of key events that triggered the rule.
	Input []*tcell.EventKey

	// Count is the parsed count parameter, if provided.
	// For example, `5x` (delete five characters) would have count set to 5.
	// If provided, the count will always be at least one and at most math.MaxInt64.
	// If no count parameter was provided, this will be nil.
	Count *uint64

	// ClipboardPageName is the name of the page to copy or paste.
	// For example `"ap` means "paste clipboard page named 'a'"
	// If not provided, this will be nil.
	ClipboardPageName *rune
}

// maxParseInputLen is the maximum number of input key events that can be recognized.
// If more events are received without either accepting or rejecting the input,
// the parser will reject and reset.
const maxParseInputLen = 64

// parserState describes the internal state of the parser.
type parserState int

const (
	parserStateStart         = parserState(iota) // initial state.
	parserStateCountPrefix                       // parsing the "count" parameter at the start of the command.
	parserStateClipboardPage                     // parsing the clipboard page name at the start of the command.
	parserStateCommand                           // parsing the command itself.
)

// candidateState describes the state of a particular candidate rule that could accept the input.
type candidateState struct {
	ruleIdx    int // index of the rule associated with this candidate.
	patternIdx int // index of the next pattern within the rule to match the next input key.
}

// Parser parses a sequence of input key events based on a set of rules.
// It parses input incrementally, waiting for a key that will trigger some rule.
// If the input is rejected by all rules, the parser resets.
type Parser struct {
	state             parserState
	rules             []Rule
	candidates        []candidateState
	inputBuffer       []*tcell.EventKey
	countDigits       []rune
	clipboardPageName *rune
}

// NewParser constructs a new parser for a set of rules.
func NewParser(rules []Rule) *Parser {
	candidates := make([]candidateState, len(rules))
	for i := 0; i < len(rules); i++ {
		candidates[i].ruleIdx = i
	}
	return &Parser{
		state:       parserStateStart,
		rules:       rules,
		candidates:  candidates,
		inputBuffer: make([]*tcell.EventKey, 0),
	}
}

// InputBufferString returns a string describing buffered input events.
// This can be displayed to the user to help them understand the input state.
func (p *Parser) InputBufferString() string {
	var sb strings.Builder
	for _, e := range p.inputBuffer {
		if e.Key() == tcell.KeyRune {
			sb.WriteRune(e.Rune())
		}
	}
	return sb.String()
}

// ProcessInput processes an input event key.
// If a rule accepts the sequence of key presses, the parser
// returns the accepted sequence, triggered rule, and parsed count parameter.
// If all rules reject the sequence, the parser resets.
func (p *Parser) ProcessInput(event *tcell.EventKey) ParseResult {
	if len(p.inputBuffer) >= maxParseInputLen {
		// Enforce a max length on the input buffer to bound memory usage.
		p.reset()
	}

	p.inputBuffer = append(p.inputBuffer, event)
	for {
		switch p.state {
		case parserStateStart:
			if isDigit(event) && !isZeroDigit(event) {
				// Digits 1-9 at the start of a sequence are parsed as the "count" parameter.
				// Note that the character '0' is treated as a command (cursor to start of line), not a count.
				p.state = parserStateCountPrefix
			} else if isQuote(event) {
				p.state = parserStateClipboardPage
				return ParseResult{Accepted: false}
			} else {
				p.state = parserStateCommand
			}
		case parserStateCountPrefix:
			if isDigit(event) {
				p.countDigits = append(p.countDigits, event.Rune())
				return ParseResult{Accepted: false}
			} else {
				p.state = parserStateStart
			}
		case parserStateClipboardPage:
			if event.Key() == tcell.KeyRune {
				r := event.Rune()
				p.clipboardPageName = &r
			}
			p.state = parserStateStart
			return ParseResult{Accepted: false}
		case parserStateCommand:
			return p.processCommandInput(event)
		}
	}
}

func (p *Parser) processCommandInput(event *tcell.EventKey) ParseResult {
	var i int
	for j := 0; j < len(p.candidates); j++ {
		c := &(p.candidates[j])
		accept, reject := p.processCandidate(c, event)
		if reject {
			// The candidate rejected the input, so we remove it from the current set of candidates.
			continue
		} else if accept {
			// The candidate accepted the input, so we accept the input and reset the parser.
			rule := p.rules[c.ruleIdx]
			input := append([]*tcell.EventKey{}, p.inputBuffer...)
			count := p.calculateCount() // This checks for overflow.
			clipboardPageName := p.clipboardPageName
			p.reset()
			return ParseResult{
				Accepted:          true,
				Rule:              rule,
				Input:             input,
				Count:             count,
				ClipboardPageName: clipboardPageName,
			}
		} else {
			// The candidate neither accepted nor rejected the input, so we keep the candidate
			// in the current set of candidates.
			p.candidates[i] = p.candidates[j]
			i++
		}
	}

	// Truncate the list of candidates to the ones we preserved earlier.
	p.candidates = p.candidates[:i]

	// If all candidates rejected, the parser resets so it can accept subsequent input.
	if len(p.candidates) == 0 {
		p.reset()
	}
	return ParseResult{Accepted: false}
}

func (p *Parser) processCandidate(c *candidateState, event *tcell.EventKey) (accept, reject bool) {
	r := p.rules[c.ruleIdx]
	matcher := r.Pattern[c.patternIdx]
	if !matcher.Matches(event) {
		reject = true
		return
	}

	c.patternIdx++
	if c.patternIdx >= len(r.Pattern) {
		accept = true
		return
	}

	return false, false
}

func (p *Parser) reset() {
	p.state = parserStateStart
	p.inputBuffer = p.inputBuffer[:0]
	p.countDigits = p.countDigits[:0]
	p.clipboardPageName = nil
	p.candidates = p.candidates[:0]
	for i := 0; i < len(p.rules); i++ {
		p.candidates = append(p.candidates, candidateState{
			ruleIdx:    i,
			patternIdx: 0,
		})
	}
}

func (p *Parser) calculateCount() *uint64 {
	if len(p.countDigits) == 0 {
		return nil
	}

	var count int64
	for _, d := range p.countDigits {
		count = (count * 10) + int64(d-'0')
		if count < 0 {
			// overflow
			count = math.MaxInt64
			break
		}
	}

	unsignedCount := uint64(count) // safe because of overflow check.
	return &unsignedCount
}

func isDigit(event *tcell.EventKey) bool {
	return event.Key() == tcell.KeyRune && event.Rune() >= '0' && event.Rune() <= '9'
}

func isZeroDigit(event *tcell.EventKey) bool {
	return event.Key() == tcell.KeyRune && event.Rune() == '0'
}

func isQuote(event *tcell.EventKey) bool {
	return event.Key() == tcell.KeyRune && event.Rune() == '"'
}
