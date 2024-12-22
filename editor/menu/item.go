package menu

// Item represents an item in the editor's menu.
type Item struct {
	// Name is the displayed name of the item.
	// This is also used when searching for menu items.
	Name string

	// Aliases are a search terms for which this item will always rank first.
	Aliases []string

	// Action is the action to perform when the user selects the menu item.
	// This should be a function that accepts a single *EditorState arg.
	Action any
}
