package text

import (
	"io"
	"unicode/utf8"
)

// CloneableRuneIter iterates over UTF-8 codepoints (runes).
// It can be cloned to produce a new, independent iterator at the same position
// as the original iterator.
type CloneableRuneIter interface {
	// NextRune returns the next available rune.  If no rune is available, it returns the error io.EOF.
	NextRune() (rune, error)

	// Clone returns a new, independent iterator at the same position as the original iterator.
	Clone() CloneableRuneIter
}

// runeIter implements CloneableRuneIter for a stream of UTF-8 bytes.
type runeIter struct {
	inputReversed      bool
	in                 CloneableReader
	pendingRunes       []rune
	pendingRunesOffset int
	overflow           [3]byte
	overflowLen        int
	eof                bool
}

// NewForwardRuneIter creates a CloneableRuneIter for a stream of UTF-8 bytes.
// It assumes the provided reader produces a stream of valid UTF-8 bytes.
func NewForwardRuneIter(in CloneableReader) CloneableRuneIter {
	return &runeIter{
		in:           in,
		pendingRunes: make([]rune, 0),
	}
}

// NewBackwardRuneIter creates a CloneableRuneIter for a stream of UTF-8 bytes received in reverse order.
// It iterates through the runes in reverse order.
func NewBackwardRuneIter(in CloneableReader) CloneableRuneIter {
	return &runeIter{
		inputReversed: true,
		in:            in,
		pendingRunes:  make([]rune, 0),
	}
}

// NextRune implements CloneableRuneIter#NextRune.
// It panics if the input bytes contain invalid UTF-8 codepoints.
func (ri *runeIter) NextRune() (rune, error) {
	if ri.pendingRunesOffset >= len(ri.pendingRunes) && !ri.eof {
		ri.pendingRunesOffset = 0
		ri.pendingRunes = ri.pendingRunes[:0]
		if err := ri.loadRunesFromReader(); err != nil {
			return '\x00', err
		}
	}

	if ri.pendingRunesOffset >= len(ri.pendingRunes) && ri.eof {
		return '\x00', io.EOF
	}

	r := ri.pendingRunes[ri.pendingRunesOffset]
	ri.pendingRunesOffset++
	return r, nil
}

func (ri *runeIter) loadRunesFromReader() error {
	var buf [64]byte
	for len(ri.pendingRunes) == 0 && !ri.eof {
		n := copy(buf[:], ri.overflow[:ri.overflowLen])

		numRead, err := ri.in.Read(buf[ri.overflowLen:])
		if err == io.EOF {
			ri.eof = true
		} else if err != nil {
			return err
		}

		n += numRead

		var bytesConsumed int
		if ri.inputReversed {
			bytesConsumed = ri.loadRunesFromBufferReverseOrder(buf[:n])
		} else {
			bytesConsumed = ri.loadRunesFromBuffer(buf[:n])
		}

		ri.overflowLen = copy(ri.overflow[:], buf[bytesConsumed:n])
	}

	return nil
}

func (ri *runeIter) loadRunesFromBuffer(buf []byte) (bytesConsumed int) {
	for bytesConsumed < len(buf) {
		r, size := utf8.DecodeRune(buf[bytesConsumed:])
		if r == utf8.RuneError && size == 1 {
			// We're assuming that the reader provides valid UTF-8 bytes,
			// so if there's a decoding error it means we need to read more bytes and try again.
			break
		}

		ri.pendingRunes = append(ri.pendingRunes, r)
		bytesConsumed += size
	}
	return bytesConsumed
}

func (ri *runeIter) loadRunesFromBufferReverseOrder(buf []byte) (bytesConsumed int) {
	for bytesConsumed < len(buf) {
		var charWidth int
		var nextRuneBytes [4]byte
		for i := 0; i < len(nextRuneBytes) && i+bytesConsumed < len(buf); i++ {
			b := buf[i+bytesConsumed]
			nextRuneBytes[i] = b
			charWidth = int(utf8CharWidth[b])
			if charWidth > 0 {
				// found a start byte
				bytesConsumed += charWidth
				break
			}
		}

		if charWidth == 0 {
			// failed to find a start byte, so wait for more bytes
			break
		}

		// Bytes were read in reverse order, so reverse them before decoding the rune.
		switch charWidth {
		case 4:
			nextRuneBytes[0], nextRuneBytes[3] = nextRuneBytes[3], nextRuneBytes[0]
			nextRuneBytes[1], nextRuneBytes[2] = nextRuneBytes[2], nextRuneBytes[1]

		case 3:
			nextRuneBytes[0], nextRuneBytes[2] = nextRuneBytes[2], nextRuneBytes[0]

		case 2:
			nextRuneBytes[0], nextRuneBytes[1] = nextRuneBytes[1], nextRuneBytes[0]
		}

		// Now that the rune bytes are in sequential order, we can decode the rune.
		r, size := utf8.DecodeRune(nextRuneBytes[:charWidth])
		if r == utf8.RuneError && size == 1 {
			panic("Invalid UTF-8")
		}

		ri.pendingRunes = append(ri.pendingRunes, r)
	}

	return bytesConsumed
}

// Clone implements CloneableRuneIter#Clone
func (ri *runeIter) Clone() CloneableRuneIter {
	pendingRunes := make([]rune, 0, len(ri.pendingRunes))
	copy(pendingRunes, ri.pendingRunes)

	return &runeIter{
		in:                 ri.in.Clone(),
		inputReversed:      ri.inputReversed,
		pendingRunes:       pendingRunes,
		pendingRunesOffset: ri.pendingRunesOffset,
		overflow:           ri.overflow,
		overflowLen:        ri.overflowLen,
		eof:                ri.eof,
	}
}
