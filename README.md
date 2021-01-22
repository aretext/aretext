# aretext
Minimalist terminal-based text editor, mostly vim-compatible.

Design principles:

1. Prioritize simplicity and performance over flexibility.
2. Compose seamlessly with the rest of the \*nix ecosystem (tmux, grep, code formatters, compilers, git, zsh, etc).


## Project Status

**Pre-alpha!**

* Important features are planned, but not yet implemented.
* Documentation has not yet been written.
* The user interface may change in fundamental ways.


## Getting Started

To build aretext and run tests, use `make`.

You can then open a file in the editor: `./aretext path/to/file.txt`

* The editor supports most key sequences from vim's insert and normal modes.
* Type ":" in normal mode to open a searchable menu of commands (save, quit, etc.)


## Roadmap

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
- [x] vim newline command ('o')
- [x] autoindent
- [x] replace tabs with spaces
- [ ] display tabs
- [ ] custom menu items that invoke external programs
- [ ] undo/redo
- [ ] line numbers
- [ ] vim replace/change commands
- [ ] vim word navigation
- [ ] vim section navigation
- [ ] visual mode / selection
- [ ] selection clipboard (delete/yank/put)
- [ ] copy/paste using system clipboard integration
- [ ] forward search through file
- [ ] backward search through file
- [ ] match parens
- [ ] repeat last action ('.')
- [ ] vim repeat commands ('10x')
- [ ] jump to definition
