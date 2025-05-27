package clipboard

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

// PageContent represents the content of a page in the clipboard.
type PageContent struct {
	Text     string
	Linewise bool
}

// Clipboard represents a clipboard.
// The clipboard consists of distinct pages, each of which can store string content.
type Clipboard struct {
	pages map[PageId]PageContent
}

// New constructs a new, empty clipboard.
func New() *Clipboard {
	pages := make(map[PageId]PageContent, 0)
	return &Clipboard{pages}
}

// Set stores a string in a page, replacing the prior contents.
func (c *Clipboard) Set(p PageId, pc PageContent) {
	if p == PageNull {
		return
	}
	c.pages[p] = pc
}

// Get retrieves the contents of a page.
func (c *Clipboard) Get(p PageId) PageContent {
	return c.pages[p]
}
