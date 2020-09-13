package exec

import "log"

// Layout controls how buffers are displayed in the editor.
type Layout int

const (
	// LayoutDocumentOnly means that only the document is displayed; the REPL is hidden.
	LayoutDocumentOnly = Layout(iota)

	// LayoutDocumentAndRepl means that both the document and REPL are displayed.
	LayoutDocumentAndRepl
)

func (l Layout) String() string {
	if l == LayoutDocumentOnly {
		return "DocumentOnly"
	} else if l == LayoutDocumentAndRepl {
		return "DocumentAndRepl"
	} else {
		log.Fatalf("Unrecognized layout: %d\n", l)
		return ""
	}
}
