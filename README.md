# aretext
Terminal-based text editor par excellence.

Design principles:

1. The user knows what they're doing, so give them total control.  This is a stick-shift car, not an automatic.
2. Compose seamlessly with the rest of the \*nix ecosystem (tmux, grep, code formatters, compilers, git, zsh, etc).
3. Run fast.



## Roadmap

- [ ] display file contents
- [ ] exit
- [ ] handle terminal resize
- [ ] navigate up/down/left/right
- [ ] scroll file contents vertically
- [ ] scroll file contents horizontally, handle long lines
- [ ] insert text
- [ ] delete text
- [ ] save changes
- [ ] check for unsaved changes on exit
- [ ] detect when file changed automatically, prompt for reload
- [ ] open a file with a particular cursor position
- [ ] status line showing file edited
- [ ] copy/paste using system clipboard integration
- [ ] line numbers
- [ ] switch to a different file (with prompt for save)
- [ ] handle double-width unicode characters
- [ ] handle tabs
- [ ] soft-wrap long lines
- [ ] undo/redo
- [ ] autocomplete through language server integration
- [ ] jump to definition through language server integration
- [ ] find usages through language server integration
- [ ] highlight errors from language server
- [ ] hover hints from language server
- [ ] forward search through file
- [ ] backward search through file
- [ ] option to display non-printable characters
- [ ] show help on bottom of screen, enable/disable with option
- [ ] config file to set default options
- [ ] optional vim-style keybindings (normal mode, visual mode, insert mode)
- [ ] syntax highlighting using textmate definitions
- [ ] autoindent
- [ ] comment/uncomment
- [ ] code folding
- [ ] setting overrides for file type
