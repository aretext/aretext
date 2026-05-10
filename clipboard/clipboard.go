package clipboard

import (
	"fmt"
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

	// The system page delegates to the operating system's clipboard.
	PageSystem

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
	// Unlike vim, we don't distinguish between different system clipboards.
	if r == '+' || r == '*' {
		return PageSystem
	}

	if r < 'a' || r > 'z' {
		return PageNull
	}

	offset := r - 'a'
	return PageId(rune(PageLetterA) + offset)
}

// pageContent represents the content of a page in the clipboard.
type pageContent struct {
	data     []byte
	linewise bool
}

// Clipboard represents a clipboard.
// The clipboard consists of distinct pages, each of which can store string content.
type Clipboard struct {
	systemClipboard *SystemClipboard
	pages           map[PageId]pageContent
}

// New constructs a new, empty clipboard.
func New(systemClipboard *SystemClipboard) *Clipboard {
	pages := make(map[PageId]pageContent, 0)
	return &Clipboard{systemClipboard, pages}
}

// SetSystemClipboard configures the clipboard used for system clipboard pages.
func (c *Clipboard) SetSystemClipboard(systemClipboard *SystemClipboard) {
	c.systemClipboard = systemClipboard
}

// Set stores a string in a page, replacing the prior contents.
func (c *Clipboard) Set(p PageId, r io.Reader, linewise bool) error {
	if p == PageNull {
		return nil
	}

	if c.systemClipboard.ShouldUseForPage(p) {
		if c.systemClipboard == nil {
			return fmt.Errorf("system clipboard not configured")
		}
		return c.systemClipboard.Set(r, linewise)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}
	c.pages[p] = pageContent{data, linewise}
	return nil
}

// Get retrieves the contents of a page.
func (c *Clipboard) Get(p PageId, w io.Writer) (bool, error) {
	if c.systemClipboard.ShouldUseForPage(p) {
		if c.systemClipboard == nil {
			return false, fmt.Errorf("system clipboard not configured")
		}
		return c.systemClipboard.Get(w)
	}

	pc := c.pages[p]
	data := pc.data

	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return false, fmt.Errorf("io.WriteString: %w", err)
		}
		data = data[n:]
	}

	return pc.linewise, nil
}
