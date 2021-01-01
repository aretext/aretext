# aretext
Minimalist terminal-based text editor, mostly vim-compatible.

Design principles:

1. Choose speed over flexibility.
2. Compose seamlessly with the rest of the \*nix ecosystem (tmux, grep, code formatters, compilers, git, zsh, etc).


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
- [x] command menu with fuzzy search
- [x] save changes
- [x] close and open new document
- [ ] custom menu items that invoke external programs
- [ ] undo/redo
- [ ] line numbers
- [ ] copy/paste using system clipboard integration
- [ ] vim newline command ('o')
- [ ] vim replace/change commands
- [ ] vim word navigation
- [ ] vim section navigation
- [ ] visual mode / selection
- [ ] forward search through file
- [ ] backward search through file


### Beyond 1.0

- [ ] autoindent
- [ ] match parens
- [ ] repeat last action ('.')
- [ ] vim repeat commands ('10x')
- [ ] jump to definition (without an index)
