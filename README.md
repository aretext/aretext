# aretext
Minimalist terminal-based text editor, mostly vim-compatible.

Aretext prioritizes stability and speed over flexibility and features.  It works reliably and consistently across heterogeneous projects and machines.  If you live in a terminal, have multiple hosts in your ssh config, and rarely install vim plugins, this might be the editor for you.


## Project Status

**Pre-alpha!**

* Important features are planned, but not yet implemented.
* Documentation has not yet been written.
* The user interface may change in fundamental ways.

See the [Version 1.0 board](https://github.com/aretext/aretext/projects/1) for more details!


## Key Features

* (Mostly) vim-compatible key bindings.
* Built-in fuzzy search for commands and files.
* Fast and accurate incremental syntax highlighting.
* Intuitive and flexible configuration in a single JSON file.


## Supported Platforms

| Platform | Status             |
|----------|--------------------|
| Linux    | Fully supported    |
| OpenBSD  | Will probably work |
| macOS    | Will probably work |
| Windows  | Not supported      |


## Getting Started

To build aretext and run tests, use `make`.

You can then open a file in the editor: `./aretext path/to/file.txt`

* The editor supports most key sequences from vim's insert and normal modes.
* Type ":" in normal mode to open a searchable menu of commands (save, quit, etc.)


## Contributing

The project isn't yet ready to accept external contributions.  If you're interested in contributing, please star the repo and check back soon!


## Security

Please do NOT post security issues in public.  To report a vulnerability, please send an email to [security@aretext.org](mailto:security@aretext.org).


## Copyright and License

Copyright (C) 2021 Will Daly

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
