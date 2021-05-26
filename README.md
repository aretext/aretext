aretext
=======

Minimalist text editor that never slows you down.

See the [User Documentation](docs/index.md) to get started!

![screenshot of aretext](screenshot.gif)

Project Status
--------------

**Beta!**

-	The user interface is stable, and I am using it as my primary editor.
-	Syntax highlighting is available for a few languages.
-	Many important vim commands are supported, but there are still many more to add!

See [open milestones](https://github.com/aretext/aretext/milestones?direction=asc&sort=title&state=open) for the current roadmap.

Key Features
------------

-	Vim-compatible\* key bindings.
-	Built-in fuzzy search for commands and files.
-	Auto-reload when files are modified outside the editor.
-	Fast and accurate incremental syntax highlighting.
-	Powerful and intuitive configuration in a single file.

*\* Aretext key bindings are compatible with vim's normal, insert, and visual modes. See [Command Reference](docs/command-reference.md) for details.*

Getting Started
---------------

-	[Install](docs/install.md)
-	[Quickstart](docs/quickstart.md)
-	[User Documentation](docs/index.md)

Contributing
------------

Contributions are welcome! Please read the [Contribution Guidelines](CONTRIBUTING.md) to get started.

Build and Run Tests
-------------------

To build aretext, you will first need to [install go](https://golang.org/doc/install).

Next, install `goimports` and `markdownfmt`:

```
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/shurcooL/markdownfmt@latest
```

You can then build aretext and run tests using `make`. See the [Makefile](Makefile) for available commands.

Debugging
---------

First, you will need to [install dlv](https://github.com/go-delve/delve/tree/master/Documentation/installation).

Then build aretext with debug symbols:

```
make build-debug
```

You can then start aretext and attach a debugger:

```
# Start aretext in one terminal.
./aretext

# Switch to another terminal and attach a debugger.
# If there are multiple aretext processes running,
# replace `pgrep aretext` with the exact process ID.
dlv attach `pgrep aretext`
```

Copyright and License
---------------------

Copyright (C) 2021 Will Daly

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see https://www.gnu.org/licenses/.
