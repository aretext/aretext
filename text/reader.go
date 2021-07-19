package text

import (
	"errors"
	"io"
)

// CloneableReader is an io.Reader that can be cloned to produce a new, independent reader at the same position.
type CloneableReader interface {
	io.Reader

	// Clone creates a new, independent reader at the same position as the original reader.
	Clone() CloneableReader
}

// stringReader implements CloneableReader for a string.
// It is useful for testing.
type stringReader struct {
	s      string
	offset int
}

func NewCloneableReaderFromString(s string) CloneableReader {
	return &stringReader{s, 0}
}

func (r *stringReader) Read(b []byte) (int, error) {
	if r.offset == len(r.s) {
		return 0, io.EOF
	}

	n := copy(b, r.s[r.offset:])
	r.offset += n
	return n, nil
}

func (r *stringReader) Clone() CloneableReader {
	return &stringReader{
		s:      r.s,
		offset: r.offset,
	}
}

type ReadDirection int

const (
	ReadDirectionForward = ReadDirection(iota)
	ReadDirectionBackward
)

func (d ReadDirection) Reverse() ReadDirection {
	switch d {
	case ReadDirectionForward:
		return ReadDirectionBackward
	case ReadDirectionBackward:
		return ReadDirectionForward
	default:
		panic("invalid direction")
	}
}

func (d ReadDirection) String() string {
	switch d {
	case ReadDirectionForward:
		return "forward"
	case ReadDirectionBackward:
		return "backward"
	default:
		panic("invalid direction")
	}
}

// TreeReader reads UTF-8 bytes from a text.Tree.
// It implements io.Reader and CloneableReader.
// text.Tree is NOT thread-safe, so reading from a tree while modifying it is undefined behavior!
type TreeReader struct {
	group          *leafNodeGroup
	nodeIdx        uint64
	textByteOffset uint64
	direction      ReadDirection
}

func newTreeReader(group *leafNodeGroup, nodeIdx uint64, textByteOffset uint64, direction ReadDirection) *TreeReader {
	return &TreeReader{
		group:          group,
		nodeIdx:        nodeIdx,
		textByteOffset: textByteOffset,
		direction:      direction,
	}
}

// Read implements io.Reader#Read
func (r *TreeReader) Read(b []byte) (int, error) {
	if r.direction == ReadDirectionBackward {
		return r.readBackward(b)
	}
	return r.readForward(b)
}

func (r *TreeReader) readForward(b []byte) (int, error) {
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
		r.textByteOffset += uint64(bytesWritten)
		i += bytesWritten

		if r.textByteOffset == uint64(node.numBytes) {
			r.nodeIdx++
			r.textByteOffset = 0
		}

		if r.nodeIdx == r.group.numNodes && r.group.next != nil {
			r.group = r.group.next
			r.nodeIdx = 0
			r.textByteOffset = 0
		}
	}
}

func (r *TreeReader) readBackward(b []byte) (int, error) {
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

// Clone implements CloneableReader#Clone
func (r *TreeReader) Clone() CloneableReader {
	return &TreeReader{
		group:          r.group,
		nodeIdx:        r.nodeIdx,
		textByteOffset: r.textByteOffset,
		direction:      r.direction,
	}
}

// SeekBackward implements parser.InputReader#SeekBackward
func (r *TreeReader) SeekBackward(offset uint64) error {
	if r.direction != ReadDirectionForward {
		return errors.New("SeekBackward is implemented only for readers with direction ReadDirectionForward")
	}

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
