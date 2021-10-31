Customization
=============

This guide describes how to customize aretext for your workflows.

Aretext uses a [rule-based system for configuration](#configuration-rules). This allows you to easily customize the editor for different programming languages and projects.

In addition, you can define [custom menu commands](#custom-menu-commands) that invoke arbitrary programs. This provides a simple yet powerful way to extend the editor. For example, you can create custom menu commands to:

-	[Build a project with make](#example-make)
-	[Copy and paste using the system clipboard](#example-copy-and-paste-using-the-system-clipboard)
-	[Format a file](#example-format-current-file)
-	[Insert a snippet](#example-insert-snippet)
-	[Search a project with grep](#example-grep)
-	[Open the current document in a new tmux window](#example-split-tmux-window)

... and much more!

Configuration rules
-------------------

Aretext stores its configuration in a single YAML file. You can edit the config file using the `-editconfig` flag:

```
aretext -editconfig
```

The configuration file is located at `$XDG_CONFIG_HOME/aretext/config.yaml`, where `XDG_CONFIG_HOME` is configured according to the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html). On Linux, this defaults to `~/.config`, and on macOS it defaults to `~/Library/Application Support`.

When you open the config file, you should see something like:

```
- name: default
  pattern: "**"
  config:
    autoIndent: false
    hideDirectories: [".git"]
    syntaxLanguage: plaintext
    tabExpand: false
    tabSize: 4
    showLineNumbers: false

- name: json
  pattern: "**/*.json"
  config:
    autoIndent: true
    syntaxLanguage: json
    tabExpand: true
    tabSize: 2
    showLineNumbers: true
```

Each item in the configuration file describes a *rule*. For example, in the snippet above, the first rule is named "default" and the second rule is named "json".

Each rule has a *pattern*. The "\*\*" is a wildcard that matches any subdirectory, and "\*" is a wildcard that matches zero or more characters in a file or directory name.

When aretext loads a file, it checks each rule in order. If the rule's pattern matches the file's absolute path, it applies the rule to update the configuration.

For example, if aretext loaded the file "foo/bar.json" using the above configuration, both rules would match the filename. The resulting configuration would be:

```
config:
    autoIndent: true           # from the "json" rule
    hideDirectories: [".git"]  # from the "default" rule
    syntaxLanguage: json       # from the "json" rule
    tabExpand: true            # from the "json" rule
    tabSize: 2                 # from the "json" rule
    showLineNumbers: true      # from the "json" rule
```

When merging configurations from different rules:

-	For strings and numbers, the values from later rules overwrite the values from previous rules.
-	For lists, the values from all rules are combined.
-	For dictionaries, the keys from later rules are added to the merged dictionary, potentially overwriting keys set by previous rules.

This is a powerful mechanism for customizing configuration based on filename extension and/or project location. For example, suppose that one project you work on uses four spaces to indent JSON files. You could add a new rule to your config that overwrites the tabSize for JSON files in that specific project:

```
# ... other rules above ...
- name: myproject-json
  pattern: "**/myproject/**/*.json"
  config:
    tabSize: 4
```

Configuration Reference
-----------------------

For a complete list of available configuration options, see [Configuration Reference](config-reference.md).

Fixing configuration errors
---------------------------

If your YAML config file has errors, aretext will exit with an error message. You can force aretext to ignore the config file by passing the "-noconfig" flag:

```
aretext -editconfig -noconfig
```

This allows you to start the editor so you can fix the configuration.

Custom menu commands
--------------------

Aretext allows you to define custom menu items to run shell commands. This provides a simple, yet powerful, way to extend aretext.

You can add new menu commands by editing the config file:

```
- name: custom-menu-rule
  pattern: "**/myproject/**"
  config:
    menuCommands:
      - name: my custom menu command
        shellCmd: echo 'hello world!' | less
        mode: terminal  # or "silent" or "insert" or "fileLocations"
```

After restarting the editor, the new command will be available in the command menu. Selecting the new command will launch a shell (configured by the `$SHELL` environment variable) and execute the shell command (in this case, echoing "hello world").

The "mode" parameter controls how aretext handles the command's input and output. There are four modes:

| Mode          | Input | Output               | Use Cases                                                                     |
|---------------|-------|----------------------|-------------------------------------------------------------------------------|
| terminal      | tty   | tty                  | `make`, `git commit`, `go test`, `man`, ...                                   |
| silent        | none  | none                 | `go fmt`, tmux commands, copy to system clipboard, ...                        |
| insert        | none  | insert into document | paste from system clipboard, insert snippet, comment/uncomment selection, ... |
| fileLocations | none  | file location menu   | grep for word under cursor, ...                                               |

In addition, the following environment variables are provided to the shell command:

-	`$FILEPATH` is the absolute path to the current file.
-	`$WORD` is the current word under the cursor.
-	`$SELECTION` is the currently selected text (if any).

### Example: Make

Add a menu command to build a project using `make`. Piping to `less` allows us to page through the output.

```
- name: custom-make-cmd
  pattern: "**/myproject/**"
  config:
    menuCommands:
      - name: build
        shellCmd: make | less
        # default mode is "terminal"
```

### Example: Copy and paste using the system clipboard

Most systems provide command-line utilities for interacting with the system clipboard.

| System               | Commands                               |
|----------------------|----------------------------------------|
| Linux using XWindows | `xclip`                                |
| Linux using Wayland  | `wl-copy`, `wl-paste`                  |
| macOS                | `pbcopy`, `pbpaste`                    |
| WSL on Windows       | `clip.exe`, `powershell Get-Clipboard` |
| tmux                 | `tmux set-buffer`, `tmux show-buffer`  |

We can add custom menu commands to copy the current selection to the system clipboard and paste from the system clipboard into the document.

On Linux (Wayland):

```
- name: linux-wayland-clipboard-commands
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

```
- name: macos-clipboard-commands
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

```
- name: tmux-clipboard-commands
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

### Example: Format current file

Many programming languages provide command line tools to automatically format code. You can add a custom menu command to run these tools on the current file.

For example, this command uses `go fmt` to format a Go file:

```
- name: custom-fmt-command
  pattern: "**/*.go"
  config:
    menuCommands:
      - name: go fmt current file
        shellCmd: go fmt $FILEPATH | less
```

If there are no unsaved changes, aretext will automatically reload the file after it has been formatted.

### Example: Insert snippet

You can add a custom menu command to insert a snippet of code.

For example, suppose you have written a template for a Go test. You can then create a menu command to `cat` the contents of the file into the document:

```
- name: custom-snippet-command
  pattern: "**/*.go"
  config:
    menuCommands:
      - name: insert test snippet
        shellCmd: cat ~/snippets/go-test.go
        mode: insert
```

### Example: Grep

You can add a custom menu command to grep for the word under the cursor. The following example uses [ripgrep](https://github.com/BurntSushi/ripgrep) to perform the search:

```
- name: custom-grep-command
  pattern: "**"
  config:
    menuCommands:
      - name: rg word
        shellCmd: rg $WORD --vimgrep  # or `grep $WORD -n -R .`
        mode: fileLocations
```

Once the search has completed, aretext loads the locations into a searchable menu. This allows you to easily navigate to a particular result.

The "fileLocations" mode works with any command that outputs file locations as lines with the format: `<file>:<line>:<snippet>` or `<file>:<line>:<col>:<snippet>`. You can use grep, ripgrep, or a script you write yourself!

### Example: Split tmux window

If you use [tmux](https://wiki.archlinux.org/title/Tmux), you can add a custom menu command to open the current document in a new window.

```
- name: split window horizontal
  shellCmd: tmux split-window -h "aretext $FILEPATH"
  mode: silent
- name: split window vertical
  shellCmd: tmux split-window -v "aretext $FILEPATH"
  mode: silent
```
