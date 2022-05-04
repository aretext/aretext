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

Logging
-------

You can tell aretext to log debug information to a file like this:

```
aretext -log debug.log
```

You can then tail the log file in a separate terminal session to see what aretext is doing:

```
tail -f debug.log
```

Testing
-------

To run all unit tests:

```
make test
```

Formatting
----------

To format Go and Markdown files:

```
make fmt
```

Generating
----------

Some files in the project are generated. To regenerate these:

```
make generate
```

Generated files should be checked into the git repository.

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

Profiling
---------

You can tell aretext to profile its CPU usage like this:

```
aretext -cpuprofile cpu.prof
```

This will create a [pprof](https://pkg.go.dev/runtime/pprof) profile that you can analyze using `go tool pprof cpu.prof`
