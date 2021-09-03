package text

import (
	"io"
	"unicode/utf8"

	textUtf8 "github.com/aretext/aretext/text/utf8"
)

// Reader reads UTF-8 bytes from a text.Tree.
// It implements io.Reader.
// Copying the struct produces a new, independent reader.
// text.Tree is NOT thread-safe, so reading from a tree while modifying it is undefined behavior!
type Reader struct {
	group          *leafNodeGroup
	nodeIdx        uint64
	textByteOffset uint64
}

// Read implements io.Reader#Read
func (r *Reader) Read(b []byte) (int, error) {
	i := 0
	for {
		if i == len(b) {
			return i, nil
		}

		if r.group.next == nil && r.nodeIdx == r.group.numNodes {
			return i, io.EOF
		}

		node := &r.group.nodes[r.nodeIdx]
		bytesWritten := copy(b[i:], node.textBytes[r.textByteOffset:node.numBytes])
		i += bytesWritten
		r.advance(uint64(bytesWritten))
	}
}

func (r *Reader) advance(n uint64) {
	// Assumes that there are at least n bytes in the current leaf.
	r.textByteOffset += n
	if r.textByteOffset == uint64(r.group.nodes[r.nodeIdx].numBytes) {
		r.nodeIdx++
		r.textByteOffset = 0
	}
	if r.nodeIdx == r.group.numNodes && r.group.next != nil {
		r.group = r.group.next
		r.nodeIdx = 0
		r.textByteOffset = 0
	}
}

func (r *Reader) readNextByte() (byte, error) {
	// Fast path: next byte is in current leaf.
	if r.nodeIdx < r.group.numNodes && r.textByteOffset < uint64(r.group.nodes[r.nodeIdx].numBytes) {
		b := r.group.nodes[r.nodeIdx].textBytes[r.textByteOffset]
		r.advance(1)
		return b, nil
	}

	// Slow path: fallback to default read.
	var buf [1]byte
	_, err := r.Read(buf[:])
	return buf[0], err
}

// ReadRune implements io.RuneReader#ReadRune
// If the next bytes in the reader are not valid UTF8, it returns InvalidUtf8Error.
// If there are no more bytes to read, it returns io.EOF.
func (r *Reader) ReadRune() (rune, int, error) {
	var buf [4]byte

	// Read the next byte to determine the number of bytes in the next rune.
	firstByte, err := r.readNextByte()
	if err != nil {
		return '\x00', 0, err
	}

	n := textUtf8.CharWidth[firstByte]
	if n == 0 {
		return '\x00', 0, InvalidUtf8Error
	} else if n == 1 {
		// Fast path for ASCII.
		return rune(firstByte), 1, nil
	}

	// Read remaining bytes in the rune.
	buf[0] = firstByte
	if _, err := r.Read(buf[1:n]); err != nil {
		return '\x00', 0, InvalidUtf8Error
	}

	// Decode the multi-byte rune.
	rn, sz := utf8.DecodeRune(buf[:n])
	if sz != int(n) {
		return '\x00', 0, InvalidUtf8Error
	}
	return rn, sz, nil
}

// SeekBackward implements parser.InputReader#SeekBackward
func (r *Reader) SeekBackward(offset uint64) error {
	for offset > 0 && !(r.group.prev == nil && r.nodeIdx == 0 && r.textByteOffset == 0) {
		if r.textByteOffset >= offset {
			r.textByteOffset -= offset
			break
		} else if r.textByteOffset > 0 {
			offset -= r.textByteOffset
			r.textByteOffset = 0
			continue
		}

		if r.nodeIdx > 0 {
			r.nodeIdx--
			r.textByteOffset = uint64(r.group.nodes[r.nodeIdx].numBytes)
			continue
		}

		if r.group.prev != nil {
			r.group = r.group.prev
			r.nodeIdx = r.group.numNodes - 1
			r.textByteOffset = uint64(r.group.nodes[r.nodeIdx].numBytes)
		}
	}

	return nil
}

// ReverseReader reads bytes in reverse order.
type ReverseReader struct {
	Reader
}

// Read implements io.Reader#Read
func (r *ReverseReader) Read(b []byte) (int, error) {
	i := 0
	for {
		if i == len(b) {
			return i, nil
		}

		if r.group.prev == nil && r.nodeIdx == 0 && r.textByteOffset == 0 {
			return i, io.EOF
		}

		node := &r.group.nodes[r.nodeIdx]
		bytesWritten := 0
		for i+bytesWritten < len(b) && r.textByteOffset > uint64(bytesWritten) {
			b[i+bytesWritten] = node.textBytes[r.textByteOffset-1-uint64(bytesWritten)]
			bytesWritten++
		}
		r.textByteOffset -= uint64(bytesWritten)
		i += bytesWritten

		if r.textByteOffset > 0 {
			continue
		}

		if r.nodeIdx > 0 {
			r.nodeIdx--
			r.textByteOffset = uint64(r.group.nodes[r.nodeIdx].numBytes)
			continue
		}

		if r.group.prev != nil {
			r.group = r.group.prev
			r.nodeIdx = r.group.numNodes - 1
			r.textByteOffset = uint64(r.group.nodes[r.nodeIdx].numBytes)
		}
	}
}

// ReadRune implements io.RuneReader#ReadRune
func (r *ReverseReader) ReadRune() (rune, int, error) {
	n, err := r.lookaheadToRuneStartByte()
	if err != nil {
		return '\x00', 0, err
	}

	var buf [4]byte
	if _, err := r.Read(buf[:n]); err != nil {
		return '\x00', 0, err
	}

	// Bytes were read in reverse order, so we need to swap them to decode as UTF-8.
	if n == 2 {
		buf[0], buf[1] = buf[1], buf[0]
	} else if n == 3 {
		buf[0], buf[2] = buf[2], buf[0]
	} else if n == 4 {
		buf[0], buf[3] = buf[3], buf[0]
		buf[1], buf[2] = buf[2], buf[1]
	}

	rn, sz := utf8.DecodeRune(buf[:n])
	if sz != n {
		return '\x00', 0, InvalidUtf8Error
	}
	return rn, sz, nil
}

func (r *ReverseReader) lookaheadToRuneStartByte() (int, error) {
	rcopy := *r     // Copy the struct to produce a new, independent reader for lookahead.
	var buf [4]byte // At most 4 bytes to the start of the next rune in valid UTF-8 encoding.
	n, _ := rcopy.Read(buf[:])
	if n == 0 {
		return 0, io.EOF
	}

	for i := 0; i < n; i++ {
		if textUtf8.StartByteIndicator[buf[i]] > 0 {
			// Found the start byte.
			return i + 1, nil
		}
	}

	// Could not find the start byte, so this is not a valid UTF-8 encoding.
	return 0, InvalidUtf8Error
}
