package utf8

type state uint8

const (
	stateValid = state(iota)
	stateInvalid
	stateAwaitingOneByte
	stateAwaitingTwoBytesA
	stateAwaitingTwoBytesB
	stateAwaitingTwoBytesC
	stateAwaitingThreeBytesA
	stateAwaitingThreeBytesB
	stateAwaitingThreeBytesC
)

// Validator checks whether a byte string is valid UTF-8 text.
// It rejects invalid start bytes, missing continuation bytes, surrogate code points, overlong byte sequences, and 4-byte sequences outside the Unicode range.
type Validator struct {
	state state
}

func NewValidator() *Validator {
	return &Validator{
		state: stateValid,
	}
}

// ValidateBytes checks if appending bytes to the bytes processed so far would produce invalid UTF-8 text.
func (v *Validator) ValidateBytes(buf []byte) bool {
	// If all bytes are ASCII (and we're not in the middle of a multi-byte sequence), then we're done.  Try this first because it's the common case and quick to verify.
	if v.state == stateValid && isAscii(buf) {
		return true
	}

	// Fallback to slow path for non-ASCII.
	for _, b := range buf {
		v.processByte(b)
	}

	return v.state != stateInvalid
}

// ValidateEnd checks whether the bytes are valid UTF-8 once all bytes have been processed.
func (v *Validator) ValidateEnd() bool {
	return v.state == stateValid
}

func (v *Validator) processByte(b byte) {
	// This implements the state machine as described in http://bjoern.hoehrmann.de/utf-8/decoder/dfa/
	switch v.state {

	case stateValid:
		if b >= 0x00 && b <= 0x7f {
			v.state = stateValid
		} else if b >= 0xc2 && b <= 0xdf {
			v.state = stateAwaitingOneByte
		} else if (b >= 0xe1 && b <= 0xec) || (b >= 0xee && b <= 0xef) {
			v.state = stateAwaitingTwoBytesA
		} else if b == 0xe0 {
			v.state = stateAwaitingTwoBytesB
		} else if b == 0xed {
			v.state = stateAwaitingTwoBytesC
		} else if b == 0xf0 {
			v.state = stateAwaitingThreeBytesA
		} else if b >= 0xf1 && b <= 0xf3 {
			v.state = stateAwaitingThreeBytesB
		} else if b == 0xf4 {
			v.state = stateAwaitingThreeBytesC
		} else {
			v.state = stateInvalid
		}

	case stateAwaitingOneByte:
		if b >= 0x80 && b <= 0xbf {
			v.state = stateValid
		} else {
			v.state = stateInvalid
		}

	case stateAwaitingTwoBytesA:
		if b >= 0x80 && b <= 0xbf {
			v.state = stateAwaitingOneByte
		} else {
			v.state = stateInvalid
		}

	case stateAwaitingTwoBytesB:
		if b >= 0xa0 && b <= 0xbf {
			v.state = stateAwaitingOneByte
		} else {
			v.state = stateInvalid
		}

	case stateAwaitingTwoBytesC:
		if b >= 0x80 && b <= 0x9f {
			v.state = stateAwaitingOneByte
		} else {
			v.state = stateInvalid
		}

	case stateAwaitingThreeBytesA:
		if b >= 0x90 && b <= 0xbf {
			v.state = stateAwaitingTwoBytesA
		} else {
			v.state = stateInvalid
		}

	case stateAwaitingThreeBytesB:
		if b >= 0x80 && b <= 0xbf {
			v.state = stateAwaitingTwoBytesA
		} else {
			v.state = stateInvalid
		}

	case stateAwaitingThreeBytesC:
		if b >= 0x80 && b <= 0x8f {
			v.state = stateAwaitingTwoBytesA
		} else {
			v.state = stateInvalid
		}

	default:
		v.state = stateInvalid
	}
}

func isAscii(buf []byte) bool {
	var x byte
	for _, b := range buf {
		x |= (b & 0x80)
	}
	return x == 0
}
