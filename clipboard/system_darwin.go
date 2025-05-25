//go:build darwin

package clipboard

var systemClipboardTools = []systemClipboardTool{
	{
		copyCmd:  []string{"pbcopy"},
		pasteCmd: []string{"pbpaste"},
	},
	{
		copyCmd:  []string{"tmux", "load-buffer", "-"},
		pasteCmd: []string{"tmux", "show-buffer"},
	},
}
