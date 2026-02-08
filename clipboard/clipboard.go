package clipboard

import (
	"bytes"
	"io"
)

// PageId represents a page in the clipboard.
// This is equivalent to what vim calls a "register".
type PageId int

const (
	// Data written to the null page is discarded.
	PageNull = PageId(iota)

	// The default page stores the contents of the most recent delete or yank operation.
	PageDefault

	// Named pages "a" through "z".
	PageLetterA
	PageLetterB
	PageLetterC
	PageLetterD
	PageLetterE
	PageLetterF
	PageLetterG
	PageLetterH
	PageLetterI
	PageLetterJ
	PageLetterK
	PageLetterL
	PageLetterM
	PageLetterN
	PageLetterO
	PageLetterP
	PageLetterQ
	PageLetterR
	PageLetterS
	PageLetterT
	PageLetterU
	PageLetterV
	PageLetterW
	PageLetterX
	PageLetterY
	PageLetterZ
)

// PageIdForInputRune returns the page named by a rune typed after '"'
// (e.g. `"ay" to yank from page "a").
// If the input rune does not match any clipboard page, return PageNull.
func PageIdForInputRune(r rune) PageId {
	if r < 'a' || r > 'z' {
		return PageNull
	}
	offset := r - 'a'
	return PageId(rune(PageLetterA) + offset)
}

// pageData holds the internal content of a clipboard page.
type pageData struct {
	buf      bytes.Buffer
	linewise bool
}

// Clipboard represents a clipboard.
// The clipboard consists of distinct pages, each of which can store string content.
type Clipboard struct {
	pages map[PageId]*pageData
}

// New constructs a new, empty clipboard.
func New() *Clipboard {
	pages := make(map[PageId]*pageData, 0)
	return &Clipboard{pages}
}

// Set clears a page and returns an io.Writer for writing new content.
// The linewise parameter indicates whether the content represents whole lines.
// Writing to the null page discards data.
func (c *Clipboard) Set(p PageId, linewise bool) io.Writer {
	if p == PageNull {
		return io.Discard
	}
	pd := &pageData{linewise: linewise}
	c.pages[p] = pd
	return &pd.buf
}

// Get returns an io.Reader for reading the contents of a page
// and a boolean indicating whether the content is linewise.
// If the page has not been set, the reader will be empty and linewise will be false.
func (c *Clipboard) Get(p PageId) (io.Reader, bool) {
	pd, ok := c.pages[p]
	if !ok {
		return bytes.NewReader(nil), false
	}
	return bytes.NewReader(pd.buf.Bytes()), pd.linewise
}
