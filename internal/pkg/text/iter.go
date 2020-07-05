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

// forwardRuneIter implements CloneableRuneIter for a stream of UTF-8 bytes.
type forwardRuneIter struct {
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
	return &forwardRuneIter{
		in:                 in,
		pendingRunes:       make([]rune, 0),
		pendingRunesOffset: 0,
		overflow:           [3]byte{},
		overflowLen:        0,
		eof:                false,
	}
}

// NextRune implements CloneableRuneIter#NextRune.
// It panics if the input bytes contain invalid UTF-8 codepoints.
func (f *forwardRuneIter) NextRune() (rune, error) {
	if f.pendingRunesOffset >= len(f.pendingRunes) && !f.eof {
		f.pendingRunesOffset = 0
		f.pendingRunes = f.pendingRunes[:0]
		if err := f.loadRunesFromReader(); err != nil {
			return '\x00', err
		}
	}

	if f.pendingRunesOffset >= len(f.pendingRunes) && f.eof {
		return '\x00', io.EOF
	}

	r := f.pendingRunes[f.pendingRunesOffset]
	f.pendingRunesOffset++
	return r, nil
}

func (f *forwardRuneIter) loadRunesFromReader() error {
	var buf [64]byte
	for len(f.pendingRunes) == 0 && !f.eof {
		n := copy(buf[:], f.overflow[:f.overflowLen])

		numRead, err := f.in.Read(buf[f.overflowLen:])
		if err == io.EOF {
			f.eof = true
		} else if err != nil {
			return err
		}

		n += numRead
		i := 0
		for i < n {
			r, size := utf8.DecodeRune(buf[i:n])
			if r == utf8.RuneError && size == 1 {
				// We're assuming that the reader provides valid UTF-8 bytes,
				// so if there's a decoding error it means we need to read more bytes and try again.
				break
			}

			f.pendingRunes = append(f.pendingRunes, r)
			i += size
		}

		f.overflowLen = copy(f.overflow[:], buf[i:n])
	}

	return nil
}

// Clone implements CloneableRuneIter#Clone
func (f *forwardRuneIter) Clone() CloneableRuneIter {
	pendingRunes := make([]rune, 0, len(f.pendingRunes))
	copy(pendingRunes, f.pendingRunes)

	return &forwardRuneIter{
		in:                 f.in.Clone(),
		pendingRunes:       pendingRunes,
		pendingRunesOffset: f.pendingRunesOffset,
		overflow:           f.overflow,
		overflowLen:        f.overflowLen,
		eof:                f.eof,
	}
}
