package locate

import (
	"io"

	"github.com/aretext/aretext/editor/text"
	"github.com/aretext/aretext/editor/text/segment"
)

// NextParagraph locates the start of the next paragraph after the cursor.
// Paragraph boundaries occur at empty lines.
func NextParagraph(tree *text.Tree, pos uint64) uint64 {
	reader := tree.ReaderAtPosition(pos)
	segmentIter := segment.NewGraphemeClusterIter(reader)
	seg := segment.Empty()
	var prevWasNewlineFlag, nonNewlineFlag bool
	var offset, prevOffset uint64
	for {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF {
			// End of document.
			return pos + prevOffset
		} else if err != nil {
			panic(err)
		}

		if seg.HasNewline() {
			if prevWasNewlineFlag && nonNewlineFlag {
				// An empty line is a paragraph boundary.
				// Choose the first one after we see a non-newline.
				break
			}
			prevWasNewlineFlag = true
		} else {
			nonNewlineFlag = true
			prevWasNewlineFlag = false
		}

		prevOffset = offset
		offset += seg.NumRunes()
	}
	return pos + offset
}

// PrevParagraph locates the start of the first paragraph before the cursor.
// Paragraph boundaries occur at empty lines.
func PrevParagraph(tree *text.Tree, pos uint64) uint64 {
	reader := tree.ReverseReaderAtPosition(pos)
	segmentIter := segment.NewReverseGraphemeClusterIter(reader)
	seg := segment.Empty()
	var prevWasNewlineFlag, nonNewlineFlag bool
	var offset uint64
	for {
		err := segmentIter.NextSegment(seg)
		if err == io.EOF {
			// Start of the document.
			return 0
		} else if err != nil {
			panic(err)
		}

		if seg.HasNewline() {
			if prevWasNewlineFlag && nonNewlineFlag {
				// An empty line is a paragraph boundary.
				return pos - offset
			}
			prevWasNewlineFlag = true
		} else {
			prevWasNewlineFlag = false
			nonNewlineFlag = true
		}

		offset += seg.NumRunes()
	}
}
