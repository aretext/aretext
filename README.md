# aretext
Terminal-based text editor par excellence.

Design principles:

1. The user knows what they're doing, so give them total control.  This is a stick-shift car, not an automatic.
2. Compose seamlessly with the rest of the \*nix ecosystem (tmux, grep, code formatters, compilers, git, zsh, etc).
3. Run efficiently.


## Roadmap

### 1.0: "just enough to use it as an editor"

- [x] display file contents, with support for wide characters and grapheme clustering
- [x] exit
- [x] handle terminal resize
- [x] navigate up/down/left/right
- [x] scroll file contents vertically
- [x] scroll file contents horizontally, handle long lines
- [x] insert text
- [x] delete text
- [x] handle tabs
- [x] soft-wrap long lines
- [x] automatic reload when file changes on disk
- [x] syntax highlighting
- [ ] save changes
- [ ] close and open new document in REPL
- [ ] undo/redo
- [ ] hotkeys
- [ ] line numbers
- [ ] copy/paste using system clipboard integration
- [ ] run external program as subprocess on current document/directory
- [ ] vim newline command ('o')
- [ ] vim replace/change commands
- [ ] vim word navigation
- [ ] vim section navigation
- [ ] visual mode / selection
- [ ] forward search through file
- [ ] backward search through file
- [ ] REPL autocomplete
- [ ] autoindent
- [ ] repeat last action ('.')
- [ ] vim repeat commands ('10x')
