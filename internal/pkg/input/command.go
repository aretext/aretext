package input

import (
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/wedaly/aretext/internal/pkg/exec"
	"github.com/wedaly/aretext/internal/pkg/file"
	"github.com/wedaly/aretext/internal/pkg/syntax"
)

func commandMenuItems() []exec.MenuItem {
	return []exec.MenuItem{
		{
			Name:   "quit",
			Action: exec.NewQuitMutator(false),
		},
		{
			Name:   "force quit",
			Action: exec.NewQuitMutator(true),
		},
		{
			Name:   "save",
			Action: exec.NewSaveDocumentMutator(false),
		},
		{
			Name:   "force save",
			Action: exec.NewSaveDocumentMutator(true),
		},
		{
			Name:   "force reload",
			Action: exec.NewReloadDocumentMutator(true),
		},
		{
			Name:   "find and open",
			Action: exec.NewShowMenuMutator("file path", loadOpenMenuItems, true),
		},
		{
			Name:   "set syntax json",
			Action: exec.NewSetSyntaxMutator(syntax.LanguageJson),
		},
		{
			Name:   "set syntax none",
			Action: exec.NewSetSyntaxMutator(syntax.LanguageUndefined),
		},
	}
}

func loadOpenMenuItems() []exec.MenuItem {
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Error loading menu items: %v\n", errors.Wrapf(err, "os.GetCwd"))
		return nil
	}

	items := make([]exec.MenuItem, 0, 0)
	file.Walk(dir, func(path string) {
		items = append(items, exec.MenuItem{
			Name:   file.RelativePathCwd(path),
			Action: exec.NewLoadDocumentMutator(path, true),
		})
	})

	return items
}
