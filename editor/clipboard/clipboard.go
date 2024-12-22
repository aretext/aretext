package clipboard

// PageId represents a page in the clipboard.
// This is equivalent to what vim calls a "register".
type PageId int

const (
	// Data written to the null page is discarded.
	PageNull = PageId(iota)

	// The default page stores the contents of the most recent delete or yank operation.
	PageDefault

	// Output of the last shell command inserted into the document.
	PageShellCmdOutput

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

// PageIdForLetter returns the page named by a letter "a" to "z".
// If the rune is non-alphabetical, this returns the null page.
func PageIdForLetter(r rune) PageId {
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

// C represents a clipboard.
// The clipboard consists of distinct pages, each of which can store string content.
type C struct {
	pages map[PageId]PageContent
}

// New constructs a new, empty clipboard.
func New() *C {
	pages := make(map[PageId]PageContent, 0)
	return &C{pages}
}

// Set stores a string in a page, replacing the prior contents.
func (c *C) Set(p PageId, pc PageContent) {
	if p == PageNull {
		return
	}
	c.pages[p] = pc
}

// Get retrieves the contents of a page.
func (c *C) Get(p PageId) PageContent {
	return c.pages[p]
}
