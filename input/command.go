package input

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"

	"github.com/aretext/aretext/file"
	"github.com/aretext/aretext/menu"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/syntax"
)

func commandMenuItems(config Config) func() []menu.Item {
	return func() []menu.Item {
		items := []menu.Item{
			{
				Name:    "quit",
				Aliases: []string{"q"},
				Action: func(s *state.EditorState) {
					state.AbortIfUnsavedChanges(s, state.Quit, true)
				},
			},
			{
				Name:    "force quit",
				Aliases: []string{"q!"},
				Action:  state.Quit,
			},
			{
				Name:    "save document",
				Aliases: []string{"s", "w"},
				Action: func(s *state.EditorState) {
					state.AbortIfFileChanged(s, state.SaveDocument)
				},
			},
			{
				Name:    "save document and quit",
				Aliases: []string{"sq", "wq"},
				Action: func(s *state.EditorState) {
					state.AbortIfFileChanged(s, func(s *state.EditorState) {
						state.SaveDocument(s)
						state.Quit(s)
					})
				},
			},
			{
				Name:    "force save document",
				Aliases: []string{"s!", "w!"},
				Action:  state.SaveDocument,
			},
			{
				Name:    "force save document and quit",
				Aliases: []string{"sq!", "wq!"},
				Action: func(s *state.EditorState) {
					state.SaveDocument(s)
					state.Quit(s)
				},
			},
			{
				Name:    "force reload",
				Aliases: []string{"r!"},
				Action:  state.ReloadDocument,
			},
			{
				Name:    "find and open",
				Aliases: []string{"f"},
				Action: func(s *state.EditorState) {
					state.AbortIfUnsavedChanges(s, ShowFileMenu(config), true)
				},
			},
			{
				Name:    "open previous document",
				Aliases: []string{"p"},
				Action: func(s *state.EditorState) {
					state.AbortIfUnsavedChanges(s, state.LoadPrevDocument, true)
				},
			},
			{
				Name:    "open next document",
				Aliases: []string{"n"},
				Action: func(s *state.EditorState) {
					state.AbortIfUnsavedChanges(s, state.LoadNextDocument, true)
				},
			},
			{
				Name:    "toggle line numbers",
				Aliases: []string{"nu"},
				Action:  state.ToggleShowLineNumbers,
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
					state.LoadDocument(s, path, true, func(state.LocatorParams) uint64 {
						return 0
					})
				},
			})
		})

		return items
	}

	return func(s *state.EditorState) {
		state.ShowMenu(s, state.MenuStyleFilePath, findFileMenuItems)
	}
}
