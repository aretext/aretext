package text

import (
	"io"
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
	in            CloneableReader
	inputReversed bool
	buf           [64]byte
	buflen        int
	offset        int
}

// NewCloneableForwardRuneIter creates a CloneableRuneIter for a stream of UTF-8 bytes.
// It assumes the provided reader produces a stream of valid UTF-8 bytes.
func NewCloneableForwardRuneIter(in CloneableReader) CloneableRuneIter {
	return &decodingRuneIter{in: in}
}

// NewCloneableBackwardRuneIter creates a CloneableRuneIter for a stream of UTF-8 bytes in reverse order.
func NewCloneableBackwardRuneIter(in CloneableReader) CloneableRuneIter {
	return &decodingRuneIter{in: in, inputReversed: true}
}

// NextRune implements RuneIter#NextRune.
// It exits with an error if the input bytes contain invalid UTF-8 codepoints.
func (ri *decodingRuneIter) NextRune() (rune, error) {
	// Assuming that well-behaved readers produce at least one byte
	// for each read, and a UTF-8 encoded rune occupies at most 4 bytes,
	// we need at most 4 reads, plus one more iteration
	// to decode the bytes into a rune.
	for i := 0; i < 5; i++ {
		// Try to decode a rune from the current buffer.
		if ri.offset < ri.buflen {
			r, size := ri.decodeNextRuneFromBuffer()
			if size == 0 || (r == utf8.RuneError && size == 1) {
				// We're assuming that the reader provides valid UTF-8 bytes,
				// so if there's a decoding error we retry after reading more bytes.
			} else {
				ri.offset += size
				return r, nil
			}
		}

		// Refill the buffer.
		ri.buflen = copy(ri.buf[:], ri.buf[ri.offset:ri.buflen])
		ri.offset = 0
		n, err := ri.in.Read(ri.buf[ri.buflen:len(ri.buf)])
		ri.buflen += n
		if err == io.EOF && ri.buflen > 0 {
			continue
		} else if err != nil {
			return '\x00', err
		}
	}

	return '\x00', io.EOF
}

func (ri *decodingRuneIter) decodeNextRuneFromBuffer() (rune, int) {
	bytes := ri.buf[ri.offset:ri.buflen]
	if ri.inputReversed {
		return decodeRuneInputReversed(bytes)
	} else {
		return utf8.DecodeRune(bytes)
	}
}

func decodeRuneInputReversed(bytes []byte) (rune, int) {
	var size int
	for i := 0; i < 4 && i < len(bytes); i++ {
		b := bytes[i]
		if textUtf8.StartByteIndicator[b] > 0 {
			// Found a start byte.
			size = i + 1
			break
		}
	}

	// Bytes were read in reverse order, so swap them before decoding.
	switch size {
	case 4:
		bytes[0], bytes[3] = bytes[3], bytes[0]
		bytes[1], bytes[2] = bytes[2], bytes[1]
	case 3:
		bytes[0], bytes[2] = bytes[2], bytes[0]
	case 2:
		bytes[0], bytes[1] = bytes[1], bytes[0]
	}

	return utf8.DecodeRune(bytes[:size])
}

// Clone implements CloneableRuneIter#Clone
func (ri *decodingRuneIter) Clone() CloneableRuneIter {
	return &decodingRuneIter{
		in:            ri.in.Clone(),
		inputReversed: ri.inputReversed,
		buf:           ri.buf,
		buflen:        ri.buflen,
		offset:        ri.offset,
	}
}
