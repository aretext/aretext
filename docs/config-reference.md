Configuration Reference
=======================

This document lists every configuration option in aretext.

| Attribute       | Type             | Description                                                                                                                                                          |
|-----------------|------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| syntaxLanguage  | enum             | Language used for syntax highlighting. Must be a valid [syntax language](#syntax-languages).                                                                         |
| tabSize         | integer          | Maximum number of cells occupied by a tab. Must be greater than zero.                                                                                                |
| tabExpand       | boolean          | If true, replace inserted tabs with the equivalent number of spaces.                                                                                                 |
| showTabs        | boolean          | If true, display tabs in the document.                                                                                                                               |
| showSpaces      | boolean          | If true, display spaces in the document.                                                                                                                             |
| autoIndent      | boolean          | If true, indent new lines to match indentation of the previous line.                                                                                                 |
| showLineNumbers | boolean          | If true, display line numbers.                                                                                                                                       |
| lineNumberMode  | enum             | Control how line numbers are displayed. Either "absolute" or "relative" to the cursor.                                                                               |
| lineWrap        | enum             | Control soft line wrapping behavior. Either "character" for breaking at any character boundary or "word" to break only at word boundaries.                           |
| menuCommands    | array of objects | Additional menu items that can run arbitrary shell commands. See [Menu Command Object](#menu-command-object) below for the expected fields.                          |
| hidePatterns    | array of strings | Glob patterns matching files or directories to hide from file search. Patterns are matched against absolute paths.                                                   |
| hideDirectories | array of strings | (DEPRECATED, use hidePatterns instead) Glob patterns matching directories to hide from file search. Patterns are matched against the absolute path to the directory. |
| styles          | dict             | Styles control how UI elements are displayed. See [Styles](#styles) below for details.                                                                               |

Syntax Languages
----------------

| Value        | Description                                                                              |
|--------------|------------------------------------------------------------------------------------------|
| bash         | [bash](https://www.gnu.org/software/bash/manual/bash.html)                               |
| c            | [C](http://www.gnu.org/software/gnu-c-manual/gnu-c-manual.html)                          |
| criticmarkup | [CriticMarkup](https://github.com/CriticMarkup/CriticMarkup-toolkit)                     |
| gitcommit    | Format for editing a git commit                                                          |
| gitrebase    | Format for git interactive rebase                                                        |
| go           | [Go](https://golang.org/ref/spec)                                                        |
| gotemplate   | [Go template](https://pkg.go.dev/text/template)                                          |
| json         | [JSON](https://www.json.org/json-en.html)                                                |
| makefile     | [Makefile](https://www.gnu.org/software/make/manual/make.html)                           |
| markdown     | [Markdown](https://commonmark.org/)                                                      |
| p4           | [p4](https://p4.org)                                                                     |
| plaintext    | Do not apply any syntax highlighting.                                                    |
| protobuf     | [Protocol Buffers Version 3](https://developers.google.com/protocol-buffers/docs/proto3) |
| python       | [Python](https://docs.python.org/3/reference/)                                           |
| rust         | [Rust](https://doc.rust-lang.org/stable/reference/)                                      |
| todotxt      | [todo.txt](https://github.com/todotxt/todo.txt)                                          |
| xml          | [xml](https://www.w3.org/TR/2006/REC-xml11-20060816/)                                    |
| yaml         | [YAML](https://yaml.org/spec/)                                                           |

Menu Command Object
-------------------

| Attribute | Type   | Description                                                                                                                      |
|-----------|--------|----------------------------------------------------------------------------------------------------------------------------------|
| name      | string | Displayed name of the menu item.                                                                                                 |
| shellCmd  | string | Shell command to execute when the menu item is selected.                                                                         |
| mode      | enum   | Either "silent", "terminal", "insert", or "fileLocations". See [Custom Menu Commands](custom-menu-commands.md) for more details. |
| save      | bool   | If true, attempt to save the document before executing the command.                                                              |

Styles
------

The `styles` configuration is an object with keys:

-	`lineNum`: the line numbers displayed in the left margin of the document.
-	`tokenOperator`: an operator token recognized by the syntax language.
-	`tokenKeyword`: a keyword token recognized by the syntax language.
-	`tokenNumber`: a number token recognized by the syntax language.
-	`tokenString`: a string token recognized by the syntax language.
-	`tokenComment`: a comment token recognized by the syntax language.
-	`tokenCustom1` through `tokenCustom16`: language-specific tokens recognized by the syntax language.

Each style object supports the following (optional) attributes:

| Attribute       | Type   | Description                  |
|-----------------|--------|------------------------------|
| color           | string | Foreground (text) color.     |
| backgroundColor | string | Background color.            |
| bold            | bool   | Set bold attribute.          |
| italic          | bool   | Set italic attribute.        |
| underline       | bool   | Set underline attribute.     |
| strikethrough   | bool   | Set strikethrough attribute. |

Colors can be either [a W3C color keyword](https://www.w3.org/wiki/CSS/Properties/color/keywords) or a hexadecimal RGB code. For example, both `red` and `#ff0000` represent the color red.

When using named colors, the terminal emulator may override the displayed color. For example, the [solarized dark theme in Alacritty](https://github.com/eendroroy/alacritty-theme/blob/06c3920d35dbbe3de35183b0512f9406041d681b/themes/solarized_dark.yaml) overrides the color `red` to a specific hex code. If you want to ignore the terminal emulator palette, specify colors using hexadecimal RGB codes instead of named colors.

Not all terminal emulators support every style attribute (bold, italic, etc.). If styles are displayed incorrectly, try changing the value of the `$TERM` environment variable. If you are using tmux, try [`set -g default-terminal "tmux"`](https://github.com/tmux/tmux/wiki/FAQ#i-dont-see-italics-or-italics-and-reverse-are-the-wrong-way-round).
