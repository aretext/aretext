package menu

// Item represents an item in the editor's menu.
type Item struct {
	// Name is the displayed name of the item.
	// This is also used when searching for menu items.
	Name string

	// Action is the action to perform when the user selects the menu item.
	// TODO: replace this with the mutator function type.
	Action interface{}
}
