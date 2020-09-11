"""
Main entry point for the aretext Python REPL.

The host program (aretext) starts a subprocess for the Python interpreter
that runs this script.  It sends input from the user to stdin and displays
text written to stdout.

The host program sends a SIGINT signal to interrupt the currently running
Python program, which causes the interpreter to raise a KeyboardInterrupt
exception.

The host program sends SIGTERM signal to terminate the Python interpreter.

If the Python interpeter exits before the host program, the host program
will start a new interpreter.
"""
import code
import sys
import pyaretext.api.editor


BANNER_MSG = """
Aretext REPL, Python {}
    Type "help()" for interactive help.
    Type "quit()" to exit the program.
    Use Ctrl-D to close the REPL.""".format(
    sys.version
).strip()


def main():
    # Redirect stderr to stdout so the host program receives all input through stdout.
    sys.stderr = sys.stdout

    # Run the REPL.
    code.interact(banner=BANNER_MSG, local={
        "__name__": "__console__",
        "__doc__": None,
        "quit": pyaretext.api.editor.quit,
    })


if __name__ == "__main__":
    main()
