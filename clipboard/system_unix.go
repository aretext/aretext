//go:build freebsd || linux || netbsd || openbsd || solaris || dragonfly

package clipboard

var systemClipboardTools = []systemClipboardTool{
	{
		copyCmd:  []string{"wl-copy"},
		pasteCmd: []string{"wl-paste", "--no-newline"},
	},
	{
		copyCmd:  []string{"tmux", "load-buffer", "-"},
		pasteCmd: []string{"tmux", "show-buffer"},
	},
}
