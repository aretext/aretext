Configuration Reference
=======================

This document lists every configuration option in aretext.

| Attribute         | Type             | Description                                                                                                                                 |
|-------------------|------------------|---------------------------------------------------------------------------------------------------------------------------------------------|
| `syntaxLanguage`  | enum             | Language used for syntax highlighting. Must be a valid [syntax language](#syntax-languages).                                                |
| `tabSize`         | integer          | Maximum number of cells occupied by a tab. Must be greater than zero.                                                                       |
| `tabExpand`       | boolean          | If true, replace inserted tabs with the equivalent number of spaces.                                                                        |
| `showTabs`        | boolean          | If true, display tabs in the document.                                                                                                      |
| `autoIndent`      | boolean          | If true, indent new lines to match indentation of the previous line.                                                                        |
| `showLineNumbers` | boolean          | If true, display line numbers.                                                                                                              |
| `menuCommands`    | array of objects | Additional menu items that can run arbitrary shell commands. See [Menu Command Object](#menu-command-object) below for the expected fields. |
| `hideDirectories` | array of strings | Glob patterns matching directories to hide from file search. Patterns are matched against the absolute path to the directory.               |
| `styles`          | dict             | Styles control how UI elements are displayed. See [Styles](#styles) below for details.                                                      |

Syntax Languages
----------------

| Value       | Description                               |
|-------------|-------------------------------------------|
| `plaintext` | Do not apply any syntax highlighting.     |
| `json`      | [JSON](https://www.json.org/json-en.html) |
| `yaml`      | [YAML](https://yaml.org/spec/)            |
| `go`        | [Go](https://golang.org/ref/spec)         |
| `gitcommit` | Format for editing a git commit           |
| `gitrebase` | Format for git interactive rebase         |

Menu Command Object
-------------------

| Attribute  | Type   | Description                                                                                                                                    |
|------------|--------|------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`     | string | Displayed name of the menu item.                                                                                                               |
| `shellCmd` | string | Shell command to execute when the menu item is selected.                                                                                       |
| `mode`     | enum   | Either "silent", "terminal", "insert", or "fileLocations". See [Custom menu commands](customization.md#custom-menu-commands) for more details. |

Styles
------

The `styles` configuration is a dictionary. Each key represents a kind of UI element:

-	`lineNum`: the line numbers displayed in the left margin of the document.
-	`tokenOperator`: an operator token recognized by the syntax language.
-	`tokenKeyword`: a keyword token recognized by the syntax language.
-	`tokenNumber`: a number token recognized by the syntax language.
-	`tokenString`: a string token recognized by the syntax language.
-	`tokenComment`: a comment token recognized by the syntax language.
-	`tokenCustom1` through `tokenCustom8`: language-specific tokens recognized by the syntax language.

The values are objects with a single key `color` that controls the color of the UI element. The color can be either [a W3C color keyword](https://www.w3.org/wiki/CSS/Properties/color/keywords) or a hexadecimal RGB code. For example, both `red` and `#ff0000` represent the color red.

When using named colors, the terminal emulator may override the displayed color. For example, the [solarized dark theme in Alacritty](https://github.com/eendroroy/alacritty-theme/blob/06c3920d35dbbe3de35183b0512f9406041d681b/themes/solarized_dark.yaml) overrides the color `red` to a specific hex code. If you want to ignore the terminal emulator palette, specify colors using hexadecimal RGB codes instead of named colors.
