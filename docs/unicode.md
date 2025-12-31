Unicode
=======

Configuration
-------------

Aretext supports two ways of rendering non-ASCII Unicode:

-	When `showUnicode` is false (default), the Unicode characters encoded in the document are sent directly to the terminal for display.
-	When `showUnicode` is true, the Unicode characters are escaped. For example, the smiling face emoji (😀) will display as `<U+1F600>`.

The `showUnicode` setting can be set in aretext's [configuration file](configuration.md) or toggled with a menu command ("Toggle Show Unicode").

Terminal Support
----------------

Terminal emulator support for non-ASCII Unicode varies widely. In some cases, the width aretext calculates for a Unicode character can differ from the width of the glyph rendered by the terminal emulator. This can cause rendering artifacts, misaligned cursor position, and other strange behavior.

If you see problems with a document containing symbols or emojis, try using a different terminal emulator or enabling `showUnicode`.
