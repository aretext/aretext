Navigation
==========

This guide explains how to navigate efficiently within a document. It assumes you know how to start the editor and switch between normal/insert mode (if not, please read the [Quickstart](quickstart.md) first).

Scrolling
---------

To scroll up by half a screen, press Ctrl-u ("up") in normal mode.

To scroll down by half a screen, press Ctrl-d ("down") in normal mode.

Line movement
-------------

To move the cursor to the last line, type "G" in normal mode.

To move the cursor to the first line, type "gg" in normal mode.

To move the cursor to a specific line number, type "<number>gg" in normal mode. For example, "123gg" moves the cursor to the start of line 123.

To move the cursor to the start of the current line (after any indentation), use "^". Use "0" to move to the start of the current line *before* any indentation.

To move the cursor to the end of the current line, type "$" in normal mode.

Word movement
-------------

A "word" in aretext is a sequence of characters separated by whitespace.

If syntax highlighting is enabled, then the syntax language will define additional word boundaries. For example, when editing Go code, the text `foo.Bar` would treat `foo`, `.`, and `Bar` as separate words because the Go compiler would recognize these as separate tokens.

To move the cursor forward to the next word, press "w" in normal mode. Use "e" to move the cursor to the *end* of the current word, and "b" to move the cursor *back* to the start of the previous word.

Paragraph movement
------------------

A "paragraph" in aretext is a contiguous sequence of non-empty lines. To move the cursor to the next paragraph, type "}" in normal mode; to move to the previous paragraph, type "{".

Text search
-----------

Aretext supports case-sensitive forward and backward search within a document. (It does not currently support case-insensitive or regular expression searches, but these features may be added in the future.)

To search forward, type "/" in normal mode, then type your search query. To move the cursor to the search result, press enter; to abort the search, press escape.

To search backward, type "?" in normal mode, then type your search query.

To repeat a search, type "n" in normal mode (this moves the cursor to the "next" result). To move the cursor back to the previous result, type "N" in normal mode.
