package text

import (
	"io"
	"log"
	"unicode/utf8"

	textUtf8 "github.com/aretext/aretext/text/utf8"
)

// RuneIter iterates over unicode codepoints (runes).
type RuneIter interface {
	// NextRune returns the next available rune.  If no rune is available, it returns the error io.EOF.
	NextRune() (rune, error)
}

// CloneableRuneIter is a RuneIter that can be cloned to produce a new, independent iterator at the same position as the original iterator.
type CloneableRuneIter interface {
	RuneIter

	// Clone returns a new, independent iterator at the same position as the original iterator.
	Clone() CloneableRuneIter
}

// runeIterForSlice iterates over the runes in a slice.
type runeIterForSlice struct {
	idx       int
	runeSlice []rune
}

// NewRuneIterForSlice returns a RuneIter over the given slice.
// This assumes that runeSlice will be immutable for the lifetime of the iterator.
func NewRuneIterForSlice(runeSlice []rune) CloneableRuneIter {
	return &runeIterForSlice{0, runeSlice}
}

// NextRune implements RuneIter#NextRune()
func (iter *runeIterForSlice) NextRune() (rune, error) {
	if iter.idx >= len(iter.runeSlice) {
		iter.runeSlice = nil
		return '\x00', io.EOF
	}

	r := iter.runeSlice[iter.idx]
	iter.idx++
	return r, nil
}

// Clone implements CloneableRuneIter#Clone()
func (iter *runeIterForSlice) Clone() CloneableRuneIter {
	return &runeIterForSlice{
		idx:       iter.idx,
		runeSlice: iter.runeSlice,
	}
}

// decodingRuneIter implements CloneableRuneIter for a stream of UTF-8 bytes.
type decodingRuneIter struct {
	inputReversed      bool
	in                 CloneableReader
	pendingRunes       []rune
	pendingRunesOffset int
	overflow           [3]byte
	overflowLen        int
	eof                bool
	buf                [64]byte
}

// NewCloneableForwardRuneIter creates a CloneableRuneIter for a stream of UTF-8 bytes.
// It assumes the provided reader produces a stream of valid UTF-8 bytes.
func NewCloneableForwardRuneIter(in CloneableReader) CloneableRuneIter {
	return &decodingRuneIter{
		inputReversed: false,
		in:            in,
		pendingRunes:  make([]rune, 0),
	}
}

// If inputReversed is true, then it interprets the reader output in reverse order.
func NewCloneableBackwardRuneIter(in CloneableReader) CloneableRuneIter {
	return &decodingRuneIter{
		inputReversed: true,
		in:            in,
		pendingRunes:  make([]rune, 0),
	}
}

// NextRune implements RuneIter#NextRune.
// It exits with an error if the input bytes contain invalid UTF-8 codepoints.
func (ri *decodingRuneIter) NextRune() (rune, error) {
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

func (ri *decodingRuneIter) loadRunesFromReader() error {
	for len(ri.pendingRunes) == 0 && !ri.eof {
		n := copy(ri.buf[:], ri.overflow[:ri.overflowLen])

		numRead, err := ri.in.Read(ri.buf[ri.overflowLen:])
		if err == io.EOF {
			ri.eof = true
		} else if err != nil {
			return err
		}

		n += numRead

		var bytesConsumed int
		if ri.inputReversed {
			bytesConsumed = ri.loadRunesFromBufferReverseOrder(ri.buf[:n])
		} else {
			bytesConsumed = ri.loadRunesFromBuffer(ri.buf[:n])
		}

		ri.overflowLen = copy(ri.overflow[:], ri.buf[bytesConsumed:n])
	}

	return nil
}

func (ri *decodingRuneIter) loadRunesFromBuffer(buf []byte) (bytesConsumed int) {
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

func (ri *decodingRuneIter) loadRunesFromBufferReverseOrder(buf []byte) (bytesConsumed int) {
	for bytesConsumed < len(buf) {
		var charWidth int
		var nextRuneBytes [4]byte
		for i := 0; i < len(nextRuneBytes) && i+bytesConsumed < len(buf); i++ {
			b := buf[i+bytesConsumed]
			nextRuneBytes[i] = b
			charWidth = int(textUtf8.CharWidth[b])
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
			log.Fatalf("Invalid UTF-8")
		}

		ri.pendingRunes = append(ri.pendingRunes, r)
	}

	return bytesConsumed
}

// Clone implements CloneableRuneIter#Clone
func (ri *decodingRuneIter) Clone() CloneableRuneIter {
	pendingRunes := make([]rune, len(ri.pendingRunes))
	copy(pendingRunes, ri.pendingRunes)

	return &decodingRuneIter{
		in:                 ri.in.Clone(),
		inputReversed:      ri.inputReversed,
		pendingRunes:       pendingRunes,
		pendingRunesOffset: ri.pendingRunesOffset,
		overflow:           ri.overflow,
		overflowLen:        ri.overflowLen,
		eof:                ri.eof,
	}
}
