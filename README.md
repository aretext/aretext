# aretext
Terminal-based text editor par excellence.

Design principles:

1. The user knows what they're doing, so give them total control.  This is a stick-shift car, not an automatic.
2. Compose seamlessly with the rest of the \*nix ecosystem (tmux, grep, code formatters, compilers, git, zsh, etc).
3. Run fast.


## Roadmap

- [x] display file contents, with support for wide characters and grapheme clustering
- [x] exit
- [x] handle terminal resize
- [x] navigate up/down/left/right
- [x] scroll file contents vertically
- [x] scroll file contents horizontally, handle long lines
- [x] insert text
- [x] delete text
- [ ] save changes
- [ ] check for unsaved changes on exit
- [ ] detect when file changed automatically, prompt for reload
- [ ] open a file with a particular cursor position
- [ ] status line showing file edited
- [ ] copy/paste using system clipboard integration
- [ ] line numbers
- [x] handle tabs
- [x] soft-wrap long lines
- [ ] undo/redo
- [ ] vim-style navigation for words, lines, and sections
- [ ] vim-style edit commands
- [ ] forward search through file
- [ ] backward search through file
- [ ] option to display non-printable characters
- [ ] syntax highlighting
- [ ] repeat last action
- [ ] macros
- [ ] autoindent
- [ ] comment/uncomment
