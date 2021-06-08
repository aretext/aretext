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
| `menuCommands`    | object           | Additional menu items that can run arbitrary shell commands. See [Menu Command Object](#menu-command-object) below for the expected fields. |
| `hideDirectories` | array of strings | Glob patterns matching directories to hide from file search. Patterns are matched against the absolute path to the directory.               |

Syntax Languages
----------------

| Value       | Description                                                                                               |
|-------------|-----------------------------------------------------------------------------------------------------------|
| `undefined` | Do not apply any syntax highlighting.                                                                     |
| `plaintext` | Parse punctuation characters as separate tokens. This affects word movement, but not syntax highlighting. |
| `json`      | [JSON](https://www.json.org/json-en.html)                                                                 |
| `yaml`      | [YAML](https://yaml.org/spec/)                                                                            |
| `go`        | [Go](https://golang.org/ref/spec)                                                                         |
| `gitcommit` | Format for editing a git commit                                                                           |
| `gitrebase` | Format for git interactive rebase                                                                         |

Menu Command Object
-------------------

| Attribute  | Type   | Description                                                                                                                                    |
|------------|--------|------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`     | string | Displayed name of the menu item.                                                                                                               |
| `shellCmd` | string | Shell command to execute when the menu item is selected.                                                                                       |
| `mode`     | enum   | Either "silent", "terminal", "insert", or "fileLocations". See [Custom menu commands](customization.md#custom-menu-commands) for more details. |
