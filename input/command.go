package input

import (
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
		return []menu.Item{
			{
				Name: "quit",
				Action: func(s *state.EditorState) {
					state.AbortIfUnsavedChanges(s, state.Quit, true)
				},
			},
			{
				Name:   "force quit",
				Action: state.Quit,
			},
			{
				Name: "save",
				Action: func(s *state.EditorState) {
					state.SaveDocument(s, false)
				},
			},
			{
				Name: "force save",
				Action: func(s *state.EditorState) {
					state.SaveDocument(s, true)
				},
			},
			{
				Name: "force reload",
				Action: func(s *state.EditorState) {
					state.ReloadDocument(s, true)
				},
			},
			{
				Name: "find and open",
				Action: func(s *state.EditorState) {
					state.AbortIfUnsavedChanges(s, ShowFileMenu(config), true)
				},
			},
			{
				Name: "set syntax json",
				Action: func(s *state.EditorState) {
					state.SetSyntax(s, syntax.LanguageJson)
				},
			},
			{
				Name: "set syntax go",
				Action: func(s *state.EditorState) {
					state.SetSyntax(s, syntax.LanguageGo)
				},
			},
			{
				Name: "set syntax none",
				Action: func(s *state.EditorState) {
					state.SetSyntax(s, syntax.LanguageUndefined)
				},
			},
		}
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
				Name: file.RelativePathCwd(path),
				Action: func(s *state.EditorState) {
					state.LoadDocument(s, path, true, true)
				},
			})
		})

		return items
	}

	return func(s *state.EditorState) {
		state.ShowMenu(s, state.MenuStyleFile, findFileMenuItems)
	}
}
