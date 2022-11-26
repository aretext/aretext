Custom Menu Commands
====================

Aretext allows you to define custom menu items to run shell commands. This provides a simple, yet powerful, way to extend the editor.

Adding Custom Menu Comands
--------------------------

You can add new menu commands by [editing the config file](configuration.md) to add this rule:

```yaml
- name: custom menu rule
  pattern: "**/myproject/**"
  config:
    menuCommands:
    - name: my custom menu command
      shellCmd: echo 'hello world!' | less
      mode: terminal  # or "silent" or "insert" or "fileLocations"
```

After restarting the editor, the new command will be available in the command menu. Selecting the new command will launch a shell (configured by the `$SHELL` environment variable) and execute the shell command (in this case, echoing "hello world").

The "mode" parameter controls how aretext handles the command's input and output. The table below shows the available modes:

| Mode          | Input | Output                 | Use Cases                                                                     |
|---------------|-------|------------------------|-------------------------------------------------------------------------------|
| terminal      | tty   | tty                    | `make`, `git commit`, `go test`, `man`, ...                                   |
| silent        | none  | none                   | `go fmt`, tmux commands, copy to system clipboard, ...                        |
| insert        | none  | insert into document   | paste from system clipboard, insert snippet, comment/uncomment selection, ... |
| insertChoice  | none  | insert choice menu     | choose a word to insert from a dictionary like `/usr/share/dict/words`, ...   |
| fileLocations | none  | file location menu     | grep for word under cursor, ...                                               |
| workingDir    | none  | working directory menu | select the current working directory from a preset list                       |

In addition, the following environment variables are provided to the shell command:

-	`$FILEPATH` is the absolute path to the current file.
-	`$WORD` is the current word under the cursor.
-	`$LINE` is the line number of the cursor, starting from one.
-	`$COLUMN` is the column position of the cursor in bytes, starting from one.
-	`$SELECTION` is the currently selected text (if any).

If there are multiple commands with the same name, only the last of these commands will appear in the menu.

Examples
--------

### Build a project with make

Add a menu command to build a project using `make`. Piping to `less` allows us to page through the output.

```yaml
- name: custom make cmd
  pattern: "**/myproject/**"
  config:
    menuCommands:
    - name: build
      shellCmd: make | less
      save: true  # save the file before running `make`
```

### Copy and paste using the system clipboard

Most systems provide command-line utilities for interacting with the system clipboard. Custom menu commands can invoke these tools to copy the current selection and paste into the document.

On Linux (Wayland):

```yaml
- name: linux wayland clipboard commands
  pattern: "**"
  config:
    menuCommands:
    - name: copy to clipboard
      shellCmd: wl-copy "$SELECTION"
      mode: silent
    - name: paste from clipboard
      shellCmd: wl-paste
      mode: insert
```

On macOS:

```yaml
- name: macos clipboard commands
  pattern: "**"
  config:
    menuCommands:
      - name: copy to clipboard
        shellCmd: printenv SELECTION | pbcopy
        mode: silent
      - name: paste from clipboard
        shellCmd: pbpaste
        mode: insert
```

Using tmux:

```yaml
- name: tmux clipboard commands
  pattern: "**"
  config:
    menuCommands:
    - name: copy to clipboard
      shellCmd: printenv SELECTION | tmux load-buffer -
      mode: silent
    - name: paste from clipboard
      shellCmd: tmux show-buffer
      mode: insert
```

### Format the current file

Many programming languages provide command line tools to automatically format code. You can add a custom menu command to run these tools on the current file.

For example, this command uses `go fmt` to format a Go file:

```yaml
- name: custom fmt command
  pattern: "**/*.go"
  config:
    menuCommands:
    - name: go fmt current file
      shellCmd: go fmt $FILEPATH | less
      save: true  # save the file before running `go fmt`
```

### Git blame the current file

When working in a git repository, you might want to know who last edited a line of code. You can find this using `git blame` on the current file.

```yaml
- name: git blame command
  pattern: "**"
  config:
    menuCommands:
    - name: git blame
      shellCmd: git blame "$FILEPATH" | less +$LINE
```

### Insert a snippet

You can add a custom menu command to insert a snippet of code.

For example, suppose you have written a template for a Go test. You can then create a menu command to `cat` the contents of the file into the document:

```yaml
- name: custom snippet command
  pattern: "**/*.go"
  config:
    menuCommands:
    - name: insert test snippet
      shellCmd: cat ~/snippets/go-test.go
      mode: insert
```

### Grep for the word under the cursor

You can add a custom menu command to grep for the word under the cursor. The following example uses [ripgrep](https://github.com/BurntSushi/ripgrep) to perform the search:

```yaml
- name: custom grep command
  pattern: "**"
  config:
    menuCommands:
    - name: rg word
      shellCmd: rg $WORD --vimgrep  # or `grep $WORD -n -R .`
      mode: fileLocations
```

Once the search has completed, aretext loads the locations into a searchable menu. This allows you to easily navigate to a particular result.

The "fileLocations" mode works with any command that outputs file locations as lines with the format: `<file>:<line>:<snippet>` or `<file>:<line>:<col>:<snippet>`. You can use grep, ripgrep, or a script you write yourself!

### Open a document in a new tmux window

If you use [tmux](https://wiki.archlinux.org/title/Tmux), you can add a custom menu command to open the current document in a new window.

```yaml
- name: tmux window commands
  pattern: "**"
  config:
    menuCommands:
    - name: split window horizontal
      shellCmd: tmux split-window -h "aretext -line $LINE $FILEPATH"
      mode: silent
    - name: split window vertical
      shellCmd: tmux split-window -v "aretext -line $LINE $FILEPATH
      mode: silent
```
