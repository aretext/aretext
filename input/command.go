package input

import (
	"fmt"
	"log"
	"os"

	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/syntax"
	"github.com/pkg/errors"
)

func commandMenuItems(config Config) func() []menu.Item {
	return func() []menu.Item {
		items := []menu.Item{
			{
				Name:  "quit",
				Alias: "q",
				Action: func(s *state.EditorState) {
					state.AbortIfUnsavedChanges(s, state.Quit, true)
				},
			},
			{
				Name:   "force quit",
				Alias:  "q!",
				Action: state.Quit,
			},
			{
				Name:  "write",
				Alias: "w",
				Action: func(s *state.EditorState) {
					state.AbortIfFileChanged(s, state.SaveDocument)
				},
			},
			{
				Name:  "write and quit",
				Alias: "wq",
				Action: func(s *state.EditorState) {
					state.AbortIfFileChanged(s, func(s *state.EditorState) {
						state.SaveDocument(s)
						state.Quit(s)
					})
				},
			},
			{
				Name:   "force write",
				Alias:  "w!",
				Action: state.SaveDocument,
			},
			{
				Name:   "force reload",
				Alias:  "r!",
				Action: state.ReloadDocument,
			},
			{
				Name:  "find and open",
				Alias: "f",
				Action: func(s *state.EditorState) {
					state.AbortIfUnsavedChanges(s, ShowFileMenu(config), true)
				},
			},
		}

		for _, language := range syntax.AllLanguages {
			l := language // ensure each item refers to a different language
			items = append(items, menu.Item{
				Name: fmt.Sprintf("set syntax %s", language),
				Action: func(s *state.EditorState) {
					state.SetSyntax(s, l)
				},
			})
		}

		return items
	}
}

func ShowFileMenu(config Config) Action {
	findFileMenuItems := func() []menu.Item {
		dir, err := os.Getwd()
		if err != nil {
			log.Printf("Error loading menu items: %v\n", errors.Wrapf(err, "os.GetCwd"))
			return nil
		}

		items := make([]menu.Item, 0, 0)
		file.Walk(dir, config.DirNamesToHide, func(path string) {
			items = append(items, menu.Item{
				Name: file.RelativePath(path, dir),
				Action: func(s *state.EditorState) {
					state.LoadDocument(s, path, true)
				},
			})
		})

		return items
	}

	return func(s *state.EditorState) {
		state.ShowMenu(s, state.MenuStyleFile, findFileMenuItems)
	}
}
