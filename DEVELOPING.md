Developing
==========

Building
--------

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
