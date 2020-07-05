package text

import (
	"io"
)

// CloneableReader is an io.Read that can be cloned to produce a new, independent reader at the same position.
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

// treeReader reads UTF-8 bytes from a text.Tree.
// It implements io.Reader and CloneableReader.
// text.Tree is NOT thread-safe, so reading from a tree while modifying it is undefined behavior!
type treeReader struct {
	group          *leafNodeGroup
	nodeIdx        uint64
	textByteOffset uint64
}

// Read implements io.Reader#Read
func (r *treeReader) Read(b []byte) (int, error) {
	i := 0
	for {
		if i == len(b) {
			return i, nil
		}

		if r.group == nil {
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

		if r.nodeIdx == r.group.numNodes {
			r.group = r.group.next
			r.nodeIdx = 0
			r.textByteOffset = 0
		}
	}
}

// Clone implements CloneableReader#Clone
func (r *treeReader) Clone() CloneableReader {
	return &treeReader{
		group:          r.group,
		nodeIdx:        r.nodeIdx,
		textByteOffset: r.textByteOffset,
	}
}
