aretext
=======

Minimalist text editor that never slows you down.

Project Status
--------------

**Pre-alpha!**

-	Important features are planned, but not yet implemented.
-	Documentation has not yet been written.
-	The user interface may change in fundamental ways.

See [open milestones](https://github.com/aretext/aretext/milestones?direction=asc&sort=title&state=open) for the current roadmap.

Key Features
------------

-	Vim-compatible\* key bindings.
-	Built-in fuzzy search for commands and files.
-	Auto-reload when files are modified outside the editor.
-	Fast and accurate incremental syntax highlighting.
-	Intuitive yet powerful configuration in a single file.

*\* Aretext key bindings are compatible with vim's normal, insert, and visual modes. Not all keybindings are implemented yet.*

Supported Platforms
-------------------

| Platform | Status             |
|----------|--------------------|
| Linux    | Fully supported    |
| \*BSD    | Will probably work |
| macOS    | Will probably work |
| Windows  | Not supported      |

Installation
------------

### From source

To install aretext from source, checkout this repository then run `make install`.

### From the ArchLinux AUR

aretext-git is available as an [AUR Package](https://aur.archlinux.org/packages/aretext-git/) If you use [yay](https://github.com/Jguer/yay) just run this to install it:

```shell
yay -S aretext-git
```

Getting Started
---------------

For development, you will need to install two formatting tools:

```
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/shurcooL/markdownfmt@latest
```

To build aretext and run tests, use `make`.

You can then open a file in the editor: `./aretext path/to/file.txt`

-	The editor supports most key bindings from vim's normal, insert, and visual modes.
-	Type ":" in normal mode to open a searchable menu of commands (save, quit, etc.)

Contributing
------------

Contributions are welcome! Please read the [Contribution Guidelines](CONTRIBUTING.md) to get started.

Security
--------

Please do NOT post security issues in public. To report a vulnerability, please send an email to [security@aretext.org](mailto:security@aretext.org).

Copyright and License
---------------------

Copyright (C) 2021 Will Daly

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see https://www.gnu.org/licenses/.
