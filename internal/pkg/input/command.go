package input

import (
	"github.com/wedaly/aretext/internal/pkg/exec"
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
			Action: exec.NewSaveDocumentMutator(),
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
