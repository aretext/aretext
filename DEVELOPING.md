Developing
==========

Setup
-----

To build aretext, you will first need to [install go](https://golang.org/doc/install).

Next, install development tools:

```
make install-devtools
```

Building and Testing
--------------------

To build aretext, format code, and run tests:

```
make
```

To only run tests:

```
make test
```

To format Go and Markdown files:

```
make fmt
```

Some files in the project are created by [go generate](https://go.dev/blog/generate) and checked into the git repository. This includes the state machine compiled from commands in [input/commands.go](input/commands.go). If you add a new command there, you should regenerate the state machine by running:

```
make generate
```

Please see the [Makefile](Makefile) for all available targets.

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
