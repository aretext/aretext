System Clipboard
================

Aretext can use command-line tools to copy and paste with the system clipboard. The specific tools vary across operating systems; see below for details.

Configure `systemClipboard` by [editing the config file](configuration.md). The `copyCmd` command receives copied text on stdin. The `pasteCmd` command writes pasted text to stdout.

After configuring the system clipboard, use the `"+` or `"*` clipboard page prefix with any copy, delete, change, or paste command. For example:

-	`"+yy` copies the current line to the system clipboard.
-	`"+p` pastes from the system clipboard.

Set `useByDefault` to `true` to make commands like `yy` and `p` use the system clipboard without the `"+` prefix.

Examples
--------

### Linux Wayland

```yaml
- name: linux wayland system clipboard
  pattern: "**"
  config:
    systemClipboard:
      copyCmd: wl-copy
      pasteCmd: wl-paste --no-newline
      useByDefault: true
```

### Linux X Windows

```yaml
- name: linux xwindows system clipboard
  pattern: "**"
  config:
    systemClipboard:
      copyCmd: xclip -selection clipboard
      pasteCmd: xclip -selection clipboard -out
      useByDefault: true
```

### macOS

```yaml
- name: macos system clipboard
  pattern: "**"
  config:
    systemClipboard:
      copyCmd: pbcopy
      pasteCmd: pbpaste
      useByDefault: true
```
