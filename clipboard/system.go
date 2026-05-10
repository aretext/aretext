package clipboard

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aretext/aretext/shellcmd"
)

// SystemClipboard uses the operating system's clipboard.
type SystemClipboard struct {
	copyCmd      string
	pasteCmd     string
	useByDefault bool
}

func NewSystemClipboard(copyCmd string, pasteCmd string, useByDefault bool) *SystemClipboard {
	return &SystemClipboard{
		copyCmd:      copyCmd,
		pasteCmd:     pasteCmd,
		useByDefault: useByDefault,
	}

}

// ShouldUseForPage returns whether to use the system clipboard for the given clipboard page.
// The rules are that "+" and "*" pages always go to the system clipboard,
// and the default page also goes to the system clipboard if useByDefault=true.
func (c *SystemClipboard) ShouldUseForPage(p PageId) bool {
	return p == PageSystem || (c != nil && c.useByDefault && p == PageDefault)
}

// Set writes the contents of io.Reader to the system clipboard.
// The indicator for linewise/charwise is encoded as a final trailing line feed (\n),
// which is the same trick that vim uses to track this state across processes.
func (c *SystemClipboard) Set(r io.Reader, linewise bool) error {
	stdin := r
	if linewise {
		stdin = io.MultiReader(r, strings.NewReader("\n"))
	}

	if err := shellcmd.Run(context.Background(), c.copyCmd, nil, stdin, nil, nil); err != nil {
		return fmt.Errorf("copy command failed: %w", err)
	}
	return nil
}

// Get writes the contents of the system clipboard to the provided io.Writer.
// The trailing newline encoding linewise/charwise is removed from the output.
func (c *SystemClipboard) Get(w io.Writer) (bool, error) {
	var buf bytes.Buffer
	if err := shellcmd.Run(context.Background(), c.pasteCmd, nil, nil, &buf, nil); err != nil {
		return false, fmt.Errorf("paste command failed: %w", err)
	}

	data := buf.Bytes()
	linewise := len(data) > 0 && data[len(data)-1] == '\n'
	if linewise {
		data = data[:len(data)-1]
	}

	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return false, fmt.Errorf("write pasted contents: %w", err)
		}
		data = data[n:]
	}

	return linewise, nil
}
